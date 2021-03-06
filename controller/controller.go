package controller

import (
	"fmt"
	"github.com/arjunrn/simple-scaler/pkg/apis/scaler/v1alpha1"
	clientset "github.com/arjunrn/simple-scaler/pkg/client/clientset/versioned"
	scalescheme "github.com/arjunrn/simple-scaler/pkg/client/clientset/versioned/scheme"
	informers "github.com/arjunrn/simple-scaler/pkg/client/informers/externalversions/scaler/v1alpha1"
	"github.com/arjunrn/simple-scaler/pkg/replicacalculator"
	prometheus "github.com/prometheus/client_golang/api"
	log "github.com/sirupsen/logrus"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	scaleclient "k8s.io/client-go/scale"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"time"
)

const (
	controllerAgentName = "scaler-controller"
	ErrComputeMetrics   = "ErrComputeMetrics"
	ErrUpdateTarget     = "ErrUpdateTarget"
	TargetUpdateSuccess = "TargetUpdateSuccess"
)

// Controller is the controller implementation for Foo resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// scalerclientset is a clientset for our own API group
	scalerclientset clientset.Interface

	queue            workqueue.RateLimitingInterface
	scalersSynced    cache.InformerSynced
	mapper           apimeta.RESTMapper
	scaleNamespacer  scaleclient.ScalesGetter
	replicaCalc      *replicacalculator.ReplicaCalculator
	prometheusClient prometheus.Client
	recorder         record.EventRecorder
}

// NewController returns a new sample controller
func NewController(kubeclientset kubernetes.Interface, scalerclientset clientset.Interface,
	scalerInformer informers.ScalerInformer, podInformer coreinformers.PodInformer,
	scaleNamespacer scaleclient.ScalesGetter, mapper apimeta.RESTMapper, prometheusClient prometheus.Client,
	resyncInterval time.Duration) *Controller {

	utilruntime.Must(scalescheme.AddToScheme(scheme.Scheme))
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	controller := &Controller{
		kubeclientset:    kubeclientset,
		scalerclientset:  scalerclientset,
		queue:            workqueue.NewNamedRateLimitingQueue(NewDefaultScalerRateLimiter(resyncInterval), "scalers"),
		scalersSynced:    scalerInformer.Informer().HasSynced,
		scaleNamespacer:  scaleNamespacer,
		prometheusClient: prometheusClient,
		recorder:         recorder,
	}
	controller.mapper = mapper
	podLister := podInformer.Lister()

	metricsSource := replicacalculator.NewPrometheusMetricsSource(prometheusClient)
	controller.replicaCalc = replicacalculator.NewReplicaCalculator(podLister, metricsSource)
	log.Info("Setting up event handlers")
	scalerInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueScaler,
		UpdateFunc: func(oldObj, newObj interface{}) {
			controller.enqueueScaler(newObj)
		},
	}, resyncInterval)

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	log.Info("Starting Scaler controller")

	log.Info("Waiting for informer caches to be synced")
	if ok := cache.WaitForCacheSync(stopCh, c.scalersSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	log.Info("starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	log.Info("Started workers")
	<-stopCh
	log.Info("Shutting down workers")

	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {

	}
}

func (c *Controller) processNextWorkItem() bool {
	key, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(key)

	err := c.reconcileKey(key.(string))
	if err == nil {
		// don't "forget" here because we want to only process a given HPA once per resync interval
		return true
	}

	c.queue.AddRateLimited(key)
	utilruntime.HandleError(err)
	return true
}

func (c *Controller) reconcileKey(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	scaler, err := c.scalerclientset.ArjunnaikV1alpha1().Scalers(namespace).Get(name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		log.Errorf("Scaler %s has been deleted", name)
		return nil
	}
	return c.reconcileScaler(scaler)
}

func (c *Controller) enqueueScaler(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.queue.AddRateLimited(key)
}

func (c *Controller) reconcileScaler(scalerShared *v1alpha1.Scaler) error {
	log.Infof("now processing scaler: %s", scalerShared.Name)
	scaler := scalerShared.DeepCopy()
	version, err := schema.ParseGroupVersion(scaler.Spec.Target.APIVersion)
	if err != nil {
		return err
	}
	targetGK := schema.GroupKind{
		Group: version.Group,
		Kind:  scaler.Spec.Target.Kind,
	}
	mappings, err := c.mapper.RESTMappings(targetGK)
	if err != nil {
		return err
	}
	log.Debugf("Found mappings: %v", mappings)
	scale, targetGR, err := c.scaleForResourceMappings(scaler.Namespace, scaler.Spec.Target.Name, mappings)
	if err != nil {
		return err
	}
	log.Debugf("Found scale: %v target group: %v", scale.Name, targetGR.Resource)

	currentReplicas := scale.Status.Replicas
	desiredReplicas := int32(0)

	if scale.Spec.Replicas == 0 {
		log.Infof("autoscaling disabled by target. %v", scale)
		return nil
	}

	replicas, err := c.computeReplicasForMetrics(scaler, scale)

	if err == nil {
		desiredReplicas = replicas
	} else {
		c.recorder.Eventf(scaler, corev1.EventTypeWarning, ErrComputeMetrics, "failed to compute replicas: %v", err)
		return err
	}

	log.Infof("target: %s currentReplicas: %d desiredReplicas: %d", scaler.Name, currentReplicas, desiredReplicas)

	if desiredReplicas < scaler.Spec.MinReplicas {
		log.Infof("cannot scaled down more than min replicas")
		return nil
	}

	if desiredReplicas > scaler.Spec.MaxReplicas {
		log.Infof("cannot scale up more than max replicas")
		return nil
	}

	if desiredReplicas == scale.Spec.Replicas {
		log.Infof("the current replicas and required replicas are the same")
		return nil
	}

	lastUpdated, err := time.Parse(time.RFC3339, scaler.Status.LastScalingTimestamp)
	if err != nil {
		log.Debugf("failed to find last updated time")
		lastUpdated = time.Now().Add(-time.Minute * 5)
	}

	//if lastUpdated.Add(time.Minute * 1).After(time.Now()) {
	//	log.Infof("still in cooldown period since last scaling period")
	//	return nil
	//}

	// Added support for coolDown
	coolDown := time.Duration(scaler.Spec.CoolDownPeriod) * time.Second
	log.Debugf("coolDownperiod is: %v", coolDown)
	if lastUpdated.Add(coolDown).After(time.Now()) {
		log.Infof("still in cooldown period since last scaling period")
		return nil
	}

	scale.Spec.Replicas = desiredReplicas
	_, err = c.scaleNamespacer.Scales(scale.Namespace).Update(targetGR, scale)

	if err != nil {
		c.recorder.Eventf(scaler, corev1.EventTypeWarning, ErrUpdateTarget, "failed to update target %s/%s",
			scale.Namespace, scale.Name)
		return err
	}

	scaler.Status.Condition = fmt.Sprintf("Scaled to %d replicas", desiredReplicas)
	scaler.Status.LastScalingTimestamp = time.Now().Format(time.RFC3339)
	scaler.Status.CurrentReplicas = desiredReplicas
	_, err = c.scalerclientset.ArjunnaikV1alpha1().Scalers(scaler.Namespace).UpdateStatus(scaler)
	if err != nil {
		log.Errorf("Failed to Update Scaler Status %v", err)
	}
	c.recorder.Eventf(scaler, corev1.EventTypeNormal, TargetUpdateSuccess,
		"successfully updated target %s/%s with replicas %d", scale.Namespace, scale.Name, desiredReplicas)
	return nil
}

func (c *Controller) scaleForResourceMappings(namespace, name string, mappings []*apimeta.RESTMapping) (*autoscalingv1.Scale, schema.GroupResource, error) {
	var firstErr error
	for i, mapping := range mappings {
		targetGR := mapping.Resource.GroupResource()
		scale, err := c.scaleNamespacer.Scales(namespace).Get(targetGR, name)
		if err == nil {
			return scale, targetGR, nil
		}

		// if this is the first error, remember it,
		// then go on and try other mappings until we find a good one
		if i == 0 {
			firstErr = err
		}
	}

	// make sure we handle an empty set of mappings
	if firstErr == nil {
		firstErr = fmt.Errorf("unrecognized resource")
	}

	return nil, schema.GroupResource{}, firstErr

}
func (c *Controller) computeReplicasForMetrics(scaler *v1alpha1.Scaler, scale *autoscalingv1.Scale, ) (replicas int32, err error) {
	currentReplicas := scale.Status.Replicas

	if scale.Status.Selector == "" {
		log.Errorf("Target needs a selector: %v", scale)
		return 0, fmt.Errorf("selector required")
	}

	selector, err := labels.Parse(scale.Status.Selector)
	if err != nil {
		return -1, fmt.Errorf("couldn't convert selector into a corresponding internal selector object: %v", err)
	}

	replicaCountProposal, err := c.replicaCalc.GetResourceReplicas(scaler.Namespace, scaler.Spec.Evaluations,
		currentReplicas, scaler.Spec.ScaleDown, scaler.Spec.ScaleUp, scaler.Spec.ScaleUpSize,
		scaler.Spec.ScaleDownSize, selector)
	if err != nil {
		return 0, err
	}
	return replicaCountProposal, nil
}

func (c *Controller) updateStatus(scaler *v1alpha1.Scaler) error {

	return nil
}

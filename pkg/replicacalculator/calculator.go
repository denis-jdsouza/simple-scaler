package replicacalculator

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
	corelisters "k8s.io/client-go/listers/core/v1"
)

type ReplicaCalculator struct {
	podLister         corelisters.PodLister
	prometheusMetrics MetricsSource
}

// NewReplicaCalculator Creates a  new replica calculator
func NewReplicaCalculator(lister corelisters.PodLister, prometheusMetrics MetricsSource) *ReplicaCalculator {
	return &ReplicaCalculator{
		podLister:         lister,
		prometheusMetrics: prometheusMetrics,
	}
}

// GetResourceReplicas get number of replicas for the deployment
func (c *ReplicaCalculator) GetResourceReplicas(namespace string, evaluations, currentReplicas,
downThreshold, upThreshold, scaleUpSize, scaleDownSize int32, selector labels.Selector) (int32, error) {
	pods, err := c.podLister.Pods(namespace).List(selector)
	if err != nil {
		return -1, err
	}

	podNames := make([]string, len(pods))
	for i, p := range pods {
		podNames[i] = p.Name
	}

	log.Debugf("pod names: %v", podNames)

	metrics, err := c.prometheusMetrics.GetPodMetrics(namespace, podNames, evaluations)
	if err != nil {
		return -1, err
	}

	log.Debugf("pod metrics: %v", metrics)

	scaleUp, scaleDown := c.shouldScale(podNames, metrics, upThreshold, downThreshold, evaluations)

	if scaleUp && scaleDown {
		scaleDown = false
	}

	proposedReplicas := currentReplicas

	if scaleUp {
		proposedReplicas += scaleUpSize
	}
	if scaleDown {
		proposedReplicas -= scaleDownSize
	}

	return proposedReplicas, nil
}

func (c *ReplicaCalculator) shouldScale(podNames []string, podMetrics map[string][]int, scaleUpThreshold,
scaleDownThreshold, evaluations int32) (bool, bool) {
	scaleUp := false
	scaleDown := false
	for _, p := range podNames {
		ignorepod := false
		var (
			pMetrics []int
			ok       bool
		)

		// If metrics are not present then continue
		if pMetrics, ok = podMetrics[p]; !ok {
			log.Debugf("metrics are not present !!")
			continue
		}

		// If metrics are not sufficient then continue
		if len(pMetrics) < int(evaluations) {
			log.Debugf("metrics are not sufficient !!")
			continue
		}

		// Fix for unwanted scaledown operation if any pod has CPU usage == 0
		// Pods can have 0 CPU usage for sometime before being added to service (for serving traffic) and after removing from service endpoint
		// Skip pods for which metrics value is 0
		for _, p := range pMetrics {
			if p == 0 {
				ignorepod = true
				break
			}
		}
		if ignorepod {
			log.Debugf("metrics contains 0 value hence skipping pod ...")
			continue
		}

		if !scaleUp {
			pScaleUp := true
			for _, p := range pMetrics {
				if p < int(scaleUpThreshold) {
					pScaleUp = false
					break
				}
			}
			scaleUp = pScaleUp
		}

		if !scaleDown {
			pScaleDown := true
			for _, p := range pMetrics {
				if p > int(scaleDownThreshold) {
					pScaleDown = false
					break
				}
			}
			scaleDown = pScaleDown
		}
	}

	return scaleUp, scaleDown
}

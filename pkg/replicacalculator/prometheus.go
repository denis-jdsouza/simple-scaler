package replicacalculator

import (
	"context"
	"fmt"
	prometheusclient "github.com/prometheus/client_golang/api"
	prometheusapi "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

const (
	prometheusQuery = `sum(rate(container_cpu_usage_seconds_total{job="kubernetes-cadvisor",image!="",container!="POD",pod=~"%s",namespace="%s"}[1m])) by(pod) / 
		sum(kube_pod_container_resource_requests_cpu_cores{job="kube-state-metrics",pod=~"%s",namespace="%s"}) by (pod)`
)

type MetricsSource interface {
	GetPodMetrics(namespace string, podIDs []string, evaluations int32) (map[string][]int, error)
}

func NewPrometheusMetricsSource(prometheusClient prometheusclient.Client) MetricsSource {
	prometheusAPI := prometheusapi.NewAPI(prometheusClient)
	return &prometheusMetricsSource{prometheusClient: prometheusClient, prometheusAPI: prometheusAPI}
}

type prometheusMetricsSource struct {
	prometheusClient prometheusclient.Client
	prometheusAPI    prometheusapi.API
}

func (m *prometheusMetricsSource) GetPodMetrics(namespace string, podIDs []string, evaluations int32) (map[string][]int, error) {
	todo := context.TODO()
	nameList := strings.Join(podIDs, "|")
	query := fmt.Sprintf(prometheusQuery, nameList, namespace, nameList, namespace)
	log.Debugf("prometheus query: %s", query)

	now := time.Now()
	end := now.Truncate(time.Minute)
	start := end.Add(-time.Minute * time.Duration(evaluations-1))
	queryRange := prometheusapi.Range{Start: start, End: end, Step: time.Minute}

	log.Debugf("query: %v", queryRange)
	results, err := m.prometheusAPI.QueryRange(todo, query, queryRange)
	if err != nil {
		return nil, err
	}
	var (
		matrixResult model.Matrix
		ok           bool
	)

	if matrixResult, ok = results.(model.Matrix); !ok {
		return nil, fmt.Errorf("unexpected return type from the prometheus api call: %v", results.Type())
	}
	mapResults := make(map[string][]int)
	log.Debugf("matrixResult ----: %v", matrixResult)
	for _, r := range matrixResult {
		podName := string(r.Metric["pod"])
		mapResults[podName] = make([]int, len(r.Values))
		for i, v := range r.Values {
			mapResults[podName][i] = int(v.Value * 100)
		}
	}
	return mapResults, nil
}

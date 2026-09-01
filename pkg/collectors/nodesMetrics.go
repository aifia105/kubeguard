package collectors

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

func NodesMetrics(ctx context.Context, mclientset *metricsclientset.Clientset) ([]metricsv1beta1.NodeMetrics, error) {
	nodesMetrics, err := mclientset.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, HandleScanError("Node Metrics", err)
	} else {
		return nodesMetrics.Items, nil
	}
}

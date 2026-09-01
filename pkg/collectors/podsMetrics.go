package collectors

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

func ListPodsMetrics(ctx context.Context, mclientset *metricsclientset.Clientset, namespace string) ([]metricsv1beta1.PodMetrics, error) {
	podsMetrics, err := mclientset.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, HandleScanError("Pods Metrics", err)
	}

	return podsMetrics.Items, nil
}

package collectors

import (
	"context"

	"github.com/aifia105/kubeguard/pkg/logger"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListConfigMaps(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]v1.ConfigMap, error) {
	configMaps, err := clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		logger.LogError("failed to list configmaps: %v", err)
		return nil, err
	}
	return configMaps.Items, nil
}

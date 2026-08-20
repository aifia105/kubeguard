package collectors

import (
	"context"

	"github.com/aifia105/kubeguard/pkg/logger"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListServices(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]v1.Service, error) {
	services, err := clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		logger.LogError("failed to list services: %v", err)
		return nil, err
	}

	return services.Items, nil
}

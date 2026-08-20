package collectors

import (
	"context"

	"github.com/aifia105/kubeguard/pkg/logger"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListNamespaces(ctx context.Context, clientset *kubernetes.Clientset) ([]v1.Namespace, error) {
	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		logger.LogError("failed to list namespaces: %v", err)
		return nil, err
	}

	return namespaces.Items, nil
}

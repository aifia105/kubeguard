package collectors

import (
	"context"

	"github.com/aifia105/kubeguard/pkg/logger"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListIngresses(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]networkingv1.Ingress, error) {
	ingresses, err := clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		logger.LogError("failed to list ingresses: %v", err)
		return nil, err
	}

	return ingresses.Items, nil
}

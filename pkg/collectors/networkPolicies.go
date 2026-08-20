package collectors

import (
	"context"

	"github.com/aifia105/kubeguard/pkg/logger"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListNetworkPolicies(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]v1.NetworkPolicy, error) {
	networkPolicies, err := clientset.NetworkingV1().NetworkPolicies(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		logger.LogError("failed to list network policies: %v", err)
		return nil, err
	}

	return networkPolicies.Items, nil
}

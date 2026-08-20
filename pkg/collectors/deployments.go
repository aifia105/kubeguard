package collectors

import (
	"context"

	"github.com/aifia105/kubeguard/pkg/logger"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListDeployments(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]v1.Deployment, error) {
	deployments, err := clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		logger.LogError("failed to list deployments: %v", err)
		return nil, err
	}

	return deployments.Items, nil
}

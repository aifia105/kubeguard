package collectors

import (
	"context"

	"github.com/aifia105/kubeguard/pkg/logger"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListNodes(ctx context.Context, clientset *kubernetes.Clientset) ([]v1.Node, error) {
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		logger.LogError("failed to list nodes: %v", err)
		return nil, err
	}
	return nodes.Items, nil

}

package collectors

import (
	"context"

	"github.com/aifia105/kubeguard/pkg/logger"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListSecrets(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]v1.Secret, error) {
	secrets, err := clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		logger.LogError("failed to list secrets: %v", err)
		return nil, err
	}

	return secrets.Items, nil
}

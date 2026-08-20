package collectors

import (
	"github.com/aifia105/kubeguard/pkg/logger"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
)

func ClusterInfo(clientset *kubernetes.Clientset) (*version.Info, error) {
	clusterInfo, err := clientset.Discovery().ServerVersion()
	if err != nil {
		logger.LogError("failed to get cluster version: %v", err)
		return nil, err
	} else {
		return clusterInfo, nil
	}

}

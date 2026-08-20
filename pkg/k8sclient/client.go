package k8sclient

import (
	"path/filepath"

	"github.com/aifia105/kubeguard/pkg/logger"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

func NewK8sClient() (*kubernetes.Clientset, *metricsclientset.Clientset, error) {
	logger.LogInfo("Initiating connection to the cluster...")

	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")

	config, err := rest.InClusterConfig()
	if err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			logger.LogFatal("failed to build kubeconfig: %v", err)
		}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.LogFatal("failed to create clientset: %v", err)
	}

	mclientset, err := metricsclientset.NewForConfig(config)
	if err != nil {
		logger.LogFatal("failed to create metrics clientset: %v", err)
	}

	logger.LogInfo("Successfully connected to the cluster.")

	return clientset, mclientset, nil
}

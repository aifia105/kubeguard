package cmd

import (
	"fmt"
	"sync"

	"github.com/aifia105/kubeguard/pkg/collectors"
	"github.com/aifia105/kubeguard/pkg/resourceprinter"

	"github.com/aifia105/kubeguard/pkg/logger"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/version"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan the Kubernetes cluster for security risks and misconfigurations",
	Run: func(cmd *cobra.Command, args []string) {
		ns := resolveNamespace(args)
		runFullScan(ns)
	},
}

type scanResults struct {
	cluster      *version.Info
	nodes        []v1.Node
	namespaces   []v1.Namespace
	pods         []v1.Pod
	events       []v1.Event
	deployments  []appsv1.Deployment
	secrets      []v1.Secret
	services     []v1.Service
	configmaps   []v1.ConfigMap
	ingresses    []networkingv1.Ingress
	nodesMetrics []metricsv1beta1.NodeMetrics
}

func init() {
	scanCmd.AddCommand(
		newScanResourceCmd("pods", func(ns string) ([]v1.Pod, error) {
			return collectors.ListPods(ctx, k8sClient, ns)
		}),
		newScanResourceCmd("nodes", func(ns string) ([]v1.Node, error) {
			return collectors.ListNodes(ctx, k8sClient)
		}),
		newScanResourceCmd("namespaces", func(ns string) ([]v1.Namespace, error) {
			return collectors.ListNamespaces(ctx, k8sClient)
		}),
		newScanResourceCmd("secrets", func(ns string) ([]v1.Secret, error) {
			return collectors.ListSecrets(ctx, k8sClient, ns)
		}),
		newScanResourceCmd("services", func(ns string) ([]v1.Service, error) {
			return collectors.ListServices(ctx, k8sClient, ns)
		}),
		newScanResourceCmd("deployments", func(ns string) ([]appsv1.Deployment, error) {
			return collectors.ListDeployments(ctx, k8sClient, ns)
		}),
		newScanResourceCmd("configmaps", func(ns string) ([]v1.ConfigMap, error) {
			return collectors.ListConfigMaps(ctx, k8sClient, ns)
		}),
		newScanResourceCmd("ingresses", func(ns string) ([]networkingv1.Ingress, error) {
			return collectors.ListIngresses(ctx, k8sClient, ns)
		}),
		newScanResourceCmd("events", func(ns string) ([]v1.Event, error) {
			return collectors.ListEvents(ctx, k8sClient, namespaceFlag)
		}),
		newScanResourceCmd("metrics", func(ns string) ([]metricsv1beta1.NodeMetrics, error) {
			return collectors.NodesMetrics(ctx, mclientset)
		}),
		newScanResourceCmd("cluster", func(ns string) (*version.Info, error) {
			return collectors.ClusterInfo(k8sClient)
		}),
	)
}

func newScanResourceCmd[T any](name string, collect func(ns string) (T, error)) *cobra.Command {
	return &cobra.Command{
		Use:   name + " [namespace]",
		Short: "Scan " + name,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ns := resolveNamespace(args)
			results, err := collect(ns)
			if err != nil {
				logger.LogFatal("failed to collect %s: %v", name, err)
				return
			}

			printResults(results, name)
		},
	}
}

func printResults(results interface{}, resourceName string) {
	switch resourceName {
	case "pods":
		resourceprinter.PrintPods(results.([]v1.Pod))
	case "nodes":
		resourceprinter.PrintNodes(results.([]v1.Node))
	case "namespaces":
		resourceprinter.PrintNamespaces(results.([]v1.Namespace))
	case "secrets":
		resourceprinter.PrintSecrets(results.([]v1.Secret))
	case "services":
		resourceprinter.PrintServices(results.([]v1.Service))
	case "deployments":
		resourceprinter.PrintDeployments(results.([]appsv1.Deployment))
	case "configmaps":
		resourceprinter.PrintConfigMaps(results.([]v1.ConfigMap))
	case "ingresses":
		resourceprinter.PrintIngresses(results.([]networkingv1.Ingress))
	case "events":
		resourceprinter.PrintEvents(results.([]v1.Event))
	case "metrics":
		resourceprinter.PrintNodeMetrics(results.([]metricsv1beta1.NodeMetrics))
	case "cluster":
		resourceprinter.PrintClusterInfo(results.(*version.Info))
	}
}

func runFullScan(namespace string) {
	logger.LogInfo("Fetching cluster information...")
	var results scanResults
	var wg sync.WaitGroup

	run := func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn()
		}()
	}

	// cluster
	run(func() {
		cluster, err := collectors.ClusterInfo(k8sClient)
		if err != nil {
			logger.LogError("failed to get cluster info: %v", err)
		}
		results.cluster = cluster
	})
	// nodes
	run(func() {
		nodes, err := collectors.ListNodes(ctx, k8sClient)
		if err != nil {
			logger.LogError("failed to list nodes: %v", err)
		}
		results.nodes = nodes
	})
	// namespaces
	run(func() {
		namespaces, err := collectors.ListNamespaces(ctx, k8sClient)
		if err != nil {
			logger.LogError("failed to list namespaces: %v", err)
		}
		results.namespaces = namespaces
	})
	// pods
	run(func() {
		pods, err := collectors.ListPods(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to list pods: %v", err)
		}
		results.pods = pods
	})
	// events
	run(func() {
		events, err := collectors.ListEvents(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to list events: %v", err)
		}
		results.events = events
	})
	// deployments
	run(func() {
		deployments, err := collectors.ListDeployments(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to list deployments: %v", err)
		}
		results.deployments = deployments
	})
	// secrets
	run(func() {
		secrets, err := collectors.ListSecrets(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to list secrets: %v", err)
		}
		results.secrets = secrets
	})
	// services
	run(func() {
		services, err := collectors.ListServices(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to list services: %v", err)
		}
		results.services = services
	})
	// configmaps
	run(func() {
		configmaps, err := collectors.ListConfigMaps(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to list configmaps: %v", err)
		}
		results.configmaps = configmaps
	})
	// ingresses
	run(func() {
		ingresses, err := collectors.ListIngresses(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to list ingresses: %v", err)
		}
		results.ingresses = ingresses
	})
	// node metrics
	run(func() {
		nodesMetrics, err := collectors.NodesMetrics(ctx, mclientset)
		if err != nil {
			logger.LogError("failed to list node metrics: %v", err)
		}
		results.nodesMetrics = nodesMetrics
	})

	wg.Wait()

	fmt.Println("")
	printResults(results.cluster, "cluster")
	printResults(results.nodes, "nodes")
	printResults(results.namespaces, "namespaces")
	printResults(results.pods, "pods")
	printResults(results.events, "events")
	printResults(results.deployments, "deployments")
	printResults(results.secrets, "secrets")
	printResults(results.services, "services")
	printResults(results.configmaps, "configmaps")
	printResults(results.ingresses, "ingresses")
	printResults(results.nodesMetrics, "metrics")

}

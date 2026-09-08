package cmd

import (
	"fmt"
	"sync"

	"github.com/aifia105/kubeguard/pkg/collectors"
	"github.com/aifia105/kubeguard/pkg/jsonoutput"
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
	Cluster         *version.Info                `json:"cluster,omitempty"`
	Nodes           []v1.Node                    `json:"nodes,omitempty"`
	Namespaces      []v1.Namespace               `json:"namespaces,omitempty"`
	Pods            []v1.Pod                     `json:"pods,omitempty"`
	Events          []v1.Event                   `json:"events,omitempty"`
	Deployments     []appsv1.Deployment          `json:"deployments,omitempty"`
	Secrets         []v1.Secret                  `json:"-"`
	Services        []v1.Service                 `json:"services,omitempty"`
	ConfigMaps      []v1.ConfigMap               `json:"configmaps,omitempty"`
	Ingresses       []networkingv1.Ingress       `json:"ingresses,omitempty"`
	NodesMetrics    []metricsv1beta1.NodeMetrics `json:"nodeMetrics,omitempty"`
	LimitRanges     []v1.LimitRange              `json:"limitRanges,omitempty"`
	ResourceQuotas  []v1.ResourceQuota           `json:"resourceQuotas,omitempty"`
	NetworkPolicies []networkingv1.NetworkPolicy `json:"networkPolicies,omitempty"`
	PodsMetrics     []metricsv1beta1.PodMetrics  `json:"podMetrics,omitempty"`
}

type scanResultsJSON struct {
	scanResults
	SecretsRedacted []secretSummary `json:"secrets,omitempty"`
}

type secretSummary struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Type      string   `json:"type"`
	DataKeys  []string `json:"dataKeys,omitempty"`
}

func redactSecrets(secrets []v1.Secret) []secretSummary {
	out := make([]secretSummary, 0, len(secrets))
	for _, s := range secrets {
		keys := make([]string, 0, len(s.Data))
		for k := range s.Data {
			keys = append(keys, k)
		}
		out = append(out, secretSummary{
			Name:      s.Name,
			Namespace: s.Namespace,
			Type:      string(s.Type),
			DataKeys:  keys,
		})
	}
	return out
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
			return collectors.ListEvents(ctx, k8sClient, ns)
		}),
		newScanResourceCmd("metrics", func(ns string) ([]metricsv1beta1.NodeMetrics, error) {
			return collectors.NodesMetrics(ctx, mclientset)
		}),
		newScanResourceCmd("cluster", func(ns string) (*version.Info, error) {
			return collectors.ClusterInfo(k8sClient)
		}),
		newScanResourceCmd("limitranges", func(ns string) ([]v1.LimitRange, error) {
			return collectors.ListLimitRanges(ctx, k8sClient, ns)
		}),
		newScanResourceCmd("resourcequotas", func(ns string) ([]v1.ResourceQuota, error) {
			return collectors.ListResourceQuotas(ctx, k8sClient, ns)
		}),
		newScanResourceCmd("networkpolicies", func(ns string) ([]networkingv1.NetworkPolicy, error) {
			return collectors.ListNetworkPolicies(ctx, k8sClient, ns)
		}),
		newScanResourceCmd("podsmetrics", func(ns string) ([]metricsv1beta1.PodMetrics, error) {
			return collectors.ListPodsMetrics(ctx, mclientset, ns)
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

			if outputFlag == "json" {
				path := fmt.Sprintf("output/scan_%s_results.json", name)
				var payload interface{} = results
				if secrets, ok := any(results).([]v1.Secret); ok {
					payload = redactSecrets(secrets)
				}
				if err := jsonoutput.Write(path, payload); err != nil {
					logger.LogFatal("failed to write JSON output: %v", err)
					return
				}
				logger.LogInfo("wrote %s to %s", name, path)
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
		if info, ok := results.(*version.Info); ok && info != nil {
			resourceprinter.PrintClusterInfo(info)
		}
	case "limitranges":
		resourceprinter.PrintLimitRanges(results.([]v1.LimitRange))
	case "resourcequotas":
		resourceprinter.PrintResourceQuotas(results.([]v1.ResourceQuota))
	case "networkpolicies":
		resourceprinter.PrintNetworkPolicies(results.([]networkingv1.NetworkPolicy))
	case "podsmetrics":
		resourceprinter.PrintPodMetrics(results.([]metricsv1beta1.PodMetrics))
	}
}

func runFullScan(namespace string) {
	logger.LogInfo("Fetching cluster information...")
	var results scanResults
	var wg sync.WaitGroup
	var mu sync.Mutex

	run := func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn()
		}()
	}

	run(func() {
		cluster, err := collectors.ClusterInfo(k8sClient)
		if err != nil {
			logger.LogError("failed to get cluster info: %v", err)
		}
		mu.Lock()
		results.Cluster = cluster
		mu.Unlock()
	})
	run(func() {
		nodes, err := collectors.ListNodes(ctx, k8sClient)
		if err != nil {
			logger.LogError("failed to list nodes: %v", err)
		}
		mu.Lock()
		results.Nodes = nodes
		mu.Unlock()
	})
	run(func() {
		namespaces, err := collectors.ListNamespaces(ctx, k8sClient)
		if err != nil {
			logger.LogError("failed to list namespaces: %v", err)
		}
		mu.Lock()
		results.Namespaces = namespaces
		mu.Unlock()
	})
	run(func() {
		pods, err := collectors.ListPods(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to list pods: %v", err)
		}
		mu.Lock()
		results.Pods = pods
		mu.Unlock()
	})
	run(func() {
		events, err := collectors.ListEvents(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to list events: %v", err)
		}
		mu.Lock()
		results.Events = events
		mu.Unlock()
	})
	run(func() {
		deployments, err := collectors.ListDeployments(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to list deployments: %v", err)
		}
		mu.Lock()
		results.Deployments = deployments
		mu.Unlock()
	})
	run(func() {
		secrets, err := collectors.ListSecrets(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to list secrets: %v", err)
		}
		mu.Lock()
		results.Secrets = secrets
		mu.Unlock()
	})
	run(func() {
		services, err := collectors.ListServices(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to list services: %v", err)
		}
		mu.Lock()
		results.Services = services
		mu.Unlock()
	})
	run(func() {
		configmaps, err := collectors.ListConfigMaps(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to list configmaps: %v", err)
		}
		mu.Lock()
		results.ConfigMaps = configmaps
		mu.Unlock()
	})
	run(func() {
		ingresses, err := collectors.ListIngresses(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to list ingresses: %v", err)
		}
		mu.Lock()
		results.Ingresses = ingresses
		mu.Unlock()
	})
	run(func() {
		nodesMetrics, err := collectors.NodesMetrics(ctx, mclientset)
		if err != nil {
			logger.LogError("failed to list node metrics: %v", err)
		}
		mu.Lock()
		results.NodesMetrics = nodesMetrics
		mu.Unlock()
	})
	run(func() {
		limitRanges, err := collectors.ListLimitRanges(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to list limit ranges: %v", err)
		}
		mu.Lock()
		results.LimitRanges = limitRanges
		mu.Unlock()
	})
	run(func() {
		resourceQuotas, err := collectors.ListResourceQuotas(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to list resource quotas: %v", err)
		}
		mu.Lock()
		results.ResourceQuotas = resourceQuotas
		mu.Unlock()
	})
	run(func() {
		networkPolicies, err := collectors.ListNetworkPolicies(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to list network policies: %v", err)
		}
		mu.Lock()
		results.NetworkPolicies = networkPolicies
		mu.Unlock()
	})
	run(func() {
		podsMetrics, err := collectors.ListPodsMetrics(ctx, mclientset, namespace)
		if err != nil {
			logger.LogError("failed to list pod metrics: %v", err)
		}
		mu.Lock()
		results.PodsMetrics = podsMetrics
		mu.Unlock()
	})

	wg.Wait()

	persistScan(namespace, results)

	if outputFlag == "json" {
		path := "output/scan_results.json"
		payload := scanResultsJSON{
			scanResults:     results,
			SecretsRedacted: redactSecrets(results.Secrets),
		}
		if err := jsonoutput.Write(path, payload); err != nil {
			logger.LogFatal("failed to write JSON output: %v", err)
			return
		}
		logger.LogInfo("wrote full scan results to %s", path)
		return
	}

	fmt.Println("")
	printResults(results.Cluster, "cluster")
	printResults(results.Nodes, "nodes")
	printResults(results.Namespaces, "namespaces")
	printResults(results.Pods, "pods")
	printResults(results.Events, "events")
	printResults(results.Deployments, "deployments")
	printResults(results.Secrets, "secrets")
	printResults(results.Services, "services")
	printResults(results.ConfigMaps, "configmaps")
	printResults(results.Ingresses, "ingresses")
	printResults(results.NodesMetrics, "metrics")
	printResults(results.LimitRanges, "limitranges")
	printResults(results.ResourceQuotas, "resourcequotas")
	printResults(results.NetworkPolicies, "networkpolicies")
	printResults(results.PodsMetrics, "podsmetrics")

}

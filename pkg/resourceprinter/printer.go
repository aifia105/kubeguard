package resourceprinter

import (
	"github.com/aifia105/kubeguard/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/version"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func PrintPods(pods []v1.Pod) {
	logger.LogSuccess("Pods: %d", len(pods))
	for _, p := range pods {
		ready := 0
		for _, cs := range p.Status.ContainerStatuses {
			if cs.Ready {
				ready++
			}
		}
		logger.LogInfo("  %s/%s [%s] ready:%d/%d restarts:%d",
			p.Namespace, p.Name, logger.PhaseColor(p.Status.Phase),
			ready, len(p.Spec.Containers), totalRestarts(p))
	}
}

func totalRestarts(p v1.Pod) int32 {
	var total int32
	for _, cs := range p.Status.ContainerStatuses {
		total += cs.RestartCount
	}
	return total
}

func PrintNodes(nodes []v1.Node) {
	logger.LogSuccess("Nodes: %d", len(nodes))
	for _, n := range nodes {
		status := "NotReady"
		for _, cond := range n.Status.Conditions {
			if cond.Type == v1.NodeReady && cond.Status == v1.ConditionTrue {
				status = "Ready"
			}
		}
		logger.LogInfo("  %s [%s] version:%s", n.Name, status, n.Status.NodeInfo.KubeletVersion)
	}
}

func PrintNamespaces(namespaces []v1.Namespace) {
	logger.LogSuccess("Namespaces: %d", len(namespaces))
	for _, n := range namespaces {
		logger.LogInfo("  %s [%s]", n.Name, n.Status.Phase)
	}
}

func PrintSecrets(secrets []v1.Secret) {
	logger.LogSuccess("Secrets: %d", len(secrets))
	for _, s := range secrets {
		logger.LogInfo("  %s/%s type:%s keys:%d", s.Namespace, s.Name, s.Type, len(s.Data))
	}
}

func PrintServices(services []v1.Service) {
	logger.LogSuccess("Services: %d", len(services))
	for _, s := range services {
		logger.LogInfo("  %s/%s type:%s clusterIP:%s", s.Namespace, s.Name, s.Spec.Type, s.Spec.ClusterIP)
	}
}

func PrintDeployments(deployments []appsv1.Deployment) {
	logger.LogSuccess("Deployments: %d", len(deployments))
	for _, d := range deployments {
		logger.LogInfo("  %s/%s ready:%d/%d", d.Namespace, d.Name, d.Status.ReadyReplicas, d.Status.Replicas)
	}
}

func PrintConfigMaps(configmaps []v1.ConfigMap) {
	logger.LogSuccess("ConfigMaps: %d", len(configmaps))
	for _, c := range configmaps {
		logger.LogInfo("  %s/%s keys:%d", c.Namespace, c.Name, len(c.Data))
	}
}

func PrintIngresses(ingresses []networkingv1.Ingress) {
	logger.LogSuccess("Ingresses: %d", len(ingresses))
	for _, i := range ingresses {
		hosts := ""
		for idx, rule := range i.Spec.Rules {
			if idx > 0 {
				hosts += ", "
			}
			hosts += rule.Host
		}
		logger.LogInfo("  %s/%s hosts:[%s]", i.Namespace, i.Name, hosts)
	}
}

func PrintEvents(events []v1.Event) {
	logger.LogSuccess("Events: %d", len(events))
	for _, e := range events {
		logger.LogInfo("  %s/%s [%s] %s: %s", e.Namespace, e.Name, e.Type, e.Reason, e.Message)
	}
}

func PrintNodeMetrics(metrics []metricsv1beta1.NodeMetrics) {
	logger.LogSuccess("Node Metrics: %d", len(metrics))
	for _, m := range metrics {
		logger.LogInfo("  %s CPU:%s Memory:%s", m.Name, m.Usage.Cpu().String(), m.Usage.Memory().String())
	}
}

func PrintClusterInfo(info *version.Info) {
	if info == nil {
		logger.LogError("cluster info unavailable")
		return
	}
	logger.LogSuccess("Cluster version: %s", info.String())
}

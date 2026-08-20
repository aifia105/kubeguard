package cmd

import (
	"sort"
	"sync"

	"github.com/aifia105/kubeguard/pkg/audit"
	"github.com/aifia105/kubeguard/pkg/audit/registry"
	"github.com/aifia105/kubeguard/pkg/collectors"
	"github.com/aifia105/kubeguard/pkg/logger"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

var serverityRank = map[audit.Severity]int{
	audit.SeverityCritical: 0,
	audit.SeverityHigh:     1,
	audit.SeverityMedium:   2,
	audit.SeverityLow:      3,
}

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Audit the Kubernetes cluster for security risks and misconfigurations",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ns := resolveNamespace(args)
		runFullAudit(ns)
	},
}

func init() {
	auditCmd.AddCommand(
		newAuditResourceCmd("pods", "Pod", func(ns string) ([]v1.Pod, error) {
			return collectors.ListPods(ctx, k8sClient, ns)
		}),
		newAuditResourceCmd("services", "Service", func(ns string) ([]v1.Service, error) {
			return collectors.ListServices(ctx, k8sClient, ns)
		}),
		newAuditResourceCmd("nodes", "Node", func(ns string) ([]v1.Node, error) {
			return collectors.ListNodes(ctx, k8sClient)
		}),
		newAuditResourceCmd("namespaces", "Namespace", func(ns string) ([]v1.Namespace, error) {
			return collectors.ListNamespaces(ctx, k8sClient)
		}),
		newAuditResourceCmd("secrets", "Secret", func(ns string) ([]v1.Secret, error) {
			return collectors.ListSecrets(ctx, k8sClient, ns)
		}),

		newAuditResourceCmd("deployments", "Deployment", func(ns string) ([]appsv1.Deployment, error) {
			return collectors.ListDeployments(ctx, k8sClient, ns)
		}),
		newAuditResourceCmd("configmaps", "ConfigMap", func(ns string) ([]v1.ConfigMap, error) {
			return collectors.ListConfigMaps(ctx, k8sClient, ns)
		}),
		newAuditResourceCmd("ingresses", "Ingress", func(ns string) ([]networkingv1.Ingress, error) {
			return collectors.ListIngresses(ctx, k8sClient, ns)
		}))
}

func newAuditResourceCmd[T any](resourceName string, resourceKey string, collect func(ns string) ([]T, error)) *cobra.Command {
	return &cobra.Command{
		Use:   resourceName + " [namespace]",
		Short: "Audit " + resourceName,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ns := resolveNamespace(args)
			results, err := collect(ns)
			if err != nil {
				logger.LogError("failed to collect %s: %v", resourceName, err)
				return
			}
			rules := registry.GetRulesForResource(resourceKey)
			findings := audit.RunAudit(results, rules)
			printFindings(findings, resourceKey)
		},
	}
}

func printFindings(findings []audit.Finding, resourceName string) {
	if len(findings) == 0 {
		logger.LogSuccess("%s: no issues found", resourceName)
		return
	}

	sortFindingsBySeverity(findings)

	logger.LogWarning("%s: %d issue(s) found", resourceName, len(findings))
	for _, f := range findings {
		logFinding(f)
	}
}

func logFinding(f audit.Finding) {
	switch f.Severity {
	case audit.SeverityCritical, audit.SeverityHigh:
		logger.LogError("[%s] %s/%s - %s: %s", f.Severity, f.Namespace, f.Name, f.RuleID, f.Description)
	case audit.SeverityMedium:
		logger.LogWarning("[%s] %s/%s - %s: %s", f.Severity, f.Namespace, f.Name, f.RuleID, f.Description)
	case audit.SeverityLow:
		logger.LogInfo("[%s] %s/%s - %s: %s", f.Severity, f.Namespace, f.Name, f.RuleID, f.Description)
	}
}

func sortFindingsBySeverity(findings []audit.Finding) {
	sort.SliceStable(findings, func(i, j int) bool {
		return serverityRank[findings[i].Severity] < serverityRank[findings[j].Severity]
	})
}

func runFullAudit(ns string) {
	logger.LogInfo("Running full cluster audit...")
	var (
		mu          sync.Mutex
		wg          sync.WaitGroup
		allFindings []audit.Finding
	)

	run := func(resouceName string, fn func() []audit.Finding) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			findings := fn()
			mu.Lock()
			allFindings = append(allFindings, findings...)
			mu.Unlock()
		}()
	}

	run("pods", func() []audit.Finding {
		pods, err := collectors.ListPods(ctx, k8sClient, ns)
		if err != nil {
			logger.LogError("failed to collect pods: %v", err)
			return nil
		}
		return audit.RunAudit(pods, registry.GetRulesForResource("Pod"))
	})

	run("services", func() []audit.Finding {
		services, err := collectors.ListServices(ctx, k8sClient, ns)
		if err != nil {
			logger.LogError("failed to collect services: %v", err)
			return nil
		}
		return audit.RunAudit(services, registry.GetRulesForResource("Service"))
	})

	run("nodes", func() []audit.Finding {
		nodes, err := collectors.ListNodes(ctx, k8sClient)
		if err != nil {
			logger.LogError("failed to collect nodes: %v", err)
			return nil
		}
		return audit.RunAudit(nodes, registry.GetRulesForResource("Node"))
	})

	run("namespaces", func() []audit.Finding {
		namespaces, err := collectors.ListNamespaces(ctx, k8sClient)
		if err != nil {
			logger.LogError("failed to collect namespaces: %v", err)
			return nil
		}
		return audit.RunAudit(namespaces, registry.GetRulesForResource("Namespace"))
	})

	run("secrets", func() []audit.Finding {
		secrets, err := collectors.ListSecrets(ctx, k8sClient, ns)
		if err != nil {
			logger.LogError("failed to collect secrets: %v", err)
			return nil
		}
		return audit.RunAudit(secrets, registry.GetRulesForResource("Secret"))
	})

	run("deployments", func() []audit.Finding {
		deployments, err := collectors.ListDeployments(ctx, k8sClient, ns)
		if err != nil {
			logger.LogError("failed to collect deployments: %v", err)
			return nil
		}
		return audit.RunAudit(deployments, registry.GetRulesForResource("Deployment"))
	})

	run("configmaps", func() []audit.Finding {
		configmaps, err := collectors.ListConfigMaps(ctx, k8sClient, ns)
		if err != nil {
			logger.LogError("failed to collect configmaps: %v", err)
			return nil
		}
		return audit.RunAudit(configmaps, registry.GetRulesForResource("ConfigMap"))
	})

	run("ingresses", func() []audit.Finding {
		ingresses, err := collectors.ListIngresses(ctx, k8sClient, ns)
		if err != nil {
			logger.LogError("failed to collect ingresses: %v", err)
			return nil
		}
		return audit.RunAudit(ingresses, registry.GetRulesForResource("Ingress"))
	})

	wg.Wait()
	printFullAuditReport(allFindings)

}

func printFullAuditReport(findings []audit.Finding) {
	if len(findings) == 0 {
		logger.LogSuccess("Full cluster audit: no issues found")
		return
	}

	sortFindingsBySeverity(findings)
	counts := map[audit.Severity]int{}
	for _, f := range findings {
		counts[f.Severity]++
	}

	logger.LogWarning("Full audit complete: %d issue(s) found "+
		"(critical:%d high:%d medium:%d low:%d)",
		len(findings),
		counts[audit.SeverityCritical], counts[audit.SeverityHigh],
		counts[audit.SeverityMedium], counts[audit.SeverityLow])

	for _, f := range findings {
		logFinding(f)
	}
}

package cmd

import (
	"fmt"
	"sort"
	"sync"

	"github.com/aifia105/kubeguard/pkg/audit"
	"github.com/aifia105/kubeguard/pkg/audit/checks"
	"github.com/aifia105/kubeguard/pkg/audit/registry"
	"github.com/aifia105/kubeguard/pkg/collectors"
	"github.com/aifia105/kubeguard/pkg/jsonoutput"
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

type auditReport struct {
	Resource      string                 `json:"resource"`
	TotalFindings int                    `json:"totalFindings"`
	Counts        map[audit.Severity]int `json:"counts"`
	Findings      []audit.Finding        `json:"findings"`
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
		newAuditResourceCmd("deployments", "Deployment", func(ns string) ([]appsv1.Deployment, error) {
			return collectors.ListDeployments(ctx, k8sClient, ns)
		}),
		newAuditResourceCmd("ingresses", "Ingress", func(ns string) ([]networkingv1.Ingress, error) {
			return collectors.ListIngresses(ctx, k8sClient, ns)
		}))

	auditCmd.AddCommand(newSecretsAuditCmd())
	auditCmd.AddCommand(newConfigMapsAuditCmd())
	auditCmd.AddCommand(newNamespacesAuditCmd())
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
			printFindings(findings, resourceName)
			persistRun(clusterID, "audit", ns, findings)
		},
	}
}

func newSecretsAuditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "secrets [namespace]",
		Short: "Audit secrets",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ns := resolveNamespace(args)
			secrets, err := collectors.ListSecrets(ctx, k8sClient, ns)
			if err != nil {
				logger.LogError("failed to list secrets: %v", err)
				return
			}
			pods, err := collectors.ListPods(ctx, k8sClient, ns)
			if err != nil {
				logger.LogError("failed to list pods: %v", err)
				return
			}

			findings := audit.RunAudit(secrets, registry.GetRulesForResource("Secret"))
			findings = append(findings, audit.RunMultiResourceAudit([]interface{}{secrets, pods}, []audit.Rule{checks.UnreferencedSecret})...)
			printFindings(findings, "secrets")
		},
	}
}

func newConfigMapsAuditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "configmaps [namespace]",
		Short: "Audit configmaps",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ns := resolveNamespace(args)
			configmaps, err := collectors.ListConfigMaps(ctx, k8sClient, ns)
			if err != nil {
				logger.LogError("failed to list configmaps: %v", err)
				return
			}
			pods, err := collectors.ListPods(ctx, k8sClient, ns)
			if err != nil {
				logger.LogError("failed to list pods: %v", err)
				return
			}

			findings := audit.RunAudit(configmaps, registry.GetRulesForResource("ConfigMap"))
			findings = append(findings, audit.RunMultiResourceAudit([]interface{}{configmaps, pods}, []audit.Rule{checks.UnreferencedConfigMap})...)
			printFindings(findings, "configmaps")
		},
	}
}

func newNamespacesAuditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "namespaces [namespace]",
		Short: "Audit namespaces",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			ns := resolveNamespace(args)
			namespaces, err := collectors.ListNamespaces(ctx, k8sClient)
			if ns != "" {
				var filtered []v1.Namespace
				for _, n := range namespaces {
					if n.Name == ns {
						filtered = append(filtered, n)
					}
				}
				namespaces = filtered
			}
			if err != nil {
				logger.LogError("failed to list namespaces: %v", err)
				return
			}
			networkPolicies, _ := collectors.ListNetworkPolicies(ctx, k8sClient, ns)
			resourceQuotas, _ := collectors.ListResourceQuotas(ctx, k8sClient, ns)
			limitRanges, _ := collectors.ListLimitRanges(ctx, k8sClient, ns)

			var findings []audit.Finding
			findings = append(findings, audit.RunMultiResourceAudit([]interface{}{namespaces, networkPolicies}, []audit.Rule{checks.NoDefaultDenyNetworkPolicy})...)
			findings = append(findings, audit.RunMultiResourceAudit([]interface{}{namespaces, resourceQuotas}, []audit.Rule{checks.NoResourceQuotas})...)
			findings = append(findings, audit.RunMultiResourceAudit([]interface{}{namespaces, limitRanges}, []audit.Rule{checks.NoLimitRanges})...)

			printFindings(findings, "namespaces")
		},
	}
}

func printFindings(findings []audit.Finding, resourceName string) {
	if outputFlag == "json" {
		writeFindingsJSON(findings, resourceName)
		return
	}

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

func CollectFullAuditFindings(namespace string) []audit.Finding {
	var (
		mu          sync.Mutex
		wg          sync.WaitGroup
		allFindings []audit.Finding
	)

	run := func(fn func() []audit.Finding) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			findings := fn()
			mu.Lock()
			allFindings = append(allFindings, findings...)
			mu.Unlock()
		}()
	}

	run(func() []audit.Finding {
		pods, err := collectors.ListPods(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to collect pods: %v", err)
			return nil
		}
		return audit.RunAudit(pods, registry.GetRulesForResource("Pod"))
	})
	run(func() []audit.Finding {
		services, err := collectors.ListServices(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to collect services: %v", err)
			return nil
		}
		return audit.RunAudit(services, registry.GetRulesForResource("Service"))
	})
	run(func() []audit.Finding {
		nodes, err := collectors.ListNodes(ctx, k8sClient)
		if err != nil {
			logger.LogError("failed to collect nodes: %v", err)
			return nil
		}
		return audit.RunAudit(nodes, registry.GetRulesForResource("Node"))
	})
	run(func() []audit.Finding {
		deployments, err := collectors.ListDeployments(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to collect deployments: %v", err)
			return nil
		}
		return audit.RunAudit(deployments, registry.GetRulesForResource("Deployment"))
	})
	run(func() []audit.Finding {
		ingresses, err := collectors.ListIngresses(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to collect ingresses: %v", err)
			return nil
		}
		return audit.RunAudit(ingresses, registry.GetRulesForResource("Ingress"))
	})
	run(func() []audit.Finding {
		secrets, err := collectors.ListSecrets(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to list secrets: %v", err)
			return nil
		}
		pods, err := collectors.ListPods(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to list pods: %v", err)
			return nil
		}
		findings := audit.RunAudit(secrets, registry.GetRulesForResource("Secret"))
		findings = append(findings, audit.RunMultiResourceAudit([]interface{}{secrets, pods}, []audit.Rule{checks.UnreferencedSecret})...)
		return findings
	})
	run(func() []audit.Finding {
		configmaps, err := collectors.ListConfigMaps(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to collect configmaps: %v", err)
			return nil
		}
		pods, err := collectors.ListPods(ctx, k8sClient, namespace)
		if err != nil {
			logger.LogError("failed to list pods: %v", err)
			return nil
		}
		findings := audit.RunAudit(configmaps, registry.GetRulesForResource("ConfigMap"))
		findings = append(findings, audit.RunMultiResourceAudit([]interface{}{configmaps, pods}, []audit.Rule{checks.UnreferencedConfigMap})...)
		return findings
	})
	run(func() []audit.Finding {
		namespaces, err := collectors.ListNamespaces(ctx, k8sClient)
		if err != nil {
			logger.LogError("failed to collect namespaces: %v", err)
			return nil
		}
		if namespace != "" {
			var filtered []v1.Namespace
			for _, n := range namespaces {
				if n.Name == namespace {
					filtered = append(filtered, n)
				}
			}
			namespaces = filtered
		}
		networkPolicies, _ := collectors.ListNetworkPolicies(ctx, k8sClient, namespace)
		resourceQuotas, _ := collectors.ListResourceQuotas(ctx, k8sClient, namespace)
		limitRanges, _ := collectors.ListLimitRanges(ctx, k8sClient, namespace)

		var findings []audit.Finding
		findings = append(findings, audit.RunMultiResourceAudit([]interface{}{namespaces, networkPolicies}, []audit.Rule{checks.NoDefaultDenyNetworkPolicy})...)
		findings = append(findings, audit.RunMultiResourceAudit([]interface{}{namespaces, resourceQuotas}, []audit.Rule{checks.NoResourceQuotas})...)
		findings = append(findings, audit.RunMultiResourceAudit([]interface{}{namespaces, limitRanges}, []audit.Rule{checks.NoLimitRanges})...)
		return findings
	})

	wg.Wait()
	return allFindings
}

func runFullAudit(ns string) {
	logger.LogInfo("Running full cluster audit...")

	allFindings := CollectFullAuditFindings(ns)
	printFullAuditReport(allFindings)
	persistRun(clusterID, "audit", ns, allFindings)

}

func writeFindingsJSON(findings []audit.Finding, resourceName string) {
	sortFindingsBySeverity(findings)

	counts := map[audit.Severity]int{}
	for _, f := range findings {
		counts[f.Severity]++
	}

	report := auditReport{
		Resource:      resourceName,
		TotalFindings: len(findings),
		Counts:        counts,
		Findings:      findings,
	}

	path := fmt.Sprintf("output/audit_%s_results.json", resourceName)
	if resourceName == "full" {
		path = "output/audit_results.json"
	}

	if err := jsonoutput.Write(path, report); err != nil {
		logger.LogFatal("failed to write JSON output: %v", err)
		return
	}
	logger.LogInfo("wrote %s audit results to %s", resourceName, path)
}

func printFullAuditReport(findings []audit.Finding) {
	if outputFlag == "json" {
		writeFindingsJSON(findings, "full")
		return
	}

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

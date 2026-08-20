package checks

// import (
// 	"github.com/aifia105/kubeguard/pkg/audit"
// 	v1 "k8s.io/api/core/v1"
// )

// func namespaceFindingBuilder(namespace v1.Namespace, ruleID string, severity audit.Severity, description string) audit.Finding {
// 	return audit.Finding{
// 		Namespace:   namespace.Name,
// 		RuleID:      ruleID,
// 		Severity:    severity,
// 		Description: description,
// 	}
// }

// var NoDefaultDenyNetworkPolicy = audit.Rule{
// 	RuleID:   "NO_DEFAULT_DENY_NETWORK_POLICY",
// 	Severity: audit.SeverityHigh,
// 	Resource: "Namespace",
// 	Check: func(resource interface{}) []audit.Finding {
// 		namespace := resource.(v1.Namespace)
// 		var findings []audit.Finding
// 		for _, annotation := range namespace.NetworkPolicies {
// 			if annotation == "default-deny" {
// 				findings = append(findings, namespaceFindingBuilder(namespace, "NO_DEFAULT_DENY_NETWORK_POLICY", audit.SeverityHigh, "Namespace has no default deny network policy"))
// 			}
// 		}

// 		return findings
// 	},
// }

// var NoResourceQuotas = audit.Rule{
// 	RuleID:   "NO_RESOURCE_QUOTAS",
// 	Severity: audit.SeverityMedium,
// 	Resource: "Namespace",
// 	Check: func(resource interface{}) []audit.Finding {
// 		namespace := resource.(v1.Namespace)
// 		var findings []audit.Finding
// 		if len(namespace.Spec.ResourceQuotas) == 0 {
// 			findings = append(findings, namespaceFindingBuilder(namespace, "NO_RESOURCE_QUOTAS", audit.SeverityMedium, "Namespace has no resource quotas"))
// 		}
// 		return findings
// 	},
// }

// var NoLimitRanges = audit.Rule{
// 	RuleID:   "NO_LIMIT_RANGES",
// 	Severity: audit.SeverityMedium,
// 	Resource: "Namespace",
// 	Check: func(resource interface{}) []audit.Finding {
// 		namespace := resource.(v1.Namespace)
// 		var findings []audit.Finding
// 		if len(namespace.Spec.LimitRanges) == 0 {
// 			findings = append(findings, namespaceFindingBuilder(namespace, "NO_LIMIT_RANGES", audit.SeverityMedium, "Namespace has no limit ranges"))
// 		}
// 		return findings
// 	},
// }

// var NamespaceRules = []audit.Rule{
// 	NoDefaultDenyNetworkPolicy,
// 	NoResourceQuotas,
// 	NoLimitRanges,
// }

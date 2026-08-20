package checks

import (
	"github.com/aifia105/kubeguard/pkg/audit"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

func namespaceFindingBuilder(namespace v1.Namespace, ruleID string, severity audit.Severity, description string) audit.Finding {
	return audit.Finding{
		Namespace:   namespace.Name,
		RuleID:      ruleID,
		Severity:    severity,
		Description: description,
	}
}

func isSystemNamespace(name string) bool {
	return name == "kube-system" || name == "kube-public" || name == "kube-node-lease"
}

func isDefaultDenyPolicy(np networkingv1.NetworkPolicy) bool {
	selectsAllPods := len(np.Spec.PodSelector.MatchLabels) == 0 && len(np.Spec.PodSelector.MatchExpressions) == 0
	hasIngressType := false
	for _, t := range np.Spec.PolicyTypes {
		if t == networkingv1.PolicyTypeIngress {
			hasIngressType = true
		}
	}
	return selectsAllPods && hasIngressType && len(np.Spec.Ingress) == 0
}

var NoDefaultDenyNetworkPolicy = audit.Rule{
	RuleID:   "NO_DEFAULT_DENY_NETWORK_POLICY",
	Severity: audit.SeverityHigh,
	Resource: "Namespace",
	MultiCheck: func(resources []interface{}) []audit.Finding {
		var findings []audit.Finding
		if len(resources) != 2 {
			return findings
		}
		namespaces, ok1 := resources[0].([]v1.Namespace)
		policies, ok2 := resources[1].([]networkingv1.NetworkPolicy)
		if !ok1 || !ok2 {
			return findings
		}
		for _, ns := range namespaces {
			if isSystemNamespace(ns.Name) {
				continue
			}
			hasDefaultDeny := false
			for _, np := range policies {
				if np.Namespace == ns.Name && isDefaultDenyPolicy(np) {
					hasDefaultDeny = true
					break
				}
			}
			if !hasDefaultDeny {
				findings = append(findings, namespaceFindingBuilder(ns, "NO_DEFAULT_DENY_NETWORK_POLICY", audit.SeverityHigh,
					"Namespace has no default-deny NetworkPolicy"))
			}
		}
		return findings
	},
}

var NoResourceQuotas = audit.Rule{
	RuleID:   "NO_RESOURCE_QUOTAS",
	Severity: audit.SeverityMedium,
	Resource: "Namespace",
	MultiCheck: func(resources []interface{}) []audit.Finding {
		var findings []audit.Finding
		if len(resources) != 2 {
			return findings
		}
		namespaces, ok1 := resources[0].([]v1.Namespace)
		quotas, ok2 := resources[1].([]v1.ResourceQuota)
		if !ok1 || !ok2 {
			return findings
		}
		for _, ns := range namespaces {
			if isSystemNamespace(ns.Name) {
				continue
			}
			hasQuota := false
			for _, q := range quotas {
				if q.Namespace == ns.Name {
					hasQuota = true
					break
				}
			}
			if !hasQuota {
				findings = append(findings, namespaceFindingBuilder(ns, "NO_RESOURCE_QUOTAS", audit.SeverityMedium,
					"Namespace has no ResourceQuota defined"))
			}
		}
		return findings
	},
}

var NoLimitRanges = audit.Rule{
	RuleID:   "NO_LIMIT_RANGES",
	Severity: audit.SeverityMedium,
	Resource: "Namespace",
	MultiCheck: func(resources []interface{}) []audit.Finding {
		var findings []audit.Finding
		if len(resources) != 2 {
			return findings
		}
		namespaces, ok1 := resources[0].([]v1.Namespace)
		limitRanges, ok2 := resources[1].([]v1.LimitRange)
		if !ok1 || !ok2 {
			return findings
		}
		for _, ns := range namespaces {
			if isSystemNamespace(ns.Name) {
				continue
			}
			hasLimitRange := false
			for _, lr := range limitRanges {
				if lr.Namespace == ns.Name {
					hasLimitRange = true
					break
				}
			}
			if !hasLimitRange {
				findings = append(findings, namespaceFindingBuilder(ns, "NO_LIMIT_RANGES", audit.SeverityMedium,
					"Namespace has no LimitRange defined"))
			}
		}
		return findings
	},
}

var NamespaceRules = []audit.Rule{
	NoDefaultDenyNetworkPolicy,
	NoResourceQuotas,
	NoLimitRanges,
}

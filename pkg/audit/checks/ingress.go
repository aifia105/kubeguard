package checks

import (
	"github.com/aifia105/kubeguard/pkg/audit"
	v1 "k8s.io/api/networking/v1"
)

func ingressFindingBuilder(ingress v1.Ingress, ruleID string, severity audit.Severity, description string) audit.Finding {
	return audit.Finding{
		Namespace:   ingress.Namespace,
		Name:        ingress.Name,
		RuleID:      ruleID,
		Severity:    severity,
		Description: description,
	}
}

var NoTLSConfigured = audit.Rule{
	RuleID:   "NO_TLS_CONFIGURED",
	Severity: audit.SeverityHigh,
	Resource: "Ingress",
	Check: func(resource interface{}) []audit.Finding {
		ingress := resource.(v1.Ingress)
		var findings []audit.Finding
		if len(ingress.Spec.TLS) == 0 {
			for _, rule := range ingress.Spec.Rules {
				if rule.Host != "" {
					findings = append(findings, ingressFindingBuilder(ingress, "NO_TLS_CONFIGURED", audit.SeverityHigh, "Ingress has no TLS configured for host "+rule.Host))
				}
			}
		}
		return findings
	},
}

var WildcardHostConfigured = audit.Rule{
	RuleID:   "WILDCARD_HOST_CONFIGURED",
	Severity: audit.SeverityMedium,
	Resource: "Ingress",
	Check: func(resource interface{}) []audit.Finding {
		ingress := resource.(v1.Ingress)
		var findings []audit.Finding
		for _, rule := range ingress.Spec.Rules {
			if rule.Host == "" {
				findings = append(findings, ingressFindingBuilder(ingress, "WILDCARD_HOST_CONFIGURED", audit.SeverityMedium, "Ingress has a wildcard host configured"))
			}
		}
		return findings
	},
}

var IngressRules = []audit.Rule{
	NoTLSConfigured,
	WildcardHostConfigured,
}

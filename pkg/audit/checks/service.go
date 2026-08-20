package checks

import (
	"github.com/aifia105/kubeguard/pkg/audit"
	v1 "k8s.io/api/core/v1"
)

func serviceFindingBuilder(service v1.Service, ruleID string, severity audit.Severity, description string) audit.Finding {
	return audit.Finding{
		Namespace:   service.Namespace,
		Name:        service.Name,
		RuleID:      ruleID,
		Severity:    severity,
		Description: description,
	}
}

var PublicExposedViaLoadBalancer = audit.Rule{
	RuleID:   "PUBLIC_EXPOSED_VIA_LOAD_BALANCER",
	Severity: audit.SeverityHigh,
	Resource: "Service",
	Check: func(resource interface{}) []audit.Finding {
		service := resource.(v1.Service)
		var findings []audit.Finding
		if service.Spec.Type == v1.ServiceTypeLoadBalancer {
			findings = append(findings, serviceFindingBuilder(service, "PUBLIC_EXPOSED_VIA_LOAD_BALANCER", audit.SeverityHigh, "Service is exposed to the public via LoadBalancer"))
		}
		return findings
	},
}

var NodePortExposed = audit.Rule{
	RuleID:   "NODE_PORT_EXPOSED",
	Severity: audit.SeverityMedium,
	Resource: "Service",
	Check: func(resource interface{}) []audit.Finding {
		service := resource.(v1.Service)
		var findings []audit.Finding
		if service.Spec.Type == v1.ServiceTypeNodePort {
			findings = append(findings, serviceFindingBuilder(service, "NODE_PORT_EXPOSED", audit.SeverityMedium, "Service is exposed via NodePort"))
		}
		return findings
	},
}

var NoSelectorDefined = audit.Rule{
	RuleID:   "NO_SELECTOR_DEFINED",
	Severity: audit.SeverityLow,
	Resource: "Service",
	Check: func(resource interface{}) []audit.Finding {
		service := resource.(v1.Service)
		var findings []audit.Finding
		if len(service.Spec.Selector) == 0 {
			findings = append(findings, serviceFindingBuilder(service, "NO_SELECTOR_DEFINED", audit.SeverityLow, "Service has no selector defined"))
		}
		return findings
	},
}

var ServiceRules = []audit.Rule{
	PublicExposedViaLoadBalancer,
	NodePortExposed,
	NoSelectorDefined,
}

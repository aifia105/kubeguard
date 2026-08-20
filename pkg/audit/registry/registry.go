package registry

import (
	"github.com/aifia105/kubeguard/pkg/audit"
	"github.com/aifia105/kubeguard/pkg/audit/checks"
)

var AllRules = buildRegistry()

func buildRegistry() []audit.Rule {
	var rules []audit.Rule
	rules = append(rules, checks.PodRules...)
	rules = append(rules, checks.ServiceRules...)
	rules = append(rules, checks.SecretRules...)
	rules = append(rules, checks.IngressRules...)
	rules = append(rules, checks.DeploymentRules...)
	rules = append(rules, checks.ConfigMapRules...)
	rules = append(rules, checks.NodeRules...)
	rules = append(rules, checks.NamespaceRules...)
	return rules
}

func GetRulesForResource(resource string) []audit.Rule {
	var rules []audit.Rule
	for _, rule := range AllRules {
		if rule.Resource == resource {
			rules = append(rules, rule)
		}
	}
	return rules
}

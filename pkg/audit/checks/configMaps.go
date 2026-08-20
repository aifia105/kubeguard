package checks

import (
	"fmt"

	"github.com/aifia105/kubeguard/pkg/audit"
	v1 "k8s.io/api/core/v1"
)

func configMapFindingBuilder(configMap v1.ConfigMap, ruleID string, severity audit.Severity, description string) audit.Finding {
	return audit.Finding{
		Namespace:   configMap.Namespace,
		Name:        configMap.Name,
		RuleID:      ruleID,
		Severity:    severity,
		Description: description,
	}
}

var SecretLookingData = audit.Rule{
	RuleID:   "SECRET_LOOKING_DATA",
	Severity: audit.SeverityMedium,
	Resource: "ConfigMap",
	Check: func(resource interface{}) []audit.Finding {
		configMap := resource.(v1.ConfigMap)
		var findings []audit.Finding
		if configMap.Data != nil {
			for key := range configMap.Data {
				if IsSensitiveEnvName(key) {
					findings = append(findings, configMapFindingBuilder(configMap, "SECRET_LOOKING_DATA", audit.SeverityMedium, fmt.Sprintf("ConfigMap %q contains sensitive data", configMap.Name)))
				}
			}
		}
		return findings
	},
}

// var UnreferencedConfigMap = audit.Rule{
// 	RuleID:   "UNREFERENCED_CONFIGMAP",
// 	Severity: audit.SeverityMedium,
// 	Resource: "ConfigMap",
// 	Check: func(resource interface{}) []audit.Finding {
// 		configMap := resource.(v1.ConfigMap)
// 		var findings []audit.Finding
// 		if configMap.Data != nil {
// 			if len(configMap.Data) == 0 {
// 				findings = append(findings, configMapFindingBuilder(configMap, "UNREFERENCED_CONFIGMAP", audit.SeverityMedium, fmt.Sprintf("ConfigMap %q has no data", configMap.Name)))
// 			}
// 		}
// 		return findings
// 	},
// }

var ConfigMapRules = []audit.Rule{
	SecretLookingData,
	// UnreferencedConfigMap,
}

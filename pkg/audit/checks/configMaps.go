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

var UnreferencedConfigMap = audit.Rule{
	RuleID:   "UNREFERENCED_CONFIGMAP",
	Severity: audit.SeverityMedium,
	Resource: "ConfigMap",
	MultiCheck: func(resources []interface{}) []audit.Finding {
		var findings []audit.Finding
		if len(resources) != 2 {
			return findings
		}
		configMaps, ok1 := resources[0].([]v1.ConfigMap)
		pods, ok2 := resources[1].([]v1.Pod)
		if !ok1 || !ok2 {
			return findings
		}
		for _, cm := range configMaps {
			if cm.Name == "kube-root-ca.crt" {
				continue
			}
			referenced := false
			for _, pod := range pods {
				if pod.Namespace != cm.Namespace {
					continue
				}
				if configMapReferencedByPod(pod, cm.Name) {
					referenced = true
					break
				}
			}
			if !referenced {
				findings = append(findings, configMapFindingBuilder(cm, "UNREFERENCED_CONFIGMAP", audit.SeverityMedium,
					fmt.Sprintf("ConfigMap %q is not referenced by any pod in its namespace", cm.Name)))
			}
		}
		return findings
	},
}

func configMapReferencedByPod(pod v1.Pod, configMapName string) bool {
	for _, c := range pod.Spec.Containers {
		for _, ef := range c.EnvFrom {
			if ef.ConfigMapRef != nil && ef.ConfigMapRef.Name == configMapName {
				return true
			}
		}
		for _, env := range c.Env {
			if env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil && env.ValueFrom.ConfigMapKeyRef.Name == configMapName {
				return true
			}
		}
	}
	for _, v := range pod.Spec.Volumes {
		if v.ConfigMap != nil && v.ConfigMap.Name == configMapName {
			return true
		}
	}
	return false
}

var ConfigMapRules = []audit.Rule{
	SecretLookingData,
	UnreferencedConfigMap,
}

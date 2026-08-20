package checks

import (
	"fmt"
	"regexp"
	"time"

	"github.com/aifia105/kubeguard/pkg/audit"

	v1 "k8s.io/api/core/v1"
)

var opaqueSecretType = regexp.MustCompile(`(?i)(port|host|url|enabled|debug|config|setting|configuration)`)

func IsSensitiveOpaqueName(name string) bool {
	return opaqueSecretType.MatchString(name)
}

func secretFindingBuilder(secret v1.Secret, ruleID string, severity audit.Severity, description string) audit.Finding {
	return audit.Finding{
		Namespace:   secret.Namespace,
		Name:        secret.Name,
		RuleID:      ruleID,
		Severity:    severity,
		Description: description,
	}
}

var UnreferencedSecret = audit.Rule{
	RuleID:   "UNREFERENCED_SECRET",
	Severity: audit.SeverityMedium,
	Resource: "Secret",
	MultiCheck: func(resources []interface{}) []audit.Finding {
		var findings []audit.Finding
		if len(resources) != 2 {
			return findings
		}
		secrets, ok1 := resources[0].([]v1.Secret)
		pods, ok2 := resources[1].([]v1.Pod)
		if !ok1 || !ok2 {
			return findings
		}
		for _, secret := range secrets {
			if secret.Type == "helm.sh/release.v1" || secret.Type == v1.SecretTypeServiceAccountToken {
				continue
			}
			referenced := false
			for _, pod := range pods {
				if pod.Namespace != secret.Namespace {
					continue
				}
				if secretReferencedByPod(pod, secret.Name) {
					referenced = true
					break
				}
			}
			if !referenced {
				findings = append(findings, secretFindingBuilder(secret, "UNREFERENCED_SECRET", audit.SeverityMedium,
					fmt.Sprintf("secret %q is not referenced by any pod in its namespace", secret.Name)))
			}
		}
		return findings
	},
}

func secretReferencedByPod(pod v1.Pod, secretName string) bool {
	for _, c := range pod.Spec.Containers {
		for _, ef := range c.EnvFrom {
			if ef.SecretRef != nil && ef.SecretRef.Name == secretName {
				return true
			}
		}
		for _, env := range c.Env {
			if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Name == secretName {
				return true
			}
		}
	}
	for _, v := range pod.Spec.Volumes {
		if v.Secret != nil && v.Secret.SecretName == secretName {
			return true
		}
	}
	return false
}

var OldOrStaleSecret = audit.Rule{
	RuleID:   "OLD_OR_STALE_SECRET",
	Severity: audit.SeverityHigh,
	Resource: "Secret",
	Check: func(resource interface{}) []audit.Finding {
		secret := resource.(v1.Secret)
		var findings []audit.Finding
		if secret.Type == "helm.sh/release.v1" {
			return findings
		}

		if secret.CreationTimestamp.Add(30 * 24 * time.Hour).Before(time.Now()) {
			findings = append(findings, secretFindingBuilder(secret, "OLD_OR_STALE_SECRET", audit.SeverityHigh, fmt.Sprintf("secret %q is older than 30 days", secret.Name)))
		}
		return findings
	},
}

var DefaultTokenTypeOverSupplied = audit.Rule{
	RuleID:   "LEGACY_SERVICE_ACCOUNT_TOKEN",
	Severity: audit.SeverityMedium,
	Resource: "Secret",
	Check: func(resource interface{}) []audit.Finding {
		secret := resource.(v1.Secret)
		var findings []audit.Finding
		if secret.Type == v1.SecretTypeServiceAccountToken {
			findings = append(findings, secretFindingBuilder(secret, "LEGACY_SERVICE_ACCOUNT_TOKEN", audit.SeverityMedium, fmt.Sprintf("secret %q is of type ServiceAccountToken, which is legacy", secret.Name)))
		}
		return findings
	},
}

var OpaqueSecretStoringConfig = audit.Rule{
	RuleID:   "OPAQUE_SECRET_STORING_CONFIG",
	Severity: audit.SeverityLow,
	Resource: "Secret",
	Check: func(resource interface{}) []audit.Finding {
		secret := resource.(v1.Secret)
		var findings []audit.Finding
		if secret.Type == v1.SecretTypeOpaque && IsSensitiveOpaqueName(secret.Name) {
			findings = append(findings, secretFindingBuilder(secret, "OPAQUE_SECRET_STORING_CONFIG", audit.SeverityLow, fmt.Sprintf("this secret %q's name suggests it might hold config, not credentials", secret.Name)))
		}
		return findings
	},
}

var SecretRules = []audit.Rule{
	UnreferencedSecret,
	OldOrStaleSecret,
	DefaultTokenTypeOverSupplied,
	OpaqueSecretStoringConfig,
}

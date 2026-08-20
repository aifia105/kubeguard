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

// var UnreferencedSecret = audit.Rule{
// 	RuleID:   "UNREFERENCED_SECRET",
// 	Severity: audit.SeverityMedium,
// 	Resource: "Secret",
// 	Check: func(resource interface{}) []audit.Finding {
// 		secret := resource.(v1.Secret)
// 		var findings []audit.Finding
// 		if secret.Name == "" {
// 			findings = append(findings, secretFindingBuilder(secret, "UNREFERENCED_SECRET", audit.SeverityMedium, fmt.Sprintf("secret %q is unreferenced", secret.Name)))
// 		}
// 		return findings
// 	},
// }

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
	// UnreferencedSecret,
	OldOrStaleSecret,
	DefaultTokenTypeOverSupplied,
	OpaqueSecretStoringConfig,
}

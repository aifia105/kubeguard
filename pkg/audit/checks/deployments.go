package checks

import (
	"fmt"

	"github.com/aifia105/kubeguard/pkg/audit"
	v1 "k8s.io/api/apps/v1"
)

func deploymentFindingBuilder(deployment v1.Deployment, ruleID string, severity audit.Severity, description string) audit.Finding {
	return audit.Finding{
		Namespace:   deployment.Namespace,
		Name:        deployment.Name,
		RuleID:      ruleID,
		Severity:    severity,
		Description: description,
	}
}

var SingleReplicaNoRedundancy = audit.Rule{
	RuleID:   "SINGLE_REPLICA_NO_REDUNDANCY",
	Severity: audit.SeverityHigh,
	Resource: "Deployment",
	Check: func(resource interface{}) []audit.Finding {
		deployment := resource.(v1.Deployment)
		var findings []audit.Finding
		if deployment.Spec.Replicas == nil || *deployment.Spec.Replicas == 1 {
			findings = append(findings, deploymentFindingBuilder(deployment, "SINGLE_REPLICA_NO_REDUNDANCY", audit.SeverityHigh, fmt.Sprintf("Deployment %q has a single replica", deployment.Name)))
		}
		return findings
	},
}

var RollingUpdateStrategy = audit.Rule{
	RuleID:   "ROLLING_UPDATE_STRATEGY",
	Severity: audit.SeverityMedium,
	Resource: "Deployment",
	Check: func(resource interface{}) []audit.Finding {
		deployment := resource.(v1.Deployment)
		var findings []audit.Finding
		if deployment.Spec.Strategy.RollingUpdate == nil && deployment.Spec.Strategy.Type == v1.RollingUpdateDeploymentStrategyType {
			findings = append(findings, deploymentFindingBuilder(deployment, "ROLLING_UPDATE_STRATEGY", audit.SeverityMedium, fmt.Sprintf("Deployment %q uses default rolling update parameters (no explicit maxSurge/maxUnavailable)", deployment.Name)))
		}
		return findings
	},
}

var DeploymentRules = []audit.Rule{
	SingleReplicaNoRedundancy,
	RollingUpdateStrategy,
}

package checks

import (
	"github.com/aifia105/kubeguard/pkg/audit"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/version"
)

func nodeFindingBuilder(node v1.Node, ruleID string, severity audit.Severity, description string) audit.Finding {
	return audit.Finding{
		Name:        node.Name,
		RuleID:      ruleID,
		Severity:    severity,
		Description: description,
	}
}

var NodeNotReady = audit.Rule{
	RuleID:   "NODE_NOT_READY",
	Severity: audit.SeverityHigh,
	Resource: "Node",
	Check: func(resource interface{}) []audit.Finding {
		node := resource.(v1.Node)
		var findings []audit.Finding
		for _, condition := range node.Status.Conditions {
			if condition.Type == v1.NodeReady && condition.Status == v1.ConditionFalse {
				findings = append(findings, nodeFindingBuilder(node, "NODE_NOT_READY", audit.SeverityHigh, "Node is not ready"))
			}
		}

		return findings
	},
}

var DiskMemoryPIDPressure = audit.Rule{
	RuleID:   "DISK_MEMORY_PID_PRESSURE",
	Severity: audit.SeverityHigh,
	Resource: "Node",
	Check: func(resource interface{}) []audit.Finding {
		node := resource.(v1.Node)
		var findings []audit.Finding
		for _, condition := range node.Status.Conditions {
			if (condition.Type == v1.NodeDiskPressure || condition.Type == v1.NodeMemoryPressure || condition.Type == v1.NodePIDPressure) && condition.Status == v1.ConditionTrue {
				findings = append(findings, nodeFindingBuilder(node, "DISK_MEMORY_PID_PRESSURE", audit.SeverityHigh, "Node is under pressure"))
			}
		}
		return findings
	},
}

var OutDatedKubeletVersion = audit.Rule{
	RuleID:   "OUTDATED_KUBELET_VERSION",
	Severity: audit.SeverityMedium,
	Resource: "Node",
	Check: func(resource interface{}) []audit.Finding {
		node := resource.(v1.Node)
		var findings []audit.Finding
		minVersion := version.MustParseGeneric("v1.20.0")
		for _, condition := range node.Status.Conditions {
			if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
				kubeletVersion, err := version.ParseGeneric(node.Status.NodeInfo.KubeletVersion)
				if err != nil {

					findings = append(findings, nodeFindingBuilder(node, "UNPARSEABLE_KUBELET_VERSION", audit.SeverityLow, "Node kubelet version could not be parsed: "+node.Status.NodeInfo.KubeletVersion))
				} else if kubeletVersion.LessThan(minVersion) {
					findings = append(findings, nodeFindingBuilder(node, "OUTDATED_KUBELET_VERSION", audit.SeverityMedium, "Node is running an outdated kubelet version"))
				}
			}
		}
		return findings
	},
}

var NodeRules = []audit.Rule{
	NodeNotReady,
	DiskMemoryPIDPressure,
	OutDatedKubeletVersion,
}

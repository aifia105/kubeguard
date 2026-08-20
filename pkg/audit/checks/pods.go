package checks

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aifia105/kubeguard/pkg/audit"
	v1 "k8s.io/api/core/v1"
)

var SensitiveEnvPattern = regexp.MustCompile(`(?i)(password|passwd|pwd|secret|token|api[_-]?key|access[_-]?key|private[_-]?key|credential)`)

var dangerousCapabilitiesKeptStrings = []string{
	"NET_ADMIN",
	"SYS_ADMIN",
	"SYS_MODULE",
	"SYS_NICE",
	"SYS_PACCTL",
	"SYS_PTRACE",
	"SYS_RAWIO",
	"SYS_RESOURCE",
	"SYS_TIME",
}

func IsSensitiveEnvName(name string) bool {
	return SensitiveEnvPattern.MatchString(name)
}

func isDangerousCapability(cap string) bool {
	for _, c := range dangerousCapabilitiesKeptStrings {
		if cap == c {
			return true
		}
	}
	return false
}

func pondFindingBuilder(pod v1.Pod, ruleID string, severity audit.Severity, description string) audit.Finding {
	return audit.Finding{
		Namespace:   pod.Namespace,
		Name:        pod.Name,
		RuleID:      ruleID,
		Severity:    severity,
		Description: description,
	}
}

var PrivilegedContainer = audit.Rule{
	RuleID:   "PRIVILEGED_CONTAINER",
	Severity: audit.SeverityHigh,
	Resource: "Pod",
	Check: func(resource interface{}) []audit.Finding {
		pod := resource.(v1.Pod)
		var findings []audit.Finding
		for _, c := range pod.Spec.Containers {
			if c.SecurityContext == nil {
				continue
			}
			if c.SecurityContext.Privileged != nil && *c.SecurityContext.Privileged {
				findings = append(findings, pondFindingBuilder(pod, "PRIVILEGED_CONTAINER", audit.SeverityHigh, fmt.Sprintf("container %q runs in privileged mode", c.Name)))
			}
		}
		return findings
	},
}

var RunningAsRoot = audit.Rule{
	RuleID:   "RUNNING_AS_ROOT",
	Severity: audit.SeverityHigh,
	Resource: "Pod",
	Check: func(resource interface{}) []audit.Finding {
		pod := resource.(v1.Pod)
		var findings []audit.Finding
		for _, c := range pod.Spec.Containers {
			if c.SecurityContext == nil {
				continue
			}
			if c.SecurityContext.RunAsNonRoot == nil || !*c.SecurityContext.RunAsNonRoot ||
				c.SecurityContext.RunAsUser == nil || *c.SecurityContext.RunAsUser == 0 {
				findings = append(findings, pondFindingBuilder(pod, "RUNNING_AS_ROOT", audit.SeverityHigh, fmt.Sprintf("container %q is running as root user", c.Name)))
			}
		}
		return findings
	},
}

var PrivilegeEscalationAllowed = audit.Rule{
	RuleID:   "PRIVILEGE_ESCALATION_ALLOWED",
	Severity: audit.SeverityHigh,
	Resource: "Pod",
	Check: func(resource interface{}) []audit.Finding {
		pod := resource.(v1.Pod)
		var findings []audit.Finding
		for _, c := range pod.Spec.Containers {
			if c.SecurityContext == nil {
				continue
			}
			if c.SecurityContext.AllowPrivilegeEscalation == nil || *c.SecurityContext.AllowPrivilegeEscalation {
				findings = append(findings, pondFindingBuilder(pod, "PRIVILEGE_ESCALATION_ALLOWED", audit.SeverityHigh, fmt.Sprintf("container %q allows privilege escalation", c.Name)))
			}
		}
		return findings
	},
}

var WritableRootFilesystem = audit.Rule{
	RuleID:   "WRITABLE_ROOT_FILESYSTEM",
	Severity: audit.SeverityHigh,
	Resource: "Pod",
	Check: func(resource interface{}) []audit.Finding {
		pod := resource.(v1.Pod)
		var findings []audit.Finding
		for _, c := range pod.Spec.Containers {
			if c.SecurityContext == nil {
				continue
			}
			if c.SecurityContext.ReadOnlyRootFilesystem == nil || !*c.SecurityContext.ReadOnlyRootFilesystem {
				findings = append(findings, pondFindingBuilder(pod, "WRITABLE_ROOT_FILESYSTEM", audit.SeverityHigh, fmt.Sprintf("container %q allows writing to the root filesystem", c.Name)))
			}
		}
		return findings
	},
}

var DangerousCapabilitiesKept = audit.Rule{
	RuleID:   "DANGEROUS_CAPABILITIES_KEPT",
	Severity: audit.SeverityHigh,
	Resource: "Pod",
	Check: func(resource interface{}) []audit.Finding {
		pod := resource.(v1.Pod)
		var findings []audit.Finding
		for _, c := range pod.Spec.Containers {
			if c.SecurityContext == nil || c.SecurityContext.Capabilities == nil {
				continue
			}
			for _, cap := range c.SecurityContext.Capabilities.Add {
				if isDangerousCapability(string(cap)) {
					findings = append(findings, pondFindingBuilder(pod, "DANGEROUS_CAPABILITIES_KEPT", audit.SeverityHigh, fmt.Sprintf("container %q keeps NET_ADMIN or SYS_ADMIN capabilities", c.Name)))
				}
			}
		}
		return findings
	},
}

var NoResourceLimits = audit.Rule{
	RuleID:   "NO_RESOURCE_LIMITS",
	Severity: audit.SeverityMedium,
	Resource: "Pod",
	Check: func(resource interface{}) []audit.Finding {
		pod := resource.(v1.Pod)
		var findings []audit.Finding
		for _, c := range pod.Spec.Containers {
			if len(c.Resources.Limits) == 0 {
				findings = append(findings, pondFindingBuilder(pod, "NO_RESOURCE_LIMITS", audit.SeverityMedium, fmt.Sprintf("container %q has no resource limits set", c.Name)))
			}
		}
		return findings
	},
}

var NoResourceRequests = audit.Rule{
	RuleID:   "NO_RESOURCE_REQUESTS",
	Severity: audit.SeverityMedium,
	Resource: "Pod",
	Check: func(resource interface{}) []audit.Finding {
		pod := resource.(v1.Pod)
		var findings []audit.Finding
		for _, c := range pod.Spec.Containers {
			if len(c.Resources.Requests) == 0 {
				findings = append(findings, pondFindingBuilder(pod, "NO_RESOURCE_REQUESTS", audit.SeverityMedium, fmt.Sprintf("container %q has no resource requests set", c.Name)))
			}
		}
		return findings
	},
}

var LatestImageTag = audit.Rule{
	RuleID:   "LATEST_IMAGE_TAG",
	Severity: audit.SeverityMedium,
	Resource: "Pod",
	Check: func(resource interface{}) []audit.Finding {
		pod := resource.(v1.Pod)
		var findings []audit.Finding
		for _, c := range pod.Spec.Containers {
			if !strings.Contains(c.Image, ":") || strings.HasSuffix(c.Image, ":latest") {
				findings = append(findings, pondFindingBuilder(pod, "LATEST_IMAGE_TAG", audit.SeverityMedium, fmt.Sprintf("container %q uses the latest image tag", c.Name)))
			}
		}
		return findings
	},
}

var NoLivenessProbe = audit.Rule{
	RuleID:   "NO_LIVENESS_PROBE",
	Severity: audit.SeverityLow,
	Resource: "Pod",
	Check: func(resource interface{}) []audit.Finding {
		pod := resource.(v1.Pod)
		var findings []audit.Finding
		for _, c := range pod.Spec.Containers {
			if c.LivenessProbe == nil {
				findings = append(findings, pondFindingBuilder(pod, "NO_LIVENESS_PROBE", audit.SeverityLow, fmt.Sprintf("container %q has no liveness probe", c.Name)))
			}
		}
		return findings
	},
}

var NoReadinessProbe = audit.Rule{
	RuleID:   "NO_READINESS_PROBE",
	Severity: audit.SeverityLow,
	Resource: "Pod",
	Check: func(resource interface{}) []audit.Finding {
		pod := resource.(v1.Pod)
		var findings []audit.Finding
		for _, c := range pod.Spec.Containers {
			if c.ReadinessProbe == nil {
				findings = append(findings, pondFindingBuilder(pod, "NO_READINESS_PROBE", audit.SeverityLow, fmt.Sprintf("container %q has no readiness probe", c.Name)))
			}
		}
		return findings
	},
}

var HostNamespacesShared = audit.Rule{
	RuleID:   "HOST_NAMESPACES_SHARED",
	Severity: audit.SeverityHigh,
	Resource: "Pod",
	Check: func(resource interface{}) []audit.Finding {
		pod := resource.(v1.Pod)
		var findings []audit.Finding
		if pod.Spec.HostNetwork || pod.Spec.HostPID || pod.Spec.HostIPC {
			findings = append(findings, pondFindingBuilder(pod, "HOST_NAMESPACES_SHARED", audit.SeverityHigh, fmt.Sprintf("pod %q shares host namespaces", pod.Name)))
		}
		return findings
	},
}

var SensitiveHostPathMounted = audit.Rule{
	RuleID:   "SENSITIVE_HOST_PATH_MOUNTED",
	Severity: audit.SeverityHigh,
	Resource: "Pod",
	Check: func(resource interface{}) []audit.Finding {
		pod := resource.(v1.Pod)
		var findings []audit.Finding
		for _, v := range pod.Spec.Volumes {
			if v.HostPath != nil {
				if v.HostPath.Path == "/var/run/docker.sock" || v.HostPath.Path == "/etc" || v.HostPath.Path == "/" || v.HostPath.Path == "/root" || v.HostPath.Path == "/proc" {
					findings = append(findings, pondFindingBuilder(pod, "SENSITIVE_HOST_PATH_MOUNTED", audit.SeverityHigh, fmt.Sprintf("container %q mounts sensitive host path %q", pod.Name, v.HostPath.Path)))
				}
			}
		}
		return findings
	},
}

var DefaultSAAutoMounted = audit.Rule{
	RuleID:   "DEFAULT_SA_AUTO_MOUNTED",
	Severity: audit.SeverityMedium,
	Resource: "Pod",
	Check: func(resource interface{}) []audit.Finding {
		pod := resource.(v1.Pod)
		var findings []audit.Finding
		tokenMounted := pod.Spec.AutomountServiceAccountToken == nil || *pod.Spec.AutomountServiceAccountToken
		usesDefaultSA := pod.Spec.ServiceAccountName == "" || pod.Spec.ServiceAccountName == "default"
		if tokenMounted && usesDefaultSA {
			for _, c := range pod.Spec.Containers {
				findings = append(findings, pondFindingBuilder(pod, "DEFAULT_SA_AUTO_MOUNTED", audit.SeverityMedium, fmt.Sprintf("container %q uses the default service account with automount enabled", c.Name)))
			}
		}
		return findings
	},
}

var NoAppArmorProfile = audit.Rule{
	RuleID:   "NO_APP_ARMOR_PROFILE",
	Severity: audit.SeverityMedium,
	Resource: "Pod",
	Check: func(resource interface{}) []audit.Finding {
		pod := resource.(v1.Pod)
		var findings []audit.Finding
		for _, c := range pod.Spec.Containers {
			if c.SecurityContext == nil {
				continue
			}
			key := "container.apparmor.security.beta.kubernetes.io/" + c.Name
			if pod.Annotations == nil || pod.Annotations[key] == "" {
				findings = append(findings, pondFindingBuilder(pod, "NO_APP_ARMOR_PROFILE", audit.SeverityMedium, fmt.Sprintf("container %q has no AppArmor profile set", c.Name)))
			}
		}
		return findings
	},
}

var NoSeccompProfile = audit.Rule{
	RuleID:   "NO_SECCOMP_PROFILE",
	Severity: audit.SeverityMedium,
	Resource: "Pod",
	Check: func(resource interface{}) []audit.Finding {
		pod := resource.(v1.Pod)
		var findings []audit.Finding
		if pod.Spec.SecurityContext == nil || pod.Spec.SecurityContext.SeccompProfile == nil || pod.Spec.SecurityContext.SeccompProfile.Type != v1.SeccompProfileTypeRuntimeDefault {
			for _, c := range pod.Spec.Containers {
				findings = append(findings, pondFindingBuilder(pod, "NO_SECCOMP_PROFILE", audit.SeverityMedium, fmt.Sprintf("container %q has no Seccomp profile set", c.Name)))
			}
		}
		return findings
	},
}

var EnvVarsWithPlaintextSecrets = audit.Rule{
	RuleID:   "ENV_VARS_WITH_PLAINTEXT_SECRETS",
	Severity: audit.SeverityHigh,
	Resource: "Pod",
	Check: func(resource interface{}) []audit.Finding {
		pod := resource.(v1.Pod)
		var findings []audit.Finding
		for _, c := range pod.Spec.Containers {
			if c.SecurityContext == nil {
				continue
			}
			for _, env := range c.Env {
				if (env.ValueFrom == nil || env.ValueFrom.SecretKeyRef == nil) && IsSensitiveEnvName(env.Name) {
					findings = append(findings, pondFindingBuilder(pod, "ENV_VARS_WITH_PLAINTEXT_SECRETS", audit.SeverityHigh, fmt.Sprintf("container %q has an environment variable with a plaintext secret", c.Name)))
				}
			}
		}
		return findings
	},
}

var PodRules = []audit.Rule{
	PrivilegedContainer,
	RunningAsRoot,
	SensitiveHostPathMounted,
	DefaultSAAutoMounted,
	NoAppArmorProfile,
	NoSeccompProfile,
	EnvVarsWithPlaintextSecrets,
	PrivilegeEscalationAllowed,
	WritableRootFilesystem,
	NoResourceLimits,
	NoResourceRequests,
	LatestImageTag,
	NoLivenessProbe,
	NoReadinessProbe,
	DangerousCapabilitiesKept,
	HostNamespacesShared,
}

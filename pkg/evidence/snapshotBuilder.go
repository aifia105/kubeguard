package evidence

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aifia105/kubeguard/pkg/audit"
	"github.com/aifia105/kubeguard/pkg/collectors"
	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

type PodResourceData struct {
	Name       string
	Namespace  string
	CPURequest resource.Quantity
	MemRequest resource.Quantity
	CPULimit   resource.Quantity
	MemLimit   resource.Quantity
	CPUUsage   resource.Quantity
	MemUsage   resource.Quantity
}

func BuildSnapshot(ctx context.Context, clientset *kubernetes.Clientset, mclientset *metricsclientset.Clientset, runID uuid.UUID, namespace, clusterName string, findings []audit.Finding, pods []corev1.Pod, events []corev1.Event) *EvidenceSnapshot {
	snapshot := &EvidenceSnapshot{
		RunID:       runID,
		GeneratedAt: time.Now(),
		ClusterName: clusterName,
		Scope:       scopeLabel(namespace),
		Findings:    findings,
		Summary:     summarizeFindings(findings),
	}

	podsMetrics, _ := collectors.ListPodsMetrics(ctx, mclientset, namespace)
	podsEvidence := buildPodsMetricsEvidence(pods, podsMetrics)
	snapshot.Pods = BuildPodsContext(pods, podsEvidence, findings)
	snapshot.Events = buildEventsContext(events)
	snapshot.Logs = collectLogs(ctx, clientset, pods, 100)

	return snapshot
}

func scopeLabel(namespace string) string {
	if namespace == "" {
		return "cluster-wide"
	}
	return namespace
}

func summarizeFindings(findings []audit.Finding) FindingsSummary {
	var summary FindingsSummary
	summary.Total = len(findings)
	for _, f := range findings {
		switch f.Severity {
		case audit.SeverityCritical:
			summary.Critical++
		case audit.SeverityHigh:
			summary.High++
		case audit.SeverityMedium:
			summary.Medium++
		case audit.SeverityLow:
			summary.Low++
		}
	}
	return summary
}

func buildPodsMetricsEvidence(pods []corev1.Pod, podsMetrics []metricsv1beta1.PodMetrics) []PodResourceData {
	metricsByName := make(map[string]metricsv1beta1.PodMetrics, len(podsMetrics))
	for _, pm := range podsMetrics {
		metricsByName[pm.Namespace+"/"+pm.Name] = pm
	}

	var result []PodResourceData

	for _, pod := range pods {
		data := PodResourceData{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		}

		for _, c := range pod.Spec.Containers {
			if cpu, ok := c.Resources.Requests[corev1.ResourceCPU]; ok {
				data.CPURequest.Add(cpu)
			}
			if mem, ok := c.Resources.Requests[corev1.ResourceMemory]; ok {
				data.MemRequest.Add(mem)
			}
			if cpu, ok := c.Resources.Limits[corev1.ResourceCPU]; ok {
				data.CPULimit.Add(cpu)
			}
			if mem, ok := c.Resources.Limits[corev1.ResourceMemory]; ok {
				data.MemLimit.Add(mem)
			}
		}

		if pm, ok := metricsByName[pod.Namespace+"/"+pod.Name]; ok {
			for _, c := range pm.Containers {
				if cpu, ok := c.Usage[corev1.ResourceCPU]; ok {
					data.CPUUsage.Add(cpu)
				}
				if mem, ok := c.Usage[corev1.ResourceMemory]; ok {
					data.MemUsage.Add(mem)
				}
			}
		}

		result = append(result, data)
	}
	return result
}

func BuildPodContext(pod corev1.Pod, resData *PodResourceData, related []string) PodContext {
	pc := PodContext{
		Name:            pod.Name,
		Namespace:       pod.Namespace,
		Phase:           string(pod.Status.Phase),
		Node:            pod.Spec.NodeName,
		RelatedFindings: related,
	}

	var readyCount, totalCount int32
	var restartCount int32
	var waitingReason, lastTermReason string

	totalCount = int32(len(pod.Status.ContainerStatuses))
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Ready {
			readyCount++
		}
		restartCount += cs.RestartCount
		if cs.State.Waiting != nil && waitingReason == "" {
			waitingReason = cs.State.Waiting.Reason
		}
		if cs.State.Terminated != nil && lastTermReason == "" {
			lastTermReason = cs.State.Terminated.Reason
		}
	}

	pc.Ready = fmt.Sprintf("%d/%d", readyCount, totalCount)
	pc.RestartCount = restartCount
	pc.WaitingReason = waitingReason
	pc.LastTermReason = lastTermReason

	if len(pod.OwnerReferences) > 0 {
		owner := pod.OwnerReferences[0]
		pc.Owner = fmt.Sprintf("%s/%s", owner.Kind, owner.Name)
	}

	if resData != nil {
		pc.CPURequest = resData.CPURequest.String()
		pc.MemRequest = resData.MemRequest.String()
		pc.CPUUsage = resData.CPUUsage.String()
		pc.MemUsage = resData.MemUsage.String()
		pc.CPULimit = resData.CPULimit.String()
		pc.MemLimit = resData.MemLimit.String()
	}

	return pc
}

func BuildPodsContext(pods []corev1.Pod, resDataList []PodResourceData, findings []audit.Finding) []PodContext {
	resByName := make(map[string]PodResourceData, len(resDataList))
	for _, res := range resDataList {
		resByName[res.Namespace+"/"+res.Name] = res
	}
	result := make([]PodContext, 0, len(pods))
	for _, pod := range pods {
		related := findingsFor(findings, pod.Namespace, pod.Name)
		if pod.Status.Phase == corev1.PodRunning && isFullyReady(pod) && len(related) == 0 {
			continue
		}
		var resData *PodResourceData
		if res, ok := resByName[pod.Namespace+"/"+pod.Name]; ok {
			resData = &res
		}
		pc := BuildPodContext(pod, resData, related)
		result = append(result, pc)
	}
	return result
}

func findingsFor(findings []audit.Finding, namespace, name string) []string {
	var ids []string
	for _, f := range findings {
		if f.Namespace == namespace && f.Name == name {
			ids = append(ids, f.RuleID)
		}
	}
	return ids
}

func isFullyReady(pod corev1.Pod) bool {
	for _, cs := range pod.Status.ContainerStatuses {
		if !cs.Ready {
			return false
		}
	}
	return len(pod.Status.ContainerStatuses) > 0
}

func buildEventsContext(events []corev1.Event) []Event {
	var result []Event
	for _, e := range events {
		if e.Type != corev1.EventTypeWarning {
			continue
		}
		result = append(result, Event{
			Namespace: e.Namespace,
			Object:    e.InvolvedObject.Kind + "/" + e.InvolvedObject.Name,
			Type:      e.Type,
			Reason:    e.Reason,
			Message:   e.Message,
			Count:     e.Count,
			LastSeen:  e.LastTimestamp.Time,
		})
	}
	return result
}

func collectLogs(ctx context.Context, clientset *kubernetes.Clientset, pods []corev1.Pod, tailLines int64) []LogExcerpt {
	var out []LogExcerpt
	for _, pod := range pods {
		if !worthFetchingLogs(pod) {
			continue
		}
		for _, c := range pod.Spec.Containers {
			logText, err := collectors.FetchPodLogs(ctx, clientset, pod.Namespace, pod.Name, c.Name, tailLines)
			if err != nil || logText == "" {
				continue
			}
			out = append(out, LogExcerpt{
				Namespace: pod.Namespace,
				Pod:       pod.Name,
				Container: c.Name,
				TailLines: int(tailLines),
				Lines:     strings.Split(logText, "\n"),
			})
		}
	}
	return out
}

func worthFetchingLogs(pod corev1.Pod) bool {
	if totalRestarts(pod) > 0 {
		return true
	}
	waiting, lastTerm := containerStateReasons(pod)
	return waiting != "" || lastTerm != ""
}

func totalRestarts(pod corev1.Pod) int32 {
	var total int32
	for _, cs := range pod.Status.ContainerStatuses {
		total += cs.RestartCount
	}
	return total
}

func containerStateReasons(pod corev1.Pod) (waiting string, lastTerm string) {
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Waiting != nil && waiting == "" {
			waiting = cs.State.Waiting.Reason
		}
		if cs.LastTerminationState.Terminated != nil && lastTerm == "" {
			lastTerm = cs.LastTerminationState.Terminated.Reason
		}
	}
	return
}

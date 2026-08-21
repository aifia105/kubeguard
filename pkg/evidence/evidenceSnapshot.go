package evidence

import (
	"time"

	"github.com/aifia105/kubeguard/pkg/audit"
)

type EvidenceSnapshot struct {
	GeneratedAt time.Time `json:"generatedAt"`
	ClusterName string    `json:"clusterName,omitempty"`
	Scope       string    `json:"scope,omitempty"`

	Summary  FindingsSummary `json:"summary"`
	Findings []audit.Finding `json:"findings,omitempty"`

	Pods      []PodContext   `json:"pods,omitempty"`
	Events    []Event        `json:"events,omitempty"`
	Logs      []LogExcerpt   `json:"logs,omitempty"`
	Resources []ResourceSpec `json:"resources,omitempty"`
}

type FindingsSummary struct {
	Total    int `json:"total"`
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
}

type ResourceSpec struct {
	Kind            string      `json:"kind"`
	Name            string      `json:"name"`
	Namespace       string      `json:"namespace"`
	Excerpt         interface{} `json:"excerpt"`
	RelatedFindings []string    `json:"relatedFindings,omitempty"`
}
type PodContext struct {
	Name            string   `json:"name"`
	Namespace       string   `json:"namespace"`
	Phase           string   `json:"phase"`
	Ready           string   `json:"ready"`
	RestartCount    int32    `json:"restartCount"`
	WaitingReason   string   `json:"waitingReason,omitempty"`
	LastTermReason  string   `json:"lastTermReason,omitempty"`
	Owner           string   `json:"owner,omitempty"`
	Node            string   `json:"node,omitempty"`
	CPURequest      string   `json:"cpuRequest,omitempty"`
	MemRequest      string   `json:"memRequest,omitempty"`
	CPUUsage        string   `json:"cpuUsage,omitempty"`
	MemUsage        string   `json:"memUsage,omitempty"`
	CPULimit        string   `json:"cpuLimit,omitempty"`
	MemLimit        string   `json:"memLimit,omitempty"`
	RelatedFindings []string `json:"relatedFindings,omitempty"`
}

type Event struct {
	Namespace string    `json:"namespace"`
	Object    string    `json:"object"`
	Type      string    `json:"type"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Count     int32     `json:"count"`
	LastSeen  time.Time `json:"lastSeen"`
}

type LogExcerpt struct {
	Namespace string   `json:"namespace"`
	Pod       string   `json:"pod"`
	Container string   `json:"container,omitempty"`
	TailLines int      `json:"tailLines"`
	Lines     []string `json:"lines"`
}

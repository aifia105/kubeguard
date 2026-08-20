package audit

type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityHigh     Severity = "HIGH"
	SeverityMedium   Severity = "MEDIUM"
	SeverityLow      Severity = "LOW"
)

type Finding struct {
	Namespace   string
	Resource    string
	RuleID      string
	Name        string
	Description string
	Severity    Severity
}

type Rule struct {
	RuleID   string
	Resource string
	Severity Severity
	Check    func(resource interface{}) []Finding
}

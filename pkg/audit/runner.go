package audit

func RunAudit[T any](resources []T, rules []Rule) []Finding {
	var findings []Finding
	for _, resource := range resources {
		for _, rule := range rules {
			findings = append(findings, rule.Check(resource)...)
		}
	}
	return findings
}

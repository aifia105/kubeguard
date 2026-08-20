package audit

func RunAudit[T any](resources []T, rules []Rule) []Finding {
	var findings []Finding
	for _, resource := range resources {
		for _, rule := range rules {
			if rule.Check == nil {
				continue
			}
			findings = append(findings, rule.Check(resource)...)
		}
	}
	return findings
}

func RunMultiResourceAudit[T any](resources []T, rules []Rule) []Finding {
	var findings []Finding
	for _, rule := range rules {
		if rule.MultiCheck != nil {
			findings = append(findings, rule.MultiCheck(toInterfaceSlice(resources))...)
		}
	}
	return findings
}

func toInterfaceSlice[T any](in []T) []interface{} {
	out := make([]interface{}, len(in))
	for i, v := range in {
		out[i] = v
	}
	return out
}

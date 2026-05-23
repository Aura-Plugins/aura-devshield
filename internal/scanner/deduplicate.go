package scanner

func DeduplicateFindings(findings []Finding) []Finding {
	seen := make(map[string]bool)
	var deduplicated []Finding

	for _, finding := range findings {
		if seen[finding.Fingerprint] {
			continue
		}
		seen[finding.Fingerprint] = true
		deduplicated = append(deduplicated, finding)
	}

	return deduplicated
}
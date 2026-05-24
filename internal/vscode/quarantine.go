package vscode

import (
	"fmt"
	"time"

	"github.com/Aura-Plugins/aura-devshield/internal/scanner"
	"github.com/Aura-Plugins/aura-devshield/internal/state"
)

// UpdateState records first-seen timestamps for each scanned extension version.
// Subsequent calls for the same version are no-ops — the clock starts on first sight.
func UpdateState(extensions []*Extension, s *state.State) {
	now := time.Now().UTC()
	for _, ext := range extensions {
		s.RecordVSCodeExtension(ext.CanonicalID(), ext.Version, now)
	}
}

// FindQuarantineFindings returns a finding for each extension whose version was
// first seen within the quarantine window.
func FindQuarantineFindings(extensions []*Extension, s *state.State, delay time.Duration) []scanner.Finding {
	var findings []scanner.Finding

	for _, ext := range extensions {
		firstSeen, ok := s.FirstSeen(ext.CanonicalID(), ext.Version)
		if !ok {
			continue
		}

		age := time.Since(firstSeen)
		if age >= delay {
			continue
		}

		remaining := delay - age

		findings = append(findings, scanner.Finding{
			ID: "vscode.update_in_quarantine",
			Fingerprint: scanner.GenerateFingerprint(
				"vscode.update_in_quarantine",
				ext.CanonicalID(),
				ext.Version,
			),
			Severity: scanner.SeverityMedium,
			Title:    "VS Code extension update in quarantine window",
			Target:   ext.CanonicalID() + "@" + ext.Version,
			Path:     ext.Path,
			Description: fmt.Sprintf(
				"Extension %s version %s was first seen %s ago. Quarantine policy requires %s before auto-update is trusted. %s remaining.",
				ext.CanonicalID(), ext.Version, formatDuration(age), formatDuration(delay), formatDuration(remaining),
			),
			Metadata: map[string]string{
				"first_seen":     firstSeen.Format(time.RFC3339),
				"age":            formatDuration(age),
				"time_remaining": formatDuration(remaining),
				"policy":         formatDuration(delay),
			},
		})
	}

	return findings
}

// QuarantinedExtensionIDs returns the canonical IDs of extensions currently inside the quarantine window.
func QuarantinedExtensionIDs(extensions []*Extension, s *state.State, delay time.Duration) []string {
	var ids []string
	for _, ext := range extensions {
		firstSeen, ok := s.FirstSeen(ext.CanonicalID(), ext.Version)
		if !ok {
			continue
		}
		if time.Since(firstSeen) < delay {
			ids = append(ids, ext.CanonicalID())
		}
	}
	return ids
}

// formatDuration renders a duration as "X days" when it is a whole number of days,
// otherwise as "Xh". Keeps output readable regardless of whether delay came from
// the config (days) or the --delay flag (hours).
func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	if h == 0 {
		return d.String()
	}
	if h%24 == 0 {
		days := h / 24
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	}
	return fmt.Sprintf("%dh", h)
}

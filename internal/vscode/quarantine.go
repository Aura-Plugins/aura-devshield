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
func FindQuarantineFindings(extensions []*Extension, s *state.State, quarantineDays int) []scanner.Finding {
	var findings []scanner.Finding
	threshold := time.Duration(quarantineDays) * 24 * time.Hour

	for _, ext := range extensions {
		firstSeen, ok := s.FirstSeen(ext.CanonicalID(), ext.Version)
		if !ok {
			continue
		}

		age := time.Since(firstSeen)
		if age >= threshold {
			continue
		}

		daysOld := int(age.Hours() / 24)
		daysRemaining := quarantineDays - daysOld

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
				"Extension %s version %s was first seen %d day(s) ago. Quarantine policy requires %d days before auto-update is trusted. %d day(s) remaining.",
				ext.CanonicalID(), ext.Version, daysOld, quarantineDays, daysRemaining,
			),
			Metadata: map[string]string{
				"first_seen":        firstSeen.Format(time.RFC3339),
				"days_old":          fmt.Sprintf("%d", daysOld),
				"days_remaining":    fmt.Sprintf("%d", daysRemaining),
				"quarantine_policy": fmt.Sprintf("%d", quarantineDays),
			},
		})
	}

	return findings
}

// QuarantinedExtensionIDs returns the canonical IDs of extensions currently inside the quarantine window.
func QuarantinedExtensionIDs(extensions []*Extension, s *state.State, quarantineDays int) []string {
	threshold := time.Duration(quarantineDays) * 24 * time.Hour
	var ids []string
	for _, ext := range extensions {
		firstSeen, ok := s.FirstSeen(ext.CanonicalID(), ext.Version)
		if !ok {
			continue
		}
		if time.Since(firstSeen) < threshold {
			ids = append(ids, ext.CanonicalID())
		}
	}
	return ids
}

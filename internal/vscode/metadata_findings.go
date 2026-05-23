package vscode

import (
	"fmt"
	"strings"

	"github.com/matias2018/aura-devshield/internal/scanner"
)

func FindInvalidMetadataFindings(
	extensions []*Extension,
) []scanner.Finding {

	var findings []scanner.Finding

	for _, extension := range extensions {

		displayName := strings.TrimSpace(extension.DisplayName)

		if displayName == "" {
			findings = append(findings, scanner.Finding{
				ID: "vscode.invalid_metadata",

				Fingerprint: scanner.GenerateFingerprint(
					"vscode.invalid_metadata",
					extension.CanonicalID(),
					"empty_display_name",
				),

				Severity: scanner.SeverityLow,

				Title: "Extension has empty display name",

				Target: extension.CanonicalID(),

				Description: fmt.Sprintf(
					"Extension %s has an empty display name.",
					extension.CanonicalID(),
				),
			})

			continue
		}

		if strings.Contains(displayName, "%") {
			findings = append(findings, scanner.Finding{
				ID: "vscode.invalid_metadata",

				Fingerprint: scanner.GenerateFingerprint(
					"vscode.invalid_metadata",
					extension.CanonicalID(),
					displayName,
				),

				Severity: scanner.SeverityLow,

				Title: "Extension has unresolved metadata placeholders",

				Target: extension.CanonicalID(),

				Description: fmt.Sprintf(
					"Extension %s contains unresolved placeholder metadata: %s.",
					extension.CanonicalID(),
					displayName,
				),
			})
		}
	}

	return findings
}

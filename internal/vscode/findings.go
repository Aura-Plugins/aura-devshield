package vscode

import (
	"fmt"
	"strings"

	"github.com/matias2018/aura-devshield/internal/scanner"
)

func FindMultiVersionFindings(extensions []*Extension) []scanner.Finding {
	multiVersionExtensions := FindMultiVersionExtensions(extensions)

	var findings []scanner.Finding

	for id, versions := range multiVersionExtensions {
		var installedVersions []string

		for _, extension := range versions {
			installedVersions = append(installedVersions, extension.Version)
		}

		finding := scanner.Finding{
			ID: "vscode.multiple_installed_versions",

			Fingerprint: scanner.GenerateFingerprint(
				"vscode.multiple_installed_versions",
				id,
				strings.Join(installedVersions, ","),
			),

			Severity: scanner.SeverityMedium,

			Title: "Multiple installed versions of VS Code extension",

			Target: id,

			Description: fmt.Sprintf(
				"Extension %s has multiple versions installed: %s.",
				id,
				strings.Join(installedVersions, ", "),
			),
		}

		findings = append(findings, finding)
	}

	return findings
}

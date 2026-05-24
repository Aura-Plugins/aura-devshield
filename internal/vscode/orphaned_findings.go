package vscode

import "github.com/matias2018/aura-devshield/internal/scanner"

func FindOrphanedDirectoryFindings(extensionsDir string) ([]scanner.Finding, error) {
	orphanedDirs, err := FindOrphanedExtensionDirs(extensionsDir)
	if err != nil {
		return nil, err
	}

	findings := make([]scanner.Finding, 0)

	for _, dir := range orphanedDirs {
		findings = append(findings, scanner.Finding{
			ID:          "vscode.orphaned_extension_directory",
			Title:       "Orphaned VS Code extension directory",
			Severity:    scanner.SeverityLow,
			Description: "A directory exists inside the VS Code extensions folder but does not contain a valid package.json file.",
			Target:      dir,
			Path:        dir,
			Fingerprint: scanner.GenerateFingerprint(
				"vscode.orphaned_extension_directory",
				dir,
			),
		})
	}

	return findings, nil
}
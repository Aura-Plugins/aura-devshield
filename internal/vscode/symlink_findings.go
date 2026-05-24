package vscode

import "github.com/Aura-Plugins/aura-devshield/internal/scanner"

func FindSymlinkedExtensionDirectoryFindings(extensionsDir string) ([]scanner.Finding, error) {
	symlinkedDirs, err := FindSymlinkedExtensionDirs(extensionsDir)
	if err != nil {
		return nil, err
	}

	findings := make([]scanner.Finding, 0)

	for _, dir := range symlinkedDirs {
		findings = append(findings, scanner.Finding{
			ID:          "vscode.symlinked_extension_directory",
			Fingerprint: scanner.GenerateFingerprint("vscode.symlinked_extension_directory", dir),
			Severity:    scanner.SeverityLow,
			Title:       "Symlinked VS Code extension directory",
			Description: "A VS Code extension directory is a symbolic link. This may be legitimate during development, but it can also obscure where extension code is actually stored.",
			Target:      dir,
			Path:        dir,
		})
	}

	return findings, nil
}
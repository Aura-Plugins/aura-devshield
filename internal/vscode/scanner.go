package vscode

import (
	"path/filepath"
)

func ScanExtensions(extensionsDir string) ([]*Extension, error) {
	entries, err := ListExtensions(extensionsDir)
	if err != nil {
		return nil, err
	}

	extensions := make([]*Extension, 0)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		extensionPath := filepath.Join(extensionsDir, entry.Name())

		extension, err := ReadPackageJSON(extensionPath)
		if err != nil {
			// Invalid or missing package.json is handled elsewhere
			// by invalid metadata / orphaned directory checks.
			continue
		}

		if extension == nil {
			continue
		}

		extension.Path = extensionPath

		extensions = append(extensions, extension)
	}

	return extensions, nil
}

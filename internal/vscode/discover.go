package vscode

import (
	"os"
	"path/filepath"
)

func FindOrphanedExtensionDirs(extensionsDir string) ([]string, error) {
	entries, err := os.ReadDir(extensionsDir)
	if err != nil {
		return nil, err
	}

	orphaned := make([]string, 0)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		extensionPath := filepath.Join(extensionsDir, entry.Name())
		packageJSONPath := filepath.Join(extensionPath, "package.json")

		info, err := os.Stat(packageJSONPath)
		if err != nil {
			if os.IsNotExist(err) {
				orphaned = append(orphaned, extensionPath)
				continue
			}

			return nil, err
		}

		if info.IsDir() {
			orphaned = append(orphaned, extensionPath)
		}
	}

	return orphaned, nil
}
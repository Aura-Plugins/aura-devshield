package vscode

import (
	"os"
	"path/filepath"
)

func FindSymlinkedExtensionDirs(extensionsDir string) ([]string, error) {
	entries, err := os.ReadDir(extensionsDir)
	if err != nil {
		return nil, err
	}

	symlinked := make([]string, 0)

	for _, entry := range entries {
		extensionPath := filepath.Join(extensionsDir, entry.Name())

		info, err := os.Lstat(extensionPath)
		if err != nil {
			return nil, err
		}

		if info.Mode()&os.ModeSymlink != 0 {
			symlinked = append(symlinked, extensionPath)
		}
	}

	return symlinked, nil
}
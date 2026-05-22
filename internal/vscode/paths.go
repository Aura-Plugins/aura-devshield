package vscode

import (
	"os"
	"path/filepath"
)

func ExtensionsDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	extensionsDir := filepath.Join(homeDir, ".vscode", "extensions")
	return extensionsDir, nil
}

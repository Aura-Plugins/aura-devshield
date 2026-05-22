package vscode

import (
	"path/filepath"
)

func ScanExtensions(extensionsDir string) ([]*Extension, error) {
	entries, err := ListExtensions(extensionsDir)
	if err != nil {
		return nil, err
	} 

	var extensions []*Extension

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		extensionPath := filepath.Join(
			extensionsDir, 
			entry.Name(),
		)

		extension, err := ReadPackageJSON(extensionPath)
		if err != nil {
			continue
		}
		
		extensions = append(extensions, extension)
	}

	return extensions, nil
}
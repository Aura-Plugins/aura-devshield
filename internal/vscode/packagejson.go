package vscode

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func ReadPackageJSON(extensionPath string) (*Extension, error) {

	packageJSONPath := filepath.Join(extensionPath, "package.json")

	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return nil, err
	}

	var extension Extension

	err = json.Unmarshal(data, &extension)
	if err != nil {
		return nil, err
	}

	extension.Path = extensionPath

	return &extension, nil
}

package vscode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// QuarantineResult describes one change made (or previewed) in VS Code settings.
type QuarantineResult struct {
	ExtensionID string
	Action      string // "pinned" | "released"
}

// VSCodeSettingsPath returns the path to VS Code's user settings.json.
func VSCodeSettingsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Code", "User", "settings.json"), nil
	case "linux":
		configDir := os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			configDir = filepath.Join(home, ".config")
		}
		return filepath.Join(configDir, "Code", "User", "settings.json"), nil
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		return filepath.Join(appData, "Code", "User", "settings.json"), nil
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func readSettings(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, err
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("could not parse VS Code settings.json: %w", err)
	}
	return settings, nil
}

func writeSettings(path string, settings map[string]interface{}) error {
	data, err := json.MarshalIndent(settings, "", "    ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// pinnedBySettings returns the set of extension IDs where auto-update is explicitly
// disabled via the per-extension map form of extensions.autoUpdate.
func pinnedBySettings(settings map[string]interface{}) map[string]bool {
	pinned := make(map[string]bool)

	val, ok := settings["extensions.autoUpdate"]
	if !ok {
		return pinned
	}

	m, ok := val.(map[string]interface{})
	if !ok {
		// Global bool or string setting — we leave it alone and treat as no per-extension pins.
		return pinned
	}

	for id, v := range m {
		if b, ok := v.(bool); ok && !b {
			pinned[id] = true
		}
	}
	return pinned
}

// computeQuarantineResults applies pin/release changes to the autoUpdate map and
// returns what changed plus the full updated map ready to write back.
// All existing entries not in toPin or toRelease are preserved unchanged, so
// manually added per-extension settings survive a round-trip through apply.
func computeQuarantineResults(settings map[string]interface{}, toPin, toRelease []string) ([]QuarantineResult, map[string]interface{}) {
	// Copy the full existing map so unrelated entries (including manually-added
	// true values) are preserved. If the key holds a non-map value (global bool),
	// we start empty — the tool only operates on the per-extension map form.
	autoUpdate := make(map[string]interface{})
	if val, ok := settings["extensions.autoUpdate"]; ok {
		if m, ok := val.(map[string]interface{}); ok {
			for k, v := range m {
				autoUpdate[k] = v
			}
		}
	}

	currentPinned := pinnedBySettings(settings)
	var results []QuarantineResult

	for _, id := range toPin {
		if currentPinned[id] {
			continue
		}
		autoUpdate[id] = false
		results = append(results, QuarantineResult{ExtensionID: id, Action: "pinned"})
	}

	for _, id := range toRelease {
		if !currentPinned[id] {
			continue
		}
		delete(autoUpdate, id)
		results = append(results, QuarantineResult{ExtensionID: id, Action: "released"})
	}

	return results, autoUpdate
}

// PreviewQuarantine returns what ApplyQuarantine would do without writing any files.
func PreviewQuarantine(settingsPath string, toPin, toRelease []string) ([]QuarantineResult, error) {
	settings, err := readSettings(settingsPath)
	if err != nil {
		return nil, err
	}
	results, _ := computeQuarantineResults(settings, toPin, toRelease)
	return results, nil
}

// ApplyQuarantine writes the quarantine pin/release changes to VS Code settings.json.
func ApplyQuarantine(settingsPath string, toPin, toRelease []string) ([]QuarantineResult, error) {
	settings, err := readSettings(settingsPath)
	if err != nil {
		return nil, err
	}

	results, newAutoUpdate := computeQuarantineResults(settings, toPin, toRelease)
	if len(results) == 0 {
		return results, nil
	}

	if len(newAutoUpdate) == 0 {
		delete(settings, "extensions.autoUpdate")
	} else {
		settings["extensions.autoUpdate"] = newAutoUpdate
	}

	if err := writeSettings(settingsPath, settings); err != nil {
		return nil, err
	}

	return results, nil
}

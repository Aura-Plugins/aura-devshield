package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type State struct {
	VSCodeExtensions map[string]map[string]time.Time `json:"vscode_extensions"`
	VSCodePinned     map[string]bool                 `json:"vscode_pinned,omitempty"`
}

func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".aura-devshield", "state.json"), nil
}

func Load(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{
				VSCodeExtensions: make(map[string]map[string]time.Time),
				VSCodePinned:     make(map[string]bool),
			}, nil
		}
		return nil, err
	}

	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	if s.VSCodeExtensions == nil {
		s.VSCodeExtensions = make(map[string]map[string]time.Time)
	}
	if s.VSCodePinned == nil {
		s.VSCodePinned = make(map[string]bool)
	}
	return &s, nil
}

func (s *State) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// RecordVSCodeExtension records the first time a version is seen; subsequent calls are no-ops.
func (s *State) RecordVSCodeExtension(canonicalID, version string, t time.Time) {
	if s.VSCodeExtensions[canonicalID] == nil {
		s.VSCodeExtensions[canonicalID] = make(map[string]time.Time)
	}
	if _, exists := s.VSCodeExtensions[canonicalID][version]; !exists {
		s.VSCodeExtensions[canonicalID][version] = t
	}
}

func (s *State) FirstSeen(canonicalID, version string) (time.Time, bool) {
	versions, ok := s.VSCodeExtensions[canonicalID]
	if !ok {
		return time.Time{}, false
	}
	t, ok := versions[version]
	return t, ok
}

func (s *State) PinVSCodeExtension(canonicalID string) {
	s.VSCodePinned[canonicalID] = true
}

func (s *State) UnpinVSCodeExtension(canonicalID string) {
	delete(s.VSCodePinned, canonicalID)
}

func (s *State) IsVSCodeExtensionPinned(canonicalID string) bool {
	return s.VSCodePinned[canonicalID]
}

func (s *State) PinnedVSCodeExtensions() []string {
	result := make([]string, 0, len(s.VSCodePinned))
	for id := range s.VSCodePinned {
		result = append(result, id)
	}
	return result
}

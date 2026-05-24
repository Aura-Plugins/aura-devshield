package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const DefaultQuarantineDays = 7

type Config struct {
	QuarantineDays int `json:"quarantine_days"`
}

func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".aura-devshield", "config.json"), nil
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{QuarantineDays: DefaultQuarantineDays}, nil
		}
		return nil, err
	}

	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	if c.QuarantineDays <= 0 {
		c.QuarantineDays = DefaultQuarantineDays
	}
	return &c, nil
}

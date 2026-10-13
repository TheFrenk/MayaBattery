package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"maya-battery/internal/icon"
)

type Config struct {
	IconMode  string `json:"icon_mode"`
	FontPath  string `json:"font_path"`
	NotifyLow bool   `json:"notify_low"`
	Threshold int    `json:"threshold"`
}

var thresholds = []int{10, 15, 20, 30}

func Default() Config {
	return Config{IconMode: icon.ModeNumeric, NotifyLow: true, Threshold: 20}
}

func Dir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, "MayaBattery")
}

func path() string { return filepath.Join(Dir(), "config.json") }

func Load() Config {
	cfg := Default()
	if data, err := os.ReadFile(path()); err == nil {
		_ = json.Unmarshal(data, &cfg)
	}
	if cfg.IconMode != icon.ModeNumeric && cfg.IconMode != icon.ModeBattery {
		cfg.IconMode = icon.ModeNumeric
	}
	if cfg.Threshold < 1 || cfg.Threshold > 99 {
		cfg.Threshold = 20
	}
	return cfg
}

func Save(cfg Config) error {
	if err := os.MkdirAll(Dir(), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path(), data, 0o644)
}

func NextThreshold(cur int) int {
	for _, t := range thresholds {
		if t > cur {
			return t
		}
	}
	return thresholds[0]
}

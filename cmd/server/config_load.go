package main

import (
	"os"

	"speaknow/internal/assets"
	"speaknow/internal/config"
)

func loadConfig(path string) (*config.Config, error) {
	if path != "" {
		if _, err := os.Stat(path); err == nil {
			return config.Load(path)
		}
	}
	if b := assets.DefaultConfig(); len(b) > 0 {
		return config.LoadBytes(b)
	}
	return config.LoadBytes(config.BuiltinDefault())
}

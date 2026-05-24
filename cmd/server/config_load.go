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
	if len(assets.DefaultConfig()) > 0 {
		return config.LoadBytes(assets.DefaultConfig())
	}
	return config.Load("configs/config.yaml")
}

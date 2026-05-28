package main

import (
	"fmt"
	"os"
	"path/filepath"

	"speaknow/internal/assets"
	"speaknow/internal/config"
)

const defaultConfigRel = "configs/config.yaml"

// defaultConfigCandidates 按优先级尝试的默认配置文件路径。
func defaultConfigCandidates() []string {
	seen := make(map[string]struct{})
	var out []string
	add := func(p string) {
		if p == "" {
			return
		}
		if abs, err := filepath.Abs(p); err == nil {
			p = abs
		}
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}

	add(defaultConfigRel)
	add("config.yaml")
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		add(filepath.Join(dir, defaultConfigRel))
		add(filepath.Join(dir, "config.yaml"))
	}
	return out
}

func loadConfig(path string) (*config.Config, error) {
	if path != "" {
		if _, err := os.Stat(path); err != nil {
			return nil, fmt.Errorf("config file %q: %w", path, err)
		}
		return config.Load(path)
	}

	for _, candidate := range defaultConfigCandidates() {
		if _, err := os.Stat(candidate); err == nil {
			return config.Load(candidate)
		}
	}

	if b := assets.DefaultConfig(); len(b) > 0 {
		return config.LoadBytes(b)
	}
	return config.LoadBytes(config.BuiltinDefault())
}

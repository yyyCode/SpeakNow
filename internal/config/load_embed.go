package config

import (
	"bytes"
	"fmt"

	"github.com/spf13/viper"
)

func LoadBytes(data []byte) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("read embedded config: %w", err)
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	cfg.applyDefaults()
	return &cfg, nil
}

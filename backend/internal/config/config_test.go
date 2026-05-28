package config

import (
	"testing"
	"time"
)

func TestApplyDefaults_ServerTimeouts(t *testing.T) {
	cfg := &Config{}
	cfg.applyDefaults()

	if cfg.Server.ReadHeaderTimeout != 10*time.Second {
		t.Fatalf("read_header_timeout = %v, want 10s", cfg.Server.ReadHeaderTimeout)
	}
	if cfg.Server.IdleTimeout != 120*time.Second {
		t.Fatalf("idle_timeout = %v, want 120s", cfg.Server.IdleTimeout)
	}
	if cfg.Server.ShutdownTimeout != 30*time.Second {
		t.Fatalf("shutdown_timeout = %v, want 30s", cfg.Server.ShutdownTimeout)
	}
	if cfg.Server.ReadTimeout != 0 {
		t.Fatalf("read_timeout = %v, want 0 (unset)", cfg.Server.ReadTimeout)
	}
}

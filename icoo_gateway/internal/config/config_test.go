package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"icoo_gateway/internal/config"
)

func TestLoadUsesDefaultsWhenEnvMissing(t *testing.T) {
	cfg, err := config.Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "127.0.0.1" {
		t.Fatalf("unexpected host: %q", cfg.Host)
	}
	if cfg.Port != 18080 {
		t.Fatalf("unexpected port: %d", cfg.Port)
	}
	if cfg.ReadTimeout != 10*time.Second {
		t.Fatalf("unexpected read timeout: %s", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout != 15*time.Second {
		t.Fatalf("unexpected write timeout: %s", cfg.WriteTimeout)
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Fatalf("unexpected shutdown timeout: %s", cfg.ShutdownTimeout)
	}
}

func TestLoadReadsDotEnv(t *testing.T) {
	root := t.TempDir()
	content := "GATEWAY_HOST=0.0.0.0\nGATEWAY_PORT=19090\nGATEWAY_READ_TIMEOUT_SECONDS=12\nGATEWAY_WRITE_TIMEOUT_SECONDS=20\nGATEWAY_SHUTDOWN_TIMEOUT_SECONDS=8\n"
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "0.0.0.0" {
		t.Fatalf("unexpected host: %q", cfg.Host)
	}
	if cfg.Port != 19090 {
		t.Fatalf("unexpected port: %d", cfg.Port)
	}
	if cfg.ReadTimeout != 12*time.Second {
		t.Fatalf("unexpected read timeout: %s", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout != 20*time.Second {
		t.Fatalf("unexpected write timeout: %s", cfg.WriteTimeout)
	}
	if cfg.ShutdownTimeout != 8*time.Second {
		t.Fatalf("unexpected shutdown timeout: %s", cfg.ShutdownTimeout)
	}
}

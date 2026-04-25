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
	if cfg.StorageDriver != "memory" {
		t.Fatalf("unexpected storage driver: %q", cfg.StorageDriver)
	}
	if cfg.DatabaseURL != "" {
		t.Fatalf("unexpected database url: %q", cfg.DatabaseURL)
	}
	if cfg.SQLitePath == "" {
		t.Fatalf("expected default sqlite path")
	}
}

func TestLoadReadsDotEnv(t *testing.T) {
	root := t.TempDir()
	content := "GATEWAY_HOST=0.0.0.0\nGATEWAY_PORT=19090\nGATEWAY_READ_TIMEOUT_SECONDS=12\nGATEWAY_WRITE_TIMEOUT_SECONDS=20\nGATEWAY_SHUTDOWN_TIMEOUT_SECONDS=8\nGATEWAY_STORAGE_DRIVER=postgres\nGATEWAY_DATABASE_URL=postgres://gateway:secret@localhost:5432/icoo_gateway?sslmode=disable\nGATEWAY_SQLITE_PATH=tmp/test.db\n"
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
	if cfg.StorageDriver != "postgres" {
		t.Fatalf("unexpected storage driver: %q", cfg.StorageDriver)
	}
	if cfg.DatabaseURL != "postgres://gateway:secret@localhost:5432/icoo_gateway?sslmode=disable" {
		t.Fatalf("unexpected database url: %q", cfg.DatabaseURL)
	}
	if cfg.SQLitePath != "tmp/test.db" {
		t.Fatalf("unexpected sqlite path: %q", cfg.SQLitePath)
	}
}

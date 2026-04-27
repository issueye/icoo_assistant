package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadUsesEnvFileAndDefaults(t *testing.T) {
	t.Setenv("PROXY_HOST", "")
	t.Setenv("PROXY_PORT", "")
	t.Setenv("PROXY_ALLOW_UNAUTHENTICATED_LOCAL", "")
	t.Setenv("PROXY_API_KEY", "")
	t.Setenv("PROXY_API_KEYS", "")

	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	data := []byte("PROXY_PORT=19191\nPROXY_ALLOW_UNAUTHENTICATED_LOCAL=false\n")
	if err := os.WriteFile(envPath, data, 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Host != "127.0.0.1" {
		t.Fatalf("expected default host, got %q", cfg.Host)
	}
	if cfg.Port != 19191 {
		t.Fatalf("expected env port, got %d", cfg.Port)
	}
	if cfg.AllowUnauthenticatedLocal {
		t.Fatalf("expected unauth local to be false")
	}
	if cfg.AnthropicVersion != "2023-06-01" {
		t.Fatalf("expected internal anthropic version default, got %q", cfg.AnthropicVersion)
	}
}

func TestLoadNormalizesLegacyAndListAuthKeys(t *testing.T) {
	t.Setenv("PROXY_API_KEY", "")
	t.Setenv("PROXY_API_KEYS", "")

	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	data := []byte("PROXY_API_KEY=alpha\nPROXY_API_KEYS=beta,gamma,alpha\n")
	if err := os.WriteFile(envPath, data, 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	want := []string{"alpha", "beta", "gamma"}
	if !reflect.DeepEqual(cfg.ProxyAPIKeys, want) {
		t.Fatalf("expected normalized proxy api keys %#v, got %#v", want, cfg.ProxyAPIKeys)
	}
	if !reflect.DeepEqual(cfg.AuthKeys(), want) {
		t.Fatalf("expected auth keys %#v, got %#v", want, cfg.AuthKeys())
	}
}

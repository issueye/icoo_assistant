package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUsesEnvFileAndDefaults(t *testing.T) {
	t.Setenv("PROXY_HOST", "")
	t.Setenv("PROXY_PORT", "")
	t.Setenv("PROXY_ALLOW_UNAUTHENTICATED_LOCAL", "")

	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	data := []byte("PROXY_PORT=19191\nPROXY_ALLOW_UNAUTHENTICATED_LOCAL=false\nOPENAI_API_KEY=test-key\nOPENAI_ONLY_STREAM=true\n")
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
	if cfg.OpenAIApiKey != "test-key" {
		t.Fatalf("expected openai key from env file, got %q", cfg.OpenAIApiKey)
	}
	if !cfg.OpenAIOnlyStream {
		t.Fatalf("expected openai only_stream from env file")
	}
}

func TestAuthKeysMergesLegacyAndListValues(t *testing.T) {
	cfg := Config{
		ProxyAPIKey:  "alpha",
		ProxyAPIKeys: []string{"beta,gamma", "alpha", " gamma "},
	}
	got := cfg.AuthKeys()
	want := []string{"alpha", "beta", "gamma"}
	if len(got) != len(want) {
		t.Fatalf("expected %#v, got %#v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %#v, got %#v", want, got)
		}
	}
}

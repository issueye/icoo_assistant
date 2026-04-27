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

func TestLoadUsesProxyAPIKeysOnly(t *testing.T) {
	t.Setenv("PROXY_API_KEYS", "")

	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	data := []byte("PROXY_API_KEYS=beta,gamma,beta\n")
	if err := os.WriteFile(envPath, data, 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	want := []string{"beta", "gamma"}
	if !reflect.DeepEqual(cfg.ProxyAPIKeys, want) {
		t.Fatalf("expected normalized proxy api keys %#v, got %#v", want, cfg.ProxyAPIKeys)
	}
	if !reflect.DeepEqual(cfg.AuthKeys(), want) {
		t.Fatalf("expected auth keys %#v, got %#v", want, cfg.AuthKeys())
	}
}

func TestLoadIgnoresLegacyModelRoutesEnvVar(t *testing.T) {
	t.Setenv("PROXY_MODEL_ROUTES", "")

	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	data := []byte("PROXY_MODEL_ROUTES=assistant-default=openai-responses:gpt-response-real\n")
	if err := os.WriteFile(envPath, data, 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.ModelRoutes != "" {
		t.Fatalf("expected legacy model routes env to be ignored, got %q", cfg.ModelRoutes)
	}
}

func TestLoadIgnoresLegacyDefaultRouteEnvVars(t *testing.T) {
	t.Setenv("PROXY_DEFAULT_ANTHROPIC_ROUTE", "")
	t.Setenv("PROXY_DEFAULT_CHAT_ROUTE", "")
	t.Setenv("PROXY_DEFAULT_RESPONSES_ROUTE", "")

	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	data := []byte("PROXY_DEFAULT_ANTHROPIC_ROUTE=anthropic:claude-real\nPROXY_DEFAULT_CHAT_ROUTE=openai-chat:gpt-chat-real\nPROXY_DEFAULT_RESPONSES_ROUTE=openai-responses:gpt-response-real\n")
	if err := os.WriteFile(envPath, data, 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.DefaultAnthropicRoute != "" {
		t.Fatalf("expected legacy anthropic default route env to be ignored, got %q", cfg.DefaultAnthropicRoute)
	}
	if cfg.DefaultChatRoute != "" {
		t.Fatalf("expected legacy chat default route env to be ignored, got %q", cfg.DefaultChatRoute)
	}
	if cfg.DefaultResponsesRoute != "" {
		t.Fatalf("expected legacy responses default route env to be ignored, got %q", cfg.DefaultResponsesRoute)
	}
}

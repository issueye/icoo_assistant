package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"icoo_assistant/internal/config"
)

func TestLoadTOMLReadsConfig(t *testing.T) {
	dir := t.TempDir()
	tomlPath := filepath.Join(dir, "config.toml")
	content := `[agent]
system_prompt = "You are a test agent"
max_rounds = 10
command_timeout_seconds = 60

[anthropic]
api_key = "sk-ant-test-123"
model = "claude-sonnet-4-7"
max_tokens = 8000
enable_thinking = false
`
	if err := os.WriteFile(tomlPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadTOML(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SystemPrompt != "You are a test agent" {
		t.Fatalf("expected system prompt, got %q", cfg.SystemPrompt)
	}
	if cfg.MaxRounds != 10 {
		t.Fatalf("expected max_rounds=10, got %d", cfg.MaxRounds)
	}
	if cfg.AnthropicAPIKey != "sk-ant-test-123" {
		t.Fatalf("expected api key, got %q", cfg.AnthropicAPIKey)
	}
	if cfg.AnthropicModel != "claude-sonnet-4-7" {
		t.Fatalf("expected model, got %q", cfg.AnthropicModel)
	}
	if cfg.AnthropicMaxTokens != 8000 {
		t.Fatalf("expected max_tokens=8000, got %d", cfg.AnthropicMaxTokens)
	}
	if cfg.EnableThinking {
		t.Fatal("expected thinking disabled")
	}
	if cfg.CommandTimeout.Seconds() != 60 {
		t.Fatalf("expected timeout=60s, got %v", cfg.CommandTimeout)
	}
}

func TestGenerateDefaultTOMLCreatesFile(t *testing.T) {
	dir := t.TempDir()
	if err := config.GenerateDefaultTOML(dir); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "[agent]") {
		t.Fatal("missing [agent] section")
	}
	if !strings.Contains(content, "[anthropic]") {
		t.Fatal("missing [anthropic] section")
	}
	if !strings.Contains(content, "max_rounds = 20") {
		t.Fatal("missing default max_rounds")
	}
	if !strings.Contains(content, "model = \"claude-opus-4-7\"") {
		t.Fatal("missing default model")
	}
}

func TestGenerateDefaultTOMLRejectsExisting(t *testing.T) {
	dir := t.TempDir()
	_ = config.GenerateDefaultTOML(dir)
	if err := config.GenerateDefaultTOML(dir); err == nil {
		t.Fatal("expected error for existing config.toml")
	}
}

func TestLoadFallsBackToDefaultsWhenNoConfig(t *testing.T) {
	dir := t.TempDir()
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AnthropicModel == "" {
		t.Fatal("expected non-empty model")
	}
	if cfg.MaxRounds <= 0 {
		t.Fatalf("expected positive max_rounds, got %d", cfg.MaxRounds)
	}
}

func TestLoadEnvOverridesTOML(t *testing.T) {
	dir := t.TempDir()
	tomlPath := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(tomlPath, []byte(`[anthropic]
api_key = "toml-key"
model = "toml-model"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	os.Setenv("ANTHROPIC_API_KEY", "env-key")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AnthropicAPIKey != "env-key" {
		t.Fatalf("env should override toml, got %q", cfg.AnthropicAPIKey)
	}
}

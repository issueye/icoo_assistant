package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"icoo_assistant/internal/config"
)

func TestLoadReadsDotEnvWithoutOverridingExistingEnv(t *testing.T) {
	root := t.TempDir()
	content := "ANTHROPIC_API_KEY=from-dotenv\nANTHROPIC_BASE_URL=https://anthropic-proxy.example.com\nAGENT_MAX_ROUNDS=12\nANTHROPIC_ENABLE_PROMPT_CACHE=true\n"
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ANTHROPIC_API_KEY", "from-env")
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AnthropicAPIKey != "from-env" {
		t.Fatalf("expected env key to win, got %q", cfg.AnthropicAPIKey)
	}
	if cfg.MaxRounds != 12 {
		t.Fatalf("expected max rounds 12, got %d", cfg.MaxRounds)
	}
	if cfg.AnthropicBaseURL != "https://anthropic-proxy.example.com" {
		t.Fatalf("expected anthropic base url from dotenv, got %q", cfg.AnthropicBaseURL)
	}
	if !cfg.EnablePromptCache {
		t.Fatal("expected prompt cache to be enabled")
	}
	if !cfg.EnableStreaming {
		t.Fatal("expected streaming to be enabled by default")
	}
}

func TestLoadAppliesDefaults(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ANTHROPIC_MODEL", "")
	t.Setenv("AGENT_MAX_ROUNDS", "")
	t.Setenv("AGENT_COMMAND_TIMEOUT_SECONDS", "")
	t.Setenv("AGENT_SYSTEM_PROMPT", "")
	t.Setenv("ANTHROPIC_ENABLE_PROMPT_CACHE", "")
	t.Setenv("ANTHROPIC_ENABLE_THINKING", "")
	t.Setenv("ANTHROPIC_ENABLE_STREAMING", "")
	t.Setenv("ANTHROPIC_MAX_TOKENS", "")
	t.Setenv("ANTHROPIC_BASE_URL", "")
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AnthropicBaseURL != "" {
		t.Fatalf("expected empty base url default, got %q", cfg.AnthropicBaseURL)
	}
	if cfg.AnthropicModel != "claude-opus-4-7" {
		t.Fatalf("unexpected model default: %q", cfg.AnthropicModel)
	}
	if cfg.MaxRounds != 20 {
		t.Fatalf("unexpected max rounds default: %d", cfg.MaxRounds)
	}
	if cfg.CommandTimeout != 120*time.Second {
		t.Fatalf("unexpected timeout default: %s", cfg.CommandTimeout)
	}
	if cfg.SystemPrompt == "" {
		t.Fatal("expected default system prompt")
	}
	if !cfg.EnableStreaming {
		t.Fatal("expected streaming to be enabled by default")
	}
}

func TestLoadCanDisableStreaming(t *testing.T) {
	root := t.TempDir()
	content := "ANTHROPIC_ENABLE_STREAMING=false\n"
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.EnableStreaming {
		t.Fatal("expected streaming to be disabled")
	}
}

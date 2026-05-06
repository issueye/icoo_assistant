package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"icoo_assistant/internal/config"
)

func TestLoadReadsConfigTOML(t *testing.T) {
	root := t.TempDir()
	content := "[core]\nmax_rounds = 12\n\n[anthropic]\napi_key = \"from-config\"\nbase_url = \"https://anthropic-proxy.example.com\"\nenable_prompt_cache = true\n"
	if err := os.WriteFile(filepath.Join(root, "config.toml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AnthropicAPIKey != "from-config" {
		t.Fatalf("expected api key from config, got %q", cfg.AnthropicAPIKey)
	}
	if cfg.MaxRounds != 12 {
		t.Fatalf("expected max rounds 12, got %d", cfg.MaxRounds)
	}
	if cfg.AnthropicBaseURL != "https://anthropic-proxy.example.com" {
		t.Fatalf("expected anthropic base url from config, got %q", cfg.AnthropicBaseURL)
	}
	if !cfg.EnablePromptCache {
		t.Fatal("expected prompt cache to be enabled")
	}
}

func TestLoadAppliesDefaults(t *testing.T) {
	root := t.TempDir()
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
}

func TestLoadReadsClaudeProjectConventions(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("Project rule: run tests before finishing."), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".icoo", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "external"), 0o755); err != nil {
		t.Fatal(err)
	}
	settings := `{
  "permissions": {
    "deny": [
      "write(secrets/**)",
      "read(private/**)",
      "bash(rm *)"
    ],
    "additionalDirectories": ["external"],
    "defaultMode": "acceptEdits"
  }
}`
	if err := os.MkdirAll(filepath.Join(root, ".icoo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".icoo", "settings.json"), []byte(settings), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ProjectInstructions != "Project rule: run tests before finishing." {
		t.Fatalf("unexpected project instructions: %q", cfg.ProjectInstructions)
	}
	if cfg.SkillsDir != filepath.Join(root, ".icoo", "skills") {
		t.Fatalf("expected .icoo skills dir, got %q", cfg.SkillsDir)
	}
	if cfg.PermissionMode != "acceptEdits" {
		t.Fatalf("unexpected permission mode: %q", cfg.PermissionMode)
	}
	if len(cfg.AdditionalDirectories) != 1 || cfg.AdditionalDirectories[0] != "external" {
		t.Fatalf("unexpected additional directories: %#v", cfg.AdditionalDirectories)
	}
	if len(cfg.DenyWritePatterns) != 1 || cfg.DenyWritePatterns[0] != "secrets/**" {
		t.Fatalf("unexpected deny write patterns: %#v", cfg.DenyWritePatterns)
	}
	if len(cfg.DenyReadPatterns) != 1 || cfg.DenyReadPatterns[0] != "private/**" {
		t.Fatalf("unexpected deny read patterns: %#v", cfg.DenyReadPatterns)
	}
	if len(cfg.DenyCommandPatterns) != 1 || cfg.DenyCommandPatterns[0] != "rm *" {
		t.Fatalf("unexpected deny command patterns: %#v", cfg.DenyCommandPatterns)
	}
}

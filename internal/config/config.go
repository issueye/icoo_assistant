package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Workdir            string
	SystemPrompt       string
	SkillsDir          string
	MaxRounds          int
	CommandTimeout     time.Duration
	AnthropicAPIKey    string
	AnthropicBaseURL   string
	AnthropicModel     string
	AnthropicMaxTokens int64
	EnablePromptCache  bool
	EnableThinking     bool
	EnableStreaming    bool
	CompactThreshold   int
	TranscriptDir      string
}

func Load(workdir string) (Config, error) {
	tomlPath := filepath.Join(workdir, "config.toml")
	envPath := filepath.Join(workdir, ".env")
	if _, err := os.Stat(tomlPath); os.IsNotExist(err) {
		if _, err := os.Stat(envPath); os.IsNotExist(err) {
			if err := GenerateDefaultTOML(workdir); err != nil {
				return Config{}, err
			}
		}
	}
	cfg, err := LoadTOML(workdir)
	if err != nil {
		if !os.IsNotExist(err) {
			return Config{}, err
		}
		cfg = Config{
			Workdir:            workdir,
			AnthropicModel:     "claude-opus-4-7",
			MaxRounds:          20,
			AnthropicMaxTokens: 16000,
			CommandTimeout:     120 * time.Second,
			CompactThreshold:   50000,
			EnableThinking:     true,
			EnableStreaming:    true,
		}
	}
	if err := loadDotEnv(filepath.Join(workdir, ".env")); err != nil {
		return Config{}, err
	}
	if v := strings.TrimSpace(os.Getenv("AGENT_SYSTEM_PROMPT")); v != "" {
		cfg.SystemPrompt = v
	}
	if v := strings.TrimSpace(os.Getenv("AGENT_SKILLS_DIR")); v != "" {
		cfg.SkillsDir = v
	}
	if v := strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY")); v != "" {
		cfg.AnthropicAPIKey = v
	}
	if v := strings.TrimSpace(os.Getenv("ANTHROPIC_BASE_URL")); v != "" {
		cfg.AnthropicBaseURL = v
	}
	if v := strings.TrimSpace(os.Getenv("ANTHROPIC_MODEL")); v != "" {
		cfg.AnthropicModel = v
	}
	cfg.EnablePromptCache = boolFromEnv("ANTHROPIC_ENABLE_PROMPT_CACHE", cfg.EnablePromptCache)
	cfg.EnableThinking = boolFromEnv("ANTHROPIC_ENABLE_THINKING", cfg.EnableThinking)
	cfg.EnableStreaming = boolFromEnv("ANTHROPIC_ENABLE_STREAMING", cfg.EnableStreaming)
	if v := intFromEnv("AGENT_COMMAND_TIMEOUT_SECONDS", -1); v > 0 {
		cfg.CommandTimeout = time.Duration(v) * time.Second
	}
	if v := intFromEnv("AGENT_MAX_ROUNDS", -1); v > 0 {
		cfg.MaxRounds = v
	}
	if v := intFromEnv("ANTHROPIC_MAX_TOKENS", -1); v > 0 {
		cfg.AnthropicMaxTokens = int64(v)
	}
	if v := intFromEnv("AGENT_COMPACT_THRESHOLD", -1); v > 0 {
		cfg.CompactThreshold = v
	}
	if v := strings.TrimSpace(os.Getenv("AGENT_TRANSCRIPT_DIR")); v != "" {
		cfg.TranscriptDir = v
	}
	applyDefaults(&cfg)
	return cfg, nil
}

func loadDotEnv(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, "\"")
		value = strings.Trim(value, "'")
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}
	return nil
}

func boolFromEnv(key string, fallback bool) bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if raw == "" {
		return fallback
	}
	switch raw {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func intFromEnv(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func durationFromEnv(key string, fallback time.Duration) time.Duration {
	seconds := intFromEnv(key, int(fallback/time.Second))
	return time.Duration(seconds) * time.Second
}

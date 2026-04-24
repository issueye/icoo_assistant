package config

import (
	"fmt"
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
	CompactThreshold   int
	TranscriptDir      string
}

func Load(workdir string) (Config, error) {
	if err := loadDotEnv(filepath.Join(workdir, ".env")); err != nil {
		return Config{}, err
	}
	cfg := Config{
		Workdir:            workdir,
		SystemPrompt:       strings.TrimSpace(os.Getenv("AGENT_SYSTEM_PROMPT")),
		SkillsDir:          strings.TrimSpace(os.Getenv("AGENT_SKILLS_DIR")),
		AnthropicAPIKey:    strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY")),
		AnthropicBaseURL:   strings.TrimSpace(os.Getenv("ANTHROPIC_BASE_URL")),
		AnthropicModel:     strings.TrimSpace(os.Getenv("ANTHROPIC_MODEL")),
		EnablePromptCache:  boolFromEnv("ANTHROPIC_ENABLE_PROMPT_CACHE", false),
		EnableThinking:     boolFromEnv("ANTHROPIC_ENABLE_THINKING", true),
		CommandTimeout:     durationFromEnv("AGENT_COMMAND_TIMEOUT_SECONDS", 120*time.Second),
		MaxRounds:          intFromEnv("AGENT_MAX_ROUNDS", 20),
		AnthropicMaxTokens: int64(intFromEnv("ANTHROPIC_MAX_TOKENS", 16000)),
		CompactThreshold:   intFromEnv("AGENT_COMPACT_THRESHOLD", 50000),
		TranscriptDir:      strings.TrimSpace(os.Getenv("AGENT_TRANSCRIPT_DIR")),
	}
	if cfg.SystemPrompt == "" {
		cfg.SystemPrompt = fmt.Sprintf("You are a coding agent at %s. Use tools to solve tasks.", workdir)
	}
	if cfg.SkillsDir == "" {
		cfg.SkillsDir = filepath.Join(workdir, "skills")
	}
	if cfg.AnthropicModel == "" {
		cfg.AnthropicModel = "claude-opus-4-7"
	}
	if cfg.TranscriptDir == "" {
		cfg.TranscriptDir = filepath.Join(workdir, ".transcripts")
	}
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

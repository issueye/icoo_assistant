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
	values, err := loadTOMLConfig(filepath.Join(workdir, "config.toml"))
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		Workdir:            workdir,
		SystemPrompt:       strings.TrimSpace(values["core.system_prompt"]),
		SkillsDir:          strings.TrimSpace(values["core.skills_dir"]),
		AnthropicAPIKey:    strings.TrimSpace(values["anthropic.api_key"]),
		AnthropicBaseURL:   strings.TrimSpace(values["anthropic.base_url"]),
		AnthropicModel:     strings.TrimSpace(values["anthropic.model"]),
		EnablePromptCache:  boolFromValue(values["anthropic.enable_prompt_cache"], false),
		EnableThinking:     boolFromValue(values["anthropic.enable_thinking"], true),
		CommandTimeout:     durationFromValue(values["core.command_timeout_seconds"], 120*time.Second),
		MaxRounds:          intFromValue(values["core.max_rounds"], 20),
		AnthropicMaxTokens: int64(intFromValue(values["anthropic.max_tokens"], 16000)),
		CompactThreshold:   intFromValue(values["core.compact_threshold"], 50000),
		TranscriptDir:      strings.TrimSpace(values["core.transcript_dir"]),
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

func loadTOMLConfig(path string) (map[string]string, error) {
	result := map[string]string{}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, err
	}
	section := ""
	for _, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "["), "]")))
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, "\"")
		value = strings.Trim(value, "'")
		if key == "" {
			continue
		}
		fullKey := key
		if section != "" {
			fullKey = section + "." + key
		}
		result[strings.ToLower(fullKey)] = value
	}
	return result, nil
}

func boolFromValue(raw string, fallback bool) bool {
	raw = strings.TrimSpace(strings.ToLower(raw))
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

func intFromValue(raw string, fallback int) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func durationFromValue(raw string, fallback time.Duration) time.Duration {
	seconds := intFromValue(raw, int(fallback/time.Second))
	return time.Duration(seconds) * time.Second
}

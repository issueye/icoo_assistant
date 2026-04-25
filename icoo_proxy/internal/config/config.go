package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Host                      string
	Port                      int
	ReadTimeout               time.Duration
	WriteTimeout              time.Duration
	ShutdownTimeout           time.Duration
	ProxyAPIKey               string
	ProxyAPIKeys              []string
	AllowUnauthenticatedLocal bool
	AnthropicBaseURL          string
	AnthropicAPIKey           string
	AnthropicVersion          string
	OpenAIBaseURL             string
	OpenAIApiKey              string
	DefaultAnthropicRoute     string
	DefaultChatRoute          string
	DefaultResponsesRoute     string
	ModelRoutes               string
}

func Load(workdir string) (Config, error) {
	if err := loadDotEnv(filepath.Join(workdir, ".env")); err != nil {
		return Config{}, err
	}
	cfg := Config{
		Host:                      strings.TrimSpace(os.Getenv("PROXY_HOST")),
		Port:                      intFromEnv("PROXY_PORT", 18181),
		ReadTimeout:               durationFromEnv("PROXY_READ_TIMEOUT_SECONDS", 15*time.Second),
		WriteTimeout:              durationFromEnv("PROXY_WRITE_TIMEOUT_SECONDS", 300*time.Second),
		ShutdownTimeout:           durationFromEnv("PROXY_SHUTDOWN_TIMEOUT_SECONDS", 10*time.Second),
		ProxyAPIKey:               strings.TrimSpace(os.Getenv("PROXY_API_KEY")),
		ProxyAPIKeys:              csvFromEnv("PROXY_API_KEYS"),
		AllowUnauthenticatedLocal: boolFromEnv("PROXY_ALLOW_UNAUTHENTICATED_LOCAL", true),
		AnthropicBaseURL:          strings.TrimSpace(os.Getenv("ANTHROPIC_BASE_URL")),
		AnthropicAPIKey:           strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY")),
		AnthropicVersion:          strings.TrimSpace(os.Getenv("ANTHROPIC_VERSION")),
		OpenAIBaseURL:             strings.TrimSpace(os.Getenv("OPENAI_BASE_URL")),
		OpenAIApiKey:              strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
		DefaultAnthropicRoute:     strings.TrimSpace(os.Getenv("PROXY_DEFAULT_ANTHROPIC_ROUTE")),
		DefaultChatRoute:          strings.TrimSpace(os.Getenv("PROXY_DEFAULT_CHAT_ROUTE")),
		DefaultResponsesRoute:     strings.TrimSpace(os.Getenv("PROXY_DEFAULT_RESPONSES_ROUTE")),
		ModelRoutes:               strings.TrimSpace(os.Getenv("PROXY_MODEL_ROUTES")),
	}
	if cfg.Host == "" {
		cfg.Host = "127.0.0.1"
	}
	if cfg.AnthropicBaseURL == "" {
		cfg.AnthropicBaseURL = "https://api.anthropic.com"
	}
	if cfg.OpenAIBaseURL == "" {
		cfg.OpenAIBaseURL = "https://api.openai.com"
	}
	if cfg.AnthropicVersion == "" {
		cfg.AnthropicVersion = "2023-06-01"
	}
	return cfg, nil
}

func (c Config) AuthKeys() []string {
	values := make([]string, 0, len(c.ProxyAPIKeys)+1)
	for _, item := range append([]string{c.ProxyAPIKey}, c.ProxyAPIKeys...) {
		for _, part := range strings.Split(item, ",") {
			value := strings.TrimSpace(part)
			if value != "" && !slices.Contains(values, value) {
				values = append(values, value)
			}
		}
	}
	return values
}

func (c Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
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

func boolFromEnv(key string, fallback bool) bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	switch raw {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	case "":
		return fallback
	default:
		return fallback
	}
}

func csvFromEnv(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}
	values := make([]string, 0)
	for _, part := range strings.Split(raw, ",") {
		value := strings.TrimSpace(part)
		if value != "" && !slices.Contains(values, value) {
			values = append(values, value)
		}
	}
	return values
}

func durationFromEnv(key string, fallback time.Duration) time.Duration {
	seconds := intFromEnv(key, int(fallback/time.Second))
	return time.Duration(seconds) * time.Second
}

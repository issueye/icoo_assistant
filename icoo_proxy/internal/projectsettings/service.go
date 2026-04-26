package projectsettings

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Values struct {
	ProxyHost                   string `json:"proxy_host"`
	ProxyPort                   int    `json:"proxy_port"`
	ProxyReadTimeoutSeconds     int    `json:"proxy_read_timeout_seconds"`
	ProxyWriteTimeoutSeconds    int    `json:"proxy_write_timeout_seconds"`
	ProxyShutdownTimeoutSeconds int    `json:"proxy_shutdown_timeout_seconds"`
	ProxyAllowUnauthenticated   bool   `json:"proxy_allow_unauthenticated_local"`
	ProxyAPIKey                 string `json:"proxy_api_key"`
	ProxyAPIKeys                string `json:"proxy_api_keys"`
	ProxyDefaultAnthropicRoute  string `json:"proxy_default_anthropic_route"`
	ProxyDefaultChatRoute       string `json:"proxy_default_chat_route"`
	ProxyDefaultResponsesRoute  string `json:"proxy_default_responses_route"`
	ProxyModelRoutes            string `json:"proxy_model_routes"`
	ProxyChainLogPath           string `json:"proxy_chain_log_path"`
	ProxyChainLogBodies         bool   `json:"proxy_chain_log_bodies"`
	ProxyChainLogMaxBodyBytes   int    `json:"proxy_chain_log_max_body_bytes"`
	AnthropicBaseURL            string `json:"anthropic_base_url"`
	AnthropicAPIKey             string `json:"anthropic_api_key"`
	AnthropicVersion            string `json:"anthropic_version"`
	AnthropicOnlyStream         bool   `json:"anthropic_only_stream"`
	AnthropicUserAgent          string `json:"anthropic_user_agent"`
	OpenAIBaseURL               string `json:"openai_base_url"`
	OpenAIApiKey                string `json:"openai_api_key"`
	OpenAIOnlyStream            bool   `json:"openai_only_stream"`
	OpenAIUserAgent             string `json:"openai_user_agent"`
}

func Load(root string) (Values, error) {
	env, err := readEnvFile(filepath.Join(root, ".env"))
	if err != nil {
		return Values{}, err
	}
	return Values{
		ProxyHost:                   stringWithDefault(env, "PROXY_HOST", "127.0.0.1"),
		ProxyPort:                   intWithDefault(env, "PROXY_PORT", 18181),
		ProxyReadTimeoutSeconds:     intWithDefault(env, "PROXY_READ_TIMEOUT_SECONDS", 15),
		ProxyWriteTimeoutSeconds:    intWithDefault(env, "PROXY_WRITE_TIMEOUT_SECONDS", 300),
		ProxyShutdownTimeoutSeconds: intWithDefault(env, "PROXY_SHUTDOWN_TIMEOUT_SECONDS", 10),
		ProxyAllowUnauthenticated:   boolWithDefault(env, "PROXY_ALLOW_UNAUTHENTICATED_LOCAL", true),
		ProxyAPIKey:                 strings.TrimSpace(env["PROXY_API_KEY"]),
		ProxyAPIKeys:                strings.TrimSpace(env["PROXY_API_KEYS"]),
		ProxyDefaultAnthropicRoute:  strings.TrimSpace(env["PROXY_DEFAULT_ANTHROPIC_ROUTE"]),
		ProxyDefaultChatRoute:       strings.TrimSpace(env["PROXY_DEFAULT_CHAT_ROUTE"]),
		ProxyDefaultResponsesRoute:  strings.TrimSpace(env["PROXY_DEFAULT_RESPONSES_ROUTE"]),
		ProxyModelRoutes:            strings.TrimSpace(env["PROXY_MODEL_ROUTES"]),
		ProxyChainLogPath:           stringWithDefault(env, "PROXY_CHAIN_LOG_PATH", filepath.Join(root, ".data", "icoo_proxy-chain.log")),
		ProxyChainLogBodies:         boolWithDefault(env, "PROXY_CHAIN_LOG_BODIES", true),
		ProxyChainLogMaxBodyBytes:   intWithDefault(env, "PROXY_CHAIN_LOG_MAX_BODY_BYTES", 0),
		AnthropicBaseURL:            stringWithDefault(env, "ANTHROPIC_BASE_URL", "https://api.anthropic.com"),
		AnthropicAPIKey:             strings.TrimSpace(env["ANTHROPIC_API_KEY"]),
		AnthropicVersion:            stringWithDefault(env, "ANTHROPIC_VERSION", "2023-06-01"),
		AnthropicOnlyStream:         boolWithDefault(env, "ANTHROPIC_ONLY_STREAM", false),
		AnthropicUserAgent:          strings.TrimSpace(env["ANTHROPIC_USER_AGENT"]),
		OpenAIBaseURL:               stringWithDefault(env, "OPENAI_BASE_URL", "https://api.openai.com"),
		OpenAIApiKey:                strings.TrimSpace(env["OPENAI_API_KEY"]),
		OpenAIOnlyStream:            boolWithDefault(env, "OPENAI_ONLY_STREAM", false),
		OpenAIUserAgent:             strings.TrimSpace(env["OPENAI_USER_AGENT"]),
	}, nil
}

func Save(root string, values Values) error {
	if err := validate(values); err != nil {
		return err
	}
	envPath := filepath.Join(root, ".env")
	content := buildEnv(values)
	if err := os.WriteFile(envPath, []byte(content), 0o644); err != nil {
		return err
	}
	applyProcessEnv(values)
	return nil
}

func validate(values Values) error {
	if strings.TrimSpace(values.ProxyHost) == "" {
		return fmt.Errorf("proxy_host is required")
	}
	if values.ProxyPort <= 0 {
		return fmt.Errorf("proxy_port must be greater than 0")
	}
	if values.ProxyReadTimeoutSeconds <= 0 {
		return fmt.Errorf("proxy_read_timeout_seconds must be greater than 0")
	}
	if values.ProxyWriteTimeoutSeconds <= 0 {
		return fmt.Errorf("proxy_write_timeout_seconds must be greater than 0")
	}
	if values.ProxyShutdownTimeoutSeconds <= 0 {
		return fmt.Errorf("proxy_shutdown_timeout_seconds must be greater than 0")
	}
	if values.ProxyChainLogMaxBodyBytes < 0 {
		return fmt.Errorf("proxy_chain_log_max_body_bytes must be 0 or greater")
	}
	return nil
}

func buildEnv(values Values) string {
	lines := []string{
		"PROXY_HOST=" + strings.TrimSpace(values.ProxyHost),
		"PROXY_PORT=" + strconv.Itoa(values.ProxyPort),
		"PROXY_READ_TIMEOUT_SECONDS=" + strconv.Itoa(values.ProxyReadTimeoutSeconds),
		"PROXY_WRITE_TIMEOUT_SECONDS=" + strconv.Itoa(values.ProxyWriteTimeoutSeconds),
		"PROXY_SHUTDOWN_TIMEOUT_SECONDS=" + strconv.Itoa(values.ProxyShutdownTimeoutSeconds),
		"PROXY_CHAIN_LOG_PATH=" + strings.TrimSpace(values.ProxyChainLogPath),
		"PROXY_CHAIN_LOG_BODIES=" + formatBool(values.ProxyChainLogBodies),
		"PROXY_CHAIN_LOG_MAX_BODY_BYTES=" + strconv.Itoa(values.ProxyChainLogMaxBodyBytes),
		"",
		"# Optional downstream key for local clients.",
		"# If empty and PROXY_ALLOW_UNAUTHENTICATED_LOCAL=true, local clients may call without auth.",
		"PROXY_API_KEY=" + strings.TrimSpace(values.ProxyAPIKey),
		"PROXY_API_KEYS=" + strings.TrimSpace(values.ProxyAPIKeys),
		"PROXY_ALLOW_UNAUTHENTICATED_LOCAL=" + formatBool(values.ProxyAllowUnauthenticated),
		"",
		"ANTHROPIC_BASE_URL=" + strings.TrimSpace(values.AnthropicBaseURL),
		"ANTHROPIC_API_KEY=" + strings.TrimSpace(values.AnthropicAPIKey),
		"ANTHROPIC_VERSION=" + strings.TrimSpace(values.AnthropicVersion),
		"ANTHROPIC_ONLY_STREAM=" + formatBool(values.AnthropicOnlyStream),
		"ANTHROPIC_USER_AGENT=" + strings.TrimSpace(values.AnthropicUserAgent),
		"",
		"OPENAI_BASE_URL=" + strings.TrimSpace(values.OpenAIBaseURL),
		"OPENAI_API_KEY=" + strings.TrimSpace(values.OpenAIApiKey),
		"OPENAI_ONLY_STREAM=" + formatBool(values.OpenAIOnlyStream),
		"OPENAI_USER_AGENT=" + strings.TrimSpace(values.OpenAIUserAgent),
		"",
		"# Defaults: <protocol>:<real-model>",
		"PROXY_DEFAULT_ANTHROPIC_ROUTE=" + strings.TrimSpace(values.ProxyDefaultAnthropicRoute),
		"PROXY_DEFAULT_CHAT_ROUTE=" + strings.TrimSpace(values.ProxyDefaultChatRoute),
		"PROXY_DEFAULT_RESPONSES_ROUTE=" + strings.TrimSpace(values.ProxyDefaultResponsesRoute),
		"",
		"# Aliases: alias=<protocol>:<real-model>,alias2=<protocol>:<real-model>",
		"PROXY_MODEL_ROUTES=" + strings.TrimSpace(values.ProxyModelRoutes),
		"",
	}
	return strings.Join(lines, "\n")
}

func formatBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func applyProcessEnv(values Values) {
	set := func(key, value string) {
		_ = os.Setenv(key, value)
	}
	set("PROXY_HOST", strings.TrimSpace(values.ProxyHost))
	set("PROXY_PORT", strconv.Itoa(values.ProxyPort))
	set("PROXY_READ_TIMEOUT_SECONDS", strconv.Itoa(values.ProxyReadTimeoutSeconds))
	set("PROXY_WRITE_TIMEOUT_SECONDS", strconv.Itoa(values.ProxyWriteTimeoutSeconds))
	set("PROXY_SHUTDOWN_TIMEOUT_SECONDS", strconv.Itoa(values.ProxyShutdownTimeoutSeconds))
	set("PROXY_ALLOW_UNAUTHENTICATED_LOCAL", formatBool(values.ProxyAllowUnauthenticated))
	set("PROXY_API_KEY", strings.TrimSpace(values.ProxyAPIKey))
	set("PROXY_API_KEYS", strings.TrimSpace(values.ProxyAPIKeys))
	set("PROXY_DEFAULT_ANTHROPIC_ROUTE", strings.TrimSpace(values.ProxyDefaultAnthropicRoute))
	set("PROXY_DEFAULT_CHAT_ROUTE", strings.TrimSpace(values.ProxyDefaultChatRoute))
	set("PROXY_DEFAULT_RESPONSES_ROUTE", strings.TrimSpace(values.ProxyDefaultResponsesRoute))
	set("PROXY_MODEL_ROUTES", strings.TrimSpace(values.ProxyModelRoutes))
	set("PROXY_CHAIN_LOG_PATH", strings.TrimSpace(values.ProxyChainLogPath))
	set("PROXY_CHAIN_LOG_BODIES", formatBool(values.ProxyChainLogBodies))
	set("PROXY_CHAIN_LOG_MAX_BODY_BYTES", strconv.Itoa(values.ProxyChainLogMaxBodyBytes))
	set("ANTHROPIC_BASE_URL", strings.TrimSpace(values.AnthropicBaseURL))
	set("ANTHROPIC_API_KEY", strings.TrimSpace(values.AnthropicAPIKey))
	set("ANTHROPIC_VERSION", strings.TrimSpace(values.AnthropicVersion))
	set("ANTHROPIC_ONLY_STREAM", formatBool(values.AnthropicOnlyStream))
	set("ANTHROPIC_USER_AGENT", strings.TrimSpace(values.AnthropicUserAgent))
	set("OPENAI_BASE_URL", strings.TrimSpace(values.OpenAIBaseURL))
	set("OPENAI_API_KEY", strings.TrimSpace(values.OpenAIApiKey))
	set("OPENAI_ONLY_STREAM", formatBool(values.OpenAIOnlyStream))
	set("OPENAI_USER_AGENT", strings.TrimSpace(values.OpenAIUserAgent))
}

func readEnvFile(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	values := make(map[string]string)
	for _, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		values[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), "\"'")
	}
	return values, nil
}

func stringWithDefault(values map[string]string, key, fallback string) string {
	if value := strings.TrimSpace(values[key]); value != "" {
		return value
	}
	return fallback
}

func intWithDefault(values map[string]string, key string, fallback int) int {
	raw := strings.TrimSpace(values[key])
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func boolWithDefault(values map[string]string, key string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(values[key])) {
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

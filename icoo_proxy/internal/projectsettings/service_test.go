package projectsettings

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAndSave(t *testing.T) {
	root := t.TempDir()
	values := Values{
		ProxyHost:                   "127.0.0.1",
		ProxyPort:                   19191,
		ProxyReadTimeoutSeconds:     10,
		ProxyWriteTimeoutSeconds:    60,
		ProxyShutdownTimeoutSeconds: 5,
		ProxyAllowUnauthenticated:   false,
		ProxyDefaultAnthropicRoute:  "anthropic:claude-test",
		ProxyDefaultChatRoute:       "openai-chat:gpt-test",
		ProxyDefaultResponsesRoute:  "openai-responses:gpt-resp",
		ProxyModelRoutes:            "alias=openai-responses:gpt-resp",
		ProxyChainLogPath:           ".data/test.log",
		ProxyChainLogBodies:         true,
		ProxyChainLogMaxBodyBytes:   123,
		AnthropicBaseURL:            "https://anthropic.example",
		AnthropicAPIKey:             "anth-key",
		AnthropicVersion:            "2023-06-01",
		AnthropicOnlyStream:         true,
		AnthropicUserAgent:          "AnthUA/1.0",
		OpenAIBaseURL:               "https://openai.example",
		OpenAIApiKey:                "openai-key",
		OpenAIOnlyStream:            true,
		OpenAIUserAgent:             "OpenAIUA/2.0",
	}
	if err := Save(root, values); err != nil {
		t.Fatalf("save settings: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".env"))
	if err != nil {
		t.Fatalf("read env: %v", err)
	}
	text := string(data)
	for _, needle := range []string{
		"PROXY_PORT=19191",
		"ANTHROPIC_ONLY_STREAM=true",
		"ANTHROPIC_USER_AGENT=AnthUA/1.0",
		"OPENAI_USER_AGENT=OpenAIUA/2.0",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("expected env file to contain %q, got %s", needle, text)
		}
	}
	loaded, err := Load(root)
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	if loaded.ProxyPort != 19191 {
		t.Fatalf("expected port to round-trip, got %d", loaded.ProxyPort)
	}
	if loaded.OpenAIUserAgent != "OpenAIUA/2.0" {
		t.Fatalf("expected openai user agent to round-trip, got %q", loaded.OpenAIUserAgent)
	}
	if !loaded.OpenAIOnlyStream {
		t.Fatalf("expected openai only_stream to round-trip")
	}
}

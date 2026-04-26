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
		"PROXY_CHAIN_LOG_MAX_BODY_BYTES=123",
		"PROXY_DEFAULT_CHAT_ROUTE=openai-chat:gpt-test",
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
	if loaded.ProxyChainLogMaxBodyBytes != 123 {
		t.Fatalf("expected log body limit to round-trip, got %d", loaded.ProxyChainLogMaxBodyBytes)
	}
}

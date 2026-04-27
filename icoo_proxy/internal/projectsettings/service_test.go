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
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("expected env file to contain %q, got %s", needle, text)
		}
	}
	for _, needle := range []string{
		"PROXY_API_KEYS",
		"PROXY_ALLOW_UNAUTHENTICATED_LOCAL",
		"PROXY_DEFAULT_ANTHROPIC_ROUTE",
		"PROXY_DEFAULT_CHAT_ROUTE",
		"PROXY_DEFAULT_RESPONSES_ROUTE",
		"PROXY_MODEL_ROUTES",
	} {
		if strings.Contains(text, needle) {
			t.Fatalf("expected env file to omit migrated setting %q, got %s", needle, text)
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

func TestSavePreservesUnmanagedEnvEntries(t *testing.T) {
	root := t.TempDir()
	envPath := filepath.Join(root, ".env")
	existing := strings.Join([]string{
		"# keep this comment",
		"CUSTOM_FLAG=enabled",
		"PROXY_API_KEYS=alpha,beta",
		"PROXY_PORT=18181",
		"PROXY_PORT=29999",
		"",
	}, "\n")
	if err := os.WriteFile(envPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}

	values := Values{
		ProxyHost:                   "127.0.0.1",
		ProxyPort:                   19191,
		ProxyReadTimeoutSeconds:     10,
		ProxyWriteTimeoutSeconds:    60,
		ProxyShutdownTimeoutSeconds: 5,
		ProxyChainLogPath:           ".data/test.log",
		ProxyChainLogBodies:         true,
		ProxyChainLogMaxBodyBytes:   123,
	}
	if err := Save(root, values); err != nil {
		t.Fatalf("save settings: %v", err)
	}

	data, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("read env: %v", err)
	}
	text := string(data)
	for _, needle := range []string{
		"# keep this comment",
		"CUSTOM_FLAG=enabled",
		"PROXY_API_KEYS=alpha,beta",
		"PROXY_PORT=19191",
		"PROXY_CHAIN_LOG_MAX_BODY_BYTES=123",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("expected env file to preserve %q, got %s", needle, text)
		}
	}
	if strings.Count(text, "PROXY_PORT=") != 1 {
		t.Fatalf("expected managed key duplicates to be collapsed, got %s", text)
	}
}

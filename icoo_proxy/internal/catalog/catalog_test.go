package catalog

import (
	"testing"

	"icoo_proxy/internal/config"
)

func TestResolveUsesAliasAndDefaults(t *testing.T) {
	cfg := config.Config{
		DefaultAnthropicRoute: "anthropic:claude-real",
		DefaultChatRoute:      "openai-chat:gpt-chat-real",
		ModelRoutes:           "assistant-default=openai-responses:gpt-response-real,claude-sonnet=anthropic:claude-real",
	}
	cat, err := New(cfg)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}

	route, err := cat.Resolve(ProtocolAnthropic, "")
	if err != nil {
		t.Fatalf("resolve default: %v", err)
	}
	if route.Upstream != ProtocolAnthropic || route.Model != "claude-real" {
		t.Fatalf("unexpected default route: %+v", route)
	}

	route, err = cat.Resolve(ProtocolOpenAIChat, "assistant-default")
	if err != nil {
		t.Fatalf("resolve alias: %v", err)
	}
	if route.Upstream != ProtocolOpenAIResponse || route.Model != "gpt-response-real" {
		t.Fatalf("unexpected alias route: %+v", route)
	}

	route, err = cat.Resolve(ProtocolOpenAIChat, "gpt-4.1")
	if err != nil {
		t.Fatalf("resolve passthrough: %v", err)
	}
	if route.Upstream != ProtocolOpenAIChat || route.Model != "gpt-4.1" {
		t.Fatalf("unexpected passthrough route: %+v", route)
	}
}

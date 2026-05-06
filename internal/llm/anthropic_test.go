package llm

import (
	"context"
	"testing"

	"icoo_assistant/internal/config"

	anthropic "github.com/anthropics/anthropic-sdk-go"
)

func TestNewAnthropicClientKeepsBaseURLAndDefaults(t *testing.T) {
	client := NewAnthropicClient(AnthropicConfig{
		BaseURL: " https://anthropic-proxy.example.com ",
	})
	if client.config.BaseURL != "https://anthropic-proxy.example.com" {
		t.Fatalf("expected trimmed base url, got %q", client.config.BaseURL)
	}
	if client.config.Model != "claude-opus-4-7" {
		t.Fatalf("expected default model, got %q", client.config.Model)
	}
	if client.config.MaxTokens != 16000 {
		t.Fatalf("expected default max tokens, got %d", client.config.MaxTokens)
	}
}

func TestAnthropicClientUsesNonNilRequestContext(t *testing.T) {
	ctx := context.Background()
	if ctx == nil {
		t.Fatal("expected non-nil background context")
	}
}

func TestNewClientFromConfigPassesAnthropicBaseURL(t *testing.T) {
	client, mode, err := NewClientFromConfig(config.Config{
		AnthropicAPIKey:  "test-key",
		AnthropicBaseURL: "https://anthropic-proxy.example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if mode != "anthropic" {
		t.Fatalf("expected anthropic mode, got %q", mode)
	}
	anthropicClient, ok := client.(*AnthropicClient)
	if !ok {
		t.Fatalf("expected anthropic client, got %T", client)
	}
	if anthropicClient.config.BaseURL != "https://anthropic-proxy.example.com" {
		t.Fatalf("expected base url to be passed through, got %q", anthropicClient.config.BaseURL)
	}
}

func TestParseAnthropicBlockExtractsThinkingText(t *testing.T) {
	text, thinking, toolUse, err := parseAnthropicBlock(anthropic.ThinkingBlock{Thinking: "compat answer"})
	if err != nil {
		t.Fatal(err)
	}
	if text != "" {
		t.Fatalf("expected empty text, got %q", text)
	}
	if thinking != "compat answer" {
		t.Fatalf("expected thinking text, got %q", thinking)
	}
	if toolUse != nil {
		t.Fatalf("expected nil tool use, got %#v", toolUse)
	}
}

func TestShouldUseThinkingFallbackOnlyForCustomBaseURL(t *testing.T) {
	if shouldUseThinkingFallback(AnthropicConfig{}) {
		t.Fatal("expected fallback disabled without base url")
	}
	if shouldUseThinkingFallback(AnthropicConfig{BaseURL: "https://api.anthropic.com"}) {
		t.Fatal("expected fallback disabled for official anthropic base url")
	}
	if !shouldUseThinkingFallback(AnthropicConfig{BaseURL: "https://yybb.codes"}) {
		t.Fatal("expected fallback enabled for custom base url")
	}
}

func TestBuildMessagesSupportsAssistantStringHistory(t *testing.T) {
	client := NewAnthropicClient(AnthropicConfig{})
	messages, err := client.buildMessages([]Message{
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "hi there"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}
}

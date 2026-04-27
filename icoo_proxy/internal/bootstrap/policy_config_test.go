package bootstrap

import (
	"strings"
	"testing"

	"icoo_proxy/internal/config"
	"icoo_proxy/internal/routepolicy"
	"icoo_proxy/internal/supplier"
)

func TestApplyRoutePoliciesAssignsChatSupplierConfig(t *testing.T) {
	root := t.TempDir()
	suppliers, err := supplier.NewService(root)
	if err != nil {
		t.Fatalf("new suppliers: %v", err)
	}
	t.Cleanup(func() { _ = suppliers.Close() })
	record, err := suppliers.Upsert(supplier.UpsertInput{
		Name:         "Test OpenAI Chat",
		Protocol:     "openai-chat",
		BaseURL:      "https://chat.example.com",
		APIKey:       "chat-secret",
		OnlyStream:   true,
		UserAgent:    "ChatPolicyUA/1.0",
		Enabled:      true,
		Models:       "gpt-4.1-mini",
		DefaultModel: "gpt-4.1-mini",
	})
	if err != nil {
		t.Fatalf("upsert supplier: %v", err)
	}
	policies, err := routepolicy.NewService(root, suppliers)
	if err != nil {
		t.Fatalf("new policies: %v", err)
	}
	t.Cleanup(func() { _ = policies.Close() })
	if _, err := policies.Upsert(routepolicy.UpsertInput{
		DownstreamProtocol: "openai-chat",
		SupplierID:         record.ID,
		Enabled:            true,
	}); err != nil {
		t.Fatalf("upsert policy: %v", err)
	}
	cfg, err := ApplyRoutePolicies(config.Config{}, suppliers, policies)
	if err != nil {
		t.Fatalf("apply policies: %v", err)
	}
	if cfg.DefaultChatRoute != "openai-chat:gpt-4.1-mini" {
		t.Fatalf("unexpected default chat route: %q", cfg.DefaultChatRoute)
	}
	if cfg.OpenAIChatBaseURL != "https://chat.example.com" {
		t.Fatalf("unexpected openai chat base url: %q", cfg.OpenAIChatBaseURL)
	}
	if cfg.OpenAIChatAPIKey != "chat-secret" {
		t.Fatalf("unexpected openai chat api key")
	}
	if !cfg.OpenAIChatOnlyStream {
		t.Fatalf("expected openai chat only_stream from supplier policy")
	}
	if cfg.OpenAIChatUserAgent != "ChatPolicyUA/1.0" {
		t.Fatalf("expected openai chat user agent from supplier policy, got %q", cfg.OpenAIChatUserAgent)
	}
	if cfg.OpenAIResponsesBaseURL != "" || cfg.OpenAIResponsesAPIKey != "" {
		t.Fatalf("expected responses config to remain empty, got base=%q key=%q", cfg.OpenAIResponsesBaseURL, cfg.OpenAIResponsesAPIKey)
	}
}

func TestApplyRoutePoliciesAssignsResponsesSupplierConfig(t *testing.T) {
	root := t.TempDir()
	suppliers, err := supplier.NewService(root)
	if err != nil {
		t.Fatalf("new suppliers: %v", err)
	}
	t.Cleanup(func() { _ = suppliers.Close() })
	record, err := suppliers.Upsert(supplier.UpsertInput{
		Name:         "Test OpenAI Responses",
		Protocol:     "openai-responses",
		BaseURL:      "https://responses.example.com",
		APIKey:       "responses-secret",
		OnlyStream:   true,
		UserAgent:    "ResponsesPolicyUA/1.0",
		Enabled:      true,
		Models:       "gpt-4.1-mini",
		DefaultModel: "gpt-4.1-mini",
	})
	if err != nil {
		t.Fatalf("upsert supplier: %v", err)
	}
	policies, err := routepolicy.NewService(root, suppliers)
	if err != nil {
		t.Fatalf("new policies: %v", err)
	}
	t.Cleanup(func() { _ = policies.Close() })
	if _, err := policies.Upsert(routepolicy.UpsertInput{
		DownstreamProtocol: "openai-chat",
		SupplierID:         record.ID,
		Enabled:            true,
	}); err != nil {
		t.Fatalf("upsert policy: %v", err)
	}
	cfg, err := ApplyRoutePolicies(config.Config{}, suppliers, policies)
	if err != nil {
		t.Fatalf("apply policies: %v", err)
	}
	if cfg.DefaultChatRoute != "openai-responses:gpt-4.1-mini" {
		t.Fatalf("unexpected default chat route: %q", cfg.DefaultChatRoute)
	}
	if cfg.OpenAIResponsesBaseURL != "https://responses.example.com" {
		t.Fatalf("unexpected openai responses base url: %q", cfg.OpenAIResponsesBaseURL)
	}
	if cfg.OpenAIResponsesAPIKey != "responses-secret" {
		t.Fatalf("unexpected openai responses api key")
	}
	if !cfg.OpenAIResponsesOnlyStream {
		t.Fatalf("expected openai responses only_stream from supplier policy")
	}
	if cfg.OpenAIResponsesUserAgent != "ResponsesPolicyUA/1.0" {
		t.Fatalf("expected openai responses user agent from supplier policy, got %q", cfg.OpenAIResponsesUserAgent)
	}
	if cfg.OpenAIChatBaseURL != "" || cfg.OpenAIChatAPIKey != "" {
		t.Fatalf("expected chat config to remain empty, got base=%q key=%q", cfg.OpenAIChatBaseURL, cfg.OpenAIChatAPIKey)
	}
}

func TestApplyRoutePoliciesRejectsSupplierWithoutDefaultModel(t *testing.T) {
	root := t.TempDir()
	suppliers, err := supplier.NewService(root)
	if err != nil {
		t.Fatalf("new suppliers: %v", err)
	}
	t.Cleanup(func() { _ = suppliers.Close() })
	record, err := suppliers.Upsert(supplier.UpsertInput{
		Name:      "Missing Default Model",
		Protocol:  "openai-responses",
		BaseURL:   "https://responses.example.com",
		APIKey:    "responses-secret",
		Enabled:   true,
		Models:    "gpt-4.1-mini",
		UserAgent: "ResponsesPolicyUA/1.0",
	})
	if err != nil {
		t.Fatalf("upsert supplier: %v", err)
	}
	policies, err := routepolicy.NewService(root, suppliers)
	if err != nil {
		t.Fatalf("new policies: %v", err)
	}
	t.Cleanup(func() { _ = policies.Close() })
	if _, err := policies.Upsert(routepolicy.UpsertInput{
		DownstreamProtocol: "openai-chat",
		SupplierID:         record.ID,
		Enabled:            true,
	}); err != nil {
		t.Fatalf("upsert policy: %v", err)
	}
	_, err = ApplyRoutePolicies(config.Config{}, suppliers, policies)
	if err == nil {
		t.Fatalf("expected missing default model error")
	}
	if !strings.Contains(err.Error(), "default model is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

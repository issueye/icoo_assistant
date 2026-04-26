package bootstrap

import (
	"testing"

	"icoo_proxy/internal/config"
	"icoo_proxy/internal/routepolicy"
	"icoo_proxy/internal/supplier"
)

func TestApplyRoutePolicies(t *testing.T) {
	root := t.TempDir()
	suppliers, err := supplier.NewService(root)
	if err != nil {
		t.Fatalf("new suppliers: %v", err)
	}
	t.Cleanup(func() { _ = suppliers.Close() })
	record, err := suppliers.Upsert(supplier.UpsertInput{
		Name:       "Test OpenAI",
		Protocol:   "openai-responses",
		BaseURL:    "https://example.com",
		APIKey:     "secret-key",
		OnlyStream: true,
		UserAgent:  "PolicyUA/1.0",
		Enabled:    true,
		Models:     "gpt-4.1-mini",
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
		TargetModel:        "gpt-4.1-mini",
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
	if cfg.OpenAIBaseURL != "https://example.com" {
		t.Fatalf("unexpected openai base url: %q", cfg.OpenAIBaseURL)
	}
	if cfg.OpenAIApiKey != "secret-key" {
		t.Fatalf("unexpected openai api key")
	}
	if !cfg.OpenAIOnlyStream {
		t.Fatalf("expected openai only_stream from supplier policy")
	}
	if cfg.OpenAIUserAgent != "PolicyUA/1.0" {
		t.Fatalf("expected openai user agent from supplier policy, got %q", cfg.OpenAIUserAgent)
	}
}

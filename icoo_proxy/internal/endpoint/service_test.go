package endpoint

import "testing"

func TestServiceSeedsAndUpsertsEndpoints(t *testing.T) {
	svc, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	t.Cleanup(func() { _ = svc.Close() })
	items := svc.List()
	if len(items) < 6 {
		t.Fatalf("expected seeded endpoints, got %d", len(items))
	}
	if _, err := svc.Upsert(UpsertInput{
		Path:        "custom/v1/chat",
		Protocol:    "openai-chat",
		Description: "custom endpoint",
		Enabled:     true,
	}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	found := false
	for _, item := range svc.Enabled() {
		if item.Path == "/custom/v1/chat" && item.Protocol == "openai-chat" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected normalized custom endpoint in enabled list")
	}
}

func TestDeleteRejectsBuiltInEndpoint(t *testing.T) {
	svc, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	t.Cleanup(func() { _ = svc.Close() })
	var builtInID string
	for _, item := range svc.List() {
		if item.BuiltIn {
			builtInID = item.ID
			break
		}
	}
	if builtInID == "" {
		t.Fatalf("expected built-in endpoint")
	}
	if err := svc.Delete(builtInID); err == nil {
		t.Fatalf("expected built-in endpoint delete to fail")
	}
}

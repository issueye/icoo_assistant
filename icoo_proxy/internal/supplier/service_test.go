package supplier

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpsertListDelete(t *testing.T) {
	svc, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	t.Cleanup(func() { _ = svc.Close() })
	initial := svc.List()
	if len(initial) == 0 {
		t.Fatalf("expected seeded suppliers")
	}

	record, err := svc.Upsert(UpsertInput{
		Name:         "Test Vendor",
		Protocol:     "openai-chat",
		BaseURL:      "https://example.com",
		APIKey:       "secret-key-123456",
		OnlyStream:   true,
		UserAgent:    "CustomVendor/1.0",
		Enabled:      true,
		Description:  "Test vendor",
		Models:       "model-a,model-b",
		DefaultModel: "model-a",
	})
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if record.APIKeyMasked == "" {
		t.Fatalf("expected masked key")
	}
	if !record.OnlyStream {
		t.Fatalf("expected only_stream to round-trip")
	}
	if record.UserAgent != "CustomVendor/1.0" {
		t.Fatalf("expected user agent to round-trip, got %q", record.UserAgent)
	}
	if record.DefaultModel != "model-a" {
		t.Fatalf("expected default model to round-trip, got %q", record.DefaultModel)
	}

	items := svc.List()
	if len(items) != len(initial)+1 {
		t.Fatalf("expected one more supplier, got %d", len(items))
	}
	resolved, ok := svc.Resolve(record.ID)
	if !ok {
		t.Fatalf("expected supplier to resolve")
	}
	if !resolved.OnlyStream {
		t.Fatalf("expected resolved supplier snapshot to preserve only_stream")
	}
	if resolved.UserAgent != "CustomVendor/1.0" {
		t.Fatalf("expected resolved supplier snapshot to preserve user agent, got %q", resolved.UserAgent)
	}
	if resolved.DefaultModel != "model-a" {
		t.Fatalf("expected resolved supplier snapshot to preserve default model, got %q", resolved.DefaultModel)
	}

	if err := svc.Delete(record.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	items = svc.List()
	if len(items) != len(initial) {
		t.Fatalf("expected supplier count restored, got %d", len(items))
	}
}

func TestUpsertRejectsDefaultModelOutsideModels(t *testing.T) {
	svc, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	t.Cleanup(func() { _ = svc.Close() })

	_, err = svc.Upsert(UpsertInput{
		Name:         "Invalid Vendor",
		Protocol:     "openai-chat",
		BaseURL:      "https://example.com",
		Enabled:      true,
		Models:       "model-a,model-b",
		DefaultModel: "model-c",
	})
	if err == nil {
		t.Fatalf("expected invalid default model error")
	}
}

func TestHealthCheck(t *testing.T) {
	var gotUserAgent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserAgent = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	t.Cleanup(func() { _ = svc.Close() })
	record, err := svc.Upsert(UpsertInput{
		Name:      "Health Vendor",
		Protocol:  "openai-responses",
		BaseURL:   server.URL,
		UserAgent: "HealthCheckUA/1.0",
		Enabled:   true,
	})
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}

	health := NewHealthService(svc)
	result, err := health.Check(record.ID)
	if err != nil {
		t.Fatalf("health check: %v", err)
	}
	if !result.Reachable {
		t.Fatalf("expected supplier to be reachable")
	}
	if result.Status != "reachable" {
		t.Fatalf("expected reachable status, got %q", result.Status)
	}
	if gotUserAgent != "HealthCheckUA/1.0" {
		t.Fatalf("expected health check user agent, got %q", gotUserAgent)
	}
}

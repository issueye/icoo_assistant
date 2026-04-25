package supplier

import "testing"

func TestUpsertListDelete(t *testing.T) {
	svc, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	initial := svc.List()
	if len(initial) == 0 {
		t.Fatalf("expected seeded suppliers")
	}

	record, err := svc.Upsert(UpsertInput{
		Name:        "Test Vendor",
		Protocol:    "openai-chat",
		BaseURL:     "https://example.com",
		APIKey:      "secret-key-123456",
		Enabled:     true,
		Description: "Test vendor",
		Models:      "model-a,model-b",
		Tags:        "internal,test",
	})
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if record.APIKeyMasked == "" {
		t.Fatalf("expected masked key")
	}

	items := svc.List()
	if len(items) != len(initial)+1 {
		t.Fatalf("expected one more supplier, got %d", len(items))
	}

	if err := svc.Delete(record.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	items = svc.List()
	if len(items) != len(initial) {
		t.Fatalf("expected supplier count restored, got %d", len(items))
	}
}

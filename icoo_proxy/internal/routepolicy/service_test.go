package routepolicy

import "testing"

type fakeResolver struct {
	items map[string]SupplierSnapshot
}

func (f fakeResolver) Resolve(id string) (SupplierSnapshot, bool) {
	item, ok := f.items[id]
	return item, ok
}

func TestUpsertAndList(t *testing.T) {
	svc, err := NewService(t.TempDir(), fakeResolver{
		items: map[string]SupplierSnapshot{
			"openai-default": {
				ID:       "openai-default",
				Name:     "OpenAI Default",
				Protocol: "openai-responses",
			},
		},
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	t.Cleanup(func() { _ = svc.Close() })
	items := svc.List()
	if len(items) != 3 {
		t.Fatalf("expected seeded policies, got %d", len(items))
	}
	record, err := svc.Upsert(UpsertInput{
		DownstreamProtocol: "openai-chat",
		SupplierID:         "openai-default",
		TargetModel:        "gpt-4.1-mini",
		Enabled:            true,
	})
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if record.UpstreamProtocol != "openai-responses" {
		t.Fatalf("expected upstream protocol from supplier, got %q", record.UpstreamProtocol)
	}
	enabled := svc.Enabled()
	if len(enabled) != 1 {
		t.Fatalf("expected one enabled policy, got %d", len(enabled))
	}
}

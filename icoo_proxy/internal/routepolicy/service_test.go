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
				ID:           "openai-default",
				Name:         "OpenAI Default",
				Protocol:     "openai-responses",
				DefaultModel: "gpt-4.1-mini",
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

func TestFindEnabledBySupplierID(t *testing.T) {
	svc, err := NewService(t.TempDir(), fakeResolver{
		items: map[string]SupplierSnapshot{
			"openai-default": {
				ID:           "openai-default",
				Name:         "OpenAI Default",
				Protocol:     "openai-responses",
				DefaultModel: "gpt-4.1-mini",
			},
		},
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	t.Cleanup(func() { _ = svc.Close() })
	if _, ok := svc.FindEnabledBySupplierID("openai-default"); ok {
		t.Fatalf("expected disabled seeded policies to be ignored")
	}
	if _, err := svc.Upsert(UpsertInput{
		DownstreamProtocol: "openai-chat",
		SupplierID:         "openai-default",
		Enabled:            true,
	}); err != nil {
		t.Fatalf("upsert enabled policy: %v", err)
	}
	record, ok := svc.FindEnabledBySupplierID("openai-default")
	if !ok {
		t.Fatalf("expected enabled policy to be found")
	}
	if record.DownstreamProtocol != "openai-chat" {
		t.Fatalf("expected downstream protocol openai-chat, got %q", record.DownstreamProtocol)
	}
}

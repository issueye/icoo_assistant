package modelalias

import (
	"reflect"
	"testing"

	"icoo_proxy/internal/consts"
)

func TestServiceUpsertListsAndDeletesAliases(t *testing.T) {
	svc, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	t.Cleanup(func() { _ = svc.Close() })

	record, err := svc.Upsert(UpsertInput{
		Name:             "fast-model",
		UpstreamProtocol: consts.ProtocolOpenAIResponses,
		Model:            "gpt-4.1-mini",
		Enabled:          true,
	})
	if err != nil {
		t.Fatalf("upsert alias: %v", err)
	}
	if record.Name != "fast-model" {
		t.Fatalf("unexpected alias name: %q", record.Name)
	}

	items := svc.List()
	if len(items) != 1 {
		t.Fatalf("expected 1 alias, got %d", len(items))
	}
	entries := svc.EnabledEntries()
	if !reflect.DeepEqual(entries, []string{"fast-model=openai-responses:gpt-4.1-mini"}) {
		t.Fatalf("unexpected enabled entries: %#v", entries)
	}

	updated, err := svc.Upsert(UpsertInput{
		ID:               record.ID,
		Name:             "fast-model",
		UpstreamProtocol: consts.ProtocolAnthropic,
		Model:            "claude-sonnet-4-20250514",
		Enabled:          false,
	})
	if err != nil {
		t.Fatalf("update alias: %v", err)
	}
	if updated.ID != record.ID {
		t.Fatalf("expected stable alias id, got %q want %q", updated.ID, record.ID)
	}
	if got := svc.EnabledEntries(); len(got) != 0 {
		t.Fatalf("expected disabled alias to be excluded, got %#v", got)
	}

	if err := svc.Delete(record.ID); err != nil {
		t.Fatalf("delete alias: %v", err)
	}
	if got := svc.List(); len(got) != 0 {
		t.Fatalf("expected empty alias list, got %#v", got)
	}
}

func TestMergeEntriesOverridesByAliasName(t *testing.T) {
	got := MergeEntries(
		"fast-model=openai-chat:gpt-4.1-mini, smart-model=openai-responses:gpt-4.1",
		[]string{
			"fast-model=openai-responses:gpt-4.1-mini",
			"new-model=anthropic:claude-sonnet-4-20250514",
		},
	)
	want := "fast-model=openai-responses:gpt-4.1-mini,smart-model=openai-responses:gpt-4.1,new-model=anthropic:claude-sonnet-4-20250514"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

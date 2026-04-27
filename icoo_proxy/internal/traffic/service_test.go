package traffic

import (
	"testing"
	"time"

	"icoo_proxy/internal/api"
)

func TestServiceRecordsAndListsRecentRequestsNewestFirst(t *testing.T) {
	root := t.TempDir()
	service, err := NewService(root)
	if err != nil {
		t.Fatalf("new traffic service: %v", err)
	}
	defer func() { _ = service.Close() }()

	first := api.RequestView{
		RequestID:  "req-first",
		Downstream: "openai-chat",
		Upstream:   "openai-responses",
		Model:      "gpt-4.1-mini",
		StatusCode: 200,
		DurationMS: 42,
		CreatedAt:  time.Date(2026, 4, 27, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
	}
	second := api.RequestView{
		RequestID:  "req-second",
		Downstream: "anthropic",
		Upstream:   "openai-responses",
		Model:      "claude-sonnet",
		StatusCode: 502,
		DurationMS: 110,
		Error:      "upstream failed",
		CreatedAt:  time.Date(2026, 4, 27, 10, 1, 0, 0, time.UTC).Format(time.RFC3339),
	}

	if err := service.RecordRequest(first); err != nil {
		t.Fatalf("record first: %v", err)
	}
	if err := service.RecordRequest(second); err != nil {
		t.Fatalf("record second: %v", err)
	}

	got := service.ListRecent(10)
	if len(got) != 2 {
		t.Fatalf("expected 2 records, got %d", len(got))
	}
	if got[0].RequestID != "req-second" || got[1].RequestID != "req-first" {
		t.Fatalf("expected newest first, got %#v", got)
	}
	if got[0].Error != "upstream failed" {
		t.Fatalf("expected error field preserved, got %#v", got[0])
	}

	if err := service.Close(); err != nil {
		t.Fatalf("close traffic service: %v", err)
	}
	reopened, err := NewService(root)
	if err != nil {
		t.Fatalf("reopen traffic service: %v", err)
	}
	defer func() { _ = reopened.Close() }()

	got = reopened.ListRecent(10)
	if len(got) != 2 {
		t.Fatalf("expected 2 persisted records, got %d", len(got))
	}
	if got[0].RequestID != "req-second" || got[1].RequestID != "req-first" {
		t.Fatalf("expected persisted records newest first, got %#v", got)
	}
}

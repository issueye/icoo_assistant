package hookaudit_test

import (
	"path/filepath"
	"testing"
	"time"

	"icoo_assistant/internal/agent"
	"icoo_assistant/internal/hookaudit"
)

func TestReaderRecentLimitAndFilter(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".agent-hooks")
	hook, err := agent.NewJSONLHook(dir)
	if err != nil {
		t.Fatal(err)
	}
	for i, event := range []agent.Event{
		{
			Timestamp: time.Unix(1700000000, 0).UTC(),
			Name:      "agent.run.started",
			RunID:     "run-1",
		},
		{
			Timestamp: time.Unix(1700000001, 0).UTC(),
			Name:      "agent.tool.completed",
			RunID:     "run-1",
			Round:     1,
		},
		{
			Timestamp: time.Unix(1700000002, 0).UTC(),
			Name:      "agent.tool.completed",
			RunID:     "run-2",
			Round:     1,
		},
	} {
		event.Fields = map[string]interface{}{"index": i}
		hook.OnEvent(event)
	}
	reader := hookaudit.NewReader(dir)
	events, err := reader.Recent(hookaudit.Query{
		Limit: 1,
		Name:  "agent.tool.completed",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected one filtered event, got %d", len(events))
	}
	if events[0].RunID != "run-2" {
		t.Fatalf("expected most recent matching event, got %#v", events[0])
	}
}

func TestReaderRecentMissingFile(t *testing.T) {
	reader := hookaudit.NewReader(filepath.Join(t.TempDir(), ".agent-hooks"))
	events, err := reader.Recent(hookaudit.Query{Limit: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 {
		t.Fatalf("expected no events, got %d", len(events))
	}
}

package tools_test

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"icoo_assistant/internal/agent"
	"icoo_assistant/internal/hookaudit"
	"icoo_assistant/internal/tools"
)

func TestAgentHookAuditToolRecent(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".agent-hooks")
	hook, err := agent.NewJSONLHook(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range []agent.Event{
		{
			Timestamp: time.Unix(1700000000, 0).UTC(),
			Name:      "agent.run.started",
			RunID:     "run-1",
			Fields:    map[string]interface{}{"message_count": 1},
		},
		{
			Timestamp: time.Unix(1700000001, 0).UTC(),
			Name:      "agent.tool.completed",
			RunID:     "run-1",
			Round:     1,
			Fields:    map[string]interface{}{"tool_name": "bash", "result_length": 5},
		},
		{
			Timestamp: time.Unix(1700000002, 0).UTC(),
			Name:      "agent.tool.completed",
			RunID:     "run-2",
			Round:     2,
			Fields:    map[string]interface{}{"tool_name": "todo", "result_length": 2},
		},
	} {
		hook.OnEvent(event)
	}
	tool := tools.NewAgentHookAuditTool(hookaudit.NewReader(dir))
	result, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "recent",
		"limit":  float64(1),
		"name":   "agent.tool.completed",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "returned_count: 1") {
		t.Fatalf("unexpected count in result: %q", result)
	}
	if !strings.Contains(result, "filter_name: agent.tool.completed") {
		t.Fatalf("expected name filter in result: %q", result)
	}
	if !strings.Contains(result, "run_id=run-2") || !strings.Contains(result, "tool_name=todo") {
		t.Fatalf("expected most recent matching hook event, got %q", result)
	}
	if strings.Contains(result, "run_id=run-1") {
		t.Fatalf("expected limited result, got %q", result)
	}
}

func TestAgentHookAuditToolRecentEmpty(t *testing.T) {
	tool := tools.NewAgentHookAuditTool(hookaudit.NewReader(filepath.Join(t.TempDir(), ".agent-hooks")))
	result, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "recent",
		"run_id": "missing-run",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "entries: none") {
		t.Fatalf("unexpected empty result: %q", result)
	}
}

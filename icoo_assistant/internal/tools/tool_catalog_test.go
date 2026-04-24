package tools_test

import (
	"strings"
	"testing"

	"icoo_assistant/internal/tools"
)

func TestToolCatalogList(t *testing.T) {
	def := tools.NewToolCatalogTool(tools.DefaultToolCatalogEntries(true))
	result, err := def.Handler(tools.Call{
		Name:  "tool_catalog",
		Input: map[string]interface{}{"action": "list"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "available_tools:") {
		t.Fatalf("expected tool count, got %q", result)
	}
	if !strings.Contains(result, "- project_task:") {
		t.Fatalf("expected project_task in list, got %q", result)
	}
	if !strings.Contains(result, "- task:") {
		t.Fatalf("expected task in list, got %q", result)
	}
}

func TestToolCatalogDescribeHighlightsBoundary(t *testing.T) {
	def := tools.NewToolCatalogTool(tools.DefaultToolCatalogEntries(true))
	result, err := def.Handler(tools.Call{
		Name: "tool_catalog",
		Input: map[string]interface{}{
			"action": "describe",
			"name":   "project_task",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "avoid_when: Avoid for audit-style history review or subagent delegation; use task_audit or task instead.") {
		t.Fatalf("expected project_task boundary guidance, got %q", result)
	}
}

func TestToolCatalogDescribeTaskAuditMentionsStatusFiltering(t *testing.T) {
	def := tools.NewToolCatalogTool(tools.DefaultToolCatalogEntries(true))
	result, err := def.Handler(tools.Call{
		Name: "tool_catalog",
		Input: map[string]interface{}{
			"action": "describe",
			"name":   "task_audit",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `"action":"summary","id":"task-1","status":"failed"`) {
		t.Fatalf("expected failure-summary example, got %q", result)
	}
}

func TestToolCatalogAuditPaths(t *testing.T) {
	def := tools.NewToolCatalogTool(tools.DefaultToolCatalogEntries(true))
	result, err := def.Handler(tools.Call{
		Name: "tool_catalog",
		Input: map[string]interface{}{
			"action": "audit_paths",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "task_first:") {
		t.Fatalf("expected task-first audit flow, got %q", result)
	}
	if !strings.Contains(result, "latest sample for each failure reason") {
		t.Fatalf("expected task audit comparison path, got %q", result)
	}
	if !strings.Contains(result, "agent_hook_audit action=recent") {
		t.Fatalf("expected agent hook audit path, got %q", result)
	}
}

func TestDefaultToolCatalogEntriesCanExcludeTask(t *testing.T) {
	entries := tools.DefaultToolCatalogEntries(false)
	for _, entry := range entries {
		if entry.Name == "task" {
			t.Fatalf("task should be excluded from base catalog")
		}
	}
}

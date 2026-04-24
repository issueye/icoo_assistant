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
	if !strings.Contains(result, `"action":"summary","id":"task-1"`) {
		t.Fatalf("expected summary example, got %q", result)
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
	if !strings.Contains(result, "recent failure trend") {
		t.Fatalf("expected task audit trend path, got %q", result)
	}
	if !strings.Contains(result, "priority failure hints with basis, recent context, repeat-pattern hints, sample-target guidance, recent sample comparison hints, direct latest-vs-previous compare targets, focused change-point hints, recent failure trend, and lightweight stability-vs-change trend hints") {
		t.Fatalf("expected priority hint path, got %q", result)
	}
	if !strings.Contains(result, "role=previous/latest") {
		t.Fatalf("expected history role guidance, got %q", result)
	}
	if !strings.Contains(result, "reason labels") {
		t.Fatalf("expected history reason-label guidance, got %q", result)
	}
	if !strings.Contains(result, "latest_sample") {
		t.Fatalf("expected history latest-sample guidance, got %q", result)
	}
	if !strings.Contains(result, "pair_summary") {
		t.Fatalf("expected history pair summary guidance, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_basis") {
		t.Fatalf("expected priority basis flow, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_context") {
		t.Fatalf("expected priority context flow, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_pattern_hint") {
		t.Fatalf("expected priority pattern flow, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_sample_target") {
		t.Fatalf("expected priority sample-target flow, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_sample_compare") {
		t.Fatalf("expected priority sample-compare flow, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_compare_target") {
		t.Fatalf("expected priority compare-target flow, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_change_hint") {
		t.Fatalf("expected priority change flow, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_trend_hint") {
		t.Fatalf("expected priority trend flow, got %q", result)
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

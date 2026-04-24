package tools_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"icoo_assistant/internal/task"
	"icoo_assistant/internal/tools"
)

func TestTaskAuditToolHistory(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(task.CreateInput{
		ID:    "task-a",
		Title: "Inspect audit history",
	}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if _, err := manager.RecordBackground("task-a", task.BackgroundContext{
			JobID:   fmt.Sprintf("job-%d", i),
			Status:  "completed",
			Command: fmt.Sprintf("cmd-%d", i),
		}); err != nil {
			t.Fatal(err)
		}
	}
	tool := tools.NewTaskAuditTool(manager)
	result, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "history",
		"id":     "task-a",
		"limit":  float64(2),
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "task_id: task-a") {
		t.Fatalf("unexpected audit result: %q", result)
	}
	if !strings.Contains(result, "history_count: 3") || !strings.Contains(result, "returned_count: 2") {
		t.Fatalf("unexpected history counters: %q", result)
	}
	if !strings.Contains(result, "filtered_count: 3") {
		t.Fatalf("expected filtered count, got %q", result)
	}
	if !strings.Contains(result, "job_id=job-1") || !strings.Contains(result, "job_id=job-2") {
		t.Fatalf("expected recent entries, got %q", result)
	}
	if strings.Contains(result, "job_id=job-0") {
		t.Fatalf("expected limited audit history, got %q", result)
	}
	if !strings.Contains(result, "latest_task_view: project_task action=get id=task-a") {
		t.Fatalf("expected latest task view hint, got %q", result)
	}
	if !strings.Contains(result, "runtime_view_hint: use agent_hook_audit action=summary or action=recent") {
		t.Fatalf("expected runtime-side hint, got %q", result)
	}
}

func TestTaskAuditToolSummary(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(task.CreateInput{
		ID:    "task-a",
		Title: "Summarize failures",
	}); err != nil {
		t.Fatal(err)
	}
	for _, entry := range []task.BackgroundContext{
		{JobID: "job-1", Status: "completed", Command: "cmd-1"},
		{JobID: "job-2", Status: "failed", Command: "cmd-2", Error: "boom"},
		{JobID: "job-3", Status: "completed", Command: "cmd-3"},
		{JobID: "job-4", Status: "failed", Command: "cmd-4", Error: "timeout after 5s"},
	} {
		if _, err := manager.RecordBackground("task-a", entry); err != nil {
			t.Fatal(err)
		}
	}
	tool := tools.NewTaskAuditTool(manager)
	result, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "summary",
		"id":     "task-a",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "history_count: 4") || !strings.Contains(result, "filtered_count: 4") {
		t.Fatalf("expected summary counters, got %q", result)
	}
	if !strings.Contains(result, "- completed=2") || !strings.Contains(result, "- failed=2") {
		t.Fatalf("expected status counts, got %q", result)
	}
	if !strings.Contains(result, "- command_error=1") || !strings.Contains(result, "- timeout=1") {
		t.Fatalf("expected failure reason counts, got %q", result)
	}
	if !strings.Contains(result, "latest_failure_by_reason:") {
		t.Fatalf("expected per-reason failure samples, got %q", result)
	}
	if !strings.Contains(result, "- command_error => job_id=job-2 status=failed") || !strings.Contains(result, "- timeout => job_id=job-4 status=failed") {
		t.Fatalf("expected latest sample for each failure reason, got %q", result)
	}
	if !strings.Contains(result, "recent_failure_trend:") {
		t.Fatalf("expected failure trend section, got %q", result)
	}
	if !strings.Contains(result, "- reason=command_error job_id=job-2 status=failed") || !strings.Contains(result, "- reason=timeout job_id=job-4 status=failed") {
		t.Fatalf("expected recent failure trend lines, got %q", result)
	}
	if !strings.Contains(result, "latest_failure: job_id=job-4 status=failed") {
		t.Fatalf("expected latest failure summary, got %q", result)
	}
	if !strings.Contains(result, "latest_failure_reason: timeout") {
		t.Fatalf("expected latest failure reason, got %q", result)
	}
	if !strings.Contains(result, "failure_history_hint: use task_audit action=history id=task-a status=failed") {
		t.Fatalf("expected failure history hint, got %q", result)
	}
}

func TestTaskAuditToolHistoryEmpty(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(task.CreateInput{
		ID:    "task-a",
		Title: "No runs yet",
	}); err != nil {
		t.Fatal(err)
	}
	tool := tools.NewTaskAuditTool(manager)
	result, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "history",
		"id":     "task-a",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "entries: none") {
		t.Fatalf("unexpected empty audit result: %q", result)
	}
	if !strings.Contains(result, "latest_task_view: project_task action=get id=task-a") {
		t.Fatalf("expected latest task view hint, got %q", result)
	}
}

func TestTaskAuditToolHistoryCanFilterByStatus(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(task.CreateInput{
		ID:    "task-a",
		Title: "Inspect filtered audit history",
	}); err != nil {
		t.Fatal(err)
	}
	for _, entry := range []task.BackgroundContext{
		{JobID: "job-1", Status: "completed", Command: "cmd-1"},
		{JobID: "job-2", Status: "failed", Command: "cmd-2", Error: "boom"},
		{JobID: "job-3", Status: "completed", Command: "cmd-3"},
	} {
		if _, err := manager.RecordBackground("task-a", entry); err != nil {
			t.Fatal(err)
		}
	}
	tool := tools.NewTaskAuditTool(manager)
	result, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "history",
		"id":     "task-a",
		"status": "failed",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "filter_status: failed") {
		t.Fatalf("expected status filter, got %q", result)
	}
	if !strings.Contains(result, "filtered_count: 1") || !strings.Contains(result, "returned_count: 1") {
		t.Fatalf("expected filtered counters, got %q", result)
	}
	if !strings.Contains(result, "job_id=job-2 status=failed") {
		t.Fatalf("expected failed job in result, got %q", result)
	}
	if strings.Contains(result, "job_id=job-1") || strings.Contains(result, "job_id=job-3") {
		t.Fatalf("expected only failed jobs, got %q", result)
	}
}

func TestTaskAuditToolSummaryCanFilterByStatus(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(task.CreateInput{
		ID:    "task-a",
		Title: "Summarize filtered failures",
	}); err != nil {
		t.Fatal(err)
	}
	for _, entry := range []task.BackgroundContext{
		{JobID: "job-1", Status: "completed", Command: "cmd-1"},
		{JobID: "job-2", Status: "failed", Command: "cmd-2", Error: "boom"},
		{JobID: "job-3", Status: "failed", Command: "cmd-3", Error: "timeout after 5s"},
	} {
		if _, err := manager.RecordBackground("task-a", entry); err != nil {
			t.Fatal(err)
		}
	}
	tool := tools.NewTaskAuditTool(manager)
	result, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "summary",
		"id":     "task-a",
		"status": "failed",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "filter_status: failed") || !strings.Contains(result, "filtered_count: 2") {
		t.Fatalf("expected filtered summary counters, got %q", result)
	}
	if !strings.Contains(result, "- command_error=1") || !strings.Contains(result, "- timeout=1") {
		t.Fatalf("expected filtered failure reason counts, got %q", result)
	}
	if !strings.Contains(result, "- timeout => job_id=job-3 status=failed") {
		t.Fatalf("expected filtered latest sample per reason, got %q", result)
	}
	if !strings.Contains(result, "- reason=timeout job_id=job-3 status=failed") {
		t.Fatalf("expected filtered failure trend line, got %q", result)
	}
	if !strings.Contains(result, "matched_latest_entry: job_id=job-3 status=failed") {
		t.Fatalf("expected filtered latest entry, got %q", result)
	}
	if !strings.Contains(result, "latest_failure_reason: timeout") {
		t.Fatalf("expected latest failure reason, got %q", result)
	}
	if !strings.Contains(result, "filtered_history_hint: use task_audit action=history id=task-a status=failed") {
		t.Fatalf("expected filtered history hint, got %q", result)
	}
}

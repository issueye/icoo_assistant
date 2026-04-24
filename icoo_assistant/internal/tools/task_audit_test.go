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

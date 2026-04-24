package tools_test

import (
	"path/filepath"
	"strings"
	"testing"

	"icoo_assistant/internal/task"
	"icoo_assistant/internal/tools"
)

func TestProjectTaskToolCreateGetAndList(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	tool := tools.NewProjectTaskTool(manager)
	createResult, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "create",
		"id":     "task-a",
		"title":  "Build CLI task entrypoint",
		"owner":  "lead",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(createResult, "id: task-a") || !strings.Contains(createResult, "status: pending") {
		t.Fatalf("unexpected create result: %q", createResult)
	}
	getResult, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "get",
		"id":     "task-a",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(getResult, "title: Build CLI task entrypoint") {
		t.Fatalf("unexpected get result: %q", getResult)
	}
	listResult, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "list",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(listResult, "task-a [pending] Build CLI task entrypoint") {
		t.Fatalf("unexpected list result: %q", listResult)
	}
}

func TestProjectTaskToolUpdateAndUpdateStatus(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(task.CreateInput{
		ID:    "task-a",
		Title: "Initial title",
	}); err != nil {
		t.Fatal(err)
	}
	tool := tools.NewProjectTaskTool(manager)
	updateResult, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action":     "update",
		"id":         "task-a",
		"title":      "Updated title",
		"owner":      "alice",
		"blocked_by": []interface{}{"task-b"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(updateResult, "title: Updated title") || !strings.Contains(updateResult, "blocked_by: task-b") {
		t.Fatalf("unexpected update result: %q", updateResult)
	}
	statusResult, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "update_status",
		"id":     "task-a",
		"status": "completed",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(statusResult, "status: completed") {
		t.Fatalf("unexpected status result: %q", statusResult)
	}
}

func TestProjectTaskToolListEmpty(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	tool := tools.NewProjectTaskTool(manager)
	result, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "list",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if result != "No project tasks." {
		t.Fatalf("unexpected result: %q", result)
	}
}

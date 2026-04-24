package tools_test

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"icoo_assistant/internal/background"
	"icoo_assistant/internal/tools"
)

func TestBackgroundToolStartAndGet(t *testing.T) {
	manager, err := background.NewManager(filepath.Join(t.TempDir(), ".background"), t.TempDir(), 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	tool := tools.NewBackgroundTool(manager)
	command := "printf hello"
	if runtime.GOOS == "windows" {
		command = "Write-Output hello"
	}
	startResult, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action":  "start",
		"id":      "job-1",
		"command": command,
		"task_id": "task-a",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(startResult, "Started background job job-1") {
		t.Fatalf("unexpected start result: %q", startResult)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		getResult, err := tool.Handler(tools.Call{Input: map[string]interface{}{
			"action": "get",
			"id":     "job-1",
		}})
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(getResult, "status: completed") {
			if !strings.Contains(getResult, "hello") {
				t.Fatalf("expected output in get result, got %q", getResult)
			}
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("background job did not complete in time")
}

func TestBackgroundToolListEmpty(t *testing.T) {
	manager, err := background.NewManager(filepath.Join(t.TempDir(), ".background"), t.TempDir(), 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	tool := tools.NewBackgroundTool(manager)
	result, err := tool.Handler(tools.Call{Input: map[string]interface{}{"action": "list"}})
	if err != nil {
		t.Fatal(err)
	}
	if result != "No background jobs." {
		t.Fatalf("unexpected result: %q", result)
	}
}

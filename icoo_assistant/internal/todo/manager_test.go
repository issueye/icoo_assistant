package todo_test

import (
	"strings"
	"testing"

	"icoo_assistant/internal/todo"
)

func TestManagerUpdateRendersTodos(t *testing.T) {
	manager := todo.NewManager()
	output, err := manager.Update([]todo.Item{{ID: "1", Text: "Plan work", Status: "in_progress"}, {ID: "2", Text: "Run tests", Status: "pending"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "[>] #1: Plan work") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestManagerRejectsMultipleInProgress(t *testing.T) {
	manager := todo.NewManager()
	_, err := manager.Update([]todo.Item{{ID: "1", Text: "A", Status: "in_progress"}, {ID: "2", Text: "B", Status: "in_progress"}})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

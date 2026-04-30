package task_test

import (
	"path/filepath"
	"testing"
	"time"

	"icoo_assistant/internal/background"
	"icoo_assistant/internal/task"
)

func TestBackgroundLifecycleLinkBeforeStartMovesTaskInProgress(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(task.CreateInput{
		ID:    "task-a",
		Title: "Run tests",
	}); err != nil {
		t.Fatal(err)
	}
	link := task.NewBackgroundLifecycleLink(manager)
	err = link.BeforeStart(background.Job{
		ID:        "job-1",
		TaskID:    "task-a",
		Status:    background.StatusRunning,
		Command:   "go test ./...",
		StartedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	item, err := manager.Get("task-a")
	if err != nil {
		t.Fatal(err)
	}
	if item.Status != task.StatusInProgress {
		t.Fatalf("expected in_progress, got %q", item.Status)
	}
	if item.LastBackground == nil || item.LastBackground.JobID != "job-1" {
		t.Fatalf("expected running background context, got %#v", item.LastBackground)
	}
}

func TestBackgroundLifecycleLinkAfterFinishFailureReturnsTaskPending(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(task.CreateInput{
		ID:     "task-a",
		Title:  "Run tests",
		Status: task.StatusInProgress,
	}); err != nil {
		t.Fatal(err)
	}
	link := task.NewBackgroundLifecycleLink(manager)
	finished := time.Now().UTC()
	err = link.AfterFinish(background.Job{
		ID:         "job-1",
		TaskID:     "task-a",
		Status:     background.StatusFailed,
		Command:    "go test ./...",
		Error:      "exit status 1",
		FinishedAt: &finished,
	})
	if err != nil {
		t.Fatal(err)
	}
	item, err := manager.Get("task-a")
	if err != nil {
		t.Fatal(err)
	}
	if item.Status != task.StatusPending {
		t.Fatalf("expected pending after failed background job, got %q", item.Status)
	}
	if item.LastBackground == nil || item.LastBackground.Status != background.StatusFailed {
		t.Fatalf("expected failed background context, got %#v", item.LastBackground)
	}
	if item.LastBackground.Error != "exit status 1" {
		t.Fatalf("unexpected background error: %#v", item.LastBackground)
	}
}

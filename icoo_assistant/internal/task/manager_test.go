package task_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"icoo_assistant/internal/task"
)

func TestNewManagerCreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".tasks")
	manager, err := task.NewManager(dir)
	if err != nil {
		t.Fatal(err)
	}
	if manager.Dir != dir {
		t.Fatalf("expected dir %q, got %q", dir, manager.Dir)
	}
}

func TestCreateGetAndListTasks(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	created, err := manager.Create(task.CreateInput{
		ID:    "setup-runtime",
		Title: "Set up runtime",
		Owner: "lead",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Status != task.StatusPending {
		t.Fatalf("expected pending status, got %q", created.Status)
	}
	fetched, err := manager.Get("setup-runtime")
	if err != nil {
		t.Fatal(err)
	}
	if fetched.Title != "Set up runtime" {
		t.Fatalf("unexpected title: %q", fetched.Title)
	}
	list, err := manager.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected one task, got %d", len(list))
	}
	if list[0].ID != "setup-runtime" {
		t.Fatalf("unexpected listed task id: %q", list[0].ID)
	}
}

func TestCreateBlockedTaskUsesBlockedStatus(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	created, err := manager.Create(task.CreateInput{
		ID:        "write-tests",
		Title:     "Write tests",
		BlockedBy: []string{"setup-runtime"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Status != task.StatusBlocked {
		t.Fatalf("expected blocked status, got %q", created.Status)
	}
}

func TestUpdateStatusCompletingTaskUnblocksDependents(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(task.CreateInput{
		ID:    "task-a",
		Title: "Finish foundation",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(task.CreateInput{
		ID:        "task-b",
		Title:     "Build dependent feature",
		BlockedBy: []string{"task-a"},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateStatus("task-a", task.StatusCompleted); err != nil {
		t.Fatal(err)
	}
	dependent, err := manager.Get("task-b")
	if err != nil {
		t.Fatal(err)
	}
	if dependent.Status != task.StatusPending {
		t.Fatalf("expected dependent to become pending, got %q", dependent.Status)
	}
	if len(dependent.BlockedBy) != 0 {
		t.Fatalf("expected dependency to clear, got %#v", dependent.BlockedBy)
	}
}

func TestUpdateRewritesTaskFields(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	created, err := manager.Create(task.CreateInput{
		ID:    "task-a",
		Title: "Initial title",
	})
	if err != nil {
		t.Fatal(err)
	}
	created.Title = "Refined title"
	created.Owner = "alice"
	created.Worktree = "wt-task-a"
	updated, err := manager.Update(created)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != "Refined title" {
		t.Fatalf("expected updated title, got %q", updated.Title)
	}
	if updated.Owner != "alice" {
		t.Fatalf("expected owner alice, got %q", updated.Owner)
	}
	if updated.Worktree != "wt-task-a" {
		t.Fatalf("expected worktree wt-task-a, got %q", updated.Worktree)
	}
}

func TestRecordBackgroundStoresLatestExecutionContext(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(task.CreateInput{
		ID:    "task-a",
		Title: "Track background run",
	}); err != nil {
		t.Fatal(err)
	}
	recorded, err := manager.RecordBackground("task-a", task.BackgroundContext{
		JobID:   "job-1",
		Status:  "running",
		Command: "go test ./...",
	})
	if err != nil {
		t.Fatal(err)
	}
	if recorded.LastBackground == nil {
		t.Fatal("expected last background context")
	}
	if recorded.LastBackground.JobID != "job-1" || recorded.LastBackground.Status != "running" {
		t.Fatalf("unexpected background context: %#v", recorded.LastBackground)
	}
	if len(recorded.BackgroundHistory) != 1 {
		t.Fatalf("expected one history entry, got %d", len(recorded.BackgroundHistory))
	}
}

func TestRecordBackgroundKeepsBoundedHistory(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(task.CreateInput{
		ID:    "task-a",
		Title: "Track repeated runs",
	}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < task.MaxBackgroundHistory+2; i++ {
		_, err := manager.RecordBackground("task-a", task.BackgroundContext{
			JobID:   fmt.Sprintf("job-%d", i),
			Status:  "completed",
			Command: "go test ./...",
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	item, err := manager.Get("task-a")
	if err != nil {
		t.Fatal(err)
	}
	if len(item.BackgroundHistory) != task.MaxBackgroundHistory {
		t.Fatalf("expected history length %d, got %d", task.MaxBackgroundHistory, len(item.BackgroundHistory))
	}
	if item.BackgroundHistory[0].JobID != "job-2" {
		t.Fatalf("expected oldest retained job to be job-2, got %q", item.BackgroundHistory[0].JobID)
	}
	if item.BackgroundHistory[len(item.BackgroundHistory)-1].JobID != "job-6" {
		t.Fatalf("expected newest retained job to be job-6, got %q", item.BackgroundHistory[len(item.BackgroundHistory)-1].JobID)
	}
}

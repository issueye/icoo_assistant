package background_test

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"icoo_assistant/internal/background"
)

func TestStartAndCompleteBackgroundJob(t *testing.T) {
	manager, err := background.NewManager(filepath.Join(t.TempDir(), ".background"), t.TempDir(), 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	command := "printf hello"
	if runtime.GOOS == "windows" {
		command = "Write-Output hello"
	}
	job, err := manager.Start(background.StartInput{
		ID:      "job-1",
		Command: command,
		TaskID:  "task-a",
		Owner:   "lead",
	})
	if err != nil {
		t.Fatal(err)
	}
	if job.Status != background.StatusRunning {
		t.Fatalf("expected running status, got %q", job.Status)
	}
	final := waitForJob(t, manager, job.ID)
	if final.Status != background.StatusCompleted {
		t.Fatalf("expected completed status, got %q", final.Status)
	}
	if !strings.Contains(final.Output, "hello") {
		t.Fatalf("expected output to contain hello, got %q", final.Output)
	}
}

func TestPollNotificationsReturnsCompletedJobsOnce(t *testing.T) {
	manager, err := background.NewManager(filepath.Join(t.TempDir(), ".background"), t.TempDir(), 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	command := "printf hello"
	if runtime.GOOS == "windows" {
		command = "Write-Output hello"
	}
	job, err := manager.Start(background.StartInput{
		ID:      "job-1",
		Command: command,
		TaskID:  "task-a",
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = waitForJob(t, manager, job.ID)
	completions, err := manager.PollNotifications()
	if err != nil {
		t.Fatal(err)
	}
	if len(completions) != 1 {
		t.Fatalf("expected one completion, got %d", len(completions))
	}
	if !strings.Contains(completions[0].Summary, "<background_result>") {
		t.Fatalf("unexpected summary: %q", completions[0].Summary)
	}
	completions, err = manager.PollNotifications()
	if err != nil {
		t.Fatal(err)
	}
	if len(completions) != 0 {
		t.Fatalf("expected no duplicate completions, got %d", len(completions))
	}
}

func TestStartRejectsDangerousCommand(t *testing.T) {
	manager, err := background.NewManager(filepath.Join(t.TempDir(), ".background"), t.TempDir(), 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Start(background.StartInput{ID: "job-1", Command: "sudo rm -rf /"}); err == nil {
		t.Fatal("expected dangerous command to be rejected")
	}
}

func waitForJob(t *testing.T, manager *background.Manager, id string) background.Job {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		job, err := manager.Get(id)
		if err != nil {
			t.Fatal(err)
		}
		if job.Status != background.StatusRunning {
			return job
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("job %s did not finish in time", id)
	return background.Job{}
}

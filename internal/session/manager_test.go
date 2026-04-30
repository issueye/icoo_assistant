package session_test

import (
	"path/filepath"
	"testing"

	"icoo_assistant/internal/session"
)

func TestNewManagerCreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".sessions")
	manager, err := session.NewManager(dir)
	if err != nil {
		t.Fatal(err)
	}
	if manager.Dir != dir {
		t.Fatalf("expected dir %q, got %q", dir, manager.Dir)
	}
}

func TestCreateAndGetSession(t *testing.T) {
	manager, err := session.NewManager(filepath.Join(t.TempDir(), ".sessions"))
	if err != nil {
		t.Fatal(err)
	}
	sess, err := manager.Create(session.CreateInput{
		ID:    "session-001",
		Title: "Refactoring auth module",
		Tags:  []string{"refactoring", "auth"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if sess.Status != session.StatusActive {
		t.Fatalf("expected active status, got %q", sess.Status)
	}
	if sess.Title != "Refactoring auth module" {
		t.Fatalf("expected title, got %q", sess.Title)
	}

	got, err := manager.Get("session-001")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "session-001" {
		t.Fatalf("expected session-001, got %q", got.ID)
	}
}

func TestEnsureActiveCreatesDefaultSession(t *testing.T) {
	manager, err := session.NewManager(filepath.Join(t.TempDir(), ".sessions"))
	if err != nil {
		t.Fatal(err)
	}
	sess, err := manager.EnsureActive()
	if err != nil {
		t.Fatal(err)
	}
	if sess.Status != session.StatusActive {
		t.Fatalf("expected active status, got %q", sess.Status)
	}
}

func TestCloseSession(t *testing.T) {
	manager, err := session.NewManager(filepath.Join(t.TempDir(), ".sessions"))
	if err != nil {
		t.Fatal(err)
	}
	sess, _ := manager.Create(session.CreateInput{ID: "s1", Title: "Test"})
	_ = sess

	closed, err := manager.Close("s1")
	if err != nil {
		t.Fatal(err)
	}
	if closed.Status != session.StatusClosed {
		t.Fatalf("expected closed status, got %q", closed.Status)
	}
	if closed.ClosedAt == nil {
		t.Fatal("expected closed_at to be set")
	}
}

func TestSwitchSession(t *testing.T) {
	manager, err := session.NewManager(filepath.Join(t.TempDir(), ".sessions"))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = manager.Create(session.CreateInput{ID: "s1", Title: "Session One"})
	_, _ = manager.Create(session.CreateInput{ID: "s2", Title: "Session Two"})

	_, _, _ = manager.Switch("s1")

	target, previous, err := manager.Switch("s2")
	if err != nil {
		t.Fatal(err)
	}
	if target.ID != "s2" {
		t.Fatalf("expected switched to s2, got %q", target.ID)
	}
	if previous.ID != "s1" {
		t.Fatalf("expected previous s1, got %q", previous.ID)
	}
	if previous.Status != session.StatusClosed {
		t.Fatalf("expected previous session closed, got %q", previous.Status)
	}
}

func TestSwitchReopensClosedSession(t *testing.T) {
	manager, err := session.NewManager(filepath.Join(t.TempDir(), ".sessions"))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = manager.Create(session.CreateInput{ID: "s1", Title: "One"})
	_, _ = manager.Create(session.CreateInput{ID: "s2", Title: "Two"})
	_, _ = manager.Close("s2")

	_, _, err = manager.Switch("s2")
	if err != nil {
		t.Fatal(err)
	}
	active, _ := manager.GetActive()
	if active.ID != "s2" {
		t.Fatalf("expected active s2, got %q", active.ID)
	}
	if active.Status != session.StatusActive {
		t.Fatalf("expected reopened to active, got %q", active.Status)
	}
}

func TestListAndArchive(t *testing.T) {
	manager, err := session.NewManager(filepath.Join(t.TempDir(), ".sessions"))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = manager.Create(session.CreateInput{ID: "s1", Title: "One"})
	_, _ = manager.Create(session.CreateInput{ID: "s2", Title: "Two"})
	_, _ = manager.Close("s2")

	all, err := manager.List("")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(all))
	}

	closed, _ := manager.List(session.StatusClosed)
	if len(closed) != 1 || closed[0].ID != "s2" {
		t.Fatal("closed session filter failed")
	}

	archived, err := manager.Archive("s2")
	if err != nil {
		t.Fatal(err)
	}
	if archived.Status != session.StatusArchived {
		t.Fatalf("expected archived, got %q", archived.Status)
	}
}

func TestUpdateStats(t *testing.T) {
	manager, err := session.NewManager(filepath.Join(t.TempDir(), ".sessions"))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = manager.Create(session.CreateInput{ID: "s1", Title: "One"})

	if err := manager.UpdateStats("s1", 5, 42); err != nil {
		t.Fatal(err)
	}
	got, _ := manager.Get("s1")
	if got.RoundCount != 5 {
		t.Fatalf("expected rounds=5, got %d", got.RoundCount)
	}
	if got.MessageCount != 42 {
		t.Fatalf("expected messages=42, got %d", got.MessageCount)
	}
}

func TestUpdateSummary(t *testing.T) {
	manager, err := session.NewManager(filepath.Join(t.TempDir(), ".sessions"))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = manager.Create(session.CreateInput{ID: "s1", Title: "One"})

	if err := manager.UpdateSummary("s1", "Completed refactoring", []string{"mem-1", "mem-2"}); err != nil {
		t.Fatal(err)
	}
	got, _ := manager.Get("s1")
	if got.Summary != "Completed refactoring" {
		t.Fatalf("expected summary, got %q", got.Summary)
	}
	if len(got.MemoryIDs) != 2 {
		t.Fatalf("expected 2 memory ids, got %d", len(got.MemoryIDs))
	}
}

func TestHistory(t *testing.T) {
	manager, err := session.NewManager(filepath.Join(t.TempDir(), ".sessions"))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = manager.Create(session.CreateInput{ID: "s1", Title: "Oldest"})
	_, _ = manager.Create(session.CreateInput{ID: "s2", Title: "Newest"})

	history, err := manager.History(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) < 2 {
		t.Fatalf("expected at least 2 in history, got %d", len(history))
	}
	if history[0].ID != "s2" {
		t.Fatalf("expected newest first, got %q", history[0].ID)
	}
}

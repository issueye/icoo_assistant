package memory_test

import (
	"path/filepath"
	"strings"
	"testing"

	"icoo_assistant/internal/memory"
)

func TestNewManagerCreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".memory")
	manager, err := memory.NewManager(dir)
	if err != nil {
		t.Fatal(err)
	}
	if manager.Dir != dir {
		t.Fatalf("expected dir %q, got %q", dir, manager.Dir)
	}
}

func TestStoreAndRecallLongTerm(t *testing.T) {
	manager, err := memory.NewManager(filepath.Join(t.TempDir(), ".memory"))
	if err != nil {
		t.Fatal(err)
	}
	mem, err := manager.Store(memory.StoreInput{
		ID:         "mem-1",
		Type:       "long_term",
		Content:    "User prefers Go language for backend work",
		Tags:       []string{"preference", "language"},
		Importance: 0.9,
	})
	if err != nil {
		t.Fatal(err)
	}
	if mem.ID != "mem-1" {
		t.Fatalf("expected id mem-1, got %q", mem.ID)
	}
	if mem.Type != "long_term" {
		t.Fatalf("expected type long_term, got %q", mem.Type)
	}

	results, err := manager.Recall(memory.QueryInput{
		Type:  "long_term",
		Query: "Go",
		Limit: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Content != "User prefers Go language for backend work" {
		t.Fatalf("unexpected content: %q", results[0].Content)
	}
}

func TestShortTermMemoryRingBuffer(t *testing.T) {
	manager, err := memory.NewManager(filepath.Join(t.TempDir(), ".memory"))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 110; i++ {
		_, err := manager.Store(memory.StoreInput{
			Type:    "short_term",
			Content: "temp memory entry",
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	results, err := manager.Recall(memory.QueryInput{Type: "short_term"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) > 100 {
		t.Fatalf("short_term exceeded ring buffer max: got %d", len(results))
	}
}

func TestRecallWithTagsAndImportanceFilter(t *testing.T) {
	manager, err := memory.NewManager(filepath.Join(t.TempDir(), ".memory"))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = manager.Store(memory.StoreInput{ID: "a", Type: "long_term", Content: "Important fact", Tags: []string{"critical"}, Importance: 1.0})
	_, _ = manager.Store(memory.StoreInput{ID: "b", Type: "long_term", Content: "Minor note", Tags: []string{"low"}, Importance: 0.1})

	results, err := manager.Recall(memory.QueryInput{Type: "long_term", MinImportance: 0.5})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 high-importance result, got %d", len(results))
	}
	if results[0].ID != "a" {
		t.Fatalf("expected mem a, got %s", results[0].ID)
	}

	results, _ = manager.Recall(memory.QueryInput{Type: "long_term", Tags: []string{"critical"}})
	if len(results) != 1 || results[0].ID != "a" {
		t.Fatalf("tag filter failed, got %d results", len(results))
	}
}

func TestUpdateMemory(t *testing.T) {
	manager, err := memory.NewManager(filepath.Join(t.TempDir(), ".memory"))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = manager.Store(memory.StoreInput{ID: "m1", Type: "long_term", Content: "Old content", Importance: 0.5})

	updated, err := manager.Update("m1", "New content", []string{"updated", "tag"}, 0.8)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Content != "New content" {
		t.Fatalf("content not updated: %q", updated.Content)
	}
	if updated.Importance != 0.8 {
		t.Fatalf("importance not updated: %f", updated.Importance)
	}
	if len(updated.Tags) != 2 {
		t.Fatalf("tags not updated: %v", updated.Tags)
	}
}

func TestDeleteMemory(t *testing.T) {
	manager, err := memory.NewManager(filepath.Join(t.TempDir(), ".memory"))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = manager.Store(memory.StoreInput{ID: "del-me", Type: "long_term", Content: "Delete me"})

	if err := manager.Delete("del-me", "long_term"); err != nil {
		t.Fatal(err)
	}
	results, _ := manager.Recall(memory.QueryInput{Type: "long_term", Query: "Delete"})
	if len(results) != 0 {
		t.Fatalf("memory should have been deleted, got %d results", len(results))
	}
}

func TestAIPersonalityAndUserProfile(t *testing.T) {
	manager, err := memory.NewManager(filepath.Join(t.TempDir(), ".memory"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = manager.Store(memory.StoreInput{ID: "ai", Type: "ai_personality", Content: "Be concise and direct"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = manager.Store(memory.StoreInput{ID: "up", Type: "user_profile", Content: "User likes tabs over spaces"})
	if err != nil {
		t.Fatal(err)
	}

	results, _ := manager.Recall(memory.QueryInput{Type: "ai_personality"})
	if len(results) != 1 || results[0].Content != "Be concise and direct" {
		t.Fatal("ai_personality recall failed")
	}

	results, _ = manager.Recall(memory.QueryInput{Type: "user_profile"})
	if len(results) != 1 || results[0].Content != "User likes tabs over spaces" {
		t.Fatal("user_profile recall failed")
	}
}

func TestGenerateSessionContext(t *testing.T) {
	manager, err := memory.NewManager(filepath.Join(t.TempDir(), ".memory"))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = manager.Store(memory.StoreInput{Type: "ai_personality", Content: "Be helpful"})
	_, _ = manager.Store(memory.StoreInput{Type: "user_profile", Content: "Prefers Python"})
	_, _ = manager.Store(memory.StoreInput{Type: "long_term", Content: "Project uses postgres", Tags: []string{"db"}, Importance: 0.9})

	ctx := manager.GenerateSessionContext()
	if !strings.Contains(ctx, "Be helpful") {
		t.Fatal("context missing personality")
	}
	if !strings.Contains(ctx, "Prefers Python") {
		t.Fatal("context missing user profile")
	}
	if !strings.Contains(ctx, "postgres") {
		t.Fatal("context missing long term memory")
	}
}

func TestSetSessionID(t *testing.T) {
	manager, err := memory.NewManager(filepath.Join(t.TempDir(), ".memory"))
	if err != nil {
		t.Fatal(err)
	}
	manager.SetSessionID("session-123")
	mem, _ := manager.Store(memory.StoreInput{Type: "short_term", Content: "In session"})
	if mem.SessionID != "session-123" {
		t.Fatalf("expected session-123, got %q", mem.SessionID)
	}
}

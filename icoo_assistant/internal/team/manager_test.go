package team

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestNewManagerCreatesDefaultConfigAndRegistryDir(t *testing.T) {
	manager, err := NewManager(filepath.Join(t.TempDir(), ".team"))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := manager.GetConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.LeadID != "lead" {
		t.Fatalf("expected default lead id, got %#v", cfg)
	}
	if manager.RegistryDir == "" || !strings.HasSuffix(manager.RegistryDir, "teammates") {
		t.Fatalf("expected teammates registry dir, got %q", manager.RegistryDir)
	}
	if manager.InboxDir == "" || !strings.HasSuffix(manager.InboxDir, "inbox") {
		t.Fatalf("expected inbox dir, got %q", manager.InboxDir)
	}
}

func TestManagerCreateListAndUpdateTeammate(t *testing.T) {
	manager, err := NewManager(filepath.Join(t.TempDir(), ".team"))
	if err != nil {
		t.Fatal(err)
	}
	item, err := manager.Create(CreateInput{
		ID:     "alice",
		Role:   "reviewer",
		Status: StatusIdle,
		Model:  "claude-opus-4-7",
	})
	if err != nil {
		t.Fatal(err)
	}
	if item.ID != "alice" || item.Role != "reviewer" || item.Status != StatusIdle {
		t.Fatalf("unexpected teammate: %#v", item)
	}
	items, err := manager.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != "alice" {
		t.Fatalf("unexpected list: %#v", items)
	}
	item.Status = StatusBusy
	updated, err := manager.Update(item)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != StatusBusy {
		t.Fatalf("expected busy status, got %#v", updated)
	}
}

func TestManagerUpdateConfig(t *testing.T) {
	manager, err := NewManager(filepath.Join(t.TempDir(), ".team"))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := manager.UpdateConfig(ConfigUpdateInput{
		LeadID:  "captain",
		Mission: "Build a reviewer pair",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.LeadID != "captain" || cfg.Mission != "Build a reviewer pair" {
		t.Fatalf("unexpected config: %#v", cfg)
	}
}

func TestManagerSendMessageAndListInbox(t *testing.T) {
	manager, err := NewManager(filepath.Join(t.TempDir(), ".team"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(CreateInput{
		ID:   "alice",
		Role: "reviewer",
	}); err != nil {
		t.Fatal(err)
	}
	msg, err := manager.SendMessage(SendMessageInput{
		FromID: "lead",
		ToID:   "alice",
		Kind:   "request",
		Body:   "Please review the latest plan.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if msg.ToID != "alice" || msg.FromID != "lead" || msg.Kind != "request" {
		t.Fatalf("unexpected message: %#v", msg)
	}
	items, err := manager.ListInbox("alice", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Body != "Please review the latest plan." {
		t.Fatalf("unexpected inbox items: %#v", items)
	}
}

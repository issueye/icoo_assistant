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

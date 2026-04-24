package tools_test

import (
	"path/filepath"
	"strings"
	"testing"

	"icoo_assistant/internal/team"
	"icoo_assistant/internal/tools"
)

func TestTeamRegistryToolConfigCreateAndList(t *testing.T) {
	manager, err := team.NewManager(filepath.Join(t.TempDir(), ".team"))
	if err != nil {
		t.Fatal(err)
	}
	tool := tools.NewTeamRegistryTool(manager)
	configResult, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action":  "update_config",
		"lead_id": "captain",
		"mission": "Stand up a reviewer teammate",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(configResult, "lead_id: captain") || !strings.Contains(configResult, "mission: Stand up a reviewer teammate") {
		t.Fatalf("unexpected config result: %q", configResult)
	}
	createResult, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "create",
		"id":     "alice",
		"role":   "reviewer",
		"model":  "claude-opus-4-7",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(createResult, "id: alice") || !strings.Contains(createResult, "status: idle") {
		t.Fatalf("unexpected create result: %q", createResult)
	}
	listResult, err := tool.Handler(tools.Call{Input: map[string]interface{}{"action": "list"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(listResult, "alice [idle] role=reviewer model=claude-opus-4-7") {
		t.Fatalf("unexpected list result: %q", listResult)
	}
}

func TestTeamRegistryToolUpdateAndGet(t *testing.T) {
	manager, err := team.NewManager(filepath.Join(t.TempDir(), ".team"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(team.CreateInput{
		ID:   "bob",
		Role: "implementer",
	}); err != nil {
		t.Fatal(err)
	}
	tool := tools.NewTeamRegistryTool(manager)
	updateResult, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "update",
		"id":     "bob",
		"status": "busy",
		"model":  "claude-sonnet",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(updateResult, "status: busy") || !strings.Contains(updateResult, "model: claude-sonnet") {
		t.Fatalf("unexpected update result: %q", updateResult)
	}
	getResult, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "get",
		"id":     "bob",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(getResult, "role: implementer") || !strings.Contains(getResult, "status: busy") {
		t.Fatalf("unexpected get result: %q", getResult)
	}
}

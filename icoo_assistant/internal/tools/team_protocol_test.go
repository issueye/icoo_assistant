package tools_test

import (
	"path/filepath"
	"strings"
	"testing"

	"icoo_assistant/internal/team"
	"icoo_assistant/internal/tools"
)

func TestTeamProtocolToolGetListAndSummary(t *testing.T) {
	manager, err := team.NewManager(filepath.Join(t.TempDir(), ".team"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(team.CreateInput{ID: "alice", Role: "reviewer"}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(team.CreateInput{ID: "bob", Role: "implementer"}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.SendMessage(team.SendMessageInput{
		FromID:    "lead",
		ToID:      "alice",
		Kind:      "request",
		Body:      "Please review the patch.",
		RequestID: "req-1",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.SendMessage(team.SendMessageInput{
		FromID:    "lead",
		ToID:      "bob",
		Kind:      "request",
		Body:      "Please implement the fix.",
		RequestID: "req-2",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.ReplyToRequest(team.ReplyInput{
		FromID:    "bob",
		RequestID: "req-2",
		Body:      "Implemented.",
	}); err != nil {
		t.Fatal(err)
	}

	tool := tools.NewTeamProtocolTool(manager)
	getResult, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action":     "get",
		"request_id": "req-2",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(getResult, "request_id: req-2") || !strings.Contains(getResult, "status: responded") {
		t.Fatalf("unexpected get result: %q", getResult)
	}

	listResult, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "list",
		"status": "pending",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(listResult, "request_count: 1") || !strings.Contains(listResult, "- req-1") {
		t.Fatalf("unexpected list result: %q", listResult)
	}

	summaryResult, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "summary",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(summaryResult, "pending_count: 1") || !strings.Contains(summaryResult, "responded_count: 1") {
		t.Fatalf("unexpected summary result: %q", summaryResult)
	}
}

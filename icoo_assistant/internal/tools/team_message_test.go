package tools_test

import (
	"path/filepath"
	"strings"
	"testing"

	"icoo_assistant/internal/team"
	"icoo_assistant/internal/tools"
)

func TestTeamMessageToolSendAndInbox(t *testing.T) {
	manager, err := team.NewManager(filepath.Join(t.TempDir(), ".team"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(team.CreateInput{
		ID:   "alice",
		Role: "reviewer",
	}); err != nil {
		t.Fatal(err)
	}
	tool := tools.NewTeamMessageTool(manager)
	sendResult, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action":     "send",
		"to":         "alice",
		"kind":       "request",
		"body":       "Please review the latest patch.",
		"request_id": "req-1",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sendResult, "from: lead") || !strings.Contains(sendResult, "to: alice") {
		t.Fatalf("unexpected send result: %q", sendResult)
	}
	inboxResult, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action":    "inbox",
		"recipient": "alice",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(inboxResult, "recipient: alice") || !strings.Contains(inboxResult, "message_count: 1") {
		t.Fatalf("unexpected inbox result: %q", inboxResult)
	}
	if !strings.Contains(inboxResult, "request_id=req-1") {
		t.Fatalf("expected request id in inbox output, got %q", inboxResult)
	}
}

func TestTeamMessageToolRequestReplyAndThread(t *testing.T) {
	manager, err := team.NewManager(filepath.Join(t.TempDir(), ".team"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(team.CreateInput{
		ID:   "alice",
		Role: "reviewer",
	}); err != nil {
		t.Fatal(err)
	}
	tool := tools.NewTeamMessageTool(manager)
	requestResult, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action":     "request",
		"to":         "alice",
		"body":       "Please review the latest plan.",
		"request_id": "req-42",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(requestResult, "kind: request") || !strings.Contains(requestResult, "request_id: req-42") {
		t.Fatalf("unexpected request result: %q", requestResult)
	}
	replyResult, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action":     "reply",
		"from":       "alice",
		"request_id": "req-42",
		"body":       "Reviewed and approved.",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(replyResult, "kind: response") || !strings.Contains(replyResult, "to: lead") {
		t.Fatalf("unexpected reply result: %q", replyResult)
	}
	threadResult, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action":     "thread",
		"request_id": "req-42",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(threadResult, "message_count: 2") {
		t.Fatalf("unexpected thread result: %q", threadResult)
	}
	if !strings.Contains(threadResult, "kind=request") || !strings.Contains(threadResult, "kind=response") {
		t.Fatalf("expected request/response thread, got %q", threadResult)
	}
}

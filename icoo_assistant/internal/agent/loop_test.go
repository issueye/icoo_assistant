package agent_test

import (
	"os"
	"path/filepath"
	"testing"

	"icoo_assistant/internal/agent"
	"icoo_assistant/internal/compact"
	"icoo_assistant/internal/llm"
	"icoo_assistant/internal/todo"
	"icoo_assistant/internal/tools"
)

func TestRunnerCompletesToolUseLoop(t *testing.T) {
	client := &llm.FakeClient{Responses: []llm.Response{
		{StopReason: "tool_use", ToolUses: []llm.ToolUse{{ID: "call-1", Name: "demo", Input: map[string]interface{}{"value": "x"}}}},
		{StopReason: "end", Text: "done"},
	}}
	registry, err := tools.NewRegistry(tools.Definition{
		Tool: llm.Tool{Name: "demo", Description: "demo", InputSchema: map[string]interface{}{}},
		Handler: func(call tools.Call) (string, error) {
			return "tool output", nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	runner := &agent.Runner{Client: client, Registry: registry, Config: agent.Config{SystemPrompt: "test", MaxRounds: 5}}
	messages, err := runner.Run([]llm.Message{{Role: "user", Content: "run demo"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 4 {
		t.Fatalf("unexpected message count: %d", len(messages))
	}
	if messages[len(messages)-1].Content != "done" {
		t.Fatalf("unexpected final content: %#v", messages[len(messages)-1].Content)
	}
	if client.Calls != 2 {
		t.Fatalf("expected 2 llm calls, got %d", client.Calls)
	}
}

func TestRunnerAddsTodoReminderAfterThreeRounds(t *testing.T) {
	client := &llm.FakeClient{Responses: []llm.Response{
		{StopReason: "tool_use", ToolUses: []llm.ToolUse{{ID: "1", Name: "demo", Input: map[string]interface{}{}}}},
		{StopReason: "tool_use", ToolUses: []llm.ToolUse{{ID: "2", Name: "demo", Input: map[string]interface{}{}}}},
		{StopReason: "tool_use", ToolUses: []llm.ToolUse{{ID: "3", Name: "demo", Input: map[string]interface{}{}}}},
		{StopReason: "end", Text: "done"},
	}}
	registry, err := tools.NewRegistry(tools.Definition{
		Tool: llm.Tool{Name: "demo", Description: "demo", InputSchema: map[string]interface{}{}},
		Handler: func(call tools.Call) (string, error) {
			return "ok", nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	runner := &agent.Runner{Client: client, Registry: registry, TodoManager: todo.NewManager(), Config: agent.Config{SystemPrompt: "test", MaxRounds: 6}}
	messages, err := runner.Run([]llm.Message{{Role: "user", Content: "run demo"}})
	if err != nil {
		t.Fatal(err)
	}
	foundReminder := false
	for _, msg := range messages {
		if results, ok := msg.Content.([]tools.Result); ok {
			for _, result := range results {
				if result.Content == "<reminder>Update your todos.</reminder>" {
					foundReminder = true
				}
			}
		}
	}
	if !foundReminder {
		t.Fatal("expected todo reminder after three non-todo rounds")
	}
}

func TestRunnerAutoCompactsWhenThresholdExceeded(t *testing.T) {
	root := t.TempDir()
	client := &llm.FakeClient{Responses: []llm.Response{{StopReason: "end", Text: "done"}}}
	registry, err := tools.NewRegistry(tools.Definition{
		Tool: llm.Tool{Name: "demo", Description: "demo", InputSchema: map[string]interface{}{}},
		Handler: func(call tools.Call) (string, error) {
			return "ok", nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	manager := &compact.Manager{Threshold: 1, KeepRecent: 3, Dir: root}
	runner := &agent.Runner{Client: client, Registry: registry, CompactManager: manager, Config: agent.Config{SystemPrompt: "test", MaxRounds: 2}}
	_, err = runner.Run([]llm.Message{{Role: "user", Content: "this is a very long message that should trigger compaction"}})
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("expected transcript file after auto compact")
	}
	if filepath.Ext(entries[0].Name()) != ".jsonl" {
		t.Fatalf("unexpected transcript file: %s", entries[0].Name())
	}
}

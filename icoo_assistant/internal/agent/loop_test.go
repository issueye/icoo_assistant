package agent_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"icoo_assistant/internal/agent"
	"icoo_assistant/internal/background"
	"icoo_assistant/internal/compact"
	"icoo_assistant/internal/llm"
	"icoo_assistant/internal/todo"
	"icoo_assistant/internal/tools"
)

type fakeSubagent struct {
	summary string
}

func (f fakeSubagent) Run(prompt string) (string, error) {
	return f.summary + ": " + prompt, nil
}

type fakeBackgroundNotifier struct {
	completions []background.Completion
	polled      bool
}

type captureHook struct {
	events []agent.Event
}

type captureTranscriptRecorder struct {
	records []agent.TranscriptRecord
}

func (h *captureHook) OnEvent(event agent.Event) {
	h.events = append(h.events, event)
}

func (r *captureTranscriptRecorder) Record(record agent.TranscriptRecord) error {
	r.records = append(r.records, record)
	return nil
}

func (f *fakeBackgroundNotifier) PollNotifications() ([]background.Completion, error) {
	if f.polled {
		return nil, nil
	}
	f.polled = true
	return f.completions, nil
}

func TestRunnerCompletesToolUseLoop(t *testing.T) {
	client := &llm.FakeClient{Responses: []llm.Response{
		{StopReason: "tool_use", ToolUses: []llm.ToolUse{{ID: "call-1", Name: "demo", Input: map[string]interface{}{"value": "x"}}}},
		{StopReason: "end", Text: "done"},
	}}
	registry, err := tools.NewRegistry(tools.Definition{
		Tool:    llm.Tool{Name: "demo", Description: "demo", InputSchema: map[string]interface{}{}},
		Handler: func(call tools.Call) (string, error) { return "tool output", nil },
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
}

func TestRunnerAddsTodoReminderAfterThreeRounds(t *testing.T) {
	client := &llm.FakeClient{Responses: []llm.Response{
		{StopReason: "tool_use", ToolUses: []llm.ToolUse{{ID: "1", Name: "demo", Input: map[string]interface{}{}}}},
		{StopReason: "tool_use", ToolUses: []llm.ToolUse{{ID: "2", Name: "demo", Input: map[string]interface{}{}}}},
		{StopReason: "tool_use", ToolUses: []llm.ToolUse{{ID: "3", Name: "demo", Input: map[string]interface{}{}}}},
		{StopReason: "end", Text: "done"},
	}}
	registry, err := tools.NewRegistry(tools.Definition{
		Tool:    llm.Tool{Name: "demo", Description: "demo", InputSchema: map[string]interface{}{}},
		Handler: func(call tools.Call) (string, error) { return "ok", nil },
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
		if text, ok := msg.Content.(string); ok && text == "<reminder>Update your todos.</reminder>" {
			foundReminder = true
		}
	}
	if !foundReminder {
		t.Fatal("expected todo reminder after three non-todo rounds")
	}
}

func TestRunnerManualCompactReplacesMessages(t *testing.T) {
	root := t.TempDir()
	client := &llm.FakeClient{Responses: []llm.Response{
		{StopReason: "tool_use", ToolUses: []llm.ToolUse{{ID: "1", Name: "compact", Input: map[string]interface{}{}}}},
		{StopReason: "end", Text: "done"},
	}}
	registry, err := tools.NewRegistry(tools.NewCompactTool())
	if err != nil {
		t.Fatal(err)
	}
	manager := &compact.Manager{Threshold: 100000, KeepRecent: 3, Dir: root}
	runner := &agent.Runner{Client: client, Registry: registry, CompactManager: manager, Config: agent.Config{SystemPrompt: "test", MaxRounds: 4}}
	messages, err := runner.Run([]llm.Message{{Role: "user", Content: "please compact"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) < 2 {
		t.Fatalf("expected compacted conversation then final answer, got %d messages", len(messages))
	}
}

func TestRunnerDelegatesTaskToSubagent(t *testing.T) {
	client := &llm.FakeClient{Responses: []llm.Response{
		{StopReason: "tool_use", ToolUses: []llm.ToolUse{{ID: "1", Name: "task", Input: map[string]interface{}{"prompt": "inspect repo"}}}},
		{StopReason: "end", Text: "done"},
	}}
	registry, err := tools.NewRegistry(tools.NewTaskTool())
	if err != nil {
		t.Fatal(err)
	}
	runner := &agent.Runner{Client: client, Registry: registry, SubagentRunner: fakeSubagent{summary: "subagent summary"}, Config: agent.Config{SystemPrompt: "test", MaxRounds: 4}}
	messages, err := runner.Run([]llm.Message{{Role: "user", Content: "delegate"}})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, msg := range messages {
		if results, ok := msg.Content.([]tools.Result); ok {
			for _, result := range results {
				if result.Content == "subagent summary: inspect repo" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Fatal("expected subagent summary in tool results")
	}
}

func TestRunnerInjectsBackgroundNotifications(t *testing.T) {
	client := &llm.FakeClient{Responses: []llm.Response{
		{StopReason: "end", Text: "done"},
	}}
	registry, err := tools.NewRegistry(tools.Definition{
		Tool:    llm.Tool{Name: "demo", Description: "demo", InputSchema: map[string]interface{}{}},
		Handler: func(call tools.Call) (string, error) { return "ok", nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	notifier := &fakeBackgroundNotifier{
		completions: []background.Completion{{
			JobID:   "job-1",
			TaskID:  "task-a",
			Status:  "completed",
			Summary: "<background_result>\njob_id: job-1\nstatus: completed\n</background_result>",
		}},
	}
	runner := &agent.Runner{
		Client:     client,
		Registry:   registry,
		Background: notifier,
		Config:     agent.Config{SystemPrompt: "test", MaxRounds: 3},
	}
	if _, err := runner.Run([]llm.Message{{Role: "user", Content: "continue"}}); err != nil {
		t.Fatal(err)
	}
	if len(client.Snapshots) == 0 {
		t.Fatal("expected at least one client snapshot")
	}
	if !strings.Contains(client.Snapshots[0], "background_result") {
		t.Fatalf("expected background notification in snapshot, got %q", client.Snapshots[0])
	}
}

func TestRunnerEmitsHookEvents(t *testing.T) {
	client := &llm.FakeClient{Responses: []llm.Response{
		{StopReason: "tool_use", ToolUses: []llm.ToolUse{{ID: "call-1", Name: "demo", Input: map[string]interface{}{}}}},
		{StopReason: "end", Text: "done"},
	}}
	registry, err := tools.NewRegistry(tools.Definition{
		Tool:    llm.Tool{Name: "demo", Description: "demo", InputSchema: map[string]interface{}{}},
		Handler: func(call tools.Call) (string, error) { return "tool output", nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	hook := &captureHook{}
	runner := &agent.Runner{
		Client:   client,
		Registry: registry,
		Hooks:    []agent.Hook{hook},
		Config:   agent.Config{SystemPrompt: "test", MaxRounds: 5},
	}
	if _, err := runner.Run([]llm.Message{{Role: "user", Content: "run demo"}}); err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(hook.events))
	for _, event := range hook.events {
		names = append(names, event.Name)
	}
	for _, expected := range []string{
		"agent.run.started",
		"agent.round.started",
		"agent.model.requested",
		"agent.model.responded",
		"agent.tool.started",
		"agent.tool.completed",
		"agent.run.completed",
	} {
		if !containsString(names, expected) {
			t.Fatalf("expected hook event %q in %#v", expected, names)
		}
	}
}

func TestJSONLHookWritesEvents(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".agent-hooks")
	hook, err := agent.NewJSONLHook(dir)
	if err != nil {
		t.Fatal(err)
	}
	event := agent.Event{
		Timestamp: time.Unix(1700000000, 0).UTC(),
		Name:      "agent.run.started",
		RunID:     "run-1",
		Round:     1,
		Fields:    map[string]interface{}{"message_count": 1},
	}
	hook.OnEvent(event)
	data, err := os.ReadFile(filepath.Join(dir, "events.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected one event line, got %d", len(lines))
	}
	var parsed agent.Event
	if err := json.Unmarshal([]byte(lines[0]), &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.Name != "agent.run.started" || parsed.RunID != "run-1" {
		t.Fatalf("unexpected parsed event: %#v", parsed)
	}
}

func TestRunnerWritesTranscriptOnCompletion(t *testing.T) {
	client := &llm.FakeClient{Responses: []llm.Response{
		{StopReason: "end", Text: "done"},
	}}
	registry, err := tools.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	recorder := &captureTranscriptRecorder{}
	runner := &agent.Runner{
		Client:     client,
		Registry:   registry,
		Transcript: recorder,
		Config:     agent.Config{SystemPrompt: "test", MaxRounds: 2},
	}
	messages, err := runner.Run([]llm.Message{{Role: "user", Content: "hello"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 {
		t.Fatalf("unexpected message count: %d", len(messages))
	}
	if len(recorder.records) != 1 {
		t.Fatalf("expected one transcript record, got %d", len(recorder.records))
	}
	record := recorder.records[0]
	if record.Status != "completed" {
		t.Fatalf("expected completed transcript, got %#v", record)
	}
	if record.MessageCount != 2 {
		t.Fatalf("expected 2 messages, got %d", record.MessageCount)
	}
	if len(record.Messages) != 2 {
		t.Fatalf("expected 2 recorded messages, got %d", len(record.Messages))
	}
	if record.Messages[0].Content != "hello" {
		t.Fatalf("unexpected first message: %#v", record.Messages[0])
	}
	if record.Messages[1].Content != "done" {
		t.Fatalf("unexpected second message: %#v", record.Messages[1])
	}
}

func TestRunnerWritesTranscriptOnFailure(t *testing.T) {
	client := &llm.FakeClient{Responses: []llm.Response{
		{StopReason: "tool_use", ToolUses: []llm.ToolUse{{ID: "1", Name: "demo", Input: map[string]interface{}{}}}},
	}}
	registry, err := tools.NewRegistry(tools.Definition{
		Tool:    llm.Tool{Name: "demo", Description: "demo", InputSchema: map[string]interface{}{}},
		Handler: func(call tools.Call) (string, error) { return "ok", nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	recorder := &captureTranscriptRecorder{}
	runner := &agent.Runner{
		Client:     client,
		Registry:   registry,
		Transcript: recorder,
		Config:     agent.Config{SystemPrompt: "test", MaxRounds: 1},
	}
	_, err = runner.Run([]llm.Message{{Role: "user", Content: "hello"}})
	if err == nil {
		t.Fatal("expected max rounds exceeded error")
	}
	if len(recorder.records) != 1 {
		t.Fatalf("expected one transcript record, got %d", len(recorder.records))
	}
	record := recorder.records[0]
	if record.Status != "failed" {
		t.Fatalf("expected failed transcript, got %#v", record)
	}
	if !strings.Contains(record.Error, "max rounds exceeded") {
		t.Fatalf("unexpected failure record: %#v", record)
	}
	if record.MessageCount != len(record.Messages) {
		t.Fatalf("expected message count to match recorded messages, got %d vs %d", record.MessageCount, len(record.Messages))
	}
	if record.MessageCount < 3 {
		t.Fatalf("expected failure transcript to retain the in-flight conversation, got %d messages", record.MessageCount)
	}
	if record.Messages[0].Content != "hello" {
		t.Fatalf("expected original user message to be retained, got %#v", record.Messages[0])
	}
}

func TestJSONTranscriptRecorderWritesRecord(t *testing.T) {
	dir := t.TempDir()
	recorder, err := agent.NewJSONTranscriptRecorder(dir)
	if err != nil {
		t.Fatal(err)
	}
	record := agent.TranscriptRecord{
		Timestamp:    time.Unix(1700000000, 0).UTC(),
		RunID:        "run-1",
		Status:       "completed",
		MessageCount: 1,
		Messages:     []llm.Message{{Role: "user", Content: "hello"}},
	}
	if err := recorder.Record(record); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "conversation_run-1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var parsed agent.TranscriptRecord
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.RunID != "run-1" || parsed.MessageCount != 1 {
		t.Fatalf("unexpected parsed transcript: %#v", parsed)
	}
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

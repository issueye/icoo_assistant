package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"icoo_assistant/internal/agent"
	"icoo_assistant/internal/agents"
	"icoo_assistant/internal/commands"
	"icoo_assistant/internal/config"
	"icoo_assistant/internal/llm"
	"icoo_assistant/internal/tools"
)

func TestRunREPLExitsImmediately(t *testing.T) {
	cfg := config.Config{
		Workdir:        t.TempDir(),
		SystemPrompt:   "test",
		MaxRounds:      2,
		CommandTimeout: 1,
	}
	app, err := newApp(cfg)
	if err != nil {
		t.Fatal(err)
	}
	in := strings.NewReader("exit\n")
	var out bytes.Buffer
	if err := app.runREPL(in, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "icoo REPL started") {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestRunREPLWithFakeClientExplainsDegradedModeAndEmptyOutput(t *testing.T) {
	cfg := config.Config{
		Workdir:        t.TempDir(),
		SystemPrompt:   "test",
		MaxRounds:      2,
		CommandTimeout: 1,
	}
	app, err := newApp(cfg)
	if err != nil {
		t.Fatal(err)
	}
	in := strings.NewReader("hello\nexit\n")
	var out bytes.Buffer
	if err := app.runREPL(in, &out); err != nil {
		t.Fatal(err)
	}
	output := out.String()
	for _, snippet := range []string{
		"icoo REPL started (fake client)",
		"warning: REPL is running in fake mode",
		"warning: no model output was produced because the fake client returns empty responses by design.",
		"hint: run `go run ./cmd/assistant check` outside the REPL",
		"hint: this is expected in fake mode; set anthropic.api_key in config.toml for real answers",
	} {
		if !strings.Contains(output, snippet) {
			t.Fatalf("expected REPL output to contain %q, got %q", snippet, output)
		}
	}
}

func TestRunOnceWithFakeClientProducesGuidance(t *testing.T) {
	cfg := config.Config{
		Workdir:        t.TempDir(),
		SystemPrompt:   "test",
		MaxRounds:      2,
		CommandTimeout: 1,
	}
	app, err := newApp(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := app.runOnce(&out, "hello"); err != nil {
		t.Fatal(err)
	}
	output := out.String()
	for _, snippet := range []string{
		"warning: icoo is running in fake mode",
		"hint: run `go run ./cmd/assistant check`",
		"warning: no model output was produced because the fake client returns empty responses by design.",
	} {
		if !strings.Contains(output, snippet) {
			t.Fatalf("expected runOnce output to contain %q, got %q", snippet, output)
		}
	}
}

func TestRunOnceStreamsWithoutDuplicatingOutput(t *testing.T) {
	client := &llm.FakeClient{Responses: []llm.Response{
		{StopReason: "end", Text: "streamed hello"},
	}}
	registry, err := tools.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	app := &app{
		runner: &agent.Runner{
			Client:   client,
			Registry: registry,
			Config:   agent.Config{SystemPrompt: "test", MaxRounds: 2},
		},
		mode: "anthropic",
	}
	var out bytes.Buffer
	if err := app.runOnce(&out, "hello"); err != nil {
		t.Fatal(err)
	}
	output := out.String()
	if strings.Count(output, "streamed hello") != 1 {
		t.Fatalf("expected streamed output once, got %q", output)
	}
}

func TestRunREPLRetainsConversationHistory(t *testing.T) {
	client := &llm.FakeClient{Responses: []llm.Response{
		{StopReason: "end", Text: "记住了，你叫小明。"},
		{StopReason: "end", Text: "你叫小明。"},
	}}
	registry, err := tools.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	app := &app{
		runner: &agent.Runner{
			Client:   client,
			Registry: registry,
			Config:   agent.Config{SystemPrompt: "test", MaxRounds: 2},
		},
		mode: "anthropic",
	}
	in := strings.NewReader("我叫小明，请记住。\n我叫什么？\nexit\n")
	var out bytes.Buffer
	if err := app.runREPL(in, &out); err != nil {
		t.Fatal(err)
	}
	output := out.String()
	for _, snippet := range []string{
		"icoo REPL started (anthropic client)",
		"记住了，你叫小明。",
		"你叫小明。",
	} {
		if !strings.Contains(output, snippet) {
			t.Fatalf("expected output to contain %q, got %q", snippet, output)
		}
	}
	if len(client.Snapshots) != 2 {
		t.Fatalf("expected two client snapshots, got %d", len(client.Snapshots))
	}
	if !strings.Contains(client.Snapshots[1], "我叫小明，请记住。") {
		t.Fatalf("expected second snapshot to contain first user turn, got %q", client.Snapshots[1])
	}
	if !strings.Contains(client.Snapshots[1], "记住了，你叫小明。") {
		t.Fatalf("expected second snapshot to contain first assistant turn, got %q", client.Snapshots[1])
	}
	if !strings.Contains(client.Snapshots[1], "我叫什么？") {
		t.Fatalf("expected second snapshot to contain second user turn, got %q", client.Snapshots[1])
	}
}

func TestRunREPLStreamsWithoutDuplicatingTurnOutput(t *testing.T) {
	client := &llm.FakeClient{Responses: []llm.Response{
		{StopReason: "end", Text: "第一轮流式回复"},
		{StopReason: "end", Text: "第二轮流式回复"},
	}}
	registry, err := tools.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	app := &app{
		runner: &agent.Runner{
			Client:   client,
			Registry: registry,
			Config:   agent.Config{SystemPrompt: "test", MaxRounds: 2},
		},
		mode: "anthropic",
	}
	in := strings.NewReader("hello\nagain\nexit\n")
	var out bytes.Buffer
	if err := app.runREPL(in, &out); err != nil {
		t.Fatal(err)
	}
	output := out.String()
	for _, snippet := range []string{"第一轮流式回复", "第二轮流式回复"} {
		if strings.Count(output, snippet) != 1 {
			t.Fatalf("expected %q once, got %q", snippet, output)
		}
	}
}

func TestRunOnceExpandsProjectSlashCommand(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".icoo", "commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".icoo", "commands", "review.md"), []byte("Review the current diff carefully."), 0o644); err != nil {
		t.Fatal(err)
	}
	client := &llm.FakeClient{Responses: []llm.Response{
		{StopReason: "end", Text: "done"},
	}}
	registry, err := tools.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	app := &app{
		runner: &agent.Runner{
			Client:   client,
			Registry: registry,
			Config:   agent.Config{SystemPrompt: "test", MaxRounds: 2},
		},
		commandLoader: mustLoadCommands(t, filepath.Join(root, ".icoo", "commands")),
		mode:          "anthropic",
	}
	var out bytes.Buffer
	if err := app.runOnce(&out, "/review --focus tests"); err != nil {
		t.Fatal(err)
	}
	if len(client.Snapshots) == 0 || !strings.Contains(client.Snapshots[0], "Review the current diff carefully.") {
		t.Fatalf("expected slash command expansion in snapshot, got %#v", client.Snapshots)
	}
	if !strings.Contains(client.Snapshots[0], "--focus tests") {
		t.Fatalf("expected slash command arguments in snapshot, got %#v", client.Snapshots)
	}
}

func TestRunOnceListsProjectCommands(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".icoo", "commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".icoo", "commands", "review.md"), []byte("Review template."), 0o644); err != nil {
		t.Fatal(err)
	}
	app := &app{
		commandLoader: mustLoadCommands(t, filepath.Join(root, ".icoo", "commands")),
		mode:          "anthropic",
	}
	var out bytes.Buffer
	if err := app.runOnce(&out, "/commands"); err != nil {
		t.Fatal(err)
	}
	output := out.String()
	if !strings.Contains(output, "Project commands (1):") || !strings.Contains(output, "- /review") {
		t.Fatalf("unexpected commands output: %q", output)
	}
}

func TestRunOnceShowsProjectCommandHelp(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".icoo", "commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".icoo", "commands", "review.md"), []byte("Review template."), 0o644); err != nil {
		t.Fatal(err)
	}
	app := &app{
		commandLoader: mustLoadCommands(t, filepath.Join(root, ".icoo", "commands")),
		mode:          "anthropic",
	}
	var out bytes.Buffer
	if err := app.runOnce(&out, "/help review"); err != nil {
		t.Fatal(err)
	}
	output := out.String()
	if !strings.Contains(output, "/review") || !strings.Contains(output, "Review template.") {
		t.Fatalf("unexpected help output: %q", output)
	}
}

func TestRunOnceListsProjectAgents(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".icoo", "agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".icoo", "agents", "reviewer.md"), []byte("Review code carefully."), 0o644); err != nil {
		t.Fatal(err)
	}
	app := &app{
		agentLoader: mustLoadAgents(t, filepath.Join(root, ".icoo", "agents")),
		mode:        "anthropic",
	}
	var out bytes.Buffer
	if err := app.runOnce(&out, "/agents"); err != nil {
		t.Fatal(err)
	}
	output := out.String()
	if !strings.Contains(output, "Project agents (1):") || !strings.Contains(output, "- reviewer") {
		t.Fatalf("unexpected agents output: %q", output)
	}
}

func TestRunOnceShowsProjectAgentHelp(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".icoo", "agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".icoo", "agents", "reviewer.md"), []byte("Review code carefully."), 0o644); err != nil {
		t.Fatal(err)
	}
	app := &app{
		agentLoader: mustLoadAgents(t, filepath.Join(root, ".icoo", "agents")),
		mode:        "anthropic",
	}
	var out bytes.Buffer
	if err := app.runOnce(&out, "/help-agent reviewer"); err != nil {
		t.Fatal(err)
	}
	output := out.String()
	if !strings.Contains(output, "reviewer") || !strings.Contains(output, "Review code carefully.") {
		t.Fatalf("unexpected agent help output: %q", output)
	}
}

func mustLoadCommands(t *testing.T, dir string) *commands.Loader {
	t.Helper()
	loader, err := commands.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	return loader
}

func mustLoadAgents(t *testing.T, dir string) *agents.Loader {
	t.Helper()
	loader, err := agents.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	return loader
}

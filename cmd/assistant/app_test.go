package main

import (
	"bytes"
	"strings"
	"testing"

	"icoo_assistant/internal/agent"
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
	if !strings.Contains(out.String(), "assistant REPL started") {
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
		"assistant REPL started (fake client)",
		"warning: REPL is running in fake mode",
		"warning: no model output was produced because the fake client returns empty responses by design.",
		"hint: run `go run ./cmd/assistant check` outside the REPL",
		"hint: this is expected in fake mode",
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
		"warning: assistant is running in fake mode",
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
		"assistant REPL started (anthropic client)",
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

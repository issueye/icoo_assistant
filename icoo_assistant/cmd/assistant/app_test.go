package main

import (
	"bytes"
	"strings"
	"testing"

	"icoo_assistant/internal/config"
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
		"hint: run `assistant check`",
		"warning: no model output was produced because the fake client returns empty responses by design.",
	} {
		if !strings.Contains(output, snippet) {
			t.Fatalf("expected runOnce output to contain %q, got %q", snippet, output)
		}
	}
}

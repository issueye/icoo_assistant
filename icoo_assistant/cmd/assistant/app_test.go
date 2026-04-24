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

func TestRunOnceWithFakeClientProducesNoError(t *testing.T) {
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
}

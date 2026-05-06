package icoohook_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"icoo_assistant/internal/agent"
	"icoo_assistant/internal/icoohook"
)

func TestRunnerExecutesMatchingHook(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".icoo"), 0o755); err != nil {
		t.Fatal(err)
	}
	hooksJSON := `{
  "hooks": [
    {
      "events": ["agent.run.completed"],
      "command": "Set-Content -Path hook-fired.txt -Value done"
    }
  ]
}`
	if err := os.WriteFile(filepath.Join(root, ".icoo", "hooks.json"), []byte(hooksJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	runner, err := icoohook.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	runner.OnEvent(agent.Event{Name: "agent.run.completed"})
	if _, err := os.Stat(filepath.Join(root, "hook-fired.txt")); err != nil {
		t.Fatalf("expected hook output file, got %v", err)
	}
}

func TestRunnerInjectsEventEnv(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".icoo"), 0o755); err != nil {
		t.Fatal(err)
	}
	hooksJSON := `{
  "hooks": [
    {
      "events": ["agent.tool.started"],
      "command": "Set-Content -Path hook-env.txt -Value \"$env:ICOO_EVENT_NAME|$env:ICOO_RUN_ID|$env:ICOO_ROUND|$env:ICOO_FIELD_TOOL_NAME\""
    }
  ]
}`
	if err := os.WriteFile(filepath.Join(root, ".icoo", "hooks.json"), []byte(hooksJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	runner, err := icoohook.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	runner.OnEvent(agent.Event{
		Name:  "agent.tool.started",
		RunID: "run-42",
		Round: 3,
		Fields: map[string]interface{}{
			"tool_name": "bash",
		},
	})
	data, err := os.ReadFile(filepath.Join(root, "hook-env.txt"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if text == "" || !strings.Contains(text, "agent.tool.started|run-42|3|bash") {
		t.Fatalf("unexpected hook env output: %q", text)
	}
}

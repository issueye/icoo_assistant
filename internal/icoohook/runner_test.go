package icoohook_test

import (
	"os"
	"path/filepath"
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

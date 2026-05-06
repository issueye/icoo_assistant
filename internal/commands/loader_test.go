package commands_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"icoo_assistant/internal/commands"
)

func TestLoadReadsMarkdownCommands(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "review.md"), []byte("Review the current changes."), 0o644); err != nil {
		t.Fatal(err)
	}
	loader, err := commands.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if !loader.Has("review") {
		t.Fatal("expected review command")
	}
	rendered := loader.Render("review", "--focus tests")
	if !strings.Contains(rendered, "Review the current changes.") || !strings.Contains(rendered, "--focus tests") {
		t.Fatalf("unexpected rendered command: %q", rendered)
	}
	if loader.Body("review") != "Review the current changes." {
		t.Fatalf("unexpected command body: %q", loader.Body("review"))
	}
}

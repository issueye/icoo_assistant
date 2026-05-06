package agents_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"icoo_assistant/internal/agents"
)

func TestLoadReadsAgentTemplates(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "reviewer.md"), []byte("Review code carefully."), 0o644); err != nil {
		t.Fatal(err)
	}
	loader, err := agents.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if !loader.Has("reviewer") {
		t.Fatal("expected reviewer agent")
	}
	rendered := loader.Render("reviewer", "Inspect the auth module.")
	if !strings.Contains(rendered, "Review code carefully.") || !strings.Contains(rendered, "Inspect the auth module.") {
		t.Fatalf("unexpected rendered agent prompt: %q", rendered)
	}
	if loader.Body("reviewer") != "Review code carefully." {
		t.Fatalf("unexpected agent body: %q", loader.Body("reviewer"))
	}
}

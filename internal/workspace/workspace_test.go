package workspace_test

import (
	"path/filepath"
	"testing"

	"icoo_assistant/internal/workspace"
)

func TestResolveBlocksEscape(t *testing.T) {
	root := t.TempDir()
	ws, err := workspace.New(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ws.Resolve("../outside.txt"); err == nil {
		t.Fatal("expected escape path error")
	}
}

func TestReadWriteEditFile(t *testing.T) {
	root := t.TempDir()
	ws, err := workspace.New(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := ws.WriteFile("notes/test.txt", "hello world"); err != nil {
		t.Fatal(err)
	}
	content, err := ws.ReadFile("notes/test.txt", 0)
	if err != nil {
		t.Fatal(err)
	}
	if content != "hello world" {
		t.Fatalf("unexpected content: %q", content)
	}
	if err := ws.EditFile("notes/test.txt", "world", "golang"); err != nil {
		t.Fatal(err)
	}
	content, err = ws.ReadFile(filepath.ToSlash("notes/test.txt"), 0)
	if err != nil {
		t.Fatal(err)
	}
	if content != "hello golang" {
		t.Fatalf("unexpected edited content: %q", content)
	}
}

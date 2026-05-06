package workspace_test

import (
	"os"
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

func TestAdditionalRootsAllowExternalFiles(t *testing.T) {
	root := t.TempDir()
	external := filepath.Join(root, "external")
	if err := os.MkdirAll(external, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(external, "note.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	ws, err := workspace.NewWithOptions(root, []string{"external"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	content, err := ws.ReadFile("external/note.txt", 0)
	if err != nil {
		t.Fatal(err)
	}
	if content != "hello" {
		t.Fatalf("unexpected content: %q", content)
	}
}

func TestDeniedPatternsBlockReadAndWrite(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "private"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "secrets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "private", "token.txt"), []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	ws, err := workspace.NewWithOptions(root, nil, []string{"private/**"}, []string{"secrets/**"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ws.ReadFile("private/token.txt", 0); err == nil {
		t.Fatal("expected denied read")
	}
	if err := ws.WriteFile("secrets/new.txt", "blocked"); err == nil {
		t.Fatal("expected denied write")
	}
}

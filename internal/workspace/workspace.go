package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Workspace struct {
	Root string
}

func New(root string) (*Workspace, error) {
	if root == "" {
		return nil, errors.New("workspace root required")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	return &Workspace{Root: abs}, nil
}

func (w *Workspace) Resolve(path string) (string, error) {
	if path == "" {
		return "", errors.New("path required")
	}
	resolved := filepath.Clean(filepath.Join(w.Root, path))
	rel, err := filepath.Rel(w.Root, resolved)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes workspace: %s", path)
	}
	return resolved, nil
}

func (w *Workspace) ReadFile(path string, limit int) (string, error) {
	resolved, err := w.Resolve(path)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return "", err
	}
	text := string(data)
	if limit <= 0 {
		return text, nil
	}
	lines := strings.Split(text, "\n")
	if len(lines) <= limit {
		return text, nil
	}
	return strings.Join(append(lines[:limit], fmt.Sprintf("... (%d more lines)", len(lines)-limit)), "\n"), nil
}

func (w *Workspace) WriteFile(path, content string) error {
	resolved, err := w.Resolve(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(resolved), 0o755); err != nil {
		return err
	}
	return os.WriteFile(resolved, []byte(content), 0o644)
}

func (w *Workspace) EditFile(path, oldText, newText string) error {
	resolved, err := w.Resolve(path)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return err
	}
	content := string(data)
	if !strings.Contains(content, oldText) {
		return fmt.Errorf("text not found in %s", path)
	}
	updated := strings.Replace(content, oldText, newText, 1)
	return os.WriteFile(resolved, []byte(updated), 0o644)
}

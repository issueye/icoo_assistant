package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Workspace struct {
	Root              string
	AdditionalRoots   []string
	DenyReadPatterns  []string
	DenyWritePatterns []string
}

func New(root string) (*Workspace, error) {
	return NewWithOptions(root, nil, nil, nil)
}

func NewWithOptions(root string, additionalRoots, denyReadPatterns, denyWritePatterns []string) (*Workspace, error) {
	if root == "" {
		return nil, errors.New("workspace root required")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	roots := make([]string, 0, len(additionalRoots))
	for _, item := range additionalRoots {
		if strings.TrimSpace(item) == "" {
			continue
		}
		resolved := item
		if !filepath.IsAbs(resolved) {
			resolved = filepath.Join(abs, item)
		}
		resolved, err = filepath.Abs(resolved)
		if err != nil {
			return nil, err
		}
		roots = append(roots, filepath.Clean(resolved))
	}
	return &Workspace{
		Root:              abs,
		AdditionalRoots:   roots,
		DenyReadPatterns:  append([]string(nil), denyReadPatterns...),
		DenyWritePatterns: append([]string(nil), denyWritePatterns...),
	}, nil
}

func (w *Workspace) Resolve(path string) (string, error) {
	if path == "" {
		return "", errors.New("path required")
	}
	resolved := filepath.Clean(filepath.Join(w.Root, path))
	if !w.isAllowedPath(resolved) {
		return "", fmt.Errorf("path escapes workspace: %s", path)
	}
	return resolved, nil
}

func (w *Workspace) ReadFile(path string, limit int) (string, error) {
	resolved, err := w.Resolve(path)
	if err != nil {
		return "", err
	}
	if err := w.checkDenied("read", path, resolved, w.DenyReadPatterns); err != nil {
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
	if err := w.checkDenied("write", path, resolved, w.DenyWritePatterns); err != nil {
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
	if err := w.checkDenied("write", path, resolved, w.DenyWritePatterns); err != nil {
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

func (w *Workspace) isAllowedPath(resolved string) bool {
	for _, root := range w.allRoots() {
		rel, err := filepath.Rel(root, resolved)
		if err != nil {
			continue
		}
		if rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))) {
			return true
		}
	}
	return false
}

func (w *Workspace) allRoots() []string {
	roots := make([]string, 0, 1+len(w.AdditionalRoots))
	roots = append(roots, w.Root)
	roots = append(roots, w.AdditionalRoots...)
	return roots
}

func (w *Workspace) checkDenied(action, inputPath, resolved string, patterns []string) error {
	for _, pattern := range patterns {
		if matchWorkspacePattern(inputPath, resolved, w.allRoots(), pattern) {
			return fmt.Errorf("permission denied: %s blocked for %s", action, inputPath)
		}
	}
	return nil
}

func matchWorkspacePattern(inputPath, resolved string, roots []string, pattern string) bool {
	pattern = normalizePattern(pattern)
	candidates := []string{normalizePath(inputPath)}
	for _, root := range roots {
		rel, err := filepath.Rel(root, resolved)
		if err != nil {
			continue
		}
		if rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))) {
			rel = normalizePath(rel)
			candidates = append(candidates, rel, "./"+rel)
		}
	}
	for _, candidate := range candidates {
		if wildcardMatch(pattern, candidate) {
			return true
		}
	}
	return false
}

func normalizePattern(pattern string) string {
	pattern = strings.TrimSpace(pattern)
	pattern = strings.TrimPrefix(pattern, "./")
	return strings.ReplaceAll(pattern, "\\", "/")
}

func normalizePath(value string) string {
	value = filepath.Clean(value)
	value = strings.ReplaceAll(value, "\\", "/")
	value = strings.TrimPrefix(value, "./")
	return value
}

func wildcardMatch(pattern, value string) bool {
	pattern = normalizePattern(pattern)
	value = strings.ReplaceAll(value, "\\", "/")
	replacer := strings.NewReplacer(
		".", `\.`,
		"+", `\+`,
		"(", `\(`,
		")", `\)`,
		"[", `\[`,
		"]", `\]`,
		"{", `\{`,
		"}", `\}`,
		"^", `\^`,
		"$", `\$`,
		"|", `\|`,
	)
	regexText := replacer.Replace(pattern)
	regexText = strings.ReplaceAll(regexText, "**", "<<DOUBLESTAR>>")
	regexText = strings.ReplaceAll(regexText, "*", "[^/]*")
	regexText = strings.ReplaceAll(regexText, "<<DOUBLESTAR>>", ".*")
	regexText = strings.ReplaceAll(regexText, "?", ".")
	matched, err := regexp.MatchString("^"+regexText+"$", value)
	return err == nil && matched
}

package tools_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"icoo_assistant/internal/tools"
)

func TestSkillCreateWritesSKILLFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "skills")

	def := tools.NewSkillCreateTool(dir)
	result, err := def.Handler(tools.Call{
		Input: map[string]interface{}{
			"name":        "test-skill",
			"description": "A test skill",
			"content":     "# Test\n\nThis is a test skill.",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "skill_created: test-skill") {
		t.Fatalf("unexpected result: %q", result)
	}

	path := filepath.Join(dir, "test-skill", "SKILL.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "name: test-skill") {
		t.Fatal("missing name in frontmatter")
	}
	if !strings.Contains(content, "description: A test skill") {
		t.Fatal("missing description in frontmatter")
	}
	if !strings.Contains(content, "# Test") {
		t.Fatal("missing body content")
	}
}

func TestSkillCreateRejectsEmptyName(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "skills")
	def := tools.NewSkillCreateTool(dir)
	_, err := def.Handler(tools.Call{
		Input: map[string]interface{}{
			"name":    "",
			"content": "some content",
		},
	})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestSkillCreateRejectsEmptyContent(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "skills")
	def := tools.NewSkillCreateTool(dir)
	_, err := def.Handler(tools.Call{
		Input: map[string]interface{}{
			"name":    "test",
			"content": "",
		},
	})
	if err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestSkillCreateUsesNameAsDescriptionWhenEmpty(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "skills")
	def := tools.NewSkillCreateTool(dir)
	_, err := def.Handler(tools.Call{
		Input: map[string]interface{}{
			"name":    "auto-desc",
			"content": "Content here",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "auto-desc", "SKILL.md")
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "description: auto-desc") {
		t.Fatal("expected auto description")
	}
}

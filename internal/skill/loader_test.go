package skill_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"icoo_assistant/internal/skill"
)

func TestLoadReadsSkillMetadataAndBody(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, "code-review")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: code-review\ndescription: Review code carefully\n---\nAlways inspect risky changes first."
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	loader, err := skill.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(loader.Descriptions(), "code-review: Review code carefully") {
		t.Fatalf("unexpected descriptions: %q", loader.Descriptions())
	}
	if !strings.Contains(loader.Load("code-review"), "Always inspect risky changes first.") {
		t.Fatalf("unexpected loaded skill content: %q", loader.Load("code-review"))
	}
}

func TestLoadAllMergesMultipleSkillDirectories(t *testing.T) {
	rootA := t.TempDir()
	rootB := t.TempDir()
	if err := os.MkdirAll(filepath.Join(rootA, "review"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(rootB, "ship"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rootA, "review", "SKILL.md"), []byte("---\nname: review\ndescription: Review\n---\nReview."), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rootB, "ship", "SKILL.md"), []byte("---\nname: ship\ndescription: Ship\n---\nShip."), 0o644); err != nil {
		t.Fatal(err)
	}
	loader, err := skill.LoadAll(rootA, rootB)
	if err != nil {
		t.Fatal(err)
	}
	descriptions := loader.Descriptions()
	if !strings.Contains(descriptions, "review: Review") || !strings.Contains(descriptions, "ship: Ship") {
		t.Fatalf("unexpected merged descriptions: %q", descriptions)
	}
}

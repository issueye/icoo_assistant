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

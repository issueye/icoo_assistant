package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"icoo_assistant/internal/llm"
)

type SkillCreator struct {
	SkillsDir string
}

func NewSkillCreateTool(skillsDir string) Definition {
	creator := &SkillCreator{SkillsDir: skillsDir}
	return Definition{
		Tool: llm.Tool{
			Name:        "skill_create",
			Description: "Create a new domain skill by writing a SKILL.md file under the skills directory. Use to add specialized knowledge, workflows, or tool integrations.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name":        map[string]interface{}{"type": "string", "description": "Skill name (also used as directory name)"},
					"description": map[string]interface{}{"type": "string", "description": "Brief description shown in skill listing"},
					"content":     map[string]interface{}{"type": "string", "description": "Full markdown content of the skill (body, without YAML frontmatter)"},
				},
				"required": []string{"name", "content"},
			},
		},
		Handler: func(call Call) (string, error) {
			name, _ := call.Input["name"].(string)
			name = strings.TrimSpace(name)
			if name == "" {
				return "", fmt.Errorf("name required")
			}
			description, _ := call.Input["description"].(string)
			description = strings.TrimSpace(description)
			if description == "" {
				description = name
			}
			content, _ := call.Input["content"].(string)
			content = strings.TrimSpace(content)
			if content == "" {
				return "", fmt.Errorf("content required")
			}
			return creator.create(name, description, content)
		},
	}
}

func (c *SkillCreator) create(name, description, content string) (string, error) {
	if c.SkillsDir == "" {
		return "", fmt.Errorf("skills directory not configured")
	}
	if err := os.MkdirAll(c.SkillsDir, 0o755); err != nil {
		return "", fmt.Errorf("create skills dir: %w", err)
	}

	dir := filepath.Join(c.SkillsDir, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create skill dir: %w", err)
	}

	path := filepath.Join(dir, "SKILL.md")
	var builder strings.Builder
	builder.WriteString("---\n")
	builder.WriteString(fmt.Sprintf("name: %s\n", name))
	builder.WriteString(fmt.Sprintf("description: %s\n", description))
	builder.WriteString("---\n\n")
	builder.WriteString(content)
	builder.WriteString("\n")

	if err := os.WriteFile(path, []byte(builder.String()), 0o644); err != nil {
		return "", fmt.Errorf("write skill file: %w", err)
	}

	return fmt.Sprintf("skill_created: %s\ndescription: %s\npath: %s\nhint: use skill_create again to update, or edit the file directly.", name, description, path), nil
}

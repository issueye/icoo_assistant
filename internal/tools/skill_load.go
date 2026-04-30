package tools

import (
	"fmt"

	"icoo_assistant/internal/llm"
	"icoo_assistant/internal/skill"
)

func NewSkillLoadTool(loader *skill.Loader) Definition {
	return Definition{
		Tool: llm.Tool{
			Name:        "skill_load",
			Description: "Load specialized knowledge by name.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{"type": "string"},
				},
				"required": []string{"name"},
			},
		},
		Handler: func(call Call) (string, error) {
			name, ok := call.Input["name"].(string)
			if !ok || name == "" {
				return "", fmt.Errorf("name required")
			}
			return loader.Load(name), nil
		},
	}
}

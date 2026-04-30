package tools

import "icoo_assistant/internal/llm"

func NewSkillExecuteTool() Definition {
	return Definition{
		Tool: llm.Tool{
			Name:        "skill_execute",
			Description: "Execute a skill in a subagent. Loads the skill content and runs the task with fresh context, returning only the summary. Use this instead of skill_load when the skill task would generate many intermediate messages.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name":   map[string]interface{}{"type": "string", "description": "Skill name to load and execute"},
					"prompt": map[string]interface{}{"type": "string", "description": "Task description for the subagent"},
				},
				"required": []string{"name", "prompt"},
			},
		},
		Handler: func(call Call) (string, error) {
			return "skill_subagent_requested", nil
		},
	}
}

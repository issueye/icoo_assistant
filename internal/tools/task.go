package tools

import "icoo_assistant/internal/llm"

func NewTaskTool() Definition {
	return Definition{
		Tool: llm.Tool{
			Name:        "task",
			Description: "Spawn a subagent with fresh context and return only its summary.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"prompt": map[string]interface{}{"type": "string"},
				},
				"required": []string{"prompt"},
			},
		},
		Handler: func(call Call) (string, error) {
			return "subagent_requested", nil
		},
	}
}

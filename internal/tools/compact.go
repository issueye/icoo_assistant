package tools

import "icoo_assistant/internal/llm"

func NewCompactTool() Definition {
	return Definition{
		Tool: llm.Tool{
			Name:        "compact",
			Description: "Compact the conversation context for continuity.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		Handler: func(call Call) (string, error) {
			return "compact_requested", nil
		},
	}
}

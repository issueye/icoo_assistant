package tools

import (
	"fmt"

	"icoo_assistant/internal/llm"
	"icoo_assistant/internal/todo"
)

func NewTodoTool(manager *todo.Manager) Definition {
	return Definition{
		Tool: llm.Tool{
			Name:        "todo",
			Description: "Update task list. Track progress on multi-step tasks.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"items": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"id":     map[string]interface{}{"type": "string"},
								"text":   map[string]interface{}{"type": "string"},
								"status": map[string]interface{}{"type": "string", "enum": []string{"pending", "in_progress", "completed"}},
							},
							"required": []string{"text", "status"},
						},
					},
				},
				"required": []string{"items"},
			},
		},
		Handler: func(call Call) (string, error) {
			rawItems, ok := call.Input["items"].([]interface{})
			if !ok {
				return "", fmt.Errorf("items required")
			}
			items := make([]todo.Item, 0, len(rawItems))
			for _, raw := range rawItems {
				itemMap, ok := raw.(map[string]interface{})
				if !ok {
					return "", fmt.Errorf("invalid todo item")
				}
				item := todo.Item{}
				if value, ok := itemMap["id"].(string); ok {
					item.ID = value
				}
				if value, ok := itemMap["text"].(string); ok {
					item.Text = value
				}
				if value, ok := itemMap["status"].(string); ok {
					item.Status = value
				}
				items = append(items, item)
			}
			return manager.Update(items)
		},
	}
}

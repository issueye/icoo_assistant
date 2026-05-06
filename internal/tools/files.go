package tools

import (
	"fmt"

	"icoo_assistant/internal/llm"
	"icoo_assistant/internal/workspace"
)

func NewReadFileTool(ws *workspace.Workspace) Definition {
	return Definition{
		Tool: llm.Tool{
			Name:        "read_file",
			Description: "Read file contents.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":  map[string]interface{}{"type": "string"},
					"limit": map[string]interface{}{"type": "integer"},
				},
				"required": []string{"path"},
			},
		},
		Handler: func(call Call) (string, error) {
			path, ok := call.Input["path"].(string)
			if !ok || path == "" {
				return "", fmt.Errorf("path required")
			}
			limit := 0
			if raw, ok := call.Input["limit"]; ok {
				switch v := raw.(type) {
				case float64:
					limit = int(v)
				case int:
					limit = v
				}
			}
			return ws.ReadFile(path, limit)
		},
	}
}

func NewWriteFileTool(ws *workspace.Workspace) Definition {
	return Definition{
		Tool: llm.Tool{
			Name:        "write_file",
			Description: "Write content to file.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":    map[string]interface{}{"type": "string"},
					"content": map[string]interface{}{"type": "string"},
				},
				"required": []string{"path", "content"},
			},
		},
		Handler: func(call Call) (string, error) {
			path, ok := call.Input["path"].(string)
			if !ok || path == "" {
				return "", fmt.Errorf("path required")
			}
			content, ok := call.Input["content"].(string)
			if !ok {
				return "", fmt.Errorf("content required")
			}
			if err := ws.WriteFile(path, content); err != nil {
				return "", err
			}
			return fmt.Sprintf("Wrote %d bytes to %s", len(content), path), nil
		},
	}
}

func NewEditFileTool(ws *workspace.Workspace) Definition {
	return Definition{
		Tool: llm.Tool{
			Name:        "edit_file",
			Description: "Replace exact text in file.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":     map[string]interface{}{"type": "string"},
					"old_text": map[string]interface{}{"type": "string"},
					"new_text": map[string]interface{}{"type": "string"},
				},
				"required": []string{"path", "old_text", "new_text"},
			},
		},
		Handler: func(call Call) (string, error) {
			path, ok := call.Input["path"].(string)
			if !ok || path == "" {
				return "", fmt.Errorf("path required")
			}
			oldText, ok := call.Input["old_text"].(string)
			if !ok {
				return "", fmt.Errorf("old_text required")
			}
			newText, ok := call.Input["new_text"].(string)
			if !ok {
				return "", fmt.Errorf("new_text required")
			}
			if err := ws.EditFile(path, oldText, newText); err != nil {
				return "", err
			}
			return fmt.Sprintf("Edited %s", path), nil
		},
	}
}

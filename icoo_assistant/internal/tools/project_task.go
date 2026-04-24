package tools

import (
	"fmt"
	"strings"

	"icoo_assistant/internal/llm"
	"icoo_assistant/internal/task"
)

type ProjectTaskManager interface {
	Create(input task.CreateInput) (task.Task, error)
	Get(id string) (task.Task, error)
	List() ([]task.Task, error)
	Update(item task.Task) (task.Task, error)
	UpdateStatus(id, status string) (task.Task, error)
}

func NewProjectTaskTool(manager ProjectTaskManager) Definition {
	return Definition{
		Tool: llm.Tool{
			Name:        "project_task",
			Description: "Create, inspect, list, and update project-level persistent tasks.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action":     map[string]interface{}{"type": "string", "enum": []string{"create", "get", "list", "update", "update_status"}},
					"id":         map[string]interface{}{"type": "string"},
					"title":      map[string]interface{}{"type": "string"},
					"status":     map[string]interface{}{"type": "string"},
					"blocked_by": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					"owner":      map[string]interface{}{"type": "string"},
					"worktree":   map[string]interface{}{"type": "string"},
				},
				"required": []string{"action"},
			},
		},
		Handler: func(call Call) (string, error) {
			action, _ := call.Input["action"].(string)
			switch strings.ToLower(strings.TrimSpace(action)) {
			case "create":
				title, _ := call.Input["title"].(string)
				if strings.TrimSpace(title) == "" {
					return "", fmt.Errorf("title required for create")
				}
				id, _ := call.Input["id"].(string)
				status, _ := call.Input["status"].(string)
				owner, _ := call.Input["owner"].(string)
				worktree, _ := call.Input["worktree"].(string)
				blockedBy, err := stringListFromInput(call.Input["blocked_by"])
				if err != nil {
					return "", err
				}
				item, err := manager.Create(task.CreateInput{
					ID:        id,
					Title:     title,
					Status:    status,
					BlockedBy: blockedBy,
					Owner:     owner,
					Worktree:  worktree,
				})
				if err != nil {
					return "", err
				}
				return renderProjectTask(item), nil
			case "get":
				id, _ := call.Input["id"].(string)
				if strings.TrimSpace(id) == "" {
					return "", fmt.Errorf("id required for get")
				}
				item, err := manager.Get(id)
				if err != nil {
					return "", err
				}
				return renderProjectTask(item), nil
			case "list":
				items, err := manager.List()
				if err != nil {
					return "", err
				}
				if len(items) == 0 {
					return "No project tasks.", nil
				}
				lines := make([]string, 0, len(items))
				for _, item := range items {
					lines = append(lines, fmt.Sprintf("%s [%s] %s", item.ID, item.Status, item.Title))
				}
				return strings.Join(lines, "\n"), nil
			case "update":
				id, _ := call.Input["id"].(string)
				if strings.TrimSpace(id) == "" {
					return "", fmt.Errorf("id required for update")
				}
				current, err := manager.Get(id)
				if err != nil {
					return "", err
				}
				if title, ok := call.Input["title"].(string); ok && strings.TrimSpace(title) != "" {
					current.Title = title
				}
				if status, ok := call.Input["status"].(string); ok && strings.TrimSpace(status) != "" {
					current.Status = status
				}
				if owner, ok := call.Input["owner"].(string); ok {
					current.Owner = owner
				}
				if worktree, ok := call.Input["worktree"].(string); ok {
					current.Worktree = worktree
				}
				if raw, exists := call.Input["blocked_by"]; exists {
					blockedBy, err := stringListFromInput(raw)
					if err != nil {
						return "", err
					}
					current.BlockedBy = blockedBy
				}
				item, err := manager.Update(current)
				if err != nil {
					return "", err
				}
				return renderProjectTask(item), nil
			case "update_status":
				id, _ := call.Input["id"].(string)
				status, _ := call.Input["status"].(string)
				if strings.TrimSpace(id) == "" {
					return "", fmt.Errorf("id required for update_status")
				}
				if strings.TrimSpace(status) == "" {
					return "", fmt.Errorf("status required for update_status")
				}
				item, err := manager.UpdateStatus(id, status)
				if err != nil {
					return "", err
				}
				return renderProjectTask(item), nil
			default:
				return "", fmt.Errorf("unsupported action %q", action)
			}
		},
	}
}

func stringListFromInput(raw interface{}) ([]string, error) {
	if raw == nil {
		return nil, nil
	}
	items, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected string list")
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		value, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("expected string list")
		}
		result = append(result, value)
	}
	return result, nil
}

func renderProjectTask(item task.Task) string {
	lines := []string{
		fmt.Sprintf("id: %s", item.ID),
		fmt.Sprintf("title: %s", item.Title),
		fmt.Sprintf("status: %s", item.Status),
	}
	if len(item.BlockedBy) > 0 {
		lines = append(lines, fmt.Sprintf("blocked_by: %s", strings.Join(item.BlockedBy, ", ")))
	}
	if item.Owner != "" {
		lines = append(lines, fmt.Sprintf("owner: %s", item.Owner))
	}
	if item.Worktree != "" {
		lines = append(lines, fmt.Sprintf("worktree: %s", item.Worktree))
	}
	lines = append(lines, fmt.Sprintf("created_at: %s", item.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")))
	lines = append(lines, fmt.Sprintf("updated_at: %s", item.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z")))
	return strings.Join(lines, "\n")
}

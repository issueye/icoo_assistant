package tools

import (
	"fmt"
	"strings"

	"icoo_assistant/internal/llm"
	"icoo_assistant/internal/task"
)

type TaskAuditManager interface {
	Get(id string) (task.Task, error)
}

func NewTaskAuditTool(manager TaskAuditManager) Definition {
	return Definition{
		Tool: llm.Tool{
			Name:        "task_audit",
			Description: "Inspect project task execution audit data such as background history.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action": map[string]interface{}{"type": "string", "enum": []string{"history"}},
					"id":     map[string]interface{}{"type": "string"},
					"limit":  map[string]interface{}{"type": "integer"},
				},
				"required": []string{"action", "id"},
			},
		},
		Handler: func(call Call) (string, error) {
			action, _ := call.Input["action"].(string)
			switch strings.ToLower(strings.TrimSpace(action)) {
			case "history":
				id, _ := call.Input["id"].(string)
				if strings.TrimSpace(id) == "" {
					return "", fmt.Errorf("id required for history")
				}
				item, err := manager.Get(id)
				if err != nil {
					return "", err
				}
				limit := intFromInput(call.Input["limit"], 10)
				return renderTaskAuditHistory(item, limit), nil
			default:
				return "", fmt.Errorf("unsupported action %q", action)
			}
		},
	}
}

func renderTaskAuditHistory(item task.Task, limit int) string {
	lines := []string{
		fmt.Sprintf("task_id: %s", item.ID),
		fmt.Sprintf("title: %s", item.Title),
		fmt.Sprintf("history_count: %d", len(item.BackgroundHistory)),
		fmt.Sprintf("returned_count: %d", len(recentBackgroundHistory(item.BackgroundHistory, limit))),
	}
	if len(item.BackgroundHistory) == 0 {
		lines = append(lines, "entries: none")
		lines = append(lines, fmt.Sprintf("latest_task_view: project_task action=get id=%s", item.ID))
		lines = append(lines, `runtime_view_hint: use agent_hook_audit action=recent or tool_catalog action=audit_paths for runtime-side investigation`)
		return strings.Join(lines, "\n")
	}
	lines = append(lines, "entries:")
	for index, entry := range recentBackgroundHistory(item.BackgroundHistory, limit) {
		line := fmt.Sprintf("%d. job_id=%s status=%s updated_at=%s", index+1, entry.JobID, entry.Status, entry.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"))
		if entry.Command != "" {
			line = fmt.Sprintf("%s command=%s", line, entry.Command)
		}
		if entry.Error != "" {
			line = fmt.Sprintf("%s error=%s", line, entry.Error)
		}
		lines = append(lines, line)
	}
	lines = append(lines, fmt.Sprintf("latest_task_view: project_task action=get id=%s", item.ID))
	lines = append(lines, `runtime_view_hint: use agent_hook_audit action=recent name=agent.tool.completed to inspect runtime-side execution context`)
	return strings.Join(lines, "\n")
}

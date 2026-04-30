package tools

import (
	"fmt"
	"strings"

	"icoo_assistant/internal/background"
	"icoo_assistant/internal/llm"
)

type BackgroundManager interface {
	Start(input background.StartInput) (background.Job, error)
	Get(id string) (background.Job, error)
	List() ([]background.Job, error)
	ListByTaskID(taskID string) ([]background.Job, error)
}

func NewBackgroundTool(manager BackgroundManager) Definition {
	return Definition{
		Tool: llm.Tool{
			Name:        "background",
			Description: "Start or inspect background shell commands.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action":  map[string]interface{}{"type": "string", "enum": []string{"start", "get", "list"}},
					"id":      map[string]interface{}{"type": "string"},
					"command": map[string]interface{}{"type": "string"},
					"task_id": map[string]interface{}{"type": "string"},
					"owner":   map[string]interface{}{"type": "string"},
				},
				"required": []string{"action"},
			},
		},
		Handler: func(call Call) (string, error) {
			action, _ := call.Input["action"].(string)
			switch strings.ToLower(strings.TrimSpace(action)) {
			case "start":
				command, _ := call.Input["command"].(string)
				if strings.TrimSpace(command) == "" {
					return "", fmt.Errorf("command required for start")
				}
				id, _ := call.Input["id"].(string)
				taskID, _ := call.Input["task_id"].(string)
				owner, _ := call.Input["owner"].(string)
				job, err := manager.Start(background.StartInput{
					ID:      id,
					Command: command,
					TaskID:  taskID,
					Owner:   owner,
				})
				if err != nil {
					if strings.Contains(err.Error(), "dangerous command blocked") {
						return "Error: Dangerous command blocked", nil
					}
					return "", err
				}
				if job.TaskID != "" {
					return fmt.Sprintf("Started background job %s for task %s: %s", job.ID, job.TaskID, job.Command), nil
				}
				return fmt.Sprintf("Started background job %s for command: %s", job.ID, job.Command), nil
			case "get":
				id, _ := call.Input["id"].(string)
				if strings.TrimSpace(id) == "" {
					return "", fmt.Errorf("id required for get")
				}
				job, err := manager.Get(id)
				if err != nil {
					return "", err
				}
				return renderBackgroundJob(job), nil
			case "list":
				taskID, _ := call.Input["task_id"].(string)
				var jobs []background.Job
				var err error
				if strings.TrimSpace(taskID) != "" {
					jobs, err = manager.ListByTaskID(taskID)
				} else {
					jobs, err = manager.List()
				}
				if err != nil {
					return "", err
				}
				if len(jobs) == 0 {
					if strings.TrimSpace(taskID) != "" {
						return fmt.Sprintf("No background jobs for task %s.", taskID), nil
					}
					return "No background jobs.", nil
				}
				lines := make([]string, 0, len(jobs))
				for _, job := range jobs {
					line := fmt.Sprintf("%s [%s] %s", job.ID, job.Status, job.Command)
					if job.TaskID != "" {
						line = fmt.Sprintf("%s (task: %s)", line, job.TaskID)
					}
					lines = append(lines, line)
				}
				return strings.Join(lines, "\n"), nil
			default:
				return "", fmt.Errorf("unsupported action %q", action)
			}
		},
	}
}

func renderBackgroundJob(job background.Job) string {
	lines := []string{
		fmt.Sprintf("id: %s", job.ID),
		fmt.Sprintf("status: %s", job.Status),
		fmt.Sprintf("command: %s", job.Command),
	}
	if job.TaskID != "" {
		lines = append(lines, fmt.Sprintf("task_id: %s", job.TaskID))
	}
	if job.Owner != "" {
		lines = append(lines, fmt.Sprintf("owner: %s", job.Owner))
	}
	if job.Error != "" {
		lines = append(lines, fmt.Sprintf("error: %s", job.Error))
	}
	if job.Output != "" {
		lines = append(lines, "output:")
		lines = append(lines, job.Output)
	}
	return strings.Join(lines, "\n")
}

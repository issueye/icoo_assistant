package tools

import (
	"fmt"
	"sort"
	"strings"

	"icoo_assistant/internal/hookaudit"
	"icoo_assistant/internal/llm"
)

type AgentHookAuditReader interface {
	Recent(query hookaudit.Query) ([]hookaudit.Event, error)
}

func NewAgentHookAuditTool(reader AgentHookAuditReader) Definition {
	return Definition{
		Tool: llm.Tool{
			Name:        "agent_hook_audit",
			Description: "Inspect recorded agent hook events such as recent runs, tool calls, and notifications.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action": map[string]interface{}{"type": "string", "enum": []string{"recent"}},
					"limit":  map[string]interface{}{"type": "integer"},
					"name":   map[string]interface{}{"type": "string"},
					"run_id": map[string]interface{}{"type": "string"},
				},
				"required": []string{"action"},
			},
		},
		Handler: func(call Call) (string, error) {
			action, _ := call.Input["action"].(string)
			switch strings.ToLower(strings.TrimSpace(action)) {
			case "recent":
				query := hookaudit.Query{
					Limit: intFromInput(call.Input["limit"], 10),
					Name:  strings.TrimSpace(stringFromInput(call.Input["name"])),
					RunID: strings.TrimSpace(stringFromInput(call.Input["run_id"])),
				}
				events, err := reader.Recent(query)
				if err != nil {
					return "", err
				}
				return renderAgentHookAuditRecent(events, query), nil
			default:
				return "", fmt.Errorf("unsupported action %q", action)
			}
		},
	}
}

func renderAgentHookAuditRecent(events []hookaudit.Event, query hookaudit.Query) string {
	lines := []string{
		fmt.Sprintf("returned_count: %d", len(events)),
		fmt.Sprintf("limit: %d", query.Limit),
	}
	if query.Name != "" {
		lines = append(lines, fmt.Sprintf("filter_name: %s", query.Name))
	}
	if query.RunID != "" {
		lines = append(lines, fmt.Sprintf("filter_run_id: %s", query.RunID))
	}
	if len(events) == 0 {
		lines = append(lines, "entries: none")
		lines = append(lines, `task_history_hint: use task_audit action=history id=<task-id> for durable project task execution history`)
		lines = append(lines, `navigation_hint: use tool_catalog action=audit_paths for audit entry-point guidance`)
		return strings.Join(lines, "\n")
	}
	lines = append(lines, "entries:")
	for index, event := range events {
		line := fmt.Sprintf("%d. timestamp=%s name=%s run_id=%s", index+1, event.Timestamp.UTC().Format("2006-01-02T15:04:05Z"), event.Name, event.RunID)
		if event.Round > 0 {
			line = fmt.Sprintf("%s round=%d", line, event.Round)
		}
		if rendered := renderAgentHookFields(event.Fields); rendered != "" {
			line = fmt.Sprintf("%s fields=%s", line, rendered)
		}
		lines = append(lines, line)
	}
	lines = append(lines, `task_history_hint: use task_audit action=history id=<task-id> when a durable task history review is needed`)
	lines = append(lines, `navigation_hint: use tool_catalog action=audit_paths for audit entry-point guidance`)
	return strings.Join(lines, "\n")
}

func renderAgentHookFields(fields map[string]interface{}) string {
	if len(fields) == 0 {
		return ""
	}
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", key, fields[key]))
	}
	return strings.Join(parts, ",")
}

func stringFromInput(raw interface{}) string {
	value, _ := raw.(string)
	return value
}

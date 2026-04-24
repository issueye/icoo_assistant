package tools

import (
	"fmt"
	"strings"

	"icoo_assistant/internal/llm"
	"icoo_assistant/internal/team"
)

type TeamProtocolManager interface {
	GetRequest(requestID string) (team.RequestRecord, error)
	ListRequests(filter team.RequestFilter, limit int) ([]team.RequestRecord, error)
}

func NewTeamProtocolTool(manager TeamProtocolManager) Definition {
	return Definition{
		Tool: llm.Tool{
			Name:        "team_protocol",
			Description: "Inspect durable team request lifecycle records stored under .team/requests.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action":     map[string]interface{}{"type": "string", "enum": []string{"get", "list", "summary"}},
					"request_id": map[string]interface{}{"type": "string"},
					"status":     map[string]interface{}{"type": "string"},
					"from":       map[string]interface{}{"type": "string"},
					"to":         map[string]interface{}{"type": "string"},
					"limit":      map[string]interface{}{"type": "integer"},
				},
				"required": []string{"action"},
			},
		},
		Handler: func(call Call) (string, error) {
			action, _ := call.Input["action"].(string)
			switch strings.ToLower(strings.TrimSpace(action)) {
			case "get":
				requestID, _ := call.Input["request_id"].(string)
				if strings.TrimSpace(requestID) == "" {
					return "", fmt.Errorf("request_id required for get")
				}
				item, err := manager.GetRequest(requestID)
				if err != nil {
					return "", err
				}
				return renderTeamRequest(item), nil
			case "list":
				filter, limit := protocolFilterFromInput(call.Input)
				items, err := manager.ListRequests(filter, limit)
				if err != nil {
					return "", err
				}
				if len(items) == 0 {
					return "No protocol requests found.", nil
				}
				lines := []string{
					fmt.Sprintf("request_count: %d", len(items)),
					"requests:",
				}
				for _, item := range items {
					lines = append(lines, renderTeamRequestLine(item))
				}
				return strings.Join(lines, "\n"), nil
			case "summary":
				filter, _ := protocolFilterFromInput(call.Input)
				items, err := manager.ListRequests(filter, 0)
				if err != nil {
					return "", err
				}
				pendingCount := 0
				respondedCount := 0
				for _, item := range items {
					switch item.Status {
					case team.RequestStatusPending:
						pendingCount++
					case team.RequestStatusResponded:
						respondedCount++
					}
				}
				lines := []string{
					fmt.Sprintf("request_count: %d", len(items)),
					fmt.Sprintf("pending_count: %d", pendingCount),
					fmt.Sprintf("responded_count: %d", respondedCount),
				}
				if filter.FromID != "" {
					lines = append(lines, fmt.Sprintf("from: %s", filter.FromID))
				}
				if filter.ToID != "" {
					lines = append(lines, fmt.Sprintf("to: %s", filter.ToID))
				}
				if filter.Status != "" {
					lines = append(lines, fmt.Sprintf("status: %s", filter.Status))
				}
				return strings.Join(lines, "\n"), nil
			default:
				return "", fmt.Errorf("unsupported action %q", action)
			}
		},
	}
}

func protocolFilterFromInput(input map[string]interface{}) (team.RequestFilter, int) {
	status, _ := input["status"].(string)
	fromID, _ := input["from"].(string)
	toID, _ := input["to"].(string)
	return team.RequestFilter{
		Status: status,
		FromID: fromID,
		ToID:   toID,
	}, intFromInput(input["limit"], 20)
}

func renderTeamRequest(item team.RequestRecord) string {
	lines := []string{
		fmt.Sprintf("request_id: %s", item.RequestID),
		fmt.Sprintf("from: %s", item.FromID),
		fmt.Sprintf("to: %s", item.ToID),
		fmt.Sprintf("kind: %s", item.Kind),
		fmt.Sprintf("body: %s", item.Body),
		fmt.Sprintf("status: %s", item.Status),
		fmt.Sprintf("root_message_id: %s", item.RootMessageID),
	}
	if item.ResponseMessageID != "" {
		lines = append(lines, fmt.Sprintf("response_message_id: %s", item.ResponseMessageID))
	}
	lines = append(
		lines,
		fmt.Sprintf("created_at: %s", item.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")),
		fmt.Sprintf("updated_at: %s", item.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z")),
	)
	return strings.Join(lines, "\n")
}

func renderTeamRequestLine(item team.RequestRecord) string {
	line := fmt.Sprintf("- %s from=%s to=%s status=%s kind=%s body=%s", item.RequestID, item.FromID, item.ToID, item.Status, item.Kind, item.Body)
	if item.ResponseMessageID != "" {
		line = fmt.Sprintf("%s response_message_id=%s", line, item.ResponseMessageID)
	}
	return fmt.Sprintf("%s updated_at=%s", line, item.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"))
}

package tools

import (
	"fmt"
	"strings"
	"time"

	"icoo_assistant/internal/llm"
	"icoo_assistant/internal/team"
)

type TeamMessageManager interface {
	SendMessage(input team.SendMessageInput) (team.Message, error)
	ListInbox(recipientID string, limit int) ([]team.Message, error)
	ReplyToRequest(input team.ReplyInput) (team.Message, error)
	ListThread(requestID string, limit int) ([]team.Message, error)
	GetRequest(requestID string) (team.RequestRecord, error)
}

func NewTeamMessageTool(manager TeamMessageManager) Definition {
	return Definition{
		Tool: llm.Tool{
			Name:        "team_message",
			Description: "Write to and inspect persistent team inbox files under .team/inbox, including minimal request/response threads.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action":     map[string]interface{}{"type": "string", "enum": []string{"send", "request", "reply", "inbox", "thread"}},
					"id":         map[string]interface{}{"type": "string"},
					"from":       map[string]interface{}{"type": "string"},
					"to":         map[string]interface{}{"type": "string"},
					"recipient":  map[string]interface{}{"type": "string"},
					"kind":       map[string]interface{}{"type": "string"},
					"body":       map[string]interface{}{"type": "string"},
					"request_id": map[string]interface{}{"type": "string"},
					"limit":      map[string]interface{}{"type": "integer"},
				},
				"required": []string{"action"},
			},
		},
		Handler: func(call Call) (string, error) {
			action, _ := call.Input["action"].(string)
			switch strings.ToLower(strings.TrimSpace(action)) {
			case "send":
				fromID, _ := call.Input["from"].(string)
				toID, _ := call.Input["to"].(string)
				body, _ := call.Input["body"].(string)
				kind, _ := call.Input["kind"].(string)
				id, _ := call.Input["id"].(string)
				requestID, _ := call.Input["request_id"].(string)
				if strings.TrimSpace(fromID) == "" {
					fromID = "lead"
				}
				if strings.TrimSpace(toID) == "" {
					return "", fmt.Errorf("to required for send")
				}
				item, err := manager.SendMessage(team.SendMessageInput{
					ID:        id,
					FromID:    fromID,
					ToID:      toID,
					Kind:      kind,
					Body:      body,
					RequestID: requestID,
				})
				if err != nil {
					return "", err
				}
				return renderTeamMessage(item), nil
			case "request":
				fromID, _ := call.Input["from"].(string)
				toID, _ := call.Input["to"].(string)
				body, _ := call.Input["body"].(string)
				id, _ := call.Input["id"].(string)
				requestID, _ := call.Input["request_id"].(string)
				if strings.TrimSpace(fromID) == "" {
					fromID = "lead"
				}
				if strings.TrimSpace(toID) == "" {
					return "", fmt.Errorf("to required for request")
				}
				if strings.TrimSpace(requestID) == "" {
					requestID = fmt.Sprintf("req-%d", time.Now().UTC().UnixNano())
				}
				item, err := manager.SendMessage(team.SendMessageInput{
					ID:        id,
					FromID:    fromID,
					ToID:      toID,
					Kind:      "request",
					Body:      body,
					RequestID: requestID,
				})
				if err != nil {
					return "", err
				}
				return renderTeamMessageWithProtocol(manager, item), nil
			case "reply":
				fromID, _ := call.Input["from"].(string)
				requestID, _ := call.Input["request_id"].(string)
				body, _ := call.Input["body"].(string)
				id, _ := call.Input["id"].(string)
				if strings.TrimSpace(fromID) == "" {
					return "", fmt.Errorf("from required for reply")
				}
				item, err := manager.ReplyToRequest(team.ReplyInput{
					ID:        id,
					FromID:    fromID,
					RequestID: requestID,
					Body:      body,
				})
				if err != nil {
					return "", err
				}
				return renderTeamMessageWithProtocol(manager, item), nil
			case "inbox":
				recipient, _ := call.Input["recipient"].(string)
				if strings.TrimSpace(recipient) == "" {
					recipient = "lead"
				}
				items, err := manager.ListInbox(recipient, intFromInput(call.Input["limit"], 20))
				if err != nil {
					return "", err
				}
				if len(items) == 0 {
					return fmt.Sprintf("No inbox messages for %s.", recipient), nil
				}
				lines := []string{
					fmt.Sprintf("recipient: %s", recipient),
					fmt.Sprintf("message_count: %d", len(items)),
					"messages:",
				}
				for _, item := range items {
					lines = append(lines, renderTeamMessageLine(item))
				}
				return strings.Join(lines, "\n"), nil
			case "thread":
				requestID, _ := call.Input["request_id"].(string)
				if strings.TrimSpace(requestID) == "" {
					return "", fmt.Errorf("request_id required for thread")
				}
				items, err := manager.ListThread(requestID, intFromInput(call.Input["limit"], 20))
				if err != nil {
					return "", err
				}
				if len(items) == 0 {
					return fmt.Sprintf("No thread messages for %s.", requestID), nil
				}
				lines := []string{
					fmt.Sprintf("request_id: %s", requestID),
					fmt.Sprintf("message_count: %d", len(items)),
					"messages:",
				}
				for _, item := range items {
					lines = append(lines, renderTeamMessageLine(item))
				}
				return strings.Join(lines, "\n"), nil
			default:
				return "", fmt.Errorf("unsupported action %q", action)
			}
		},
	}
}

func renderTeamMessage(item team.Message) string {
	lines := []string{
		fmt.Sprintf("id: %s", item.ID),
		fmt.Sprintf("from: %s", item.FromID),
		fmt.Sprintf("to: %s", item.ToID),
		fmt.Sprintf("kind: %s", item.Kind),
		fmt.Sprintf("body: %s", item.Body),
	}
	if item.RequestID != "" {
		lines = append(lines, fmt.Sprintf("request_id: %s", item.RequestID))
	}
	lines = append(lines, fmt.Sprintf("created_at: %s", item.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")))
	return strings.Join(lines, "\n")
}

func renderTeamMessageWithProtocol(manager TeamMessageManager, item team.Message) string {
	result := renderTeamMessage(item)
	if strings.TrimSpace(item.RequestID) == "" {
		return result
	}
	record, err := manager.GetRequest(item.RequestID)
	if err != nil {
		return result
	}
	return result + "\n" + fmt.Sprintf("request_status: %s", record.Status)
}

func renderTeamMessageLine(item team.Message) string {
	line := fmt.Sprintf("- %s from=%s kind=%s body=%s", item.ID, item.FromID, item.Kind, item.Body)
	if item.RequestID != "" {
		line = fmt.Sprintf("%s request_id=%s", line, item.RequestID)
	}
	return fmt.Sprintf("%s created_at=%s", line, item.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"))
}

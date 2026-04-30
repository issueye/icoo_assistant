package tools

import (
	"fmt"
	"strings"

	"icoo_assistant/internal/llm"
	"icoo_assistant/internal/session"
)

type SessionManager interface {
	Create(input session.CreateInput) (session.Session, error)
	Close(id string) (session.Session, error)
	Switch(id string) (session.Session, session.Session, error)
	Get(id string) (session.Session, error)
	GetActive() (session.Session, error)
	List(status string) ([]session.Session, error)
	Archive(id string) (session.Session, error)
	History(limit int) ([]session.Session, error)
	GetActiveID() string
	UpdateStats(id string, roundCount, messageCount int) error
	UpdateSummary(id, summary string, memoryIDs []string) error
}

func NewSessionTool(manager SessionManager) Definition {
	return Definition{
		Tool: llm.Tool{
			Name:        "session",
			Description: "Manage sessions: create, close, switch, list, view status, browse history, and archive sessions for cross-session continuity.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action": map[string]interface{}{"type": "string", "enum": []string{"create", "close", "switch", "list", "status", "history", "archive"}},
					"id":     map[string]interface{}{"type": "string", "description": "Session ID for close/switch/archive"},
					"title":  map[string]interface{}{"type": "string", "description": "Session title for create"},
					"tags":   map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Tags for categorization"},
					"limit":  map[string]interface{}{"type": "integer", "description": "Max results for list/history"},
				},
				"required": []string{"action"},
			},
		},
		Handler: func(call Call) (string, error) {
			action, _ := call.Input["action"].(string)
			switch strings.ToLower(strings.TrimSpace(action)) {
			case "create":
				return sessionCreate(manager, call)
			case "close":
				return sessionClose(manager, call)
			case "switch":
				return sessionSwitch(manager, call)
			case "list":
				return sessionList(manager, call)
			case "status":
				return sessionStatus(manager)
			case "history":
				return sessionHistory(manager, call)
			case "archive":
				return sessionArchive(manager, call)
			default:
				return "", fmt.Errorf("unsupported action %q", action)
			}
		},
	}
}

func sessionCreate(manager SessionManager, call Call) (string, error) {
	title, _ := call.Input["title"].(string)
	id, _ := call.Input["id"].(string)
	tags, _ := stringListFromCall(call.Input["tags"])

	session, err := manager.Create(session.CreateInput{
		ID:    id,
		Title: title,
		Tags:  tags,
	})
	if err != nil {
		return "", err
	}
	return renderSession(session, true), nil
}

func sessionClose(manager SessionManager, call Call) (string, error) {
	id, _ := call.Input["id"].(string)
	if strings.TrimSpace(id) == "" {
		active, err := manager.GetActive()
		if err != nil {
			return "", fmt.Errorf("no active session and no id specified: %w", err)
		}
		id = active.ID
	}

	session, err := manager.Close(id)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Session closed.\n%s", renderSession(session, false)), nil
}

func sessionSwitch(manager SessionManager, call Call) (string, error) {
	id, _ := call.Input["id"].(string)
	if strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("id required for switch")
	}

	target, previous, err := manager.Switch(id)
	if err != nil {
		return "", err
	}

	lines := []string{
		fmt.Sprintf("Switched to session %s.", target.ID),
		renderSession(target, false),
	}
	if previous.ID != "" {
		lines = append(lines, fmt.Sprintf("Previous session %s auto-closed.", previous.ID))
	}
	return strings.Join(lines, "\n"), nil
}

func sessionList(manager SessionManager, call Call) (string, error) {
	limit := intFromCall(call.Input["limit"], 20)

	sessions, err := manager.List("")
	if err != nil {
		return "", err
	}
	if len(sessions) == 0 {
		return "No sessions found.", nil
	}

	if limit > 0 && len(sessions) > limit {
		sessions = sessions[:limit]
	}

	activeID := manager.GetActiveID()
	lines := make([]string, 0, len(sessions)+1)
	lines = append(lines, fmt.Sprintf("sessions: %d", len(sessions)))
	for _, s := range sessions {
		marker := " "
		if s.ID == activeID {
			marker = "*"
		}
		lines = append(lines, renderSessionLine(s, marker))
	}
	return strings.Join(lines, "\n"), nil
}

func sessionStatus(manager SessionManager) (string, error) {
	active, err := manager.GetActive()
	if err != nil {
		return fmt.Sprintf("No active session. Error: %v", err), nil
	}

	all, err := manager.List("")
	if err != nil {
		return "", err
	}

	var activeCount, closedCount, archivedCount int
	for _, s := range all {
		switch s.Status {
		case session.StatusActive:
			activeCount++
		case session.StatusClosed:
			closedCount++
		case session.StatusArchived:
			archivedCount++
		}
	}

	lines := []string{
		fmt.Sprintf("active_session: %s", active.ID),
		renderSession(active, false),
		fmt.Sprintf("total_sessions: %d (active=%d, closed=%d, archived=%d)", len(all), activeCount, closedCount, archivedCount),
	}
	return strings.Join(lines, "\n"), nil
}

func sessionHistory(manager SessionManager, call Call) (string, error) {
	limit := intFromCall(call.Input["limit"], 10)

	sessions, err := manager.History(limit)
	if err != nil {
		return "", err
	}
	if len(sessions) == 0 {
		return "No session history.", nil
	}

	activeID := manager.GetActiveID()
	lines := make([]string, 0, len(sessions)+1)
	lines = append(lines, fmt.Sprintf("recent_sessions: %d", len(sessions)))
	for _, s := range sessions {
		marker := " "
		if s.ID == activeID {
			marker = "*"
		}
		lines = append(lines, renderSessionLine(s, marker))
	}
	return strings.Join(lines, "\n"), nil
}

func sessionArchive(manager SessionManager, call Call) (string, error) {
	id, _ := call.Input["id"].(string)
	if strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("id required for archive")
	}

	session, err := manager.Archive(id)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Session archived.\n%s", renderSession(session, false)), nil
}

func renderSession(s session.Session, isNew bool) string {
	lines := []string{
		fmt.Sprintf("id: %s", s.ID),
		fmt.Sprintf("title: %s", s.Title),
		fmt.Sprintf("status: %s", s.Status),
	}
	if len(s.Tags) > 0 {
		lines = append(lines, fmt.Sprintf("tags: %s", strings.Join(s.Tags, ", ")))
	}
	lines = append(lines, fmt.Sprintf("rounds: %d", s.RoundCount))
	lines = append(lines, fmt.Sprintf("messages: %d", s.MessageCount))
	if s.Summary != "" {
		summary := s.Summary
		if len(summary) > 200 {
			summary = summary[:197] + "..."
		}
		lines = append(lines, fmt.Sprintf("summary: %s", summary))
	}
	if len(s.MemoryIDs) > 0 {
		lines = append(lines, fmt.Sprintf("linked_memories: %d", len(s.MemoryIDs)))
	}
	lines = append(lines, fmt.Sprintf("created_at: %s", s.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")))
	if s.ClosedAt != nil {
		lines = append(lines, fmt.Sprintf("closed_at: %s", s.ClosedAt.UTC().Format("2006-01-02T15:04:05Z")))
	}
	if isNew {
		lines = append(lines, "hint: use memory_summarize to persist session context before closing")
	}
	return strings.Join(lines, "\n")
}

func renderSessionLine(s session.Session, marker string) string {
	tagStr := ""
	if len(s.Tags) > 0 {
		tagStr = fmt.Sprintf(" [%s]", strings.Join(s.Tags, ", "))
	}
	closed := ""
	if s.ClosedAt != nil {
		closed = fmt.Sprintf(" closed=%s", s.ClosedAt.UTC().Format("2006-01-02"))
	}
	return fmt.Sprintf("%s %s [%s]%s%s %s%s",
		marker, s.ID, s.Status, tagStr, closed, s.Title,
		func() string {
			if s.Summary != "" {
				return fmt.Sprintf(" (has_summary rounds=%d)", s.RoundCount)
			}
			return fmt.Sprintf(" (rounds=%d)", s.RoundCount)
		}())
}

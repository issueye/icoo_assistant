package tools

import (
	"fmt"
	"strings"

	"icoo_assistant/internal/llm"
	"icoo_assistant/internal/memory"
)

type MemoryManager interface {
	SetSessionID(id string)
	Store(input memory.StoreInput) (memory.Memory, error)
	Recall(input memory.QueryInput) ([]memory.Memory, error)
	Delete(id, memType string) error
	Update(id, content string, tags []string, importance float64) (memory.Memory, error)
	GenerateSessionContext() string
}

func NewMemoryStoreTool(manager MemoryManager) Definition {
	return Definition{
		Tool: llm.Tool{
			Name:        "memory_store",
			Description: "Store information into the memory system. Use to remember facts, decisions, preferences, and context across sessions.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action": map[string]interface{}{"type": "string", "enum": []string{"set", "add"}},
					"type":   map[string]interface{}{"type": "string", "enum": []string{"short_term", "long_term", "ai_personality", "user_profile"}, "description": "Memory type: short_term (session-only), long_term (persistent fact), ai_personality (AI behavior), user_profile (user preferences)"},
					"id":     map[string]interface{}{"type": "string", "description": "Unique identifier for the memory"},
					"content":    map[string]interface{}{"type": "string", "description": "The content to store"},
					"tags":       map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Tags for categorization and search"},
					"importance": map[string]interface{}{"type": "number", "description": "Importance score 0-1 (default 0.5)"},
				},
				"required": []string{"action", "type", "content"},
			},
		},
		Handler: func(call Call) (string, error) {
			action, _ := call.Input["action"].(string)
			memType, _ := call.Input["type"].(string)
			content, _ := call.Input["content"].(string)
			id, _ := call.Input["id"].(string)
			tags, _ := stringListFromCall(call.Input["tags"])
			importance := floatFromCall(call.Input["importance"], 0.5)

			switch strings.ToLower(strings.TrimSpace(action)) {
			case "set":
				mem, err := manager.Store(memory.StoreInput{
					ID:         id,
					Type:       memType,
					Content:    content,
					Tags:       tags,
					Importance: importance,
				})
				if err != nil {
					return "", err
				}
				return renderMemory(mem), nil
			case "add":
				mem, err := manager.Store(memory.StoreInput{
					ID:         "",
					Type:       memType,
					Content:    content,
					Tags:       tags,
					Importance: importance,
				})
				if err != nil {
					return "", err
				}
				return renderMemory(mem), nil
			default:
				return "", fmt.Errorf("unsupported action %q", action)
			}
		},
	}
}

func NewMemoryRecallTool(manager MemoryManager) Definition {
	return Definition{
		Tool: llm.Tool{
			Name:        "memory_recall",
			Description: "Recall information from the memory system. Search across all memory types or filter by type, tags, importance, and keywords. Use 'context' action to get relevant memories for session initialization.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action": map[string]interface{}{"type": "string", "enum": []string{"search", "get", "list", "context"}},
					"type":   map[string]interface{}{"type": "string", "enum": []string{"", "short_term", "long_term", "session_summary", "ai_personality", "user_profile"}, "description": "Memory type to query (empty = all)"},
					"query":         map[string]interface{}{"type": "string", "description": "Keyword search query"},
					"tags":          map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Filter by tags"},
					"limit":         map[string]interface{}{"type": "integer", "description": "Max results (default 10)"},
					"min_importance": map[string]interface{}{"type": "number", "description": "Minimum importance filter (0-1)"},
					"id":            map[string]interface{}{"type": "string", "description": "Specific memory ID for 'get' action"},
					"session_id":    map[string]interface{}{"type": "string", "description": "Session ID for session_summary recall"},
				},
				"required": []string{"action"},
			},
		},
		Handler: func(call Call) (string, error) {
			action, _ := call.Input["action"].(string)
			switch strings.ToLower(strings.TrimSpace(action)) {
			case "context":
				context := manager.GenerateSessionContext()
				if context == "" {
					return "No memories available.", nil
				}
				return context, nil
			case "get":
				id, _ := call.Input["id"].(string)
				memType, _ := call.Input["type"].(string)
				return recallGet(manager, id, memType)
			case "list":
				return recallList(manager, call)
			case "search":
				return recallSearch(manager, call)
			default:
				return "", fmt.Errorf("unsupported action %q", action)
			}
		},
	}
}

func NewMemorySummarizeTool(manager MemoryManager) Definition {
	return Definition{
		Tool: llm.Tool{
			Name:        "memory_summarize",
			Description: "Generate and persist a session summary. Use at the end of a session to capture key decisions, findings, and context for future sessions.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"summary":       map[string]interface{}{"type": "string", "description": "Summary of the session"},
					"key_decisions": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Key decisions made during the session"},
					"key_findings":  map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Key findings and discoveries"},
					"tags":          map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Tags for categorization"},
					"importance":    map[string]interface{}{"type": "number", "description": "Importance score 0-1 (default 0.8)"},
					"session_id":    map[string]interface{}{"type": "string", "description": "Session identifier (auto-generated if empty)"},
				},
				"required": []string{"summary"},
			},
		},
		Handler: func(call Call) (string, error) {
			summary, _ := call.Input["summary"].(string)
			sessionID, _ := call.Input["session_id"].(string)
			importance := floatFromCall(call.Input["importance"], 0.8)

			if strings.TrimSpace(summary) == "" {
				return "", fmt.Errorf("summary required")
			}

			var contentBuilder strings.Builder
			contentBuilder.WriteString(summary)

			if decisions, ok := call.Input["key_decisions"].([]interface{}); ok && len(decisions) > 0 {
				contentBuilder.WriteString("\n\nKey Decisions:")
				for _, d := range decisions {
					contentBuilder.WriteString(fmt.Sprintf("\n- %s", d))
				}
			}

			if findings, ok := call.Input["key_findings"].([]interface{}); ok && len(findings) > 0 {
				contentBuilder.WriteString("\n\nKey Findings:")
				for _, f := range findings {
					contentBuilder.WriteString(fmt.Sprintf("\n- %s", f))
				}
			}

			tags, _ := stringListFromCall(call.Input["tags"])

			mem, err := manager.Store(memory.StoreInput{
				ID:         sessionID,
				Type:       "session_summary",
				Content:    contentBuilder.String(),
				Tags:       tags,
				Importance: importance,
				SessionID:  sessionID,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("Session summary stored.\n%s", renderMemory(mem)), nil
		},
	}
}

func NewMemoryManageTool(manager MemoryManager) Definition {
	return Definition{
		Tool: llm.Tool{
			Name:        "memory_manage",
			Description: "Manage existing memories: update content, change tags, adjust importance, delete memories, or consolidate similar memories.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action": map[string]interface{}{"type": "string", "enum": []string{"update", "delete", "tag", "consolidate"}},
					"id":     map[string]interface{}{"type": "string", "description": "Memory ID to manage"},
					"type":   map[string]interface{}{"type": "string", "enum": []string{"short_term", "long_term", "session_summary", "ai_personality", "user_profile"}, "description": "Memory type for delete"},
					"content":    map[string]interface{}{"type": "string", "description": "Updated content (for update action)"},
					"tags":       map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Updated tags (for update/tag action)"},
					"importance": map[string]interface{}{"type": "number", "description": "Updated importance (for update action)"},
				},
				"required": []string{"action"},
			},
		},
		Handler: func(call Call) (string, error) {
			action, _ := call.Input["action"].(string)
			switch strings.ToLower(strings.TrimSpace(action)) {
			case "update":
				return manageUpdate(manager, call)
			case "delete":
				return manageDelete(manager, call)
			case "tag":
				return manageTag(manager, call)
			case "consolidate":
				return manageConsolidate(manager, call)
			default:
				return "", fmt.Errorf("unsupported action %q", action)
			}
		},
	}
}

func recallGet(manager MemoryManager, id, memType string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("id required")
	}
	mems, err := manager.Recall(memory.QueryInput{
		Type: memType,
	})
	if err != nil {
		return "", err
	}
	for _, mem := range mems {
		if mem.ID == id {
			return renderMemory(mem), nil
		}
	}
	return "", fmt.Errorf("memory %s not found", id)
}

func recallList(manager MemoryManager, call Call) (string, error) {
	memType, _ := call.Input["type"].(string)
	limit := intFromCall(call.Input["limit"], 20)

	mems, err := manager.Recall(memory.QueryInput{
		Type:  memType,
		Limit: limit,
	})
	if err != nil {
		return "", err
	}
	if len(mems) == 0 {
		return "No memories found.", nil
	}
	lines := make([]string, 0, len(mems)+1)
	lines = append(lines, fmt.Sprintf("found: %d", len(mems)))
	for _, mem := range mems {
		lines = append(lines, renderMemoryLine(mem))
	}
	return strings.Join(lines, "\n"), nil
}

func recallSearch(manager MemoryManager, call Call) (string, error) {
	memType, _ := call.Input["type"].(string)
	query, _ := call.Input["query"].(string)
	tags, _ := stringListFromCall(call.Input["tags"])
	limit := intFromCall(call.Input["limit"], 10)
	minImportance := floatFromCall(call.Input["min_importance"], 0)

	mems, err := manager.Recall(memory.QueryInput{
		Type:          memType,
		Query:         query,
		Tags:          tags,
		Limit:         limit,
		MinImportance: minImportance,
	})
	if err != nil {
		return "", err
	}
	if len(mems) == 0 {
		return "No matching memories found.", nil
	}
	lines := make([]string, 0, len(mems)+1)
	lines = append(lines, fmt.Sprintf("found: %d", len(mems)))
	for _, mem := range mems {
		lines = append(lines, renderMemory(mem))
	}
	return strings.Join(lines, "\n---\n"), nil
}

func manageUpdate(manager MemoryManager, call Call) (string, error) {
	id, _ := call.Input["id"].(string)
	content, _ := call.Input["content"].(string)
	tags, _ := stringListFromCall(call.Input["tags"])
	importance := floatFromCall(call.Input["importance"], -1)

	if strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("id required")
	}

	mem, err := manager.Update(id, content, tags, importance)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Memory updated.\n%s", renderMemory(mem)), nil
}

func manageDelete(manager MemoryManager, call Call) (string, error) {
	id, _ := call.Input["id"].(string)
	memType, _ := call.Input["type"].(string)

	if strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("id required")
	}
	if strings.TrimSpace(memType) == "" {
		return "", fmt.Errorf("type required")
	}

	if err := manager.Delete(id, memType); err != nil {
		return "", err
	}
	return fmt.Sprintf("Memory %s deleted.", id), nil
}

func manageTag(manager MemoryManager, call Call) (string, error) {
	id, _ := call.Input["id"].(string)
	tags, err := stringListFromCall(call.Input["tags"])
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("id required")
	}
	mem, err := manager.Update(id, "", tags, -1)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Tags updated.\n%s", renderMemory(mem)), nil
}

func manageConsolidate(manager MemoryManager, call Call) (string, error) {
	memType, _ := call.Input["type"].(string)
	query, _ := call.Input["query"].(string)
	tags, _ := stringListFromCall(call.Input["tags"])

	mems, err := manager.Recall(memory.QueryInput{
		Type:  memType,
		Query: query,
		Tags:  tags,
		Limit: 50,
	})
	if err != nil {
		return "", err
	}
	if len(mems) < 2 {
		return "Nothing to consolidate - at least 2 memories needed.", nil
	}

	var consolidated strings.Builder
	consolidated.WriteString("Consolidated Memories:\n\n")
	for i, mem := range mems {
		consolidated.WriteString(fmt.Sprintf("[%d]: %s\n", i+1, mem.Content))
	}

	return fmt.Sprintf("Found %d memories for consolidation review.\n%s\n\nUse memory_manage action=delete to remove redundant entries after consolidation.", len(mems), consolidated.String()), nil
}

func renderMemory(mem memory.Memory) string {
	lines := []string{
		fmt.Sprintf("id: %s", mem.ID),
		fmt.Sprintf("type: %s", mem.Type),
	}
	if len(mem.Tags) > 0 {
		lines = append(lines, fmt.Sprintf("tags: %s", strings.Join(mem.Tags, ", ")))
	}
	lines = append(lines, fmt.Sprintf("importance: %.2f", mem.Importance))
	lines = append(lines, fmt.Sprintf("content: %s", mem.Content))
	lines = append(lines, fmt.Sprintf("created_at: %s", mem.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")))
	lines = append(lines, fmt.Sprintf("updated_at: %s", mem.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z")))
	if mem.SessionID != "" {
		lines = append(lines, fmt.Sprintf("session_id: %s", mem.SessionID))
	}
	return strings.Join(lines, "\n")
}

func renderMemoryLine(mem memory.Memory) string {
	tagPart := ""
	if len(mem.Tags) > 0 {
		tagPart = fmt.Sprintf(" [%s]", strings.Join(mem.Tags, ", "))
	}
	content := mem.Content
	if len(content) > 100 {
		content = content[:97] + "..."
	}
	return fmt.Sprintf("- %s %s%.2f%s %s", mem.ID, mem.Type, mem.Importance, tagPart, content)
}

func floatFromCall(raw interface{}, fallback float64) float64 {
	switch v := raw.(type) {
	case float64:
		return v
	default:
		return fallback
	}
}

func intFromCall(raw interface{}, fallback int) int {
	switch v := raw.(type) {
	case float64:
		if int(v) > 0 {
			return int(v)
		}
	case int:
		if v > 0 {
			return v
		}
	}
	return fallback
}

func stringListFromCall(raw interface{}) ([]string, error) {
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

package todo

import (
	"fmt"
	"strings"
)

type Item struct {
	ID     string `json:"id"`
	Text   string `json:"text"`
	Status string `json:"status"`
}

type Manager struct {
	items []Item
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Update(items []Item) (string, error) {
	if len(items) > 20 {
		return "", fmt.Errorf("max 20 todos allowed")
	}
	validated := make([]Item, 0, len(items))
	inProgress := 0
	for i, item := range items {
		id := strings.TrimSpace(item.ID)
		text := strings.TrimSpace(item.Text)
		status := strings.ToLower(strings.TrimSpace(item.Status))
		if id == "" {
			id = fmt.Sprintf("%d", i+1)
		}
		if text == "" {
			return "", fmt.Errorf("item %s: text required", id)
		}
		switch status {
		case "pending", "in_progress", "completed":
		default:
			return "", fmt.Errorf("item %s: invalid status %q", id, status)
		}
		if status == "in_progress" {
			inProgress++
		}
		validated = append(validated, Item{ID: id, Text: text, Status: status})
	}
	if inProgress > 1 {
		return "", fmt.Errorf("only one task can be in_progress at a time")
	}
	m.items = validated
	return m.Render(), nil
}

func (m *Manager) Render() string {
	if len(m.items) == 0 {
		return "No todos."
	}
	lines := make([]string, 0, len(m.items)+1)
	done := 0
	for _, item := range m.items {
		marker := map[string]string{
			"pending":     "[ ]",
			"in_progress": "[>]",
			"completed":   "[x]",
		}[item.Status]
		if item.Status == "completed" {
			done++
		}
		lines = append(lines, fmt.Sprintf("%s #%s: %s", marker, item.ID, item.Text))
	}
	lines = append(lines, fmt.Sprintf("\n(%d/%d completed)", done, len(m.items)))
	return strings.Join(lines, "\n")
}

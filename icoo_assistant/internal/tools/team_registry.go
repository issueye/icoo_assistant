package tools

import (
	"fmt"
	"strings"

	"icoo_assistant/internal/llm"
	"icoo_assistant/internal/team"
)

type TeamRegistryManager interface {
	GetConfig() (team.Config, error)
	UpdateConfig(input team.ConfigUpdateInput) (team.Config, error)
	Create(input team.CreateInput) (team.Teammate, error)
	Get(id string) (team.Teammate, error)
	List() ([]team.Teammate, error)
	Update(item team.Teammate) (team.Teammate, error)
}

func NewTeamRegistryTool(manager TeamRegistryManager) Definition {
	return Definition{
		Tool: llm.Tool{
			Name:        "team_registry",
			Description: "Manage persistent team config and teammate registry data under .team.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action":  map[string]interface{}{"type": "string", "enum": []string{"get_config", "update_config", "create", "get", "list", "update"}},
					"id":      map[string]interface{}{"type": "string"},
					"role":    map[string]interface{}{"type": "string"},
					"status":  map[string]interface{}{"type": "string"},
					"model":   map[string]interface{}{"type": "string"},
					"lead_id": map[string]interface{}{"type": "string"},
					"mission": map[string]interface{}{"type": "string"},
				},
				"required": []string{"action"},
			},
		},
		Handler: func(call Call) (string, error) {
			action, _ := call.Input["action"].(string)
			switch strings.ToLower(strings.TrimSpace(action)) {
			case "get_config":
				cfg, err := manager.GetConfig()
				if err != nil {
					return "", err
				}
				items, err := manager.List()
				if err != nil {
					return "", err
				}
				return renderTeamConfig(cfg, len(items)), nil
			case "update_config":
				leadID, _ := call.Input["lead_id"].(string)
				mission, _ := call.Input["mission"].(string)
				cfg, err := manager.UpdateConfig(team.ConfigUpdateInput{
					LeadID:  leadID,
					Mission: mission,
				})
				if err != nil {
					return "", err
				}
				items, err := manager.List()
				if err != nil {
					return "", err
				}
				return renderTeamConfig(cfg, len(items)), nil
			case "create":
				id, _ := call.Input["id"].(string)
				role, _ := call.Input["role"].(string)
				status, _ := call.Input["status"].(string)
				model, _ := call.Input["model"].(string)
				item, err := manager.Create(team.CreateInput{
					ID:     id,
					Role:   role,
					Status: status,
					Model:  model,
				})
				if err != nil {
					return "", err
				}
				return renderTeammate(item), nil
			case "get":
				id, _ := call.Input["id"].(string)
				if strings.TrimSpace(id) == "" {
					return "", fmt.Errorf("id required for get")
				}
				item, err := manager.Get(id)
				if err != nil {
					return "", err
				}
				return renderTeammate(item), nil
			case "list":
				items, err := manager.List()
				if err != nil {
					return "", err
				}
				if len(items) == 0 {
					return "No teammates.", nil
				}
				lines := make([]string, 0, len(items))
				for _, item := range items {
					line := fmt.Sprintf("%s [%s] role=%s", item.ID, item.Status, item.Role)
					if item.Model != "" {
						line = fmt.Sprintf("%s model=%s", line, item.Model)
					}
					lines = append(lines, line)
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
				if role, ok := call.Input["role"].(string); ok {
					current.Role = role
				}
				if status, ok := call.Input["status"].(string); ok {
					current.Status = status
				}
				if model, ok := call.Input["model"].(string); ok {
					current.Model = model
				}
				item, err := manager.Update(current)
				if err != nil {
					return "", err
				}
				return renderTeammate(item), nil
			default:
				return "", fmt.Errorf("unsupported action %q", action)
			}
		},
	}
}

func renderTeamConfig(cfg team.Config, teammateCount int) string {
	lines := []string{
		fmt.Sprintf("lead_id: %s", cfg.LeadID),
		fmt.Sprintf("teammate_count: %d", teammateCount),
	}
	if cfg.Mission != "" {
		lines = append(lines, fmt.Sprintf("mission: %s", cfg.Mission))
	}
	lines = append(lines, fmt.Sprintf("created_at: %s", cfg.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")))
	lines = append(lines, fmt.Sprintf("updated_at: %s", cfg.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z")))
	return strings.Join(lines, "\n")
}

func renderTeammate(item team.Teammate) string {
	lines := []string{
		fmt.Sprintf("id: %s", item.ID),
		fmt.Sprintf("role: %s", item.Role),
		fmt.Sprintf("status: %s", item.Status),
	}
	if item.Model != "" {
		lines = append(lines, fmt.Sprintf("model: %s", item.Model))
	}
	lines = append(lines, fmt.Sprintf("created_at: %s", item.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")))
	lines = append(lines, fmt.Sprintf("updated_at: %s", item.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z")))
	return strings.Join(lines, "\n")
}

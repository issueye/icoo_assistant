package subagent

import (
	"fmt"

	"icoo_assistant/internal/agent"
	"icoo_assistant/internal/llm"
	"icoo_assistant/internal/tools"
)

type Runner struct {
	Client      llm.Client
	Registry    *tools.Registry
	Config      agent.Config
	Hooks       []agent.Hook
	SkillLoader agent.SkillContentProvider
}

func (r *Runner) Run(prompt string) (string, error) {
	if r.Client == nil {
		return "", fmt.Errorf("client required")
	}
	if r.Registry == nil {
		return "", fmt.Errorf("registry required")
	}
	child := &agent.Runner{
		Client:      r.Client,
		Registry:    r.Registry,
		Hooks:       r.Hooks,
		SkillLoader: r.SkillLoader,
		Config: agent.Config{
			SystemPrompt: r.Config.SystemPrompt,
			MaxRounds:    r.Config.MaxRounds,
		},
	}
	messages, err := child.Run([]llm.Message{{Role: "user", Content: prompt}})
	if err != nil {
		return "", err
	}
	if len(messages) == 0 {
		return "", nil
	}
	return fmt.Sprintf("%v", messages[len(messages)-1].Content), nil
}

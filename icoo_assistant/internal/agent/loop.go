package agent

import (
	"fmt"

	"icoo_assistant/internal/llm"
	"icoo_assistant/internal/todo"
	"icoo_assistant/internal/tools"
)

type Config struct {
	SystemPrompt string
	MaxRounds    int
}

type Runner struct {
	Client      llm.Client
	Registry    *tools.Registry
	TodoManager *todo.Manager
	Config      Config
}

func (r *Runner) Run(messages []llm.Message) ([]llm.Message, error) {
	if r.Client == nil {
		return nil, fmt.Errorf("client required")
	}
	if r.Registry == nil {
		return nil, fmt.Errorf("registry required")
	}
	maxRounds := r.Config.MaxRounds
	if maxRounds <= 0 {
		maxRounds = 20
	}
	roundsSinceTodo := 0
	for i := 0; i < maxRounds; i++ {
		resp, err := r.Client.CreateMessage(r.Config.SystemPrompt, messages, r.Registry.Tools())
		if err != nil {
			return nil, err
		}
		switch {
		case resp.StopReason == "tool_use" && resp.Raw != nil:
			messages = append(messages, llm.Message{Role: "assistant", Content: resp.Raw})
		case resp.Text != "":
			messages = append(messages, llm.Message{Role: "assistant", Content: resp.Text})
		default:
			messages = append(messages, llm.Message{Role: "assistant", Content: resp.Raw})
		}
		if resp.StopReason != "tool_use" {
			return messages, nil
		}
		results := make([]tools.Result, 0, len(resp.ToolUses))
		usedTodo := false
		for _, toolUse := range resp.ToolUses {
			result, err := r.Registry.Execute(tools.Call{ID: toolUse.ID, Name: toolUse.Name, Input: toolUse.Input})
			if err != nil {
				return nil, err
			}
			if toolUse.Name == "todo" {
				usedTodo = true
			}
			results = append(results, result)
		}
		if usedTodo {
			roundsSinceTodo = 0
		} else {
			roundsSinceTodo++
		}
		if roundsSinceTodo >= 3 {
			results = append(results, tools.Result{Type: "tool_result", ToolUseID: "reminder", Content: "<reminder>Update your todos.</reminder>"})
		}
		messages = append(messages, llm.Message{Role: "user", Content: results})
	}
	return nil, fmt.Errorf("max rounds exceeded")
}

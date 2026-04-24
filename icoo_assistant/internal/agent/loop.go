package agent

import (
	"fmt"

	"icoo_assistant/internal/compact"
	"icoo_assistant/internal/llm"
	"icoo_assistant/internal/todo"
	"icoo_assistant/internal/tools"
)

type SubagentRunner interface {
	Run(prompt string) (string, error)
}

type Config struct {
	SystemPrompt string
	MaxRounds    int
}

type Runner struct {
	Client         llm.Client
	Registry       *tools.Registry
	TodoManager    *todo.Manager
	CompactManager *compact.Manager
	SubagentRunner SubagentRunner
	Config         Config
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
		if r.CompactManager != nil {
			r.CompactManager.MicroCompact(messages)
			threshold := r.CompactManager.Threshold
			if threshold > 0 && r.CompactManager.EstimateTokens(messages) > threshold {
				compressed, err := r.CompactManager.AutoCompact(r.Client, messages)
				if err != nil {
					return nil, err
				}
				messages = compressed
			}
		}
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
		manualCompact := false
		for _, toolUse := range resp.ToolUses {
			var result tools.Result
			if toolUse.Name == "task" {
				if r.SubagentRunner == nil {
					return nil, fmt.Errorf("subagent runner required")
				}
				prompt, _ := toolUse.Input["prompt"].(string)
				summary, err := r.SubagentRunner.Run(prompt)
				if err != nil {
					return nil, err
				}
				result = tools.Result{Type: "tool_result", ToolUseID: toolUse.ID, Content: summary}
			} else {
				result, err = r.Registry.Execute(tools.Call{ID: toolUse.ID, Name: toolUse.Name, Input: toolUse.Input})
				if err != nil {
					return nil, err
				}
			}
			if toolUse.Name == "todo" {
				usedTodo = true
			}
			if toolUse.Name == "compact" {
				manualCompact = true
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
		if manualCompact && r.CompactManager != nil {
			compressed, err := r.CompactManager.AutoCompact(r.Client, messages)
			if err != nil {
				return nil, err
			}
			messages = compressed
		}
	}
	return nil, fmt.Errorf("max rounds exceeded")
}

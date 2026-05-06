package agent

import (
	"fmt"
	"strings"
	"time"

	"icoo_assistant/internal/background"
	"icoo_assistant/internal/compact"
	"icoo_assistant/internal/llm"
	"icoo_assistant/internal/todo"
	"icoo_assistant/internal/tools"
)

type SubagentRunner interface {
	Run(prompt string) (string, error)
}

type BackgroundNotifier interface {
	PollNotifications() ([]background.Completion, error)
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
	Transcript     TranscriptRecorder
	SubagentRunner SubagentRunner
	Background     BackgroundNotifier
	StreamHandler  func(string)
	Hooks          []Hook
	Config         Config
	now            func() time.Time
}

func (r *Runner) Run(messages []llm.Message) (_ []llm.Message, err error) {
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
	if r.now == nil {
		r.now = time.Now
	}
	runID := fmt.Sprintf("run-%d", r.now().UTC().UnixNano())
	defer func() {
		recordErr := r.recordTranscript(runID, messages, err)
		if err == nil && recordErr != nil {
			err = recordErr
		}
	}()
	r.emit(Event{
		Name:  "agent.run.started",
		RunID: runID,
		Fields: map[string]interface{}{
			"message_count": len(messages),
			"max_rounds":    maxRounds,
		},
	})
	roundsSinceTodo := 0
	for i := 0; i < maxRounds; i++ {
		round := i + 1
		r.emit(Event{
			Name:  "agent.round.started",
			RunID: runID,
			Round: round,
			Fields: map[string]interface{}{
				"message_count": len(messages),
			},
		})
		if r.Background != nil {
			completions, err := r.Background.PollNotifications()
			if err != nil {
				return nil, err
			}
			if len(completions) > 0 {
				r.emit(Event{
					Name:  "agent.background.notifications_injected",
					RunID: runID,
					Round: round,
					Fields: map[string]interface{}{
						"completion_count": len(completions),
					},
				})
				messages = append(messages, llm.Message{Role: "user", Content: formatBackgroundNotifications(completions)})
			}
		}
		if r.CompactManager != nil {
			r.CompactManager.MicroCompact(messages)
			threshold := r.CompactManager.Threshold
			if threshold > 0 && r.CompactManager.EstimateTokens(messages) > threshold {
				r.emit(Event{Name: "agent.compact.auto_requested", RunID: runID, Round: round})
				compressed, err := r.CompactManager.AutoCompact(r.Client, messages)
				if err != nil {
					return nil, err
				}
				r.emit(Event{Name: "agent.compact.auto_completed", RunID: runID, Round: round})
				messages = compressed
			}
		}
		r.emit(Event{
			Name:  "agent.model.requested",
			RunID: runID,
			Round: round,
			Fields: map[string]interface{}{
				"message_count": len(messages),
				"tool_count":    len(r.Registry.Tools()),
			},
		})
		var resp llm.Response
		if r.StreamHandler != nil {
			resp, err = r.Client.CreateMessageStream(r.Config.SystemPrompt, messages, r.Registry.Tools(), r.StreamHandler)
		} else {
			resp, err = r.Client.CreateMessage(r.Config.SystemPrompt, messages, r.Registry.Tools())
		}
		if err != nil {
			return nil, err
		}
		r.emit(Event{
			Name:  "agent.model.responded",
			RunID: runID,
			Round: round,
			Fields: map[string]interface{}{
				"stop_reason":    resp.StopReason,
				"tool_use_count": len(resp.ToolUses),
				"has_text":       resp.Text != "",
			},
		})
		switch {
		case resp.StopReason == "tool_use" && resp.Raw != nil:
			messages = append(messages, llm.Message{Role: "assistant", Content: resp.Raw})
		case resp.Text != "":
			messages = append(messages, llm.Message{Role: "assistant", Content: resp.Text})
		default:
			messages = append(messages, llm.Message{Role: "assistant", Content: resp.Raw})
		}
		if resp.StopReason != "tool_use" {
			r.emit(Event{
				Name:  "agent.run.completed",
				RunID: runID,
				Round: round,
				Fields: map[string]interface{}{
					"message_count": len(messages),
					"stop_reason":   resp.StopReason,
				},
			})
			return messages, nil
		}
		results := make([]tools.Result, 0, len(resp.ToolUses))
		usedTodo := false
		manualCompact := false
		for _, toolUse := range resp.ToolUses {
			var result tools.Result
			r.emit(Event{
				Name:  "agent.tool.started",
				RunID: runID,
				Round: round,
				Fields: map[string]interface{}{
					"tool_name": toolUse.Name,
					"tool_id":   toolUse.ID,
				},
			})
			if toolUse.Name == "task" {
				if r.SubagentRunner == nil {
					return nil, fmt.Errorf("subagent runner required")
				}
				prompt, _ := toolUse.Input["prompt"].(string)
				r.emit(Event{
					Name:  "agent.subagent.started",
					RunID: runID,
					Round: round,
					Fields: map[string]interface{}{
						"tool_id":       toolUse.ID,
						"prompt_length": len(prompt),
					},
				})
				summary, err := r.SubagentRunner.Run(prompt)
				if err != nil {
					return nil, err
				}
				r.emit(Event{
					Name:  "agent.subagent.completed",
					RunID: runID,
					Round: round,
					Fields: map[string]interface{}{
						"tool_id":        toolUse.ID,
						"summary_length": len(summary),
					},
				})
				result = tools.Result{Type: "tool_result", ToolUseID: toolUse.ID, Content: summary}
			} else {
				result, err = r.Registry.Execute(tools.Call{ID: toolUse.ID, Name: toolUse.Name, Input: toolUse.Input})
				if err != nil {
					return nil, err
				}
			}
			r.emit(Event{
				Name:  "agent.tool.completed",
				RunID: runID,
				Round: round,
				Fields: map[string]interface{}{
					"tool_name":     toolUse.Name,
					"tool_id":       toolUse.ID,
					"result_length": len(result.Content),
				},
			})
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
		shouldInjectTodoReminder := roundsSinceTodo >= 3
		if roundsSinceTodo >= 3 {
			r.emit(Event{Name: "agent.todo.reminder_injected", RunID: runID, Round: round})
		}
		messages = append(messages, llm.Message{Role: "user", Content: results})
		if shouldInjectTodoReminder {
			messages = append(messages, llm.Message{Role: "user", Content: "<reminder>Update your todos.</reminder>"})
		}
		if manualCompact && r.CompactManager != nil {
			r.emit(Event{Name: "agent.compact.manual_requested", RunID: runID, Round: round})
			compressed, err := r.CompactManager.AutoCompact(r.Client, messages)
			if err != nil {
				return nil, err
			}
			r.emit(Event{Name: "agent.compact.manual_completed", RunID: runID, Round: round})
			messages = compressed
		}
	}
	r.emit(Event{
		Name:  "agent.run.failed",
		RunID: runID,
		Fields: map[string]interface{}{
			"error": "max rounds exceeded",
		},
	})
	return nil, fmt.Errorf("max rounds exceeded")
}

func (r *Runner) recordTranscript(runID string, messages []llm.Message, runErr error) error {
	recorder := r.Transcript
	if recorder == nil && r.CompactManager != nil && strings.TrimSpace(r.CompactManager.Dir) != "" {
		defaultRecorder, err := NewJSONTranscriptRecorder(r.CompactManager.Dir)
		if err != nil {
			return err
		}
		r.Transcript = defaultRecorder
		recorder = defaultRecorder
	}
	if recorder == nil {
		return nil
	}
	record := TranscriptRecord{
		Timestamp:    r.now().UTC(),
		RunID:        runID,
		Status:       "completed",
		MessageCount: len(messages),
		Messages:     cloneMessages(messages),
	}
	if runErr != nil {
		record.Status = "failed"
		record.Error = runErr.Error()
	}
	return recorder.Record(record)
}

func cloneMessages(messages []llm.Message) []llm.Message {
	if len(messages) == 0 {
		return nil
	}
	cloned := make([]llm.Message, len(messages))
	copy(cloned, messages)
	return cloned
}

func formatBackgroundNotifications(completions []background.Completion) string {
	parts := make([]string, 0, len(completions))
	for _, completion := range completions {
		parts = append(parts, completion.Summary)
	}
	return strings.Join(parts, "\n\n")
}

func (r *Runner) emit(event Event) {
	if len(r.Hooks) == 0 {
		return
	}
	if event.Timestamp.IsZero() {
		now := time.Now
		if r.now != nil {
			now = r.now
		}
		event.Timestamp = now().UTC()
	}
	for _, hook := range r.Hooks {
		if hook == nil {
			continue
		}
		hook.OnEvent(event)
	}
}

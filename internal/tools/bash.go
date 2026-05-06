package tools

import (
	"fmt"
	"strings"
	"time"

	"icoo_assistant/internal/llm"
)

type CommandRunner struct {
	Workdir string
	Timeout time.Duration
}

func NewBashTool(runner CommandRunner) Definition {
	return Definition{
		Tool: llm.Tool{
			Name:        "bash",
			Description: "Run a shell command in the workspace.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{"type": "string"},
				},
				"required": []string{"command"},
			},
		},
		Handler: func(call Call) (string, error) {
			command, ok := call.Input["command"].(string)
			if !ok || strings.TrimSpace(command) == "" {
				return "", fmt.Errorf("command required")
			}
			text, err := RunCommand(runner.Workdir, command, runner.Timeout)
			if err != nil {
				if strings.Contains(err.Error(), "dangerous command blocked") {
					return "Error: Dangerous command blocked", nil
				}
				return "", err
			}
			return text, nil
		},
	}
}

package tools

import (
	"fmt"
	"path"
	"strings"
	"time"

	"icoo_assistant/internal/llm"
)

type CommandRunner struct {
	Workdir      string
	Timeout      time.Duration
	DenyPatterns []string
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
			if isCommandDenied(command, runner.DenyPatterns) {
				return "Error: Command blocked by permission settings", nil
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

func isCommandDenied(command string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}
	normalized := normalizeCommandPattern(command)
	for _, pattern := range patterns {
		matched, err := path.Match(normalizeCommandPattern(pattern), normalized)
		if err == nil && matched {
			return true
		}
	}
	return false
}

func normalizeCommandPattern(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, ":", " ")
	value = strings.Join(strings.Fields(value), " ")
	return value
}

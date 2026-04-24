package tools

import (
	"context"
	"fmt"
	"os/exec"
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
			Description: "Run a shell command.",
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
			for _, dangerous := range []string{"rm -rf /", "sudo", "shutdown", "reboot", "> /dev/"} {
				if strings.Contains(command, dangerous) {
					return "Error: Dangerous command blocked", nil
				}
			}
			timeout := runner.Timeout
			if timeout <= 0 {
				timeout = 120 * time.Second
			}
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			cmd := exec.CommandContext(ctx, "bash", "-lc", command)
			cmd.Dir = runner.Workdir
			out, err := cmd.CombinedOutput()
			if ctx.Err() == context.DeadlineExceeded {
				return "Error: Timeout (120s)", nil
			}
			text := strings.TrimSpace(string(out))
			if text == "" {
				text = "(no output)"
			}
			if len(text) > 50000 {
				text = text[:50000]
			}
			if err != nil {
				return text, nil
			}
			return text, nil
		},
	}
}

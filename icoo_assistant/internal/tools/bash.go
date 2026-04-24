package tools

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
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
			shell, args := defaultShellCommand(command)
			cmd := exec.CommandContext(ctx, shell, args...)
			cmd.Dir = runner.Workdir
			out, err := cmd.CombinedOutput()
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Sprintf("Error: Timeout (%s)", timeout), nil
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

func defaultShellCommand(command string) (string, []string) {
	if runtime.GOOS == "windows" {
		return "powershell.exe", []string{"-NoLogo", "-NoProfile", "-Command", command}
	}
	return "sh", []string{"-lc", command}
}

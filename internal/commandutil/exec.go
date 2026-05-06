package commandutil

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func Execute(ctx context.Context, workdir, command string) (string, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return "", fmt.Errorf("command required")
	}
	if err := Validate(command); err != nil {
		return "", err
	}
	shell, args := DefaultShell(command)
	cmd := exec.CommandContext(ctx, shell, args...)
	cmd.Dir = workdir
	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))
	if text == "" {
		text = "(no output)"
	}
	if len(text) > 50000 {
		text = text[:50000]
	}
	return text, err
}

func Run(workdir, command string, timeout time.Duration) (string, error) {
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	text, err := Execute(ctx, workdir, command)
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Sprintf("Error: Timeout (%s)", timeout), nil
	}
	if err != nil {
		return text, nil
	}
	return text, nil
}

func Validate(command string) error {
	for _, dangerous := range []string{"rm -rf /", "sudo", "shutdown", "reboot", "> /dev/"} {
		if strings.Contains(command, dangerous) {
			return fmt.Errorf("dangerous command blocked")
		}
	}
	return nil
}

func DefaultShell(command string) (string, []string) {
	if runtime.GOOS == "windows" {
		return "powershell.exe", []string{"-NoLogo", "-NoProfile", "-Command", command}
	}
	return "sh", []string{"-lc", command}
}

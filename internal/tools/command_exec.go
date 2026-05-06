package tools

import (
	"context"
	"time"

	"icoo_assistant/internal/commandutil"
)

func ExecuteCommand(ctx context.Context, workdir, command string) (string, error) {
	return commandutil.Execute(ctx, workdir, command)
}

func RunCommand(workdir, command string, timeout time.Duration) (string, error) {
	return commandutil.Run(workdir, command, timeout)
}

func ValidateCommand(command string) error {
	return commandutil.Validate(command)
}

func DefaultShellCommand(command string) (string, []string) {
	return commandutil.DefaultShell(command)
}

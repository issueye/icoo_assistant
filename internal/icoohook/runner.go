package icoohook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"icoo_assistant/internal/agent"
	"icoo_assistant/internal/commandutil"
)

type Config struct {
	Events  []string `json:"events"`
	Command string   `json:"command"`
}

type fileConfig struct {
	Hooks []Config `json:"hooks"`
}

type Runner struct {
	workdir string
	hooks   []Config
}

func Load(workdir string) (*Runner, error) {
	path := filepath.Join(workdir, ".icoo", "hooks.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Runner{workdir: workdir}, nil
		}
		return nil, err
	}
	var file fileConfig
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	return &Runner{
		workdir: workdir,
		hooks:   file.Hooks,
	}, nil
}

func (r *Runner) OnEvent(event agent.Event) {
	if r == nil || len(r.hooks) == 0 {
		return
	}
	for _, item := range r.hooks {
		if !matchesEvent(item.Events, event.Name) {
			continue
		}
		command := strings.TrimSpace(item.Command)
		if command == "" {
			continue
		}
		_, _ = commandutil.Run(r.workdir, command, 30_000_000_000)
	}
}

func matchesEvent(events []string, name string) bool {
	for _, item := range events {
		item = strings.TrimSpace(item)
		if item == "*" || item == name {
			return true
		}
	}
	return false
}

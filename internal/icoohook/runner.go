package icoohook

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
		_, _ = commandutil.RunWithEnv(r.workdir, command, 30_000_000_000, buildHookEnv(event))
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

func buildHookEnv(event agent.Event) []string {
	env := []string{
		"ICOO_EVENT_NAME=" + event.Name,
		"ICOO_RUN_ID=" + event.RunID,
		"ICOO_ROUND=" + strconv.Itoa(event.Round),
		"ICOO_EVENT_TIMESTAMP=" + event.Timestamp.UTC().Format("2006-01-02T15:04:05Z"),
	}
	for key, value := range event.Fields {
		name := normalizeEnvKey("ICOO_FIELD_" + key)
		env = append(env, name+"="+fmt.Sprint(value))
	}
	return env
}

func normalizeEnvKey(key string) string {
	key = strings.ToUpper(strings.TrimSpace(key))
	replacer := strings.NewReplacer(
		".", "_",
		"-", "_",
		" ", "_",
		"/", "_",
		"\\", "_",
	)
	return replacer.Replace(key)
}

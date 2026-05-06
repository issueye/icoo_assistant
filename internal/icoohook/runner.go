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
	Events      []string `json:"events"`
	EventPrefix string   `json:"event_prefix"`
	ToolName    string   `json:"tool_name"`
	StopReason  string   `json:"stop_reason"`
	RoundEquals int      `json:"round_equals"`
	Command     string   `json:"command"`
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
		if !matchesConditions(item, event) {
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
	if len(events) == 0 {
		return true
	}
	for _, item := range events {
		item = strings.TrimSpace(item)
		if item == "*" || item == name {
			return true
		}
	}
	return false
}

func matchesConditions(config Config, event agent.Event) bool {
	if prefix := strings.TrimSpace(config.EventPrefix); prefix != "" && !strings.HasPrefix(event.Name, prefix) {
		return false
	}
	if config.RoundEquals > 0 && event.Round != config.RoundEquals {
		return false
	}
	if toolName := strings.TrimSpace(config.ToolName); toolName != "" {
		if !strings.EqualFold(strings.TrimSpace(fmt.Sprint(event.Fields["tool_name"])), toolName) {
			return false
		}
	}
	if stopReason := strings.TrimSpace(config.StopReason); stopReason != "" {
		if !strings.EqualFold(strings.TrimSpace(fmt.Sprint(event.Fields["stop_reason"])), stopReason) {
			return false
		}
	}
	return true
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

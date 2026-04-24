package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Event struct {
	Timestamp time.Time              `json:"timestamp"`
	Name      string                 `json:"name"`
	RunID     string                 `json:"runId"`
	Round     int                    `json:"round,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

type Hook interface {
	OnEvent(Event)
}

type HookFunc func(Event)

func (f HookFunc) OnEvent(event Event) {
	f(event)
}

type JSONLHook struct {
	path string
	mu   sync.Mutex
}

func DefaultHookDir(root string) string {
	return filepath.Join(root, ".agent-hooks")
}

func NewJSONLHook(dir string) (*JSONLHook, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &JSONLHook{
		path: filepath.Join(dir, "events.jsonl"),
	}, nil
}

func (h *JSONLHook) OnEvent(event Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	file, err := os.OpenFile(h.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer file.Close()

	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	_, _ = file.Write(append(data, '\n'))
}

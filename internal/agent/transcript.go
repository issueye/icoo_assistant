package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"icoo_assistant/internal/llm"
)

type TranscriptRecord struct {
	Timestamp    time.Time     `json:"timestamp"`
	RunID        string        `json:"run_id"`
	Status       string        `json:"status"`
	Error        string        `json:"error,omitempty"`
	MessageCount int           `json:"message_count"`
	Messages     []llm.Message `json:"messages"`
}

type TranscriptRecorder interface {
	Record(record TranscriptRecord) error
}

type JSONTranscriptRecorder struct {
	dir string
	mu  sync.Mutex
}

func NewJSONTranscriptRecorder(dir string) (*JSONTranscriptRecorder, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, fmt.Errorf("transcript dir required")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &JSONTranscriptRecorder{dir: dir}, nil
}

func (r *JSONTranscriptRecorder) Record(record TranscriptRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	filename := fmt.Sprintf("conversation_%s.json", sanitizeTranscriptRunID(record.RunID))
	path := filepath.Join(r.dir, filename)
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func sanitizeTranscriptRunID(runID string) string {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return fmt.Sprintf("run-%d", time.Now().UTC().UnixNano())
	}
	replacer := strings.NewReplacer("\\", "_", "/", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_", " ", "_")
	return replacer.Replace(runID)
}

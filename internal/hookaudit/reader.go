package hookaudit

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type Event struct {
	Timestamp time.Time              `json:"timestamp"`
	Name      string                 `json:"name"`
	RunID     string                 `json:"runId"`
	Round     int                    `json:"round,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

type Query struct {
	Limit int
	Name  string
	RunID string
}

type Reader struct {
	path string
}

func NewReader(dir string) *Reader {
	return &Reader{
		path: filepath.Join(dir, "events.jsonl"),
	}
}

func (r *Reader) Recent(query Query) ([]Event, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}
	file, err := os.Open(r.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	events := make([]Event, 0, limit)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, err
		}
		if query.Name != "" && event.Name != query.Name {
			continue
		}
		if query.RunID != "" && event.RunID != query.RunID {
			continue
		}
		events = append(events, event)
		if len(events) > limit {
			events = events[1:]
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

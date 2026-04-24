package compact

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"icoo_assistant/internal/llm"
)

type Manager struct {
	Threshold  int
	KeepRecent int
	Dir        string
}

func (m Manager) EstimateTokens(messages []llm.Message) int {
	data, err := json.Marshal(messages)
	if err != nil {
		return 0
	}
	return len(data) / 4
}

func (m Manager) MicroCompact(messages []llm.Message) {
	keepRecent := m.KeepRecent
	if keepRecent <= 0 {
		keepRecent = 3
	}
	total := 0
	for _, msg := range messages {
		results, ok := msg.Content.([]llm.ToolResultBlock)
		if !ok {
			continue
		}
		total += len(results)
	}
	if total <= keepRecent {
		return
	}
	seen := 0
	for i := len(messages) - 1; i >= 0; i-- {
		results, ok := messages[i].Content.([]llm.ToolResultBlock)
		if !ok {
			continue
		}
		for j := len(results) - 1; j >= 0; j-- {
			seen++
			if seen <= keepRecent {
				continue
			}
			if len(results[j].Content) > 100 {
				results[j].Content = "[cleared]"
			}
		}
		messages[i].Content = results
	}
}

func (m Manager) AutoCompact(messages []llm.Message) ([]llm.Message, error) {
	if err := os.MkdirAll(m.Dir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(m.Dir, fmt.Sprintf("transcript_%d.jsonl", time.Now().UnixNano()))
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	for _, msg := range messages {
		data, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}
		if _, err := file.Write(append(data, '\n')); err != nil {
			return nil, err
		}
	}
	summary := fmt.Sprintf("[Compressed. Transcript: %s]", path)
	return []llm.Message{{Role: "user", Content: summary}}, nil
}

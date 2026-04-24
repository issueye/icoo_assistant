package compact_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"icoo_assistant/internal/compact"
	"icoo_assistant/internal/llm"
)

func TestMicroCompactClearsOldToolResults(t *testing.T) {
	mgr := compact.Manager{KeepRecent: 2}
	messages := []llm.Message{
		{Role: "user", Content: []llm.ToolResultBlock{{Type: "tool_result", ToolUseID: "1", Content: strings.Repeat("a", 120)}}},
		{Role: "user", Content: []llm.ToolResultBlock{{Type: "tool_result", ToolUseID: "2", Content: strings.Repeat("b", 120)}}},
		{Role: "user", Content: []llm.ToolResultBlock{{Type: "tool_result", ToolUseID: "3", Content: strings.Repeat("c", 120)}}},
	}
	mgr.MicroCompact(messages)
	results := messages[0].Content.([]llm.ToolResultBlock)
	if results[0].Content != "[cleared]" {
		t.Fatalf("expected first result cleared, got %q", results[0].Content)
	}
}

func TestAutoCompactWritesTranscript(t *testing.T) {
	root := t.TempDir()
	mgr := compact.Manager{Dir: root}
	messages := []llm.Message{{Role: "user", Content: "hello"}}
	compressed, err := mgr.AutoCompact(messages)
	if err != nil {
		t.Fatal(err)
	}
	if len(compressed) != 1 {
		t.Fatalf("unexpected compressed message count: %d", len(compressed))
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected one transcript file, got %d", len(entries))
	}
	if filepath.Ext(entries[0].Name()) != ".jsonl" {
		t.Fatalf("unexpected transcript file: %s", entries[0].Name())
	}
}

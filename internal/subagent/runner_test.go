package subagent_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"icoo_assistant/internal/agent"
	"icoo_assistant/internal/agents"
	"icoo_assistant/internal/llm"
	"icoo_assistant/internal/subagent"
	"icoo_assistant/internal/tools"
)

func TestRunWithAgentUsesTemplate(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "reviewer.md"), []byte("Review code carefully."), 0o644); err != nil {
		t.Fatal(err)
	}
	loader, err := agents.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	client := &llm.FakeClient{Responses: []llm.Response{
		{StopReason: "end", Text: "done"},
	}}
	registry, err := tools.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	runner := &subagent.Runner{
		Client:      client,
		Registry:    registry,
		AgentLoader: loader,
		Config:      agent.Config{SystemPrompt: "test", MaxRounds: 2},
	}
	if _, err := runner.RunWithAgent("reviewer", "Inspect auth."); err != nil {
		t.Fatal(err)
	}
	if len(client.Snapshots) == 0 || !strings.Contains(client.Snapshots[0], "Review code carefully.") || !strings.Contains(client.Snapshots[0], "Inspect auth.") {
		t.Fatalf("unexpected subagent snapshot: %#v", client.Snapshots)
	}
}

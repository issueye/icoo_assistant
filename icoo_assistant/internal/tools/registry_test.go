package tools_test

import (
	"testing"

	"icoo_assistant/internal/llm"
	"icoo_assistant/internal/tools"
)

func TestRegistryExecutesByName(t *testing.T) {
	registry, err := tools.NewRegistry(tools.Definition{
		Tool: llm.Tool{Name: "demo", Description: "demo", InputSchema: map[string]interface{}{}},
		Handler: func(call tools.Call) (string, error) {
			return "ok", nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := registry.Execute(tools.Call{ID: "1", Name: "demo", Input: map[string]interface{}{}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Content != "ok" {
		t.Fatalf("unexpected result: %q", result.Content)
	}
}

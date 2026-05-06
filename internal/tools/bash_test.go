package tools_test

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"icoo_assistant/internal/tools"
)

func TestBashToolExecutesOnCurrentPlatform(t *testing.T) {
	command := "printf hello"
	if runtime.GOOS == "windows" {
		command = "Write-Output hello"
	}
	tool := tools.NewBashTool(tools.CommandRunner{
		Workdir: t.TempDir(),
		Timeout: 5 * time.Second,
	})
	result, err := tool.Handler(tools.Call{Input: map[string]interface{}{"command": command}})
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(result) != "hello" {
		t.Fatalf("expected hello, got %q", result)
	}
}

func TestBashToolBlocksDeniedCommandPatterns(t *testing.T) {
	tool := tools.NewBashTool(tools.CommandRunner{
		Workdir:      t.TempDir(),
		Timeout:      5 * time.Second,
		DenyPatterns: []string{"rm *"},
	})
	result, err := tool.Handler(tools.Call{Input: map[string]interface{}{"command": "rm temp.txt"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "Command blocked by permission settings") {
		t.Fatalf("unexpected result: %q", result)
	}
}

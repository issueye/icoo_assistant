package main

import (
	"os"
	"strings"
	"testing"
)

func TestIsHelpRequest(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want bool
	}{
		{name: "double dash", args: []string{"--help"}, want: true},
		{name: "short flag", args: []string{"-h"}, want: true},
		{name: "plain help", args: []string{"help"}, want: true},
		{name: "query content", args: []string{"summarize repo"}, want: false},
		{name: "multiple args", args: []string{"help", "extra"}, want: false},
	}
	for _, tc := range cases {
		if got := isHelpRequest(tc.args); got != tc.want {
			t.Fatalf("%s: expected %v, got %v", tc.name, tc.want, got)
		}
	}
}

func TestPrintUsage(t *testing.T) {
	path := t.TempDir() + "/usage.txt"
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	printUsage(file)
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	output := string(data)
	for _, snippet := range []string{"icoo_assistant " + Version, "go run ./cmd/assistant [query]", "go run ./cmd/assistant check", ".env.example", "Without ANTHROPIC_API_KEY, assistant runs in fake mode"} {
		if !strings.Contains(output, snippet) {
			t.Fatalf("expected usage to contain %q, got %q", snippet, output)
		}
	}
	for _, snippet := range []string{
		"go run ./cmd/assistant doctor",
		"Recommended first-run path:",
		"1. go run ./cmd/assistant check",
		"In fake mode, steps 2-4 still work as a dry run",
		"Replace `go run ./cmd/assistant` with `assistant` if the binary is already installed.",
	} {
		if !strings.Contains(output, snippet) {
			t.Fatalf("expected usage to contain %q, got %q", snippet, output)
		}
	}
}

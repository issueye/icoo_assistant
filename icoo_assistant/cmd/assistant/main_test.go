package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintUsageIncludesVersionAndExamples(t *testing.T) {
	var buf bytes.Buffer
	printUsage(&buf)
	output := buf.String()
	if !strings.Contains(output, "icoo_assistant "+Version) {
		t.Fatalf("expected version banner, got %q", output)
	}
	if !strings.Contains(output, "assistant --help") {
		t.Fatalf("expected help example, got %q", output)
	}
	if !strings.Contains(output, "assistant check") {
		t.Fatalf("expected check example, got %q", output)
	}
	if !strings.Contains(output, "assistant doctor") {
		t.Fatalf("expected doctor alias, got %q", output)
	}
	if !strings.Contains(output, "Run `assistant check` before the first real task") {
		t.Fatalf("expected first-use guidance, got %q", output)
	}
	if !strings.Contains(output, "Recommended first-run path:") {
		t.Fatalf("expected first-run path guidance, got %q", output)
	}
	if !strings.Contains(output, "1. assistant check") {
		t.Fatalf("expected explicit assistant check step, got %q", output)
	}
	if !strings.Contains(output, "In fake mode, steps 2-4 still work as a dry run") {
		t.Fatalf("expected fake-mode dry-run note, got %q", output)
	}
	if !strings.Contains(output, "Without ANTHROPIC_API_KEY, assistant runs in fake mode") {
		t.Fatalf("expected mode notes, got %q", output)
	}
	if !strings.Contains(output, "See .env.example for supported settings.") {
		t.Fatalf("expected config hint, got %q", output)
	}
}

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"icoo_assistant/internal/agent"
	"icoo_assistant/internal/background"
	"icoo_assistant/internal/config"
	"icoo_assistant/internal/task"
)

func TestIsCheckRequest(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want bool
	}{
		{name: "double dash", args: []string{"--check"}, want: true},
		{name: "plain check", args: []string{"check"}, want: true},
		{name: "doctor alias", args: []string{"doctor"}, want: true},
		{name: "query content", args: []string{"summarize repo"}, want: false},
		{name: "multiple args", args: []string{"check", "extra"}, want: false},
	}
	for _, tc := range cases {
		if got := isCheckRequest(tc.args); got != tc.want {
			t.Fatalf("%s: expected %v, got %v", tc.name, tc.want, got)
		}
	}
}

func TestBuildSelfCheckReportCreatesRuntimeDirs(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("AGENT_SKILLS_DIR", "")
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	report, err := buildSelfCheckReport(cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, snippet := range []string{
		"self_check: ready",
		"client: ready mode=fake",
		"dotenv: missing",
		"skills_dir: missing",
		"fake_client_active:",
		"first_run_status: completed step=1 command=assistant check",
		"minimal_happy_path:",
		"1. assistant check",
		`2. assistant "先用 tool_catalog 总结当前可用工具，再说明 project_task、task_audit 和 agent_hook_audit 的边界"`,
		"transcript_dir: ready",
		"task_dir: ready",
		"background_dir: ready",
		"agent_hook_dir: ready",
	} {
		if !strings.Contains(report, snippet) {
			t.Fatalf("expected report to contain %q, got %q", snippet, report)
		}
	}
	for _, dir := range []string{
		resolveConfigPath(root, cfg.TranscriptDir),
		task.DefaultDir(root),
		background.DefaultDir(root),
		agent.DefaultHookDir(root),
	} {
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			t.Fatalf("expected runtime dir %s to exist after self-check", dir)
		}
	}
}

func TestBuildSelfCheckReportSupportsAnthropicAndSkillsDir(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AGENT_SKILLS_DIR", "")
	skillsDir := filepath.Join(root, "skills", "demo")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte("# demo"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ANTHROPIC_API_KEY", "test-key")
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	report, err := buildSelfCheckReport(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(report, "client: ready mode=anthropic") {
		t.Fatalf("expected anthropic mode, got %q", report)
	}
	if !strings.Contains(report, "skills_dir: ready") || !strings.Contains(report, "skill_count=1") {
		t.Fatalf("expected skills dir to be ready, got %q", report)
	}
	if !strings.Contains(report, "minimal_happy_path:") || strings.Contains(report, "0. optional: set ANTHROPIC_API_KEY") {
		t.Fatalf("expected anthropic happy path without fake-mode preface, got %q", report)
	}
	if !strings.Contains(report, "1. assistant check") {
		t.Fatalf("expected full first-run path to include assistant check, got %q", report)
	}
	if !strings.Contains(report, "next_step: continue with minimal_happy_path step=2") {
		t.Fatalf("expected next step to continue from step 2, got %q", report)
	}
	if strings.Contains(report, "fake_client_active:") {
		t.Fatalf("did not expect fake client advisory, got %q", report)
	}
}

func TestRunSelfCheckWritesReport(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("AGENT_SKILLS_DIR", "")
	cfg, err := config.Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := runSelfCheck(&out, cfg); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "next_step:") {
		t.Fatalf("expected next step guidance, got %q", out.String())
	}
	if !strings.Contains(out.String(), "minimal_happy_path:") {
		t.Fatalf("expected minimal happy path guidance, got %q", out.String())
	}
	if !strings.Contains(out.String(), "first_run_status: completed step=1 command=assistant check") {
		t.Fatalf("expected first-run status guidance, got %q", out.String())
	}
}

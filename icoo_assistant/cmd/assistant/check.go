package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"icoo_assistant/internal/agent"
	"icoo_assistant/internal/background"
	"icoo_assistant/internal/config"
	"icoo_assistant/internal/llm"
	"icoo_assistant/internal/task"
)

func runSelfCheck(out io.Writer, cfg config.Config) error {
	report, err := buildSelfCheckReport(cfg)
	if err != nil {
		return err
	}
	_, err = io.WriteString(out, report)
	return err
}

func buildSelfCheckReport(cfg config.Config) (string, error) {
	workdir, err := filepath.Abs(cfg.Workdir)
	if err != nil {
		return "", err
	}
	infoLines := []string{
		"self_check: ready",
		fmt.Sprintf("version: %s", Version),
	}
	workdirLine, err := describeDirectory("workdir", workdir, false)
	if err != nil {
		return "", err
	}
	infoLines = append(infoLines, workdirLine)

	dotEnvPath := filepath.Join(workdir, ".env")
	if _, err := os.Stat(dotEnvPath); err == nil {
		infoLines = append(infoLines, fmt.Sprintf("dotenv: present path=%s", dotEnvPath))
	} else if os.IsNotExist(err) {
		infoLines = append(infoLines, fmt.Sprintf("dotenv: missing path=%s optional=true", dotEnvPath))
	} else {
		return "", err
	}

	_, mode, err := llm.NewClientFromConfig(cfg)
	if err != nil {
		return "", err
	}
	infoLines = append(infoLines, fmt.Sprintf(
		"client: ready mode=%s model=%s max_tokens=%d prompt_cache=%t thinking=%t",
		mode,
		cfg.AnthropicModel,
		cfg.AnthropicMaxTokens,
		cfg.EnablePromptCache,
		cfg.EnableThinking,
	))
	infoLines = append(infoLines, fmt.Sprintf(
		"config: max_rounds=%d command_timeout=%s compact_threshold=%d",
		cfg.MaxRounds,
		cfg.CommandTimeout,
		cfg.CompactThreshold,
	))

	skillsLine, skillsAdvisory, err := describeSkillsDir(resolveConfigPath(workdir, cfg.SkillsDir))
	if err != nil {
		return "", err
	}
	infoLines = append(infoLines, skillsLine)

	transcriptLine, err := describeDirectory("transcript_dir", resolveConfigPath(workdir, cfg.TranscriptDir), true)
	if err != nil {
		return "", err
	}
	taskLine, err := describeDirectory("task_dir", task.DefaultDir(workdir), true)
	if err != nil {
		return "", err
	}
	backgroundLine, err := describeDirectory("background_dir", background.DefaultDir(workdir), true)
	if err != nil {
		return "", err
	}
	hookLine, err := describeDirectory("agent_hook_dir", agent.DefaultHookDir(workdir), true)
	if err != nil {
		return "", err
	}
	infoLines = append(infoLines, transcriptLine, taskLine, backgroundLine, hookLine)

	advisories := make([]string, 0, 2)
	if mode == "fake" {
		advisories = append(advisories, "fake_client_active: set ANTHROPIC_API_KEY in .env or shell to enable real model calls")
	}
	if skillsAdvisory != "" {
		advisories = append(advisories, skillsAdvisory)
	}

	lines := append([]string{}, infoLines...)
	if len(advisories) == 0 {
		lines = append(lines, "advisories: none")
	} else {
		lines = append(lines, "advisories:")
		for _, advisory := range advisories {
			lines = append(lines, "- "+advisory)
		}
	}
	lines = append(lines, "first_run_status: completed step=1 command=assistant check")
	lines = append(lines, "minimal_happy_path:")
	for _, step := range minimalHappyPathLines(mode) {
		lines = append(lines, step)
	}
	if mode == "fake" {
		lines = append(lines, `next_step: continue with minimal_happy_path step=2 command=assistant "先用 tool_catalog 总结当前可用工具，再说明 project_task、task_audit 和 agent_hook_audit 的边界"; set ANTHROPIC_API_KEY and rerun assistant check first if you want real model calls`)
	} else {
		lines = append(lines, `next_step: continue with minimal_happy_path step=2 command=assistant "先用 tool_catalog 总结当前可用工具，再说明 project_task、task_audit 和 agent_hook_audit 的边界"`)
	}
	return strings.Join(lines, "\n") + "\n", nil
}

func minimalHappyPathLines(mode string) []string {
	lines := make([]string, 0, 5)
	lines = append(lines, "1. assistant check")
	if mode == "fake" {
		lines = append(lines, "optional: set ANTHROPIC_API_KEY in .env or shell and rerun assistant check before step 2 if you want real model calls")
	}
	lines = append(lines,
		`2. assistant "先用 tool_catalog 总结当前可用工具，再说明 project_task、task_audit 和 agent_hook_audit 的边界"`,
		`3. assistant "创建一个项目任务，用于验证后台测试"`,
		`4. assistant "使用 tool_catalog action=audit_paths 说明审计入口，再给出 task_audit 和 agent_hook_audit 的查询示例"`,
	)
	return lines
}

func resolveConfigPath(workdir, path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return workdir
	}
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(workdir, path)
}

func describeDirectory(label, path string, createIfMissing bool) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("%s path required", label)
	}
	info, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) || !createIfMissing {
			return "", err
		}
		if err := os.MkdirAll(path, 0o755); err != nil {
			return "", err
		}
		info, err = os.Stat(path)
		if err != nil {
			return "", err
		}
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory: %s", label, path)
	}
	if createIfMissing {
		tempFile, err := os.CreateTemp(path, ".self-check-*")
		if err != nil {
			return "", err
		}
		tempName := tempFile.Name()
		if err := tempFile.Close(); err != nil {
			return "", err
		}
		if err := os.Remove(tempName); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%s: ready path=%s", label, path), nil
}

func describeSkillsDir(path string) (string, string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "skills_dir: not_configured path=", "", nil
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Sprintf("skills_dir: missing path=%s skill_count=0", path),
				fmt.Sprintf("skills_dir_missing: optional; add local skills under %s if you want repository-specific skills", path),
				nil
		}
		return "", "", err
	}
	if !info.IsDir() {
		return "", "", fmt.Errorf("skills dir is not a directory: %s", path)
	}
	skillFiles, err := filepath.Glob(filepath.Join(path, "*", "SKILL.md"))
	if err != nil {
		return "", "", err
	}
	return fmt.Sprintf("skills_dir: ready path=%s skill_count=%d", path, len(skillFiles)), "", nil
}

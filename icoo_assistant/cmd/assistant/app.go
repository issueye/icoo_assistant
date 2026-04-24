package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"icoo_assistant/internal/agent"
	"icoo_assistant/internal/background"
	"icoo_assistant/internal/compact"
	"icoo_assistant/internal/config"
	"icoo_assistant/internal/hookaudit"
	"icoo_assistant/internal/llm"
	"icoo_assistant/internal/skill"
	"icoo_assistant/internal/subagent"
	"icoo_assistant/internal/task"
	"icoo_assistant/internal/team"
	"icoo_assistant/internal/todo"
	"icoo_assistant/internal/tools"
	"icoo_assistant/internal/workspace"
)

type app struct {
	runner *agent.Runner
	mode   string
}

func (a *app) isFakeMode() bool {
	return strings.EqualFold(strings.TrimSpace(a.mode), "fake")
}

func (a *app) writeDegradedModeHint(out io.Writer) {
	if !a.isFakeMode() {
		return
	}
	_, _ = fmt.Fprintln(out, "warning: assistant is running in fake mode; set ANTHROPIC_API_KEY in .env or shell for real model calls.")
	_, _ = fmt.Fprintf(out, "hint: run `%s` to confirm the current mode and follow the reported minimal_happy_path; replace it with `%s` if the binary is already installed.\n", sourceCommand("check"), binaryCommand("check"))
}

func (a *app) writeFakeModeNoOutputHint(out io.Writer) {
	if !a.isFakeMode() {
		return
	}
	_, _ = fmt.Fprintln(out, "warning: no model output was produced because the fake client returns empty responses by design.")
	_, _ = fmt.Fprintln(out, "hint: this is expected in fake mode; set ANTHROPIC_API_KEY for real answers, or keep following the minimal_happy_path as a local dry run.")
}

func newApp(cfg config.Config) (*app, error) {
	ws, err := workspace.New(cfg.Workdir)
	if err != nil {
		return nil, err
	}
	todoManager := todo.NewManager()
	compactManager := &compact.Manager{
		Threshold:  cfg.CompactThreshold,
		KeepRecent: 3,
		Dir:        cfg.TranscriptDir,
	}
	backgroundManager, err := background.NewManager(
		background.DefaultDir(cfg.Workdir),
		cfg.Workdir,
		cfg.CommandTimeout,
	)
	if err != nil {
		return nil, err
	}
	taskManager, err := task.NewManager(task.DefaultDir(cfg.Workdir))
	if err != nil {
		return nil, err
	}
	teamManager, err := team.NewManager(team.DefaultDir(cfg.Workdir))
	if err != nil {
		return nil, err
	}
	backgroundManager.SetLifecycleHooks(task.NewBackgroundLifecycleLink(taskManager))
	hookWriter, err := agent.NewJSONLHook(agent.DefaultHookDir(cfg.Workdir))
	if err != nil {
		return nil, err
	}
	hooks := []agent.Hook{hookWriter}
	eventReader := hookaudit.NewReader(agent.DefaultHookDir(cfg.Workdir))
	skillLoader, err := skill.Load(cfg.SkillsDir)
	if err != nil {
		return nil, err
	}
	systemPrompt := cfg.SystemPrompt + "\n\nSkills available:\n" + skillLoader.Descriptions()
	baseCatalog := tools.DefaultToolCatalogEntries(false)
	baseRegistry, err := tools.NewRegistry(
		tools.NewBashTool(tools.CommandRunner{Workdir: cfg.Workdir, Timeout: cfg.CommandTimeout}),
		tools.NewReadFileTool(ws),
		tools.NewWriteFileTool(ws),
		tools.NewEditFileTool(ws),
		tools.NewBackgroundTool(backgroundManager),
		tools.NewAgentHookAuditTool(eventReader),
		tools.NewProjectTaskTool(taskManager, backgroundManager),
		tools.NewTeamRegistryTool(teamManager),
		tools.NewTeamMessageTool(teamManager),
		tools.NewTaskAuditTool(taskManager),
		tools.NewToolCatalogTool(baseCatalog),
		tools.NewTodoTool(todoManager),
		tools.NewCompactTool(),
		tools.NewLoadSkillTool(skillLoader),
	)
	if err != nil {
		return nil, err
	}
	client, mode, err := llm.NewClientFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	subRunner := &subagent.Runner{
		Client:   client,
		Registry: baseRegistry,
		Hooks:    hooks,
		Config: agent.Config{
			SystemPrompt: systemPrompt,
			MaxRounds:    cfg.MaxRounds,
		},
	}
	registry, err := tools.NewRegistry(
		tools.NewBashTool(tools.CommandRunner{Workdir: cfg.Workdir, Timeout: cfg.CommandTimeout}),
		tools.NewReadFileTool(ws),
		tools.NewWriteFileTool(ws),
		tools.NewEditFileTool(ws),
		tools.NewBackgroundTool(backgroundManager),
		tools.NewAgentHookAuditTool(eventReader),
		tools.NewProjectTaskTool(taskManager, backgroundManager),
		tools.NewTeamRegistryTool(teamManager),
		tools.NewTeamMessageTool(teamManager),
		tools.NewTaskAuditTool(taskManager),
		tools.NewToolCatalogTool(tools.DefaultToolCatalogEntries(true)),
		tools.NewTodoTool(todoManager),
		tools.NewCompactTool(),
		tools.NewTaskTool(),
		tools.NewLoadSkillTool(skillLoader),
	)
	if err != nil {
		return nil, err
	}
	return &app{
		runner: &agent.Runner{
			Client:         client,
			Registry:       registry,
			TodoManager:    todoManager,
			CompactManager: compactManager,
			SubagentRunner: subRunner,
			Background:     backgroundManager,
			Hooks:          hooks,
			Config: agent.Config{
				SystemPrompt: systemPrompt,
				MaxRounds:    cfg.MaxRounds,
			},
		},
		mode: mode,
	}, nil
}

func (a *app) execute(query string) (string, error) {
	messages, err := a.runner.Run([]llm.Message{{Role: "user", Content: query}})
	if err != nil {
		return "", err
	}
	if len(messages) == 0 {
		return "", nil
	}
	content := messages[len(messages)-1].Content
	if content == nil {
		return "", nil
	}
	if text, ok := content.(string); ok {
		return text, nil
	}
	return fmt.Sprintf("%v", content), nil
}

func (a *app) runOnce(out io.Writer, query string) error {
	a.writeDegradedModeHint(out)
	result, err := a.execute(strings.TrimSpace(query))
	if err != nil {
		return err
	}
	if result != "" {
		_, _ = fmt.Fprintln(out, result)
		return nil
	}
	a.writeFakeModeNoOutputHint(out)
	return nil
}

func (a *app) runREPL(in io.Reader, out io.Writer) error {
	_, _ = fmt.Fprintf(out, "assistant REPL started (%s client). Type exit to quit.\n", a.mode)
	if a.isFakeMode() {
		_, _ = fmt.Fprintln(out, "warning: REPL is running in fake mode; model-generated answers are disabled until ANTHROPIC_API_KEY is set.")
		_, _ = fmt.Fprintf(out, "hint: run `%s` outside the REPL if you want the current minimal_happy_path and setup guidance; replace it with `%s` if the binary is already installed.\n", sourceCommand("check"), binaryCommand("check"))
	}
	scanner := bufio.NewScanner(in)
	for {
		_, _ = fmt.Fprint(out, ">> ")
		if !scanner.Scan() {
			break
		}
		query := strings.TrimSpace(scanner.Text())
		if query == "" || query == "exit" {
			break
		}
		result, err := a.execute(query)
		if err != nil {
			_, _ = fmt.Fprintf(out, "error: %v\n", err)
			continue
		}
		if result != "" {
			_, _ = fmt.Fprintln(out, result)
			continue
		}
		a.writeFakeModeNoOutputHint(out)
	}
	return scanner.Err()
}

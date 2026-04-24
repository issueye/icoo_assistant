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
	"icoo_assistant/internal/todo"
	"icoo_assistant/internal/tools"
	"icoo_assistant/internal/workspace"
)

type app struct {
	runner *agent.Runner
	mode   string
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
	return fmt.Sprintf("%v", messages[len(messages)-1].Content), nil
}

func (a *app) runOnce(out io.Writer, query string) error {
	result, err := a.execute(strings.TrimSpace(query))
	if err != nil {
		return err
	}
	if result != "" {
		_, _ = fmt.Fprintln(out, result)
	}
	return nil
}

func (a *app) runREPL(in io.Reader, out io.Writer) error {
	_, _ = fmt.Fprintf(out, "assistant REPL started (%s client). Type exit to quit.\n", a.mode)
	if a.mode == "fake" {
		_, _ = fmt.Fprintln(out, "Set ANTHROPIC_API_KEY in env or .env to use the real Anthropic client.")
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
		}
	}
	return scanner.Err()
}

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
	"icoo_assistant/internal/memory"
	"icoo_assistant/internal/session"
	"icoo_assistant/internal/skill"
	"icoo_assistant/internal/subagent"
	"icoo_assistant/internal/task"
	"icoo_assistant/internal/todo"
	"icoo_assistant/internal/tools"
	"icoo_assistant/internal/workspace"
)

type app struct {
	runner            *agent.Runner
	mode              string
	streamingDisabled bool
	sessionManager    *session.Manager
	memoryManager     *memory.Manager
}

type streamedOutput struct {
	writer      io.Writer
	wroteText   bool
	lastChunkLF bool
}

func (s *streamedOutput) Write(chunk string) {
	if s == nil || chunk == "" {
		return
	}
	_, _ = io.WriteString(s.writer, chunk)
	s.wroteText = true
	s.lastChunkLF = strings.HasSuffix(chunk, "\n")
}

func (s *streamedOutput) Finish() {
	if s == nil || !s.wroteText || s.lastChunkLF {
		return
	}
	_, _ = fmt.Fprintln(s.writer)
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

func (a *app) writeNoOutputHint(out io.Writer) {
	if a.isFakeMode() {
		a.writeFakeModeNoOutputHint(out)
		return
	}
	_, _ = fmt.Fprintln(out, "warning: model call completed but returned no text content.")
	_, _ = fmt.Fprintln(out, "hint: if you are using icoo_proxy for cross-protocol routing, restart the rebuilt proxy and check /admin/requests for the upstream route.")
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
	memoryManager, err := memory.NewManager(memory.DefaultDir(cfg.Workdir))
	if err != nil {
		return nil, err
	}
	sessionManager, err := session.NewManager(session.DefaultDir(cfg.Workdir))
	if err != nil {
		return nil, err
	}
	skillLoader, err := skill.Load(cfg.SkillsDir)
	if err != nil {
		return nil, err
	}
	memoryContext := memoryManager.GenerateSessionContext()
	systemPrompt := cfg.SystemPrompt
	if memoryContext != "" {
		systemPrompt += "\n\n<memory_context>\n" + memoryContext + "</memory_context>"
	}
	systemPrompt += "\n\nSkills available:\n" + skillLoader.Descriptions()
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
		tools.NewSkillCreateTool(cfg.SkillsDir),
		tools.NewMemoryStoreTool(memoryManager),
		tools.NewMemoryRecallTool(memoryManager),
		tools.NewMemorySummarizeTool(memoryManager),
		tools.NewMemoryManageTool(memoryManager),
		tools.NewSessionTool(sessionManager),
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
		tools.NewSkillCreateTool(cfg.SkillsDir),
		tools.NewMemoryStoreTool(memoryManager),
		tools.NewMemoryRecallTool(memoryManager),
		tools.NewMemorySummarizeTool(memoryManager),
		tools.NewMemoryManageTool(memoryManager),
		tools.NewSessionTool(sessionManager),
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
		mode:              mode,
		streamingDisabled: !cfg.EnableStreaming,
		sessionManager:    sessionManager,
		memoryManager:     memoryManager,
	}, nil
}

func (a *app) withStreamHandler(handler func(string), fn func() error) error {
	if a.streamingDisabled {
		return fn()
	}
	previous := a.runner.StreamHandler
	a.runner.StreamHandler = handler
	defer func() {
		a.runner.StreamHandler = previous
	}()
	return fn()
}

func (a *app) execute(query string) (string, error) {
	_, result, err := a.executeMessages(nil, query)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (a *app) executeMessages(history []llm.Message, query string) ([]llm.Message, string, error) {
	messages := make([]llm.Message, len(history), len(history)+1)
	copy(messages, history)
	messages = append(messages, llm.Message{Role: "user", Content: query})
	messages, err := a.runner.Run(messages)
	if err != nil {
		return nil, "", err
	}
	return messages, renderLatestAssistantContent(messages), nil
}

func renderLatestAssistantContent(messages []llm.Message) string {
	if len(messages) == 0 {
		return ""
	}
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role != "assistant" {
			continue
		}
		content := messages[i].Content
		if content == nil {
			return ""
		}
		if text, ok := content.(string); ok {
			return text
		}
		return ""
	}
	return ""
}

func (a *app) runOnce(out io.Writer, query string) error {
	if a.sessionManager != nil {
		if activeSession, err := a.sessionManager.EnsureActive(); err == nil {
			a.memoryManager.SetSessionID(activeSession.ID)
		}
	}
	a.writeDegradedModeHint(out)
	a.writeDegradedModeHint(out)
	stream := &streamedOutput{writer: out}
	var result string
	err := a.withStreamHandler(stream.Write, func() error {
		var runErr error
		result, runErr = a.execute(strings.TrimSpace(query))
		return runErr
	})
	if err != nil {
		return err
	}
	if stream.wroteText {
		stream.Finish()
		return nil
	}
	if result != "" {
		_, _ = fmt.Fprintln(out, result)
		return nil
	}
	a.writeNoOutputHint(out)
	return nil
}

func (a *app) runREPL(in io.Reader, out io.Writer) error {
	var activeSession session.Session
	hasSession := false
	if a.sessionManager != nil {
		sess, err := a.sessionManager.EnsureActive()
		if err != nil {
			a.memoryManager.SetSessionID("")
		} else {
			activeSession = sess
			hasSession = true
			a.memoryManager.SetSessionID(activeSession.ID)
			_, _ = fmt.Fprintf(out, "session: %s [%s]\n", activeSession.ID, activeSession.Title)
		}
	}
	_, _ = fmt.Fprintf(out, "assistant REPL started (%s client). Type exit to quit.\n", a.mode)
	if a.isFakeMode() {
		_, _ = fmt.Fprintln(out, "warning: REPL is running in fake mode; model-generated answers are disabled until ANTHROPIC_API_KEY is set.")
		_, _ = fmt.Fprintf(out, "hint: run `%s` outside the REPL if you want the current minimal_happy_path and setup guidance; replace it with `%s` if the binary is already installed.\n", sourceCommand("check"), binaryCommand("check"))
	}
	scanner := bufio.NewScanner(in)
	conversation := make([]llm.Message, 0)
	for {
		_, _ = fmt.Fprint(out, ">> ")
		if !scanner.Scan() {
			break
		}
		query := strings.TrimSpace(scanner.Text())
		if query == "" {
			continue
		}
		if query == "exit" {
			break
		}
		stream := &streamedOutput{writer: out}
		var (
			nextConversation []llm.Message
			result           string
		)
		err := a.withStreamHandler(stream.Write, func() error {
			var runErr error
			nextConversation, result, runErr = a.executeMessages(conversation, query)
			return runErr
		})
		if err != nil {
			_, _ = fmt.Fprintf(out, "error: %v\n", err)
			continue
		}
		conversation = nextConversation
		if hasSession {
			_ = a.sessionManager.UpdateStats(activeSession.ID, 0, len(conversation))
		}
		if stream.wroteText {
			stream.Finish()
			continue
		}
		if result != "" {
			_, _ = fmt.Fprintln(out, result)
			continue
		}
		a.writeNoOutputHint(out)
	}
	if hasSession {
		sess, err := a.sessionManager.Close(activeSession.ID)
		if err == nil {
			_, _ = fmt.Fprintf(out, "Session %s closed. (rounds=%d, messages=%d)\n", sess.ID, sess.RoundCount, sess.MessageCount)
			if sess.Summary == "" {
				_, _ = fmt.Fprintln(out, "hint: use memory_summarize to persist session context before closing next time")
			}
		}
	}
	return scanner.Err()
}

package tools

import (
	"fmt"
	"strings"

	"icoo_assistant/internal/llm"
)

type ToolCatalogEntry struct {
	Name        string
	Summary     string
	UseWhen     string
	AvoidWhen   string
	Example     string
	Description string
}

func DefaultToolCatalogEntries(includeTask bool) []ToolCatalogEntry {
	entries := []ToolCatalogEntry{
		{
			Name:        "agent_hook_audit",
			Summary:     "Inspect recent agent hook events recorded on disk.",
			UseWhen:     "Use when recent runs, tool calls, compact actions, or background notification injections should be reviewed or summarized for debugging.",
			AvoidWhen:   "Avoid for project task execution history; use task_audit instead.",
			Example:     `{"action":"summary","limit":20}`,
			Description: "Reads recorded events from .agent-hooks/events.jsonl and supports recent-event filtering plus compact summaries.",
		},
		{
			Name:        "background",
			Summary:     "Start or inspect long-running shell commands.",
			UseWhen:     "Use for commands that may take a while or should continue outside the current round.",
			AvoidWhen:   "Avoid for quick one-shot commands; use bash instead.",
			Example:     `{"action":"start","command":"go test ./...","task_id":"task-1"}`,
			Description: "Pairs well with project_task when execution progress should be tracked.",
		},
		{
			Name:        "bash",
			Summary:     "Run a short shell command in the workspace.",
			UseWhen:     "Use for quick inspection, build, test, or automation commands that should finish in the current round.",
			AvoidWhen:   "Avoid for long-running commands; use background instead.",
			Example:     `{"command":"go test ./..."}`,
			Description: "The command runs in the configured workspace with the standard timeout guard.",
		},
		{
			Name:        "compact",
			Summary:     "Request conversation compaction for continuity.",
			UseWhen:     "Use when context is getting large and a compact summary should be produced.",
			AvoidWhen:   "Avoid for normal task state; use todo or project_task for persistent work tracking.",
			Example:     `{}`,
			Description: "This is about conversation continuity, not project state.",
		},
		{
			Name:        "edit_file",
			Summary:     "Replace exact text in an existing file.",
			UseWhen:     "Use for small, targeted changes when the existing text to replace is known.",
			AvoidWhen:   "Avoid for entirely new files or full rewrites; use write_file instead.",
			Example:     `{"path":"README.md","old_text":"old","new_text":"new"}`,
			Description: "Best for precise edits after reading the target file.",
		},
		{
			Name:        "skill_load",
			Summary:     "Load a named skill into the current run.",
			UseWhen:     "Use when a specialized local skill should guide the task.",
			AvoidWhen:   "Avoid when the task is already fully supported by the current toolset.",
			Example:     `{"name":"ui-ux-pro-max"}`,
			Description: "Returns the skill content so the agent can follow its workflow. Use skill_execute to run a skill in a subagent without polluting the main conversation.",
		},
		{
			Name:        "skill_execute",
			Summary:     "Execute a skill in a subagent to avoid polluting the main conversation context.",
			UseWhen:     "Use when a skill involves many intermediate steps that would bloat the main conversation. The skill runs with fresh context and returns only the summary.",
			AvoidWhen:   "Avoid for simple skill tasks that produce minimal output; use skill_load instead.",
			Example:     `{"name":"brainstorming","prompt":"Design a new authentication flow for the project"}`,
			Description: "Loads the skill content, prepends it to the task prompt, and runs everything in a subagent. Returns a concise summary. This is the preferred way to execute heavy skills.",
		},
		{
			Name:        "skill_create",
			Summary:     "Create a new domain skill by writing a SKILL.md file.",
			UseWhen:     "Use to add specialized knowledge, workflows, or reusable instructions as persistent skills.",
			AvoidWhen:   "Avoid for one-off instructions; use memory_store long_term instead.",
			Example:     `{"name":"my-skill","description":"Custom workflow for X","content":"# Steps\n1. Do this\n2. Do that"}`,
			Description: "Creates skills/<name>/SKILL.md with YAML frontmatter. Skills persist across sessions and can be loaded with load_skill.",
		},
		{
			Name:        "project_task",
			Summary:     "Manage persistent project-level tasks.",
			UseWhen:     "Use to create, list, update, or inspect durable project tasks and their latest execution context.",
			AvoidWhen:   "Avoid for audit-style history review or subagent delegation; use task_audit or task instead.",
			Example:     `{"action":"create","title":"Polish v0.1.0 docs"}`,
			Description: "This is the main durable task management entry point.",
		},
		{
			Name:        "read_file",
			Summary:     "Read file contents from the workspace.",
			UseWhen:     "Use to inspect source files, docs, or generated artifacts before making changes.",
			AvoidWhen:   "Avoid when only a targeted text replacement is needed and the exact content is already known.",
			Example:     `{"path":"README.md","limit":200}`,
			Description: "Supports an optional line limit for lighter inspection.",
		},
		{
			Name:        "task_audit",
			Summary:     "Inspect project task execution history from an audit angle.",
			UseWhen:     "Use when task execution history should be summarized, reviewed, reported, or filtered by execution status or failure reason.",
			AvoidWhen:   "Avoid for normal task CRUD; use project_task instead.",
			Example:     `{"action":"summary","id":"task-1"}`,
			Description: "Keeps audit queries separate from day-to-day project task operations and supports failure-focused summaries, reason classification, priority failure hints with selection basis, recent context, repeat-pattern hints, sample-target guidance, recent sample comparison hints, direct latest-vs-previous compare targets, focused change-point hints, lightweight stability-vs-change trend hints, plus status- or reason-filtered inspection with latest sample hints, latest failure command hints, latest failure error hints, latest failure signature hints, latest failure updated_at hints, latest failure entry hints, direct reason labels, lightweight latest-vs-previous role markers, and pair summaries in focused history views.",
		},
		{
			Name:        "todo",
			Summary:     "Track in-session progress for multi-step work.",
			UseWhen:     "Use to keep the current run organized with pending, in_progress, and completed items.",
			AvoidWhen:   "Avoid when state should persist across runs; use project_task instead.",
			Example:     `{"items":[{"text":"implement tool catalog","status":"in_progress"}]}`,
			Description: "This is lightweight session planning rather than durable project management.",
		},
		{
			Name:        "write_file",
			Summary:     "Write a full file in the workspace.",
			UseWhen:     "Use for new files or full-file rewrites.",
			AvoidWhen:   "Avoid for surgical updates to existing files; use edit_file instead.",
			Example:     `{"path":"docs/note.md","content":"hello"}`,
			Description: "The target file will be created or overwritten with the provided content.",
		},
		{
			Name:        "memory_store",
			Summary:     "Store information into the persistent memory system.",
			UseWhen:     "Use to remember facts, decisions, user preferences, and AI personality across sessions.",
			AvoidWhen:   "Avoid for ephemeral session state; use short_term type for temporary in-session memory.",
			Example:     `{"action":"set","type":"long_term","content":"User prefers concise responses","tags":["preference"],"importance":0.8}`,
			Description: "Supports short_term (session-only), long_term (persistent), ai_personality, and user_profile types.",
		},
		{
			Name:        "memory_recall",
			Summary:     "Recall information from the memory system.",
			UseWhen:     "Use to search, list, or retrieve memories by type, tags, keywords, or importance. Use 'context' action for session initialization.",
			AvoidWhen:   "Avoid when you already have the information in current context.",
			Example:     `{"action":"search","query":"preference","type":"long_term","limit":5}`,
			Description: "Cross-session retrieval with keyword search, tag filtering, and importance-based ranking.",
		},
		{
			Name:        "memory_summarize",
			Summary:     "Generate and persist a session summary for future sessions.",
			UseWhen:     "Use at the end of a significant session to capture key decisions, findings, and context.",
			AvoidWhen:   "Avoid for trivial or very short sessions.",
			Example:     `{"summary":"Implemented memory system","key_decisions":["Used file-based JSONL storage"],"key_findings":["Memory context injection works well"],"tags":["development","architecture"]}`,
			Description: "Stores a structured session summary with key decisions and findings for cross-session continuity.",
		},
		{
			Name:        "memory_manage",
			Summary:     "Manage existing memories: update, delete, retag, or consolidate.",
			UseWhen:     "Use to update memory content, adjust importance, change tags, delete obsolete memories, or consolidate similar entries.",
			AvoidWhen:   "Avoid for creating new memories; use memory_store instead.",
			Example:     `{"action":"tag","id":"mem-123","tags":["preference","coding-style"]}`,
			Description: "Supports update, delete, tag, and consolidate actions for memory maintenance.",
		},
		{
			Name:        "session",
			Summary:     "Manage sessions: create, close, switch, list, status, history, and archive.",
			UseWhen:     "Use to organize work into named sessions, switch between contexts, review session history, or archive completed sessions.",
			AvoidWhen:   "Avoid for ephemeral in-session task tracking; use todo instead.",
			Example:     `{"action":"create","title":"Feature X Refactoring","tags":["refactoring","backend"]}`,
			Description: "Sessions provide long-lived work contexts. Each session tracks round/message counts. Use memory_summarize to persist session context before closing.",
		},
	}
	if includeTask {
		entries = append(entries, ToolCatalogEntry{
			Name:        "task",
			Summary:     "Delegate bounded work to a subagent and get back a summary.",
			UseWhen:     "Use when a separate focused subtask should run with fresh context.",
			AvoidWhen:   "Avoid for durable project tracking; use project_task instead.",
			Example:     `{"prompt":"Review the background manager and summarize risks."}`,
			Description: "This is for delegation, not for persistent project planning.",
		})
	}
	return entries
}

func NewToolCatalogTool(entries []ToolCatalogEntry) Definition {
	index := make(map[string]ToolCatalogEntry, len(entries))
	for _, entry := range entries {
		index[entry.Name] = entry
	}
	return Definition{
		Tool: llm.Tool{
			Name:        "tool_catalog",
			Description: "Explain available tools, their boundaries, and recommended usage.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action": map[string]interface{}{"type": "string", "enum": []string{"list", "describe", "audit_paths"}},
					"name":   map[string]interface{}{"type": "string"},
				},
				"required": []string{"action"},
			},
		},
		Handler: func(call Call) (string, error) {
			action, _ := call.Input["action"].(string)
			switch strings.ToLower(strings.TrimSpace(action)) {
			case "list":
				return renderToolCatalogList(entries), nil
			case "audit_paths":
				return renderToolCatalogAuditPaths(), nil
			case "describe":
				name, _ := call.Input["name"].(string)
				name = strings.TrimSpace(name)
				if name == "" {
					return "", fmt.Errorf("name required for describe")
				}
				entry, ok := index[name]
				if !ok {
					return "", fmt.Errorf("unknown tool %q", name)
				}
				return renderToolCatalogEntry(entry), nil
			default:
				return "", fmt.Errorf("unsupported action %q", action)
			}
		},
	}
}

func renderToolCatalogList(entries []ToolCatalogEntry) string {
	lines := []string{
		fmt.Sprintf("available_tools: %d", len(entries)),
		"tools:",
	}
	for _, entry := range entries {
		lines = append(lines, fmt.Sprintf("- %s: %s", entry.Name, entry.Summary))
	}
	lines = append(lines, `hint: use {"action":"describe","name":"<tool>"} for boundary guidance`)
	lines = append(lines, `audit_hint: use {"action":"audit_paths"} for task/runtime audit navigation`)
	return strings.Join(lines, "\n")
}

func renderToolCatalogEntry(entry ToolCatalogEntry) string {
	lines := []string{
		fmt.Sprintf("name: %s", entry.Name),
		fmt.Sprintf("summary: %s", entry.Summary),
		fmt.Sprintf("use_when: %s", entry.UseWhen),
		fmt.Sprintf("avoid_when: %s", entry.AvoidWhen),
	}
	if entry.Description != "" {
		lines = append(lines, fmt.Sprintf("notes: %s", entry.Description))
	}
	if entry.Example != "" {
		lines = append(lines, fmt.Sprintf("example: %s", entry.Example))
	}
	return strings.Join(lines, "\n")
}

func renderToolCatalogAuditPaths() string {
	lines := []string{
		"audit_paths:",
		`- project_task action=get: inspect the latest durable task snapshot and most recent background context`,
		`- project_task action=history: inspect a compact task-centric execution history`,
		`- task_audit action=summary: inspect status counts, failure reason counts, priority failure hints with basis, recent context, repeat-pattern hints, sample-target guidance, recent sample comparison hints, direct latest-vs-previous compare targets, focused change-point hints, recent failure trend, and lightweight stability-vs-change trend hints before drilling into detailed history or focusing on one failure reason`,
		`- task_audit action=history: inspect stable project task history for reporting, review, or reason-focused drill-down, with latest_sample hints, latest_failure_command hints, latest_failure_error hints, latest_failure_signature hints, latest_failure_updated_at hints, latest_failure_entry hints, direct reason labels, plus lightweight role=previous/latest markers and pair_summary in focused two-sample views`,
		`- agent_hook_audit action=recent: inspect agent runtime events such as model calls, tool use, compact, and notifications`,
		"recommended_flows:",
		`- task_first: project_task get -> task_audit summary -> inspect priority_failure_basis, priority_failure_context, priority_failure_pattern_hint, priority_failure_sample_target, priority_failure_sample_compare, priority_failure_compare_target, priority_failure_change_hint, and priority_failure_trend_hint -> follow priority_failure_hint or task_audit summary reason=<reason> -> task_audit history and use latest_sample, latest_failure_command, latest_failure_error, latest_failure_signature, latest_failure_updated_at, latest_failure_entry, reason labels, role=previous/latest, plus pair_summary when comparing the latest pair`,
		`- runtime_first: agent_hook_audit recent -> agent_hook_audit recent run_id=<run> -> task_audit history when a task review is needed`,
		`hint: use {"action":"describe","name":"task_audit"} or {"action":"describe","name":"agent_hook_audit"} for per-tool boundaries`,
	}
	return strings.Join(lines, "\n")
}

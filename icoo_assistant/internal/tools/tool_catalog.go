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
			Name:        "load_skill",
			Summary:     "Load a named skill into the current run.",
			UseWhen:     "Use when a specialized local skill should guide the task.",
			AvoidWhen:   "Avoid when the task is already fully supported by the current toolset.",
			Example:     `{"name":"ui-ux-pro-max"}`,
			Description: "Returns the skill content so the agent can follow its workflow.",
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
			Name:        "team_registry",
			Summary:     "Manage persistent team config and teammate registry data.",
			UseWhen:     "Use to inspect or update .team config, register teammates, or review the current teammate roster before message-bus work exists.",
			AvoidWhen:   "Avoid for task history, subagent delegation, or teammate-to-teammate messaging; use task_audit, task, or future team message tools instead.",
			Example:     `{"action":"create","id":"alice","role":"reviewer"}`,
			Description: "This is the initial Team-stage storage entry point and persists data under .team.",
		},
		{
			Name:        "team_message",
			Summary:     "Write to and inspect persistent team inbox files and minimal request/response threads.",
			UseWhen:     "Use to send a lead/teammate message, start a request, reply to a request, inspect one recipient inbox, or review one minimal communication thread before live teammate loops exist.",
			AvoidWhen:   "Avoid for roster management or task history; use team_registry or task_audit instead.",
			Example:     `{"action":"request","to":"alice","body":"Review the patch","request_id":"req-1"}`,
			Description: "This is the Stage-B message-store entry point and persists JSONL inbox files under .team/inbox, including minimal request/reply threads.",
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

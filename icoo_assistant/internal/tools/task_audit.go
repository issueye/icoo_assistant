package tools

import (
	"fmt"
	"sort"
	"strings"

	"icoo_assistant/internal/llm"
	"icoo_assistant/internal/task"
)

type TaskAuditManager interface {
	Get(id string) (task.Task, error)
}

type priorityFailureSelection struct {
	Reason   string
	Count    int
	Basis    string
	Context  string
	Pattern  string
	Latest   *task.BackgroundContext
	LatestAt int
}

func NewTaskAuditTool(manager TaskAuditManager) Definition {
	return Definition{
		Tool: llm.Tool{
			Name:        "task_audit",
			Description: "Inspect project task execution audit data such as background history.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action": map[string]interface{}{"type": "string", "enum": []string{"history", "summary"}},
					"id":     map[string]interface{}{"type": "string"},
					"limit":  map[string]interface{}{"type": "integer"},
					"reason": map[string]interface{}{"type": "string"},
					"status": map[string]interface{}{"type": "string"},
				},
				"required": []string{"action", "id"},
			},
		},
		Handler: func(call Call) (string, error) {
			action, _ := call.Input["action"].(string)
			switch strings.ToLower(strings.TrimSpace(action)) {
			case "history":
				id, _ := call.Input["id"].(string)
				if strings.TrimSpace(id) == "" {
					return "", fmt.Errorf("id required for history")
				}
				item, err := manager.Get(id)
				if err != nil {
					return "", err
				}
				limit := intFromInput(call.Input["limit"], 10)
				statusFilter := normalizeAuditStatusFilter(call.Input["status"])
				reasonFilter := normalizeAuditReasonFilter(call.Input["reason"])
				return renderTaskAuditHistory(item, limit, statusFilter, reasonFilter), nil
			case "summary":
				id, _ := call.Input["id"].(string)
				if strings.TrimSpace(id) == "" {
					return "", fmt.Errorf("id required for summary")
				}
				item, err := manager.Get(id)
				if err != nil {
					return "", err
				}
				statusFilter := normalizeAuditStatusFilter(call.Input["status"])
				reasonFilter := normalizeAuditReasonFilter(call.Input["reason"])
				return renderTaskAuditSummary(item, statusFilter, reasonFilter), nil
			default:
				return "", fmt.Errorf("unsupported action %q", action)
			}
		},
	}
}

func renderTaskAuditHistory(item task.Task, limit int, statusFilter, reasonFilter string) string {
	filtered := applyTaskAuditFilters(item.BackgroundHistory, statusFilter, reasonFilter)
	recent := recentBackgroundHistory(filtered, limit)
	lines := []string{
		fmt.Sprintf("task_id: %s", item.ID),
		fmt.Sprintf("title: %s", item.Title),
		fmt.Sprintf("history_count: %d", len(item.BackgroundHistory)),
		fmt.Sprintf("filtered_count: %d", len(filtered)),
		fmt.Sprintf("returned_count: %d", len(recent)),
	}
	if statusFilter != "" {
		lines = append(lines, fmt.Sprintf("filter_status: %s", statusFilter))
	}
	if reasonFilter != "" {
		lines = append(lines, fmt.Sprintf("filter_reason: %s", reasonFilter))
	}
	if len(filtered) == 0 {
		lines = append(lines, "entries: none")
		lines = append(lines, fmt.Sprintf("latest_task_view: project_task action=get id=%s", item.ID))
		lines = append(lines, `runtime_view_hint: use agent_hook_audit action=recent or action=summary for runtime-side investigation`)
		return strings.Join(lines, "\n")
	}
	lines = append(lines, "entries:")
	for index, entry := range recent {
		line := fmt.Sprintf("%d. job_id=%s status=%s updated_at=%s", index+1, entry.JobID, entry.Status, entry.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"))
		if entry.Command != "" {
			line = fmt.Sprintf("%s command=%s", line, entry.Command)
		}
		if entry.Error != "" {
			line = fmt.Sprintf("%s error=%s", line, entry.Error)
		}
		lines = append(lines, line)
	}
	lines = append(lines, fmt.Sprintf("latest_task_view: project_task action=get id=%s", item.ID))
	lines = append(lines, `runtime_view_hint: use agent_hook_audit action=summary or action=recent name=agent.tool.completed to inspect runtime-side execution context`)
	return strings.Join(lines, "\n")
}

func renderTaskAuditSummary(item task.Task, statusFilter, reasonFilter string) string {
	filtered := applyTaskAuditFilters(item.BackgroundHistory, statusFilter, reasonFilter)
	failureSource := item.BackgroundHistory
	if statusFilter != "" || reasonFilter != "" {
		failureSource = filtered
	}
	failures := filterBackgroundHistoryByStatus(failureSource, "failed")
	lines := []string{
		fmt.Sprintf("task_id: %s", item.ID),
		fmt.Sprintf("title: %s", item.Title),
		fmt.Sprintf("history_count: %d", len(item.BackgroundHistory)),
		fmt.Sprintf("filtered_count: %d", len(filtered)),
	}
	if statusFilter != "" {
		lines = append(lines, fmt.Sprintf("filter_status: %s", statusFilter))
	}
	if reasonFilter != "" {
		lines = append(lines, fmt.Sprintf("filter_reason: %s", reasonFilter))
	}
	if len(item.BackgroundHistory) == 0 {
		lines = append(lines, "status_counts: none")
		lines = append(lines, "failure_reason_counts: none")
		lines = append(lines, "priority_failure_reason: none")
		lines = append(lines, "priority_failure_basis: none")
		lines = append(lines, "priority_failure_context: none")
		lines = append(lines, "priority_failure_pattern_hint: none")
		lines = append(lines, "priority_failure_hint: none")
		lines = append(lines, "latest_failure_by_reason: none")
		lines = append(lines, "recent_failure_trend: none")
		lines = append(lines, "latest_entry: none")
		lines = append(lines, "latest_failure: none")
		lines = append(lines, "latest_failure_reason: none")
		lines = append(lines, fmt.Sprintf("history_hint: use task_audit action=history id=%s", item.ID))
		lines = append(lines, `runtime_view_hint: use agent_hook_audit action=summary for runtime-side troubleshooting`)
		return strings.Join(lines, "\n")
	}
	lines = append(lines, "status_counts:")
	for _, line := range sortedCountLines(backgroundStatusCounts(item.BackgroundHistory)) {
		lines = append(lines, fmt.Sprintf("- %s", line))
	}
	if len(failures) == 0 {
		lines = append(lines, "failure_reason_counts: none")
		lines = append(lines, "priority_failure_reason: none")
		lines = append(lines, "priority_failure_basis: none")
		lines = append(lines, "priority_failure_context: none")
		lines = append(lines, "priority_failure_pattern_hint: none")
		lines = append(lines, "priority_failure_hint: none")
		lines = append(lines, "latest_failure_by_reason: none")
		lines = append(lines, "recent_failure_trend: none")
	} else {
		selection := selectPriorityFailureReason(failures)
		lines = append(lines, "failure_reason_counts:")
		for _, line := range sortedCountLines(backgroundFailureReasonCounts(failures)) {
			lines = append(lines, fmt.Sprintf("- %s", line))
		}
		lines = append(lines, fmt.Sprintf("priority_failure_reason: %s count=%d", selection.Reason, selection.Count))
		lines = append(lines, fmt.Sprintf("priority_failure_basis: %s", selection.Basis))
		lines = append(lines, fmt.Sprintf("priority_failure_context: %s", selection.Context))
		lines = append(lines, fmt.Sprintf("priority_failure_pattern_hint: %s", selection.Pattern))
		lines = append(lines, fmt.Sprintf("priority_failure_hint: use task_audit action=summary id=%s reason=%s, then task_audit action=history id=%s reason=%s", item.ID, selection.Reason, item.ID, selection.Reason))
		lines = append(lines, "latest_failure_by_reason:")
		for _, line := range renderLatestFailureByReasonLines(failures) {
			lines = append(lines, fmt.Sprintf("- %s", line))
		}
		lines = append(lines, "recent_failure_trend:")
		for _, line := range renderRecentFailureTrendLines(failures, 3) {
			lines = append(lines, fmt.Sprintf("- %s", line))
		}
	}
	lines = append(lines, fmt.Sprintf("latest_entry: %s", renderBackgroundContextSummary(item.BackgroundHistory[len(item.BackgroundHistory)-1])))
	latestFailure := latestBackgroundByStatus(item.BackgroundHistory, "failed")
	if latestFailure == nil {
		lines = append(lines, "latest_failure: none")
		lines = append(lines, "latest_failure_reason: none")
	} else {
		lines = append(lines, fmt.Sprintf("latest_failure: %s", renderBackgroundContextSummary(*latestFailure)))
		lines = append(lines, fmt.Sprintf("latest_failure_reason: %s", classifyBackgroundFailureReason(*latestFailure)))
	}
	if len(filtered) == 0 {
		lines = append(lines, "matched_latest_entry: none")
	} else {
		lines = append(lines, fmt.Sprintf("matched_latest_entry: %s", renderBackgroundContextSummary(filtered[len(filtered)-1])))
	}
	lines = append(lines, fmt.Sprintf("history_hint: use task_audit action=history id=%s", item.ID))
	filterSuffix := renderTaskAuditFilterSuffix(statusFilter, reasonFilter)
	if filterSuffix != "" {
		lines = append(lines, fmt.Sprintf("filtered_history_hint: use task_audit action=history id=%s%s", item.ID, filterSuffix))
	} else {
		lines = append(lines, fmt.Sprintf("failure_history_hint: use task_audit action=history id=%s status=failed", item.ID))
	}
	lines = append(lines, `runtime_view_hint: use agent_hook_audit action=summary or action=recent for runtime-side troubleshooting`)
	return strings.Join(lines, "\n")
}

func selectPriorityFailureReason(history []task.BackgroundContext) priorityFailureSelection {
	counts := backgroundFailureReasonCounts(history)
	selection := priorityFailureSelection{
		LatestAt: -1,
	}
	for reason, count := range counts {
		latestIndex := latestBackgroundReasonIndex(history, reason)
		if count > selection.Count || (count == selection.Count && latestIndex > selection.LatestAt) || (count == selection.Count && latestIndex == selection.LatestAt && (selection.Reason == "" || reason < selection.Reason)) {
			selection.Reason = reason
			selection.Count = count
			selection.LatestAt = latestIndex
			selection.Latest = latestBackgroundByReason(history, reason)
		}
	}
	selection.Basis = renderPriorityFailureBasis(selection, counts)
	selection.Context = renderPriorityFailureContext(selection)
	selection.Pattern = renderPriorityFailurePatternHint(history, selection.Reason)
	return selection
}

func renderPriorityFailureBasis(selection priorityFailureSelection, counts map[string]int) string {
	if selection.Reason == "" {
		return "none"
	}
	chosenBy := "highest_count"
	for reason, count := range counts {
		if reason == selection.Reason {
			continue
		}
		if count == selection.Count {
			chosenBy = "latest_occurrence"
			break
		}
	}
	latestJobID := ""
	latestUpdatedAt := ""
	if selection.Latest != nil {
		latestJobID = selection.Latest.JobID
		latestUpdatedAt = selection.Latest.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	return fmt.Sprintf("chosen_by=%s count=%d latest_job_id=%s latest_updated_at=%s", chosenBy, selection.Count, latestJobID, latestUpdatedAt)
}

func renderPriorityFailureContext(selection priorityFailureSelection) string {
	if selection.Latest == nil {
		return "none"
	}
	return renderBackgroundContextSummary(*selection.Latest)
}

func renderPriorityFailurePatternHint(history []task.BackgroundContext, reason string) string {
	reasonHistory := recentBackgroundHistory(filterBackgroundHistoryByReason(history, reason), 3)
	if len(reasonHistory) == 0 {
		return "none"
	}
	type patternCandidate struct {
		signature string
		count     int
		latestAt  int
	}
	candidates := map[string]patternCandidate{}
	for index, entry := range reasonHistory {
		signature := normalizeBackgroundFailurePattern(entry)
		candidate := candidates[signature]
		candidate.signature = signature
		candidate.count++
		candidate.latestAt = index
		candidates[signature] = candidate
	}
	best := patternCandidate{}
	for _, candidate := range candidates {
		if candidate.count > best.count || (candidate.count == best.count && candidate.latestAt > best.latestAt) || (candidate.count == best.count && candidate.latestAt == best.latestAt && (best.signature == "" || candidate.signature < best.signature)) {
			best = candidate
		}
	}
	patternType := "single"
	if best.count >= 2 {
		patternType = "repeat"
	}
	return fmt.Sprintf("pattern=%s count=%d signature=%s", patternType, best.count, best.signature)
}

func applyTaskAuditFilters(history []task.BackgroundContext, statusFilter, reasonFilter string) []task.BackgroundContext {
	filtered := filterBackgroundHistoryByStatus(history, statusFilter)
	return filterBackgroundHistoryByReason(filtered, reasonFilter)
}

func filterBackgroundHistoryByStatus(history []task.BackgroundContext, status string) []task.BackgroundContext {
	if status == "" {
		return history
	}
	filtered := make([]task.BackgroundContext, 0, len(history))
	for _, entry := range history {
		if strings.EqualFold(strings.TrimSpace(entry.Status), status) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func filterBackgroundHistoryByReason(history []task.BackgroundContext, reason string) []task.BackgroundContext {
	if reason == "" {
		return history
	}
	filtered := make([]task.BackgroundContext, 0, len(history))
	for _, entry := range history {
		if !strings.EqualFold(strings.TrimSpace(entry.Status), "failed") {
			continue
		}
		if classifyBackgroundFailureReason(entry) == reason {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func normalizeAuditStatusFilter(raw interface{}) string {
	value, _ := raw.(string)
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeAuditReasonFilter(raw interface{}) string {
	value, _ := raw.(string)
	return strings.ToLower(strings.TrimSpace(value))
}

func renderTaskAuditFilterSuffix(statusFilter, reasonFilter string) string {
	parts := make([]string, 0, 2)
	if statusFilter != "" {
		parts = append(parts, fmt.Sprintf("status=%s", statusFilter))
	}
	if reasonFilter != "" {
		parts = append(parts, fmt.Sprintf("reason=%s", reasonFilter))
	}
	if len(parts) == 0 {
		return ""
	}
	return " " + strings.Join(parts, " ")
}

func backgroundStatusCounts(history []task.BackgroundContext) map[string]int {
	counts := map[string]int{}
	for _, entry := range history {
		status := strings.ToLower(strings.TrimSpace(entry.Status))
		if status == "" {
			status = "unknown"
		}
		counts[status]++
	}
	return counts
}

func latestBackgroundByStatus(history []task.BackgroundContext, status string) *task.BackgroundContext {
	status = strings.ToLower(strings.TrimSpace(status))
	for index := len(history) - 1; index >= 0; index-- {
		entry := history[index]
		if strings.EqualFold(strings.TrimSpace(entry.Status), status) {
			copied := entry
			return &copied
		}
	}
	return nil
}

func renderBackgroundContextSummary(entry task.BackgroundContext) string {
	line := fmt.Sprintf("job_id=%s status=%s updated_at=%s", entry.JobID, entry.Status, entry.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"))
	if entry.Command != "" {
		line = fmt.Sprintf("%s command=%s", line, entry.Command)
	}
	if entry.Error != "" {
		line = fmt.Sprintf("%s error=%s", line, entry.Error)
	}
	return line
}

func backgroundFailureReasonCounts(history []task.BackgroundContext) map[string]int {
	counts := map[string]int{}
	for _, entry := range history {
		counts[classifyBackgroundFailureReason(entry)]++
	}
	return counts
}

func classifyBackgroundFailureReason(entry task.BackgroundContext) string {
	errorText := strings.ToLower(strings.TrimSpace(entry.Error))
	switch {
	case errorText == "":
		return "unknown"
	case strings.HasPrefix(errorText, "timeout after"):
		return "timeout"
	default:
		return "command_error"
	}
}

func normalizeBackgroundFailurePattern(entry task.BackgroundContext) string {
	errorText := strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(entry.Error)), " "))
	switch {
	case errorText == "":
		return "unknown"
	case strings.HasPrefix(errorText, "timeout after"):
		return "timeout after <duration>"
	default:
		return errorText
	}
}

func latestBackgroundByReason(history []task.BackgroundContext, reason string) *task.BackgroundContext {
	reason = strings.ToLower(strings.TrimSpace(reason))
	for index := len(history) - 1; index >= 0; index-- {
		entry := history[index]
		if classifyBackgroundFailureReason(entry) == reason {
			copied := entry
			return &copied
		}
	}
	return nil
}

func latestBackgroundReasonIndex(history []task.BackgroundContext, reason string) int {
	reason = strings.ToLower(strings.TrimSpace(reason))
	for index := len(history) - 1; index >= 0; index-- {
		entry := history[index]
		if classifyBackgroundFailureReason(entry) == reason {
			return index
		}
	}
	return -1
}

func renderLatestFailureByReasonLines(history []task.BackgroundContext) []string {
	counts := backgroundFailureReasonCounts(history)
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		entry := latestBackgroundByReason(history, key)
		if entry == nil {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s => %s", key, renderBackgroundContextSummary(*entry)))
	}
	return lines
}

func renderRecentFailureTrendLines(history []task.BackgroundContext, limit int) []string {
	recent := recentBackgroundHistory(history, limit)
	if len(recent) == 0 {
		return []string{"none"}
	}
	lines := make([]string, 0, len(recent))
	for _, entry := range recent {
		lines = append(lines, fmt.Sprintf("reason=%s %s", classifyBackgroundFailureReason(entry), renderBackgroundContextSummary(entry)))
	}
	return lines
}

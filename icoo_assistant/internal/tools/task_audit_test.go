package tools_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"icoo_assistant/internal/task"
	"icoo_assistant/internal/tools"
)

func TestTaskAuditToolHistory(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(task.CreateInput{
		ID:    "task-a",
		Title: "Inspect audit history",
	}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if _, err := manager.RecordBackground("task-a", task.BackgroundContext{
			JobID:   fmt.Sprintf("job-%d", i),
			Status:  "completed",
			Command: fmt.Sprintf("cmd-%d", i),
		}); err != nil {
			t.Fatal(err)
		}
	}
	tool := tools.NewTaskAuditTool(manager)
	result, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "history",
		"id":     "task-a",
		"limit":  float64(2),
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "task_id: task-a") {
		t.Fatalf("unexpected audit result: %q", result)
	}
	if !strings.Contains(result, "history_count: 3") || !strings.Contains(result, "returned_count: 2") {
		t.Fatalf("unexpected history counters: %q", result)
	}
	if !strings.Contains(result, "filtered_count: 3") {
		t.Fatalf("expected filtered count, got %q", result)
	}
	if !strings.Contains(result, "job_id=job-1") || !strings.Contains(result, "job_id=job-2") {
		t.Fatalf("expected recent entries, got %q", result)
	}
	if strings.Contains(result, "reason=") {
		t.Fatalf("did not expect reason label on completed-only history, got %q", result)
	}
	if strings.Contains(result, "job_id=job-0") {
		t.Fatalf("expected limited audit history, got %q", result)
	}
	if !strings.Contains(result, "latest_task_view: project_task action=get id=task-a") {
		t.Fatalf("expected latest task view hint, got %q", result)
	}
	if !strings.Contains(result, "runtime_view_hint: use agent_hook_audit action=summary or action=recent") {
		t.Fatalf("expected runtime-side hint, got %q", result)
	}
}

func TestTaskAuditToolSummary(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(task.CreateInput{
		ID:    "task-a",
		Title: "Summarize failures",
	}); err != nil {
		t.Fatal(err)
	}
	for _, entry := range []task.BackgroundContext{
		{JobID: "job-1", Status: "completed", Command: "cmd-1"},
		{JobID: "job-2", Status: "failed", Command: "cmd-2", Error: "boom"},
		{JobID: "job-3", Status: "completed", Command: "cmd-3"},
		{JobID: "job-4", Status: "failed", Command: "cmd-4", Error: "timeout after 5s"},
	} {
		if _, err := manager.RecordBackground("task-a", entry); err != nil {
			t.Fatal(err)
		}
	}
	tool := tools.NewTaskAuditTool(manager)
	result, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "summary",
		"id":     "task-a",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "history_count: 4") || !strings.Contains(result, "filtered_count: 4") {
		t.Fatalf("expected summary counters, got %q", result)
	}
	if !strings.Contains(result, "- completed=2") || !strings.Contains(result, "- failed=2") {
		t.Fatalf("expected status counts, got %q", result)
	}
	if !strings.Contains(result, "- command_error=1") || !strings.Contains(result, "- timeout=1") {
		t.Fatalf("expected failure reason counts, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_reason: timeout count=1") {
		t.Fatalf("expected priority failure reason, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_basis: chosen_by=latest_occurrence count=1 latest_job_id=job-4") {
		t.Fatalf("expected priority failure basis, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_context: job_id=job-4 status=failed") || !strings.Contains(result, "command=cmd-4") || !strings.Contains(result, "error=timeout after 5s") {
		t.Fatalf("expected priority failure context, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_pattern_hint: pattern=single count=1 signature=timeout after <duration>") {
		t.Fatalf("expected priority failure pattern hint, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_sample_target: job_id=job-4") || !strings.Contains(result, "history_command=task_audit action=history id=<task-id> reason=timeout limit=1") {
		t.Fatalf("expected priority failure sample target, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_sample_compare: sample_count=1 latest_job_id=job-4 comparison=insufficient_samples") {
		t.Fatalf("expected priority failure sample compare, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_compare_target: sample_count=1 compare=insufficient_samples latest_job_id=job-4 history_command=task_audit action=history id=task-a reason=timeout limit=1") {
		t.Fatalf("expected priority failure compare target, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_change_hint: sample_count=1 change=insufficient_samples focus=collect_more_samples latest_job_id=job-4") {
		t.Fatalf("expected priority failure change hint, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_trend_hint: trend=emerging sample_count=1 latest_signature=timeout after <duration>") {
		t.Fatalf("expected priority failure trend hint, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_hint: use task_audit action=summary id=task-a reason=timeout, then task_audit action=history id=task-a reason=timeout limit=1") {
		t.Fatalf("expected priority failure hint, got %q", result)
	}
	if !strings.Contains(result, "latest_failure_by_reason:") {
		t.Fatalf("expected per-reason failure samples, got %q", result)
	}
	if !strings.Contains(result, "- command_error => job_id=job-2 status=failed") || !strings.Contains(result, "- timeout => job_id=job-4 status=failed") {
		t.Fatalf("expected latest sample for each failure reason, got %q", result)
	}
	if !strings.Contains(result, "recent_failure_trend:") {
		t.Fatalf("expected failure trend section, got %q", result)
	}
	if !strings.Contains(result, "- reason=command_error job_id=job-2 status=failed") || !strings.Contains(result, "- reason=timeout job_id=job-4 status=failed") {
		t.Fatalf("expected recent failure trend lines, got %q", result)
	}
	if !strings.Contains(result, "latest_failure: job_id=job-4 status=failed") {
		t.Fatalf("expected latest failure summary, got %q", result)
	}
	if !strings.Contains(result, "latest_failure_reason: timeout") {
		t.Fatalf("expected latest failure reason, got %q", result)
	}
	if !strings.Contains(result, "failure_history_hint: use task_audit action=history id=task-a status=failed") {
		t.Fatalf("expected failure history hint, got %q", result)
	}
}

func TestTaskAuditToolHistoryEmpty(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(task.CreateInput{
		ID:    "task-a",
		Title: "No runs yet",
	}); err != nil {
		t.Fatal(err)
	}
	tool := tools.NewTaskAuditTool(manager)
	result, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "history",
		"id":     "task-a",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "entries: none") {
		t.Fatalf("unexpected empty audit result: %q", result)
	}
	if !strings.Contains(result, "latest_task_view: project_task action=get id=task-a") {
		t.Fatalf("expected latest task view hint, got %q", result)
	}
}

func TestTaskAuditToolHistoryCanFilterByStatus(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(task.CreateInput{
		ID:    "task-a",
		Title: "Inspect filtered audit history",
	}); err != nil {
		t.Fatal(err)
	}
	for _, entry := range []task.BackgroundContext{
		{JobID: "job-1", Status: "completed", Command: "cmd-1"},
		{JobID: "job-2", Status: "failed", Command: "cmd-2", Error: "boom"},
		{JobID: "job-3", Status: "completed", Command: "cmd-3"},
	} {
		if _, err := manager.RecordBackground("task-a", entry); err != nil {
			t.Fatal(err)
		}
	}
	tool := tools.NewTaskAuditTool(manager)
	result, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "history",
		"id":     "task-a",
		"status": "failed",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "filter_status: failed") {
		t.Fatalf("expected status filter, got %q", result)
	}
	if !strings.Contains(result, "filtered_count: 1") || !strings.Contains(result, "returned_count: 1") {
		t.Fatalf("expected filtered counters, got %q", result)
	}
	if !strings.Contains(result, "job_id=job-2 status=failed") {
		t.Fatalf("expected failed job in result, got %q", result)
	}
	if !strings.Contains(result, "job_id=job-2 status=failed") || !strings.Contains(result, "reason=command_error") {
		t.Fatalf("expected failure reason on failed history entry, got %q", result)
	}
	if strings.Contains(result, "job_id=job-1") || strings.Contains(result, "job_id=job-3") {
		t.Fatalf("expected only failed jobs, got %q", result)
	}
}

func TestTaskAuditToolSummaryCanFilterByStatus(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(task.CreateInput{
		ID:    "task-a",
		Title: "Summarize filtered failures",
	}); err != nil {
		t.Fatal(err)
	}
	for _, entry := range []task.BackgroundContext{
		{JobID: "job-1", Status: "completed", Command: "cmd-1"},
		{JobID: "job-2", Status: "failed", Command: "cmd-2", Error: "boom"},
		{JobID: "job-3", Status: "failed", Command: "cmd-3", Error: "timeout after 5s"},
	} {
		if _, err := manager.RecordBackground("task-a", entry); err != nil {
			t.Fatal(err)
		}
	}
	tool := tools.NewTaskAuditTool(manager)
	result, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "summary",
		"id":     "task-a",
		"status": "failed",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "filter_status: failed") || !strings.Contains(result, "filtered_count: 2") {
		t.Fatalf("expected filtered summary counters, got %q", result)
	}
	if !strings.Contains(result, "- command_error=1") || !strings.Contains(result, "- timeout=1") {
		t.Fatalf("expected filtered failure reason counts, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_reason: timeout count=1") {
		t.Fatalf("expected filtered priority failure reason, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_basis: chosen_by=latest_occurrence count=1 latest_job_id=job-3") {
		t.Fatalf("expected filtered priority failure basis, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_context: job_id=job-3 status=failed") || !strings.Contains(result, "command=cmd-3") || !strings.Contains(result, "error=timeout after 5s") {
		t.Fatalf("expected filtered priority failure context, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_pattern_hint: pattern=single count=1 signature=timeout after <duration>") {
		t.Fatalf("expected filtered priority failure pattern hint, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_sample_target: job_id=job-3") {
		t.Fatalf("expected filtered priority failure sample target, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_sample_compare: sample_count=1 latest_job_id=job-3 comparison=insufficient_samples") {
		t.Fatalf("expected filtered priority failure sample compare, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_compare_target: sample_count=1 compare=insufficient_samples latest_job_id=job-3 history_command=task_audit action=history id=task-a reason=timeout limit=1") {
		t.Fatalf("expected filtered priority failure compare target, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_change_hint: sample_count=1 change=insufficient_samples focus=collect_more_samples latest_job_id=job-3") {
		t.Fatalf("expected filtered priority failure change hint, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_trend_hint: trend=emerging sample_count=1 latest_signature=timeout after <duration>") {
		t.Fatalf("expected filtered priority failure trend hint, got %q", result)
	}
	if !strings.Contains(result, "- timeout => job_id=job-3 status=failed") {
		t.Fatalf("expected filtered latest sample per reason, got %q", result)
	}
	if !strings.Contains(result, "- reason=timeout job_id=job-3 status=failed") {
		t.Fatalf("expected filtered failure trend line, got %q", result)
	}
	if !strings.Contains(result, "matched_latest_entry: job_id=job-3 status=failed") {
		t.Fatalf("expected filtered latest entry, got %q", result)
	}
	if !strings.Contains(result, "latest_failure_reason: timeout") {
		t.Fatalf("expected latest failure reason, got %q", result)
	}
	if !strings.Contains(result, "filtered_history_hint: use task_audit action=history id=task-a status=failed") {
		t.Fatalf("expected filtered history hint, got %q", result)
	}
}

func TestTaskAuditToolHistoryCanFilterByFailureReason(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(task.CreateInput{
		ID:    "task-a",
		Title: "Inspect timeout failures",
	}); err != nil {
		t.Fatal(err)
	}
	for _, entry := range []task.BackgroundContext{
		{JobID: "job-1", Status: "completed", Command: "cmd-1"},
		{JobID: "job-2", Status: "failed", Command: "cmd-2", Error: "boom"},
		{JobID: "job-3", Status: "failed", Command: "cmd-3", Error: "timeout after 5s"},
	} {
		if _, err := manager.RecordBackground("task-a", entry); err != nil {
			t.Fatal(err)
		}
	}
	tool := tools.NewTaskAuditTool(manager)
	result, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "history",
		"id":     "task-a",
		"reason": "timeout",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "filter_reason: timeout") {
		t.Fatalf("expected reason filter, got %q", result)
	}
	if !strings.Contains(result, "filtered_count: 1") || !strings.Contains(result, "returned_count: 1") {
		t.Fatalf("expected filtered counters, got %q", result)
	}
	if !strings.Contains(result, "job_id=job-3 status=failed") {
		t.Fatalf("expected timeout failure in result, got %q", result)
	}
	if !strings.Contains(result, "job_id=job-3 status=failed") || !strings.Contains(result, "role=latest") {
		t.Fatalf("expected latest role on single timeout failure, got %q", result)
	}
	if !strings.Contains(result, "job_id=job-3 status=failed") || !strings.Contains(result, "reason=timeout") {
		t.Fatalf("expected timeout reason on single filtered history entry, got %q", result)
	}
	if strings.Contains(result, "job_id=job-1") || strings.Contains(result, "job_id=job-2") {
		t.Fatalf("expected only matching reason in result, got %q", result)
	}
}

func TestTaskAuditToolHistoryCanMarkPreviousAndLatestForReasonFilteredPair(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(task.CreateInput{
		ID:    "task-a",
		Title: "Inspect timeout pair roles",
	}); err != nil {
		t.Fatal(err)
	}
	for _, entry := range []task.BackgroundContext{
		{JobID: "job-1", Status: "failed", Command: "cmd-1", Error: "boom"},
		{JobID: "job-2", Status: "failed", Command: "cmd-2", Error: "timeout after 5s"},
		{JobID: "job-3", Status: "failed", Command: "cmd-3", Error: "timeout after 8s"},
	} {
		if _, err := manager.RecordBackground("task-a", entry); err != nil {
			t.Fatal(err)
		}
	}
	tool := tools.NewTaskAuditTool(manager)
	result, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "history",
		"id":     "task-a",
		"reason": "timeout",
		"limit":  float64(2),
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "filter_reason: timeout") || !strings.Contains(result, "returned_count: 2") {
		t.Fatalf("expected reason-filtered pair counters, got %q", result)
	}
	if !strings.Contains(result, "pair_summary: compare=previous_vs_latest previous_job_id=job-2 latest_job_id=job-3 command=changed error_signature=same previous_signature=timeout after <duration> latest_signature=timeout after <duration>") {
		t.Fatalf("expected timeout pair summary, got %q", result)
	}
	if !strings.Contains(result, "job_id=job-2 status=failed") || !strings.Contains(result, "role=previous") {
		t.Fatalf("expected previous role on older timeout sample, got %q", result)
	}
	if !strings.Contains(result, "job_id=job-2 status=failed") || !strings.Contains(result, "reason=timeout") {
		t.Fatalf("expected timeout reason on previous sample, got %q", result)
	}
	if !strings.Contains(result, "job_id=job-3 status=failed") || !strings.Contains(result, "role=latest") {
		t.Fatalf("expected latest role on newest timeout sample, got %q", result)
	}
	if !strings.Contains(result, "job_id=job-3 status=failed") || !strings.Contains(result, "reason=timeout") {
		t.Fatalf("expected timeout reason on latest sample, got %q", result)
	}
}

func TestTaskAuditToolHistoryPairSummaryCanHighlightErrorOnlyChange(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(task.CreateInput{
		ID:    "task-a",
		Title: "Inspect command error pair summary",
	}); err != nil {
		t.Fatal(err)
	}
	for _, entry := range []task.BackgroundContext{
		{JobID: "job-1", Status: "failed", Command: "cmd-build", Error: "boom"},
		{JobID: "job-2", Status: "failed", Command: "cmd-build", Error: "boom again"},
	} {
		if _, err := manager.RecordBackground("task-a", entry); err != nil {
			t.Fatal(err)
		}
	}
	tool := tools.NewTaskAuditTool(manager)
	result, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "history",
		"id":     "task-a",
		"reason": "command_error",
		"limit":  float64(2),
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "pair_summary: compare=previous_vs_latest previous_job_id=job-1 latest_job_id=job-2 command=same error_signature=changed previous_signature=boom latest_signature=boom again") {
		t.Fatalf("expected error-only pair summary, got %q", result)
	}
	if !strings.Contains(result, "job_id=job-1 status=failed") || !strings.Contains(result, "role=previous") {
		t.Fatalf("expected previous role on first command_error sample, got %q", result)
	}
	if !strings.Contains(result, "job_id=job-1 status=failed") || !strings.Contains(result, "reason=command_error") {
		t.Fatalf("expected command_error reason on previous sample, got %q", result)
	}
	if !strings.Contains(result, "job_id=job-2 status=failed") || !strings.Contains(result, "role=latest") {
		t.Fatalf("expected latest role on second command_error sample, got %q", result)
	}
	if !strings.Contains(result, "job_id=job-2 status=failed") || !strings.Contains(result, "reason=command_error") {
		t.Fatalf("expected command_error reason on latest sample, got %q", result)
	}
}

func TestTaskAuditToolSummaryCanFilterByFailureReason(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(task.CreateInput{
		ID:    "task-a",
		Title: "Summarize timeout failures",
	}); err != nil {
		t.Fatal(err)
	}
	for _, entry := range []task.BackgroundContext{
		{JobID: "job-1", Status: "completed", Command: "cmd-1"},
		{JobID: "job-2", Status: "failed", Command: "cmd-2", Error: "boom"},
		{JobID: "job-3", Status: "failed", Command: "cmd-3", Error: "timeout after 5s"},
		{JobID: "job-4", Status: "failed", Command: "cmd-4", Error: "timeout after 8s"},
	} {
		if _, err := manager.RecordBackground("task-a", entry); err != nil {
			t.Fatal(err)
		}
	}
	tool := tools.NewTaskAuditTool(manager)
	result, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "summary",
		"id":     "task-a",
		"reason": "timeout",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "filter_reason: timeout") || !strings.Contains(result, "filtered_count: 2") {
		t.Fatalf("expected reason-filtered summary counters, got %q", result)
	}
	if !strings.Contains(result, "- timeout=2") || strings.Contains(result, "- command_error=") {
		t.Fatalf("expected only timeout reason counts, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_reason: timeout count=2") {
		t.Fatalf("expected reason-filtered priority failure reason, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_basis: chosen_by=highest_count count=2 latest_job_id=job-4") {
		t.Fatalf("expected reason-filtered priority failure basis, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_context: job_id=job-4 status=failed") || !strings.Contains(result, "command=cmd-4") || !strings.Contains(result, "error=timeout after 8s") {
		t.Fatalf("expected reason-filtered priority failure context, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_pattern_hint: pattern=repeat count=2 signature=timeout after <duration>") {
		t.Fatalf("expected reason-filtered priority failure pattern hint, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_sample_target: job_id=job-4") {
		t.Fatalf("expected reason-filtered priority failure sample target, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_sample_compare: latest_job_id=job-4 previous_job_id=job-3 command=changed error_signature=same latest_signature=timeout after <duration> previous_signature=timeout after <duration>") {
		t.Fatalf("expected reason-filtered priority failure sample compare, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_compare_target: sample_count=2 compare=latest_vs_previous latest_job_id=job-4 previous_job_id=job-3 history_command=task_audit action=history id=task-a reason=timeout limit=2 history_focus=job-3->job-4") {
		t.Fatalf("expected reason-filtered priority failure compare target, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_change_hint: sample_count=2 change=command_only focus=command latest_job_id=job-4 previous_job_id=job-3 signature=timeout after <duration> latest_command=cmd-4 previous_command=cmd-3") {
		t.Fatalf("expected reason-filtered priority failure change hint, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_trend_hint: trend=stable sample_count=2 signature=timeout after <duration>") {
		t.Fatalf("expected reason-filtered priority failure trend hint, got %q", result)
	}
	if !strings.Contains(result, "- timeout => job_id=job-4 status=failed") {
		t.Fatalf("expected latest timeout sample, got %q", result)
	}
	if !strings.Contains(result, "- reason=timeout job_id=job-3 status=failed") || !strings.Contains(result, "- reason=timeout job_id=job-4 status=failed") {
		t.Fatalf("expected timeout trend lines, got %q", result)
	}
	if !strings.Contains(result, "matched_latest_entry: job_id=job-4 status=failed") {
		t.Fatalf("expected matched latest entry, got %q", result)
	}
	if !strings.Contains(result, "filtered_history_hint: use task_audit action=history id=task-a reason=timeout") {
		t.Fatalf("expected reason-filtered history hint, got %q", result)
	}
}

func TestTaskAuditToolSummaryPriorityReasonPrefersHigherCount(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(task.CreateInput{
		ID:    "task-a",
		Title: "Rank failure reasons",
	}); err != nil {
		t.Fatal(err)
	}
	for _, entry := range []task.BackgroundContext{
		{JobID: "job-1", Status: "failed", Command: "cmd-1", Error: "boom"},
		{JobID: "job-2", Status: "failed", Command: "cmd-2", Error: "timeout after 5s"},
		{JobID: "job-3", Status: "failed", Command: "cmd-3", Error: "boom again"},
	} {
		if _, err := manager.RecordBackground("task-a", entry); err != nil {
			t.Fatal(err)
		}
	}
	tool := tools.NewTaskAuditTool(manager)
	result, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "summary",
		"id":     "task-a",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "priority_failure_reason: command_error count=2") {
		t.Fatalf("expected higher-count reason to be prioritized, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_basis: chosen_by=highest_count count=2 latest_job_id=job-3") {
		t.Fatalf("expected highest-count basis, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_context: job_id=job-3 status=failed") || !strings.Contains(result, "command=cmd-3") || !strings.Contains(result, "error=boom again") {
		t.Fatalf("expected highest-count context, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_pattern_hint: pattern=single count=1 signature=boom again") {
		t.Fatalf("expected highest-count pattern hint, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_sample_target: job_id=job-3") {
		t.Fatalf("expected highest-count sample target, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_sample_compare: latest_job_id=job-3 previous_job_id=job-1 command=changed error_signature=changed latest_signature=boom again previous_signature=boom") {
		t.Fatalf("expected highest-count sample compare, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_compare_target: sample_count=2 compare=latest_vs_previous latest_job_id=job-3 previous_job_id=job-1 history_command=task_audit action=history id=task-a reason=command_error limit=2 history_focus=job-1->job-3") {
		t.Fatalf("expected highest-count compare target, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_change_hint: sample_count=2 change=command_and_error_signature focus=command,error_signature latest_job_id=job-3 previous_job_id=job-1 latest_signature=boom again previous_signature=boom latest_command=cmd-3 previous_command=cmd-1") {
		t.Fatalf("expected highest-count change hint, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_trend_hint: trend=changing sample_count=2 latest_signature=boom again previous_signature=boom") {
		t.Fatalf("expected highest-count trend hint, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_hint: use task_audit action=summary id=task-a reason=command_error, then task_audit action=history id=task-a reason=command_error limit=1") {
		t.Fatalf("expected command_error hint, got %q", result)
	}
}

func TestTaskAuditToolSummaryPriorityChangeHintCanHighlightErrorOnlyChange(t *testing.T) {
	manager, err := task.NewManager(filepath.Join(t.TempDir(), ".tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(task.CreateInput{
		ID:    "task-a",
		Title: "Highlight error-only changes",
	}); err != nil {
		t.Fatal(err)
	}
	for _, entry := range []task.BackgroundContext{
		{JobID: "job-1", Status: "failed", Command: "cmd-build", Error: "boom"},
		{JobID: "job-2", Status: "failed", Command: "cmd-build", Error: "boom again"},
	} {
		if _, err := manager.RecordBackground("task-a", entry); err != nil {
			t.Fatal(err)
		}
	}
	tool := tools.NewTaskAuditTool(manager)
	result, err := tool.Handler(tools.Call{Input: map[string]interface{}{
		"action": "summary",
		"id":     "task-a",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "priority_failure_reason: command_error count=2") {
		t.Fatalf("expected command_error priority reason, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_sample_compare: latest_job_id=job-2 previous_job_id=job-1 command=same error_signature=changed latest_signature=boom again previous_signature=boom") {
		t.Fatalf("expected error-only sample compare, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_compare_target: sample_count=2 compare=latest_vs_previous latest_job_id=job-2 previous_job_id=job-1 history_command=task_audit action=history id=task-a reason=command_error limit=2 history_focus=job-1->job-2") {
		t.Fatalf("expected error-only compare target, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_change_hint: sample_count=2 change=error_signature_only focus=error_signature latest_job_id=job-2 previous_job_id=job-1 latest_signature=boom again previous_signature=boom") {
		t.Fatalf("expected error-only change hint, got %q", result)
	}
	if !strings.Contains(result, "priority_failure_trend_hint: trend=changing sample_count=2 latest_signature=boom again previous_signature=boom") {
		t.Fatalf("expected error-only trend hint, got %q", result)
	}
}

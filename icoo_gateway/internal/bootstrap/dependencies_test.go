package bootstrap_test

import (
	"os"
	"path/filepath"
	"testing"

	"icoo_gateway/internal/audit"
	"icoo_gateway/internal/bootstrap"
	"icoo_gateway/internal/config"
	"icoo_gateway/internal/conversation"
	"icoo_gateway/internal/run"
	"icoo_gateway/internal/team"
)

func TestBuildDependenciesWithSQLiteCreatesPersistentAuditStore(t *testing.T) {
	root := t.TempDir()
	dbPath := filepath.Join(root, "data", "gateway.db")
	cfg := config.Config{
		StorageDriver: "sqlite",
		SQLitePath:    dbPath,
	}

	deps, err := bootstrap.BuildDependencies(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if deps.Close != nil {
		defer func() {
			if err := deps.Close(); err != nil {
				t.Fatal(err)
			}
		}()
	}
	if deps.Audits == nil {
		t.Fatal("expected audit store")
	}

	record := deps.Audits.Record(audit.RecordInput{
		ResourceType: "skill",
		ResourceID:   "skill-1",
		EventName:    "skill.created",
		Operator:     "tester",
		Payload:      map[string]string{"name": "demo"},
	})

	items := deps.Audits.List()
	if len(items) != 1 {
		t.Fatalf("expected one audit event, got %#v", items)
	}
	if items[0].ID != record.ID || items[0].Operator != "tester" {
		t.Fatalf("unexpected audit items: %#v", items)
	}
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("expected sqlite db file, got %v", err)
	}
}

func TestBuildDependenciesWithSQLiteCreatesPersistentRunStore(t *testing.T) {
	root := t.TempDir()
	dbPath := filepath.Join(root, "data", "gateway.db")
	cfg := config.Config{
		StorageDriver: "sqlite",
		SQLitePath:    dbPath,
	}

	deps, err := bootstrap.BuildDependencies(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if deps.Close != nil {
		defer func() {
			if err := deps.Close(); err != nil {
				t.Fatal(err)
			}
		}()
	}
	if deps.Runs == nil {
		t.Fatal("expected run store")
	}

	created, err := deps.Runs.Create(run.CreateInput{
		ConversationID:   "conv-1",
		TriggerType:      "message",
		TriggerMessageID: "msg-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	completed, err := deps.Runs.Complete(created.ID, run.CompleteInput{
		Status:  "completed",
		Summary: "done",
	})
	if err != nil {
		t.Fatal(err)
	}

	items := deps.Runs.ListByConversation("conv-1")
	if len(items) != 1 {
		t.Fatalf("expected one run, got %#v", items)
	}
	if items[0].ID != completed.ID || items[0].Status != "completed" || items[0].Summary != "done" {
		t.Fatalf("unexpected run items: %#v", items)
	}
}

func TestBuildDependenciesWithSQLiteCreatesPersistentConversationStore(t *testing.T) {
	root := t.TempDir()
	dbPath := filepath.Join(root, "data", "gateway.db")
	cfg := config.Config{
		StorageDriver: "sqlite",
		SQLitePath:    dbPath,
	}

	deps, err := bootstrap.BuildDependencies(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if deps.Close != nil {
		defer func() {
			if err := deps.Close(); err != nil {
				t.Fatal(err)
			}
		}()
	}
	if deps.Conversations == nil {
		t.Fatal("expected conversation store")
	}

	created, err := deps.Conversations.Create(conversation.CreateInput{
		Mode:          "single",
		Title:         "demo",
		TargetAgentID: "agent-profile-1",
		CreatedBy:     "tester",
	})
	if err != nil {
		t.Fatal(err)
	}
	message, err := deps.Conversations.AddMessage(created.ID, conversation.AddMessageInput{
		Role:    "user",
		Content: "hello",
	})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := deps.Conversations.SetLastRunID(created.ID, "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if updated.LastRunID != "run-1" {
		t.Fatalf("expected last run id, got %#v", updated)
	}

	loaded, ok := deps.Conversations.Get(created.ID)
	if !ok {
		t.Fatal("expected conversation to load")
	}
	if loaded.MessageCount != 1 || loaded.LastRunID != "run-1" {
		t.Fatalf("unexpected loaded conversation: %#v", loaded)
	}
	items, ok := deps.Conversations.ListMessagesByScope(created.ID, "")
	if !ok {
		t.Fatal("expected messages to load")
	}
	if len(items) != 1 || items[0].ID != message.ID || items[0].Content != "hello" {
		t.Fatalf("unexpected message items: %#v", items)
	}
}

func TestBuildDependenciesWithSQLiteCreatesPersistentTeamStore(t *testing.T) {
	root := t.TempDir()
	dbPath := filepath.Join(root, "data", "gateway.db")
	cfg := config.Config{
		StorageDriver: "sqlite",
		SQLitePath:    dbPath,
	}

	deps, err := bootstrap.BuildDependencies(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if deps.Close != nil {
		defer func() {
			if err := deps.Close(); err != nil {
				t.Fatal(err)
			}
		}()
	}
	if deps.Teams == nil {
		t.Fatal("expected team store")
	}

	created, err := deps.Teams.Create(team.CreateInput{
		Name:         "core-team",
		Description:  "demo",
		EntryAgentID: "agent-instance-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	member, err := deps.Teams.AddMember(created.ID, team.AddMemberInput{
		AgentID:   "agent-instance-1",
		Role:      "lead",
		SortOrder: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	role := "architect"
	status := "inactive"
	updatedMember, err := deps.Teams.UpdateMember(created.ID, member.ID, team.UpdateMemberInput{
		Role:   &role,
		Status: &status,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updatedMember.Role != "architect" || updatedMember.Status != "inactive" {
		t.Fatalf("unexpected updated member: %#v", updatedMember)
	}

	items, ok := deps.Teams.ListMembers(created.ID)
	if !ok {
		t.Fatal("expected members to load")
	}
	if len(items) != 1 || items[0].ID != member.ID {
		t.Fatalf("unexpected members: %#v", items)
	}
	if deps.Teams.HasMember(created.ID, "agent-instance-1") {
		t.Fatal("expected inactive member to not count as active")
	}
	removed, err := deps.Teams.DeleteMember(created.ID, member.ID)
	if err != nil {
		t.Fatal(err)
	}
	if removed.ID != member.ID {
		t.Fatalf("unexpected removed member: %#v", removed)
	}
	items, ok = deps.Teams.ListMembers(created.ID)
	if !ok {
		t.Fatal("expected members to load after delete")
	}
	if len(items) != 0 {
		t.Fatalf("expected no members after delete, got %#v", items)
	}
}

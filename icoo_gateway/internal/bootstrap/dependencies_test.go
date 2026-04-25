package bootstrap_test

import (
	"os"
	"path/filepath"
	"testing"

	"icoo_gateway/internal/audit"
	"icoo_gateway/internal/bootstrap"
	"icoo_gateway/internal/config"
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

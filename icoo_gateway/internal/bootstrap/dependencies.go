package bootstrap

import (
	"fmt"
	"strings"

	"icoo_gateway/internal/agentinstance"
	"icoo_gateway/internal/agentprofile"
	"icoo_gateway/internal/audit"
	"icoo_gateway/internal/config"
	"icoo_gateway/internal/conversation"
	"icoo_gateway/internal/run"
	"icoo_gateway/internal/skill"
	"icoo_gateway/internal/storage/postgres"
	"icoo_gateway/internal/storage/sqlite"
	"icoo_gateway/internal/team"
)

type Dependencies struct {
	Audits         audit.Store
	Skills         *skill.Service
	AgentProfiles  *agentprofile.Service
	AgentInstances *agentinstance.Service
	Teams          *team.Service
	Conversations  *conversation.Service
	Runs           *run.Service
	Close          func() error
}

func NewMemoryDependencies() Dependencies {
	return Dependencies{
		Audits:         audit.NewService(),
		Skills:         skill.NewService(),
		AgentProfiles:  agentprofile.NewService(),
		AgentInstances: agentinstance.NewService(),
		Teams:          team.NewService(),
		Conversations:  conversation.NewService(),
		Runs:           run.NewService(),
	}
}

func BuildDependencies(cfg config.Config) (Dependencies, error) {
	driver := strings.TrimSpace(cfg.StorageDriver)
	if driver == "" || driver == "memory" {
		return NewMemoryDependencies(), nil
	}
	if driver == "sqlite" {
		db, err := sqlite.Open(cfg)
		if err != nil {
			return Dependencies{}, err
		}
		sqlDB, err := db.DB()
		if err != nil {
			return Dependencies{}, err
		}
		deps := NewMemoryDependencies()
		deps.Audits = audit.NewGormStore(db)
		deps.Close = sqlDB.Close
		return deps, nil
	}
	if driver == "postgres" {
		if err := postgres.ValidateConfig(cfg); err != nil {
			return Dependencies{}, err
		}
		return Dependencies{}, fmt.Errorf("postgres storage provider not implemented yet")
	}
	return Dependencies{}, fmt.Errorf("unsupported storage driver: %s", driver)
}

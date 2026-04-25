package api

import (
	"icoo_gateway/internal/agentinstance"
	"icoo_gateway/internal/agentprofile"
	"icoo_gateway/internal/audit"
	"icoo_gateway/internal/bootstrap"
	"icoo_gateway/internal/conversation"
	"icoo_gateway/internal/routing"
	"icoo_gateway/internal/run"
	"icoo_gateway/internal/skill"
	"icoo_gateway/internal/team"
)

type App struct {
	Audits         audit.Store
	Skills         *skill.Service
	AgentProfiles  *agentprofile.Service
	AgentInstances *agentinstance.Service
	Teams          team.Store
	Conversations  conversation.Store
	Runs           run.Store
	Router         routing.Router
}

func NewApp() *App {
	return NewAppWithDependencies(bootstrap.NewMemoryDependencies())
}

func NewAppWithDependencies(deps bootstrap.Dependencies) *App {
	if deps.Audits == nil {
		deps.Audits = audit.NewService()
	}
	if deps.Skills == nil {
		deps.Skills = skill.NewService()
	}
	if deps.AgentProfiles == nil {
		deps.AgentProfiles = agentprofile.NewService()
	}
	if deps.AgentInstances == nil {
		deps.AgentInstances = agentinstance.NewService()
	}
	if deps.Teams == nil {
		deps.Teams = team.NewService()
	}
	if deps.Conversations == nil {
		deps.Conversations = conversation.NewService()
	}
	if deps.Runs == nil {
		deps.Runs = run.NewService()
	}
	return &App{
		Audits:         deps.Audits,
		Skills:         deps.Skills,
		AgentProfiles:  deps.AgentProfiles,
		AgentInstances: deps.AgentInstances,
		Teams:          deps.Teams,
		Conversations:  deps.Conversations,
		Runs:           deps.Runs,
		Router: routing.Router{
			Teams:         deps.Teams,
			Conversations: deps.Conversations,
			Writer:        deps.Conversations,
		},
	}
}

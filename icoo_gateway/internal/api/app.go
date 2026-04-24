package api

import (
	"icoo_gateway/internal/agentinstance"
	"icoo_gateway/internal/agentprofile"
	"icoo_gateway/internal/conversation"
	"icoo_gateway/internal/routing"
	"icoo_gateway/internal/skill"
	"icoo_gateway/internal/team"
)

type App struct {
	Skills         *skill.Service
	AgentProfiles  *agentprofile.Service
	AgentInstances *agentinstance.Service
	Teams          *team.Service
	Conversations  *conversation.Service
	Router         routing.Router
}

func NewApp() *App {
	teams := team.NewService()
	conversations := conversation.NewService()
	return &App{
		Skills:         skill.NewService(),
		AgentProfiles:  agentprofile.NewService(),
		AgentInstances: agentinstance.NewService(),
		Teams:          teams,
		Conversations:  conversations,
		Router: routing.Router{
			Teams:         teams,
			Conversations: conversations,
			Writer:        conversations,
		},
	}
}

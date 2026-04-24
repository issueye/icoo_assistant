package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"icoo_gateway/internal/agentinstance"
	"icoo_gateway/internal/agentprofile"
	"icoo_gateway/internal/conversation"
	"icoo_gateway/internal/skill"
	"icoo_gateway/internal/team"
)

func NewMux(app *App) http.Handler {
	if app == nil {
		app = NewApp()
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/healthz", handleHealth)
	mux.HandleFunc("/api/v1/skills", app.handleSkills)
	mux.HandleFunc("/api/v1/skills/", app.handleSkillByID)
	mux.HandleFunc("/api/v1/agent-profiles", app.handleAgentProfiles)
	mux.HandleFunc("/api/v1/agent-profiles/", app.handleAgentProfileByID)
	mux.HandleFunc("/api/v1/agent-instances", app.handleAgentInstances)
	mux.HandleFunc("/api/v1/agent-instances/", app.handleAgentInstanceRoutes)
	mux.HandleFunc("/api/v1/teams", app.handleTeams)
	mux.HandleFunc("/api/v1/teams/", app.handleTeamRoutes)
	mux.HandleFunc("/api/v1/conversations", app.handleConversations)
	mux.HandleFunc("/api/v1/conversations/", app.handleConversationRoutes)
	return mux
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"service": "icoo_gateway",
		"status":  "ready",
		"routes": []string{
			"/",
			"/healthz",
			"/api/v1/skills",
			"/api/v1/agent-profiles",
			"/api/v1/agent-instances",
			"/api/v1/teams",
			"/api/v1/conversations",
		},
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"service": "icoo_gateway",
		"status":  "ok",
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{
		"error": message,
	})
}

func decodeJSON(r *http.Request, target interface{}) error {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return fmt.Errorf("invalid json body")
	}
	return nil
}

func pathID(path, prefix string) string {
	return strings.TrimSpace(strings.TrimPrefix(path, prefix))
}

func (a *App) handleSkills(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"items": a.Skills.List(),
		})
	case http.MethodPost:
		var input skill.CreateInput
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		record, err := a.Skills.Create(input)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, record)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *App) handleSkillByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id := pathID(r.URL.Path, "/api/v1/skills/")
	record, ok := a.Skills.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "skill not found")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (a *App) handleAgentProfiles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"items": a.AgentProfiles.List(),
		})
	case http.MethodPost:
		var input agentprofile.CreateInput
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		record, err := a.AgentProfiles.Create(input)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, record)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *App) handleAgentProfileByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id := pathID(r.URL.Path, "/api/v1/agent-profiles/")
	record, ok := a.AgentProfiles.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "agent profile not found")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (a *App) handleAgentInstances(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"items": a.AgentInstances.List(),
		})
	case http.MethodPost:
		var input agentinstance.CreateInput
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		record, err := a.AgentInstances.Create(input)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, record)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *App) handleAgentInstanceByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id := pathID(r.URL.Path, "/api/v1/agent-instances/")
	record, ok := a.AgentInstances.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "agent instance not found")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (a *App) handleAgentInstanceRoutes(w http.ResponseWriter, r *http.Request) {
	path := pathID(r.URL.Path, "/api/v1/agent-instances/")
	if path == "" {
		writeError(w, http.StatusNotFound, "agent instance not found")
		return
	}
	if strings.HasSuffix(path, "/heartbeat") {
		id := strings.TrimSuffix(path, "/heartbeat")
		id = strings.TrimSuffix(id, "/")
		a.handleAgentInstanceHeartbeat(w, r, id)
		return
	}
	a.handleAgentInstanceByID(w, r)
}

func (a *App) handleAgentInstanceHeartbeat(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	record, err := a.AgentInstances.Heartbeat(id)
	if err != nil {
		if err.Error() == "agent instance not found" {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (a *App) handleTeams(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"items": a.Teams.List(),
		})
	case http.MethodPost:
		var input team.CreateInput
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		record, err := a.Teams.Create(input)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, record)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *App) handleTeamByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id := pathID(r.URL.Path, "/api/v1/teams/")
	record, ok := a.Teams.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "team not found")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (a *App) handleTeamRoutes(w http.ResponseWriter, r *http.Request) {
	path := pathID(r.URL.Path, "/api/v1/teams/")
	if path == "" {
		writeError(w, http.StatusNotFound, "team not found")
		return
	}
	if strings.HasSuffix(path, "/members") {
		id := strings.TrimSuffix(path, "/members")
		id = strings.TrimSuffix(id, "/")
		a.handleTeamMembers(w, r, id)
		return
	}
	a.handleTeamByID(w, r)
}

func (a *App) handleTeamMembers(w http.ResponseWriter, r *http.Request, teamID string) {
	switch r.Method {
	case http.MethodGet:
		items, ok := a.Teams.ListMembers(teamID)
		if !ok {
			writeError(w, http.StatusNotFound, "team not found")
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"items": items,
		})
	case http.MethodPost:
		var input team.AddMemberInput
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if _, ok := a.AgentInstances.Get(input.AgentID); !ok {
			writeError(w, http.StatusBadRequest, "agent_id must reference an existing agent instance")
			return
		}
		record, err := a.Teams.AddMember(teamID, input)
		if err != nil {
			if err.Error() == "team not found" {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, record)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *App) handleConversations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"items": a.Conversations.List(),
		})
	case http.MethodPost:
		var input conversation.CreateInput
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		record, err := a.Conversations.Create(input)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, record)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *App) handleConversationRoutes(w http.ResponseWriter, r *http.Request) {
	path := pathID(r.URL.Path, "/api/v1/conversations/")
	if path == "" {
		writeError(w, http.StatusNotFound, "conversation not found")
		return
	}
	if strings.HasSuffix(path, "/messages") {
		id := strings.TrimSuffix(path, "/messages")
		id = strings.TrimSuffix(id, "/")
		a.handleConversationMessages(w, r, id)
		return
	}
	a.handleConversationByID(w, r, path)
}

func (a *App) handleConversationByID(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	record, ok := a.Conversations.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "conversation not found")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (a *App) handleConversationMessages(w http.ResponseWriter, r *http.Request, id string) {
	switch r.Method {
	case http.MethodGet:
		scope := strings.TrimSpace(r.URL.Query().Get("scope"))
		items, ok := a.Conversations.ListMessagesByScope(id, scope)
		if !ok {
			writeError(w, http.StatusNotFound, "conversation not found")
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"items": items,
		})
	case http.MethodPost:
		var input conversation.AddMessageInput
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		record, err := a.Conversations.AddMessage(id, input)
		if err != nil {
			if err.Error() == "conversation not found" {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		recordConversation, ok := a.Conversations.Get(id)
		if ok && recordConversation.Mode == "team" && record.Scope == "external" {
			if err := a.Router.RouteExternalMessage(id, record); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
		}
		writeJSON(w, http.StatusCreated, record)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

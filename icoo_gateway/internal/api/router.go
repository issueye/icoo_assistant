package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"icoo_gateway/internal/agentinstance"
	"icoo_gateway/internal/agentprofile"
	"icoo_gateway/internal/audit"
	"icoo_gateway/internal/conversation"
	"icoo_gateway/internal/run"
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
	mux.HandleFunc("/api/v1/audit-events", app.handleAuditEvents)
	mux.HandleFunc("/api/v1/audit-events/", app.handleAuditEventByID)
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
			"/api/v1/audit-events",
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

func operatorFromRequest(r *http.Request) string {
	operator := strings.TrimSpace(r.Header.Get("X-Operator"))
	if operator == "" {
		operator = "system"
	}
	return operator
}

func (a *App) recordAudit(r *http.Request, resourceType, resourceID, eventName string, payload interface{}) {
	if a == nil || a.Audits == nil {
		return
	}
	a.Audits.Record(audit.RecordInput{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		EventName:    eventName,
		Operator:     operatorFromRequest(r),
		Payload:      payload,
	})
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
		a.recordAudit(r, "skill", record.ID, "skill.created", record)
		writeJSON(w, http.StatusCreated, record)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *App) handleSkillByID(w http.ResponseWriter, r *http.Request) {
	path := pathID(r.URL.Path, "/api/v1/skills/")
	if path == "" {
		writeError(w, http.StatusNotFound, "skill not found")
		return
	}
	if strings.HasSuffix(path, "/activate") {
		id := strings.TrimSuffix(path, "/activate")
		id = strings.TrimSuffix(id, "/")
		a.handleSkillActivate(w, r, id)
		return
	}
	if strings.HasSuffix(path, "/deactivate") {
		id := strings.TrimSuffix(path, "/deactivate")
		id = strings.TrimSuffix(id, "/")
		a.handleSkillDeactivate(w, r, id)
		return
	}
	id := strings.TrimSuffix(path, "/")
	switch r.Method {
	case http.MethodGet:
		record, ok := a.Skills.Get(id)
		if !ok {
			writeError(w, http.StatusNotFound, "skill not found")
			return
		}
		writeJSON(w, http.StatusOK, record)
	case http.MethodPatch:
		var input skill.UpdateInput
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		record, err := a.Skills.Update(id, input)
		if err != nil {
			if err.Error() == "skill not found" {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		a.recordAudit(r, "skill", record.ID, "skill.updated", record)
		writeJSON(w, http.StatusOK, record)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *App) handleSkillActivate(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	record, err := a.Skills.Activate(id)
	if err != nil {
		if err.Error() == "skill not found" {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.recordAudit(r, "skill", record.ID, "skill.activated", record)
	writeJSON(w, http.StatusOK, record)
}

func (a *App) handleSkillDeactivate(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	record, err := a.Skills.Deactivate(id)
	if err != nil {
		if err.Error() == "skill not found" {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.recordAudit(r, "skill", record.ID, "skill.deactivated", record)
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
		a.recordAudit(r, "agent_profile", record.ID, "agent_profile.created", record)
		writeJSON(w, http.StatusCreated, record)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *App) handleAgentProfileByID(w http.ResponseWriter, r *http.Request) {
	id := pathID(r.URL.Path, "/api/v1/agent-profiles/")
	switch r.Method {
	case http.MethodGet:
		record, ok := a.AgentProfiles.Get(id)
		if !ok {
			writeError(w, http.StatusNotFound, "agent profile not found")
			return
		}
		writeJSON(w, http.StatusOK, record)
	case http.MethodPatch:
		var input agentprofile.UpdateInput
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		record, err := a.AgentProfiles.Update(id, input)
		if err != nil {
			if err.Error() == "agent profile not found" {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		a.recordAudit(r, "agent_profile", record.ID, "agent_profile.updated", record)
		writeJSON(w, http.StatusOK, record)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
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
		a.recordAudit(r, "agent_instance", record.ID, "agent_instance.created", record)
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
	if strings.HasSuffix(path, "/disable") {
		id := strings.TrimSuffix(path, "/disable")
		id = strings.TrimSuffix(id, "/")
		a.handleAgentInstanceDisable(w, r, id)
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
	a.recordAudit(r, "agent_instance", record.ID, "agent_instance.heartbeat", record)
	writeJSON(w, http.StatusOK, record)
}

func (a *App) handleAgentInstanceDisable(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	record, err := a.AgentInstances.Disable(id)
	if err != nil {
		if err.Error() == "agent instance not found" {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.recordAudit(r, "agent_instance", record.ID, "agent_instance.disabled", record)
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
		a.recordAudit(r, "team", record.ID, "team.created", record)
		writeJSON(w, http.StatusCreated, record)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *App) handleTeamByID(w http.ResponseWriter, r *http.Request) {
	id := pathID(r.URL.Path, "/api/v1/teams/")
	switch r.Method {
	case http.MethodGet:
		record, ok := a.Teams.Get(id)
		if !ok {
			writeError(w, http.StatusNotFound, "team not found")
			return
		}
		writeJSON(w, http.StatusOK, record)
	case http.MethodPatch:
		var input team.UpdateInput
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		record, err := a.Teams.Update(id, input)
		if err != nil {
			if err.Error() == "team not found" {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		a.recordAudit(r, "team", record.ID, "team.updated", record)
		writeJSON(w, http.StatusOK, record)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
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
	if strings.Contains(path, "/members/") {
		parts := strings.SplitN(path, "/members/", 2)
		teamID := strings.TrimSuffix(strings.TrimSpace(parts[0]), "/")
		memberID := ""
		if len(parts) > 1 {
			memberID = strings.TrimSuffix(strings.TrimSpace(parts[1]), "/")
		}
		a.handleTeamMemberByID(w, r, teamID, memberID)
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
		a.recordAudit(r, "team_member", record.ID, "team_member.created", record)
		writeJSON(w, http.StatusCreated, record)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *App) handleTeamMemberByID(w http.ResponseWriter, r *http.Request, teamID, memberID string) {
	if teamID == "" || memberID == "" {
		writeError(w, http.StatusNotFound, "team member not found")
		return
	}
	switch r.Method {
	case http.MethodPatch:
		var input team.UpdateMemberInput
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		record, err := a.Teams.UpdateMember(teamID, memberID, input)
		if err != nil {
			if err.Error() == "team not found" || err.Error() == "team member not found" {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		a.recordAudit(r, "team_member", record.ID, "team_member.updated", record)
		writeJSON(w, http.StatusOK, record)
	case http.MethodDelete:
		record, err := a.Teams.DeleteMember(teamID, memberID)
		if err != nil {
			if err.Error() == "team not found" || err.Error() == "team member not found" {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		a.recordAudit(r, "team_member", record.ID, "team_member.deleted", record)
		writeJSON(w, http.StatusOK, record)
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
		a.recordAudit(r, "conversation", record.ID, "conversation.created", record)
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
	if strings.HasSuffix(path, "/runs") {
		id := strings.TrimSuffix(path, "/runs")
		id = strings.TrimSuffix(id, "/")
		a.handleConversationRuns(w, r, id)
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
		if record.Scope == "external" {
			createdRun, err := a.Runs.Create(run.CreateInput{
				ConversationID:   id,
				TriggerType:      "message",
				TriggerMessageID: record.ID,
			})
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			if _, err := a.Conversations.SetLastRunID(id, createdRun.ID); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			recordConversation, ok := a.Conversations.Get(id)
			summary := "external message accepted"
			if ok && recordConversation.Mode == "team" {
				summary = "team external message accepted and routed"
			}
			finishedRun, err := a.Runs.Complete(createdRun.ID, run.CompleteInput{
				Status:  "completed",
				Summary: summary,
			})
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			a.recordAudit(r, "run", finishedRun.ID, "run.completed", finishedRun)
		}
		recordConversation, ok := a.Conversations.Get(id)
		if ok && recordConversation.Mode == "team" && record.Scope == "external" {
			if err := a.Router.RouteExternalMessage(id, record); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
		}
		a.recordAudit(r, "conversation_message", record.ID, "conversation_message.created", record)
		writeJSON(w, http.StatusCreated, record)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *App) handleConversationRuns(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if _, ok := a.Conversations.Get(id); !ok {
		writeError(w, http.StatusNotFound, "conversation not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": a.Runs.ListByConversation(id),
	})
}

func (a *App) handleAuditEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": a.Audits.List(),
	})
}

func (a *App) handleAuditEventByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id := pathID(r.URL.Path, "/api/v1/audit-events/")
	record, ok := a.Audits.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "audit event not found")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

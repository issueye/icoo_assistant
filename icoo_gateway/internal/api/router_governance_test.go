package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"icoo_gateway/internal/api"
)

func TestSkillPatchActivateDeactivateAndAudit(t *testing.T) {
	app := api.NewApp()
	handler := api.NewMux(app)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/skills", bytes.NewBufferString(`{"name":"code-review","version":"v1"}`))
	createRec := httptest.NewRecorder()
	handler.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("unexpected create status: %d body=%s", createRec.Code, createRec.Body.String())
	}

	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	skillID, _ := created["id"].(string)
	if skillID == "" {
		t.Fatalf("expected skill id, got %#v", created)
	}

	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/skills/"+skillID, bytes.NewBufferString(`{"description":"review skill updated","status":"inactive"}`))
	patchReq.Header.Set("X-Operator", "tester")
	patchRec := httptest.NewRecorder()
	handler.ServeHTTP(patchRec, patchReq)
	if patchRec.Code != http.StatusOK {
		t.Fatalf("unexpected patch status: %d body=%s", patchRec.Code, patchRec.Body.String())
	}

	activateReq := httptest.NewRequest(http.MethodPost, "/api/v1/skills/"+skillID+"/activate", nil)
	activateRec := httptest.NewRecorder()
	handler.ServeHTTP(activateRec, activateReq)
	if activateRec.Code != http.StatusOK {
		t.Fatalf("unexpected activate status: %d body=%s", activateRec.Code, activateRec.Body.String())
	}

	deactivateReq := httptest.NewRequest(http.MethodPost, "/api/v1/skills/"+skillID+"/deactivate", nil)
	deactivateRec := httptest.NewRecorder()
	handler.ServeHTTP(deactivateRec, deactivateReq)
	if deactivateRec.Code != http.StatusOK {
		t.Fatalf("unexpected deactivate status: %d body=%s", deactivateRec.Code, deactivateRec.Body.String())
	}

	auditReq := httptest.NewRequest(http.MethodGet, "/api/v1/audit-events", nil)
	auditRec := httptest.NewRecorder()
	handler.ServeHTTP(auditRec, auditReq)
	if auditRec.Code != http.StatusOK {
		t.Fatalf("unexpected audit list status: %d body=%s", auditRec.Code, auditRec.Body.String())
	}

	var auditList struct {
		Items []struct {
			ResourceType string `json:"resource_type"`
			ResourceID   string `json:"resource_id"`
			EventName    string `json:"event_name"`
			Operator     string `json:"operator"`
		} `json:"items"`
	}
	if err := json.Unmarshal(auditRec.Body.Bytes(), &auditList); err != nil {
		t.Fatal(err)
	}
	if len(auditList.Items) < 4 {
		t.Fatalf("expected at least four audit events, got %#v", auditList.Items)
	}
	if auditList.Items[1].Operator != "tester" {
		t.Fatalf("expected patch operator to be recorded, got %#v", auditList.Items)
	}
}

func TestAgentProfilePatchAndAgentInstanceDisable(t *testing.T) {
	app := api.NewApp()
	handler := api.NewMux(app)

	profileReq := httptest.NewRequest(http.MethodPost, "/api/v1/agent-profiles", bytes.NewBufferString(`{"name":"lead","model_provider":"anthropic","model_name":"claude-opus-4-1"}`))
	profileRec := httptest.NewRecorder()
	handler.ServeHTTP(profileRec, profileReq)
	if profileRec.Code != http.StatusCreated {
		t.Fatalf("unexpected profile create status: %d body=%s", profileRec.Code, profileRec.Body.String())
	}

	var createdProfile map[string]interface{}
	if err := json.Unmarshal(profileRec.Body.Bytes(), &createdProfile); err != nil {
		t.Fatal(err)
	}
	profileID, _ := createdProfile["id"].(string)

	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/agent-profiles/"+profileID, bytes.NewBufferString(`{"system_prompt":"be concise","status":"inactive"}`))
	patchRec := httptest.NewRecorder()
	handler.ServeHTTP(patchRec, patchReq)
	if patchRec.Code != http.StatusOK {
		t.Fatalf("unexpected profile patch status: %d body=%s", patchRec.Code, patchRec.Body.String())
	}

	instanceReq := httptest.NewRequest(http.MethodPost, "/api/v1/agent-instances", bytes.NewBufferString(`{"display_name":"lead-agent","runtime_type":"local","profile_id":"`+profileID+`"}`))
	instanceRec := httptest.NewRecorder()
	handler.ServeHTTP(instanceRec, instanceReq)
	if instanceRec.Code != http.StatusCreated {
		t.Fatalf("unexpected instance create status: %d body=%s", instanceRec.Code, instanceRec.Body.String())
	}

	var createdInstance map[string]interface{}
	if err := json.Unmarshal(instanceRec.Body.Bytes(), &createdInstance); err != nil {
		t.Fatal(err)
	}
	instanceID, _ := createdInstance["id"].(string)

	disableReq := httptest.NewRequest(http.MethodPost, "/api/v1/agent-instances/"+instanceID+"/disable", nil)
	disableRec := httptest.NewRecorder()
	handler.ServeHTTP(disableRec, disableReq)
	if disableRec.Code != http.StatusOK {
		t.Fatalf("unexpected disable status: %d body=%s", disableRec.Code, disableRec.Body.String())
	}

	var disabled struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(disableRec.Body.Bytes(), &disabled); err != nil {
		t.Fatal(err)
	}
	if disabled.Status != "disabled" {
		t.Fatalf("expected disabled status, got %#v", disabled)
	}
}

func TestTeamPatchMemberPatchDeleteAndAuditGet(t *testing.T) {
	app := api.NewApp()
	handler := api.NewMux(app)

	instanceReq := httptest.NewRequest(http.MethodPost, "/api/v1/agent-instances", bytes.NewBufferString(`{"display_name":"worker-agent","runtime_type":"local"}`))
	instanceRec := httptest.NewRecorder()
	handler.ServeHTTP(instanceRec, instanceReq)
	if instanceRec.Code != http.StatusCreated {
		t.Fatalf("unexpected instance create status: %d body=%s", instanceRec.Code, instanceRec.Body.String())
	}

	var createdInstance map[string]interface{}
	if err := json.Unmarshal(instanceRec.Body.Bytes(), &createdInstance); err != nil {
		t.Fatal(err)
	}
	instanceID, _ := createdInstance["id"].(string)

	teamReq := httptest.NewRequest(http.MethodPost, "/api/v1/teams", bytes.NewBufferString(`{"name":"core-team","description":"initial"}`))
	teamRec := httptest.NewRecorder()
	handler.ServeHTTP(teamRec, teamReq)
	if teamRec.Code != http.StatusCreated {
		t.Fatalf("unexpected team create status: %d body=%s", teamRec.Code, teamRec.Body.String())
	}

	var createdTeam map[string]interface{}
	if err := json.Unmarshal(teamRec.Body.Bytes(), &createdTeam); err != nil {
		t.Fatal(err)
	}
	teamID, _ := createdTeam["id"].(string)

	patchTeamReq := httptest.NewRequest(http.MethodPatch, "/api/v1/teams/"+teamID, bytes.NewBufferString(`{"description":"updated","entry_agent_id":"`+instanceID+`"}`))
	patchTeamRec := httptest.NewRecorder()
	handler.ServeHTTP(patchTeamRec, patchTeamReq)
	if patchTeamRec.Code != http.StatusOK {
		t.Fatalf("unexpected team patch status: %d body=%s", patchTeamRec.Code, patchTeamRec.Body.String())
	}

	memberReq := httptest.NewRequest(http.MethodPost, "/api/v1/teams/"+teamID+"/members", bytes.NewBufferString(`{"agent_id":"`+instanceID+`","role":"worker","sort_order":1}`))
	memberRec := httptest.NewRecorder()
	handler.ServeHTTP(memberRec, memberReq)
	if memberRec.Code != http.StatusCreated {
		t.Fatalf("unexpected member create status: %d body=%s", memberRec.Code, memberRec.Body.String())
	}

	var createdMember map[string]interface{}
	if err := json.Unmarshal(memberRec.Body.Bytes(), &createdMember); err != nil {
		t.Fatal(err)
	}
	memberID, _ := createdMember["id"].(string)

	patchMemberReq := httptest.NewRequest(http.MethodPatch, "/api/v1/teams/"+teamID+"/members/"+memberID, bytes.NewBufferString(`{"role":"lead","sort_order":2,"status":"inactive"}`))
	patchMemberRec := httptest.NewRecorder()
	handler.ServeHTTP(patchMemberRec, patchMemberReq)
	if patchMemberRec.Code != http.StatusOK {
		t.Fatalf("unexpected member patch status: %d body=%s", patchMemberRec.Code, patchMemberRec.Body.String())
	}

	deleteMemberReq := httptest.NewRequest(http.MethodDelete, "/api/v1/teams/"+teamID+"/members/"+memberID, nil)
	deleteMemberRec := httptest.NewRecorder()
	handler.ServeHTTP(deleteMemberRec, deleteMemberReq)
	if deleteMemberRec.Code != http.StatusOK {
		t.Fatalf("unexpected member delete status: %d body=%s", deleteMemberRec.Code, deleteMemberRec.Body.String())
	}

	auditListReq := httptest.NewRequest(http.MethodGet, "/api/v1/audit-events", nil)
	auditListRec := httptest.NewRecorder()
	handler.ServeHTTP(auditListRec, auditListReq)
	if auditListRec.Code != http.StatusOK {
		t.Fatalf("unexpected audit list status: %d body=%s", auditListRec.Code, auditListRec.Body.String())
	}

	var auditList struct {
		Items []struct {
			ID        string `json:"id"`
			EventName string `json:"event_name"`
		} `json:"items"`
	}
	if err := json.Unmarshal(auditListRec.Body.Bytes(), &auditList); err != nil {
		t.Fatal(err)
	}
	if len(auditList.Items) == 0 {
		t.Fatalf("expected audit events, got %#v", auditList.Items)
	}

	auditID := auditList.Items[0].ID
	getAuditReq := httptest.NewRequest(http.MethodGet, "/api/v1/audit-events/"+auditID, nil)
	getAuditRec := httptest.NewRecorder()
	handler.ServeHTTP(getAuditRec, getAuditReq)
	if getAuditRec.Code != http.StatusOK {
		t.Fatalf("unexpected audit get status: %d body=%s", getAuditRec.Code, getAuditRec.Body.String())
	}
}

func TestConversationRunsCreatedForExternalMessages(t *testing.T) {
	app := api.NewApp()
	handler := api.NewMux(app)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/conversations", bytes.NewBufferString(`{"mode":"single","title":"demo chat","target_agent_id":"agent-profile-1","created_by":"tester"}`))
	createRec := httptest.NewRecorder()
	handler.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("unexpected create status: %d body=%s", createRec.Code, createRec.Body.String())
	}
	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	conversationID, _ := created["id"].(string)
	if conversationID == "" {
		t.Fatalf("expected conversation id, got %#v", created)
	}

	messageReq := httptest.NewRequest(http.MethodPost, "/api/v1/conversations/"+conversationID+"/messages", bytes.NewBufferString(`{"role":"user","content":"hello gateway"}`))
	messageRec := httptest.NewRecorder()
	handler.ServeHTTP(messageRec, messageReq)
	if messageRec.Code != http.StatusCreated {
		t.Fatalf("unexpected message create status: %d body=%s", messageRec.Code, messageRec.Body.String())
	}

	runsReq := httptest.NewRequest(http.MethodGet, "/api/v1/conversations/"+conversationID+"/runs", nil)
	runsRec := httptest.NewRecorder()
	handler.ServeHTTP(runsRec, runsReq)
	if runsRec.Code != http.StatusOK {
		t.Fatalf("unexpected runs status: %d body=%s", runsRec.Code, runsRec.Body.String())
	}

	var runs struct {
		Items []struct {
			ConversationID   string `json:"conversation_id"`
			TriggerType      string `json:"trigger_type"`
			TriggerMessageID string `json:"trigger_message_id"`
			Status           string `json:"status"`
			Summary          string `json:"summary"`
		} `json:"items"`
	}
	if err := json.Unmarshal(runsRec.Body.Bytes(), &runs); err != nil {
		t.Fatal(err)
	}
	if len(runs.Items) != 1 {
		t.Fatalf("expected one run, got %#v", runs.Items)
	}
	if runs.Items[0].ConversationID != conversationID || runs.Items[0].TriggerType != "message" || runs.Items[0].TriggerMessageID == "" || runs.Items[0].Status != "completed" {
		t.Fatalf("unexpected run item: %#v", runs.Items[0])
	}

	getConversationReq := httptest.NewRequest(http.MethodGet, "/api/v1/conversations/"+conversationID, nil)
	getConversationRec := httptest.NewRecorder()
	handler.ServeHTTP(getConversationRec, getConversationReq)
	if getConversationRec.Code != http.StatusOK {
		t.Fatalf("unexpected get conversation status: %d body=%s", getConversationRec.Code, getConversationRec.Body.String())
	}
	var conversation struct {
		LastRunID string `json:"last_run_id"`
	}
	if err := json.Unmarshal(getConversationRec.Body.Bytes(), &conversation); err != nil {
		t.Fatal(err)
	}
	if conversation.LastRunID == "" {
		t.Fatalf("expected last_run_id, got %#v", conversation)
	}
}

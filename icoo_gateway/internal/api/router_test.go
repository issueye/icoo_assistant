package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"icoo_gateway/internal/api"
)

func TestHealthzEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	api.NewMux(api.NewApp()).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("unexpected content type: %q", got)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["service"] != "icoo_gateway" || body["status"] != "ok" {
		t.Fatalf("unexpected body: %#v", body)
	}
}

func TestRootEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	api.NewMux(api.NewApp()).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d", rec.Code)
	}
	var body struct {
		Service string   `json:"service"`
		Status  string   `json:"status"`
		Routes  []string `json:"routes"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Service != "icoo_gateway" || body.Status != "ready" {
		t.Fatalf("unexpected body: %#v", body)
	}
	if len(body.Routes) != 8 {
		t.Fatalf("unexpected routes: %#v", body.Routes)
	}
}

func TestSkillCreateListAndGet(t *testing.T) {
	app := api.NewApp()
	handler := api.NewMux(app)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/skills", bytes.NewBufferString(`{"name":"code-review","version":"v1","description":"review skill"}`))
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

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/skills", nil)
	listRec := httptest.NewRecorder()
	handler.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("unexpected list status: %d", listRec.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/skills/"+skillID, nil)
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("unexpected get status: %d body=%s", getRec.Code, getRec.Body.String())
	}
}

func TestAgentProfileCreateAndGet(t *testing.T) {
	app := api.NewApp()
	handler := api.NewMux(app)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/agent-profiles", bytes.NewBufferString(`{"name":"lead","model_provider":"anthropic","model_name":"claude-opus-4-1"}`))
	createRec := httptest.NewRecorder()
	handler.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("unexpected create status: %d body=%s", createRec.Code, createRec.Body.String())
	}
	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	profileID, _ := created["id"].(string)
	if profileID == "" {
		t.Fatalf("expected profile id, got %#v", created)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/agent-profiles/"+profileID, nil)
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("unexpected get status: %d body=%s", getRec.Code, getRec.Body.String())
	}
}

func TestTeamCreateAndGet(t *testing.T) {
	app := api.NewApp()
	handler := api.NewMux(app)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/teams", bytes.NewBufferString(`{"name":"core-team","description":"core delivery team","entry_agent_id":"agent-profile-1"}`))
	createRec := httptest.NewRecorder()
	handler.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("unexpected create status: %d body=%s", createRec.Code, createRec.Body.String())
	}
	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	teamID, _ := created["id"].(string)
	if teamID == "" {
		t.Fatalf("expected team id, got %#v", created)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/teams/"+teamID, nil)
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("unexpected get status: %d body=%s", getRec.Code, getRec.Body.String())
	}
}

func TestAgentInstanceCreateListAndGet(t *testing.T) {
	app := api.NewApp()
	handler := api.NewMux(app)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/agent-instances", bytes.NewBufferString(`{"display_name":"lead-agent","runtime_type":"local","status":"idle"}`))
	createRec := httptest.NewRecorder()
	handler.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("unexpected create status: %d body=%s", createRec.Code, createRec.Body.String())
	}
	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	instanceID, _ := created["id"].(string)
	if instanceID == "" {
		t.Fatalf("expected instance id, got %#v", created)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/agent-instances", nil)
	listRec := httptest.NewRecorder()
	handler.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("unexpected list status: %d body=%s", listRec.Code, listRec.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/agent-instances/"+instanceID, nil)
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("unexpected get status: %d body=%s", getRec.Code, getRec.Body.String())
	}
}

func TestAgentInstanceHeartbeatUpdatesStatusAndTimestamp(t *testing.T) {
	app := api.NewApp()
	handler := api.NewMux(app)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/agent-instances", bytes.NewBufferString(`{"display_name":"lead-agent","runtime_type":"local","status":"offline"}`))
	createRec := httptest.NewRecorder()
	handler.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("unexpected create status: %d body=%s", createRec.Code, createRec.Body.String())
	}
	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	instanceID, _ := created["id"].(string)
	if instanceID == "" {
		t.Fatalf("expected instance id, got %#v", created)
	}

	heartbeatReq := httptest.NewRequest(http.MethodPost, "/api/v1/agent-instances/"+instanceID+"/heartbeat", nil)
	heartbeatRec := httptest.NewRecorder()
	handler.ServeHTTP(heartbeatRec, heartbeatReq)
	if heartbeatRec.Code != http.StatusOK {
		t.Fatalf("unexpected heartbeat status: %d body=%s", heartbeatRec.Code, heartbeatRec.Body.String())
	}
	var heartbeat struct {
		Status          string `json:"status"`
		LastHeartbeatAt string `json:"last_heartbeat_at"`
	}
	if err := json.Unmarshal(heartbeatRec.Body.Bytes(), &heartbeat); err != nil {
		t.Fatal(err)
	}
	if heartbeat.Status != "idle" {
		t.Fatalf("expected idle after heartbeat, got %#v", heartbeat)
	}
	if heartbeat.LastHeartbeatAt == "" {
		t.Fatalf("expected last_heartbeat_at, got %#v", heartbeat)
	}
}

func TestTeamMembersAddAndList(t *testing.T) {
	app := api.NewApp()
	handler := api.NewMux(app)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/teams", bytes.NewBufferString(`{"name":"core-team","entry_agent_id":"lead-agent"}`))
	createRec := httptest.NewRecorder()
	handler.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("unexpected create status: %d body=%s", createRec.Code, createRec.Body.String())
	}
	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	teamID, _ := created["id"].(string)
	if teamID == "" {
		t.Fatalf("expected team id, got %#v", created)
	}

	instanceReq := httptest.NewRequest(http.MethodPost, "/api/v1/agent-instances", bytes.NewBufferString(`{"display_name":"lead-agent","runtime_type":"local"}`))
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
	if instanceID == "" {
		t.Fatalf("expected instance id, got %#v", createdInstance)
	}

	memberReq := httptest.NewRequest(http.MethodPost, "/api/v1/teams/"+teamID+"/members", bytes.NewBufferString(`{"agent_id":"`+instanceID+`","role":"lead","sort_order":1}`))
	memberRec := httptest.NewRecorder()
	handler.ServeHTTP(memberRec, memberReq)
	if memberRec.Code != http.StatusCreated {
		t.Fatalf("unexpected member create status: %d body=%s", memberRec.Code, memberRec.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/teams/"+teamID+"/members", nil)
	listRec := httptest.NewRecorder()
	handler.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("unexpected member list status: %d body=%s", listRec.Code, listRec.Body.String())
	}
	var list struct {
		Items []struct {
			AgentID   string `json:"agent_id"`
			Role      string `json:"role"`
			SortOrder int    `json:"sort_order"`
		} `json:"items"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("expected one member, got %#v", list.Items)
	}
	if list.Items[0].AgentID != instanceID || list.Items[0].Role != "lead" || list.Items[0].SortOrder != 1 {
		t.Fatalf("unexpected member list: %#v", list.Items)
	}
}

func TestTeamMemberRequiresExistingAgentInstance(t *testing.T) {
	app := api.NewApp()
	handler := api.NewMux(app)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/teams", bytes.NewBufferString(`{"name":"core-team","entry_agent_id":"agent-instance-1"}`))
	createRec := httptest.NewRecorder()
	handler.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("unexpected create status: %d body=%s", createRec.Code, createRec.Body.String())
	}
	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	teamID, _ := created["id"].(string)
	if teamID == "" {
		t.Fatalf("expected team id, got %#v", created)
	}

	memberReq := httptest.NewRequest(http.MethodPost, "/api/v1/teams/"+teamID+"/members", bytes.NewBufferString(`{"agent_id":"agent-instance-999","role":"lead"}`))
	memberRec := httptest.NewRecorder()
	handler.ServeHTTP(memberRec, memberReq)
	if memberRec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d body=%s", memberRec.Code, memberRec.Body.String())
	}
}

func TestCreateEndpointsValidateRequiredName(t *testing.T) {
	app := api.NewApp()
	handler := api.NewMux(app)

	for _, path := range []string{
		"/api/v1/skills",
		"/api/v1/agent-profiles",
		"/api/v1/teams",
	} {
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(`{"description":"missing name"}`))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected bad request for %s, got %d", path, rec.Code)
		}
	}
}

func TestGetEndpointsReturnNotFound(t *testing.T) {
	app := api.NewApp()
	handler := api.NewMux(app)

	for _, path := range []string{
		"/api/v1/skills/skill-999",
		"/api/v1/agent-profiles/agent-profile-999",
		"/api/v1/agent-instances/agent-instance-999",
		"/api/v1/agent-instances/agent-instance-999/heartbeat",
		"/api/v1/teams/team-999",
		"/api/v1/teams/team-999/members",
		"/api/v1/conversations/conv-999",
		"/api/v1/conversations/conv-999/messages",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if strings.HasSuffix(path, "/heartbeat") {
			req = httptest.NewRequest(http.MethodPost, path, nil)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected not found for %s, got %d", path, rec.Code)
		}
	}
}

func TestConversationCreateListGetAndMessages(t *testing.T) {
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

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/conversations", nil)
	listRec := httptest.NewRecorder()
	handler.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("unexpected list status: %d", listRec.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/conversations/"+conversationID, nil)
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("unexpected get status: %d body=%s", getRec.Code, getRec.Body.String())
	}

	messageReq := httptest.NewRequest(http.MethodPost, "/api/v1/conversations/"+conversationID+"/messages", bytes.NewBufferString(`{"role":"user","content":"hello gateway"}`))
	messageRec := httptest.NewRecorder()
	handler.ServeHTTP(messageRec, messageReq)
	if messageRec.Code != http.StatusCreated {
		t.Fatalf("unexpected message create status: %d body=%s", messageRec.Code, messageRec.Body.String())
	}

	listMessagesReq := httptest.NewRequest(http.MethodGet, "/api/v1/conversations/"+conversationID+"/messages", nil)
	listMessagesRec := httptest.NewRecorder()
	handler.ServeHTTP(listMessagesRec, listMessagesReq)
	if listMessagesRec.Code != http.StatusOK {
		t.Fatalf("unexpected message list status: %d body=%s", listMessagesRec.Code, listMessagesRec.Body.String())
	}
	var messageList struct {
		Items []struct {
			Scope      string `json:"scope"`
			Role       string `json:"role"`
			Content    string `json:"content"`
			SequenceNo int    `json:"sequence_no"`
		} `json:"items"`
	}
	if err := json.Unmarshal(listMessagesRec.Body.Bytes(), &messageList); err != nil {
		t.Fatal(err)
	}
	if len(messageList.Items) != 1 {
		t.Fatalf("expected one message, got %#v", messageList.Items)
	}
	if messageList.Items[0].Scope != "external" || messageList.Items[0].Role != "user" || messageList.Items[0].Content != "hello gateway" || messageList.Items[0].SequenceNo != 1 {
		t.Fatalf("unexpected message list: %#v", messageList.Items)
	}
}

func TestConversationCreateValidatesTargetByMode(t *testing.T) {
	app := api.NewApp()
	handler := api.NewMux(app)

	for _, body := range []string{
		`{"mode":"single","title":"missing agent"}`,
		`{"mode":"team","title":"missing team"}`,
		`{"mode":"unsupported","title":"bad mode"}`,
	} {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/conversations", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected bad request for body=%s, got %d", body, rec.Code)
		}
	}
}

func TestConversationMessageValidatesFields(t *testing.T) {
	app := api.NewApp()
	handler := api.NewMux(app)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/conversations", bytes.NewBufferString(`{"mode":"single","title":"demo chat","target_agent_id":"agent-profile-1"}`))
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

	req := httptest.NewRequest(http.MethodPost, "/api/v1/conversations/"+conversationID+"/messages", bytes.NewBufferString(`{"role":"user"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d", rec.Code)
	}
}

func TestTeamConversationSupportsInternalMessagesAndScopeFilter(t *testing.T) {
	app := api.NewApp()
	handler := api.NewMux(app)

	instanceReq := httptest.NewRequest(http.MethodPost, "/api/v1/agent-instances", bytes.NewBufferString(`{"display_name":"lead-agent","runtime_type":"local"}`))
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
	if instanceID == "" {
		t.Fatalf("expected instance id, got %#v", createdInstance)
	}

	teamCreateReq := httptest.NewRequest(http.MethodPost, "/api/v1/teams", bytes.NewBufferString(`{"name":"core-team","entry_agent_id":"`+instanceID+`"}`))
	teamCreateRec := httptest.NewRecorder()
	handler.ServeHTTP(teamCreateRec, teamCreateReq)
	if teamCreateRec.Code != http.StatusCreated {
		t.Fatalf("unexpected team create status: %d body=%s", teamCreateRec.Code, teamCreateRec.Body.String())
	}
	var createdTeam map[string]interface{}
	if err := json.Unmarshal(teamCreateRec.Body.Bytes(), &createdTeam); err != nil {
		t.Fatal(err)
	}
	teamID, _ := createdTeam["id"].(string)
	if teamID == "" {
		t.Fatalf("expected team id, got %#v", createdTeam)
	}

	memberReq := httptest.NewRequest(http.MethodPost, "/api/v1/teams/"+teamID+"/members", bytes.NewBufferString(`{"agent_id":"`+instanceID+`","role":"lead","sort_order":1}`))
	memberRec := httptest.NewRecorder()
	handler.ServeHTTP(memberRec, memberReq)
	if memberRec.Code != http.StatusCreated {
		t.Fatalf("unexpected member create status: %d body=%s", memberRec.Code, memberRec.Body.String())
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/conversations", bytes.NewBufferString(`{"mode":"team","title":"team demo","target_team_id":"`+teamID+`","created_by":"tester"}`))
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

	externalReq := httptest.NewRequest(http.MethodPost, "/api/v1/conversations/"+conversationID+"/messages", bytes.NewBufferString(`{"scope":"external","role":"user","content":"请团队协作处理"}`))
	externalRec := httptest.NewRecorder()
	handler.ServeHTTP(externalRec, externalReq)
	if externalRec.Code != http.StatusCreated {
		t.Fatalf("unexpected external message status: %d body=%s", externalRec.Code, externalRec.Body.String())
	}

	listInternalReq := httptest.NewRequest(http.MethodGet, "/api/v1/conversations/"+conversationID+"/messages?scope=internal", nil)
	listInternalRec := httptest.NewRecorder()
	handler.ServeHTTP(listInternalRec, listInternalReq)
	if listInternalRec.Code != http.StatusOK {
		t.Fatalf("unexpected internal list status: %d body=%s", listInternalRec.Code, listInternalRec.Body.String())
	}
	var internalList struct {
		Items []struct {
			Scope        string `json:"scope"`
			SenderType   string `json:"sender_type"`
			SenderID     string `json:"sender_id"`
			ReceiverType string `json:"receiver_type"`
			ReceiverID   string `json:"receiver_id"`
			Content      string `json:"content"`
			SequenceNo   int    `json:"sequence_no"`
		} `json:"items"`
	}
	if err := json.Unmarshal(listInternalRec.Body.Bytes(), &internalList); err != nil {
		t.Fatal(err)
	}
	if len(internalList.Items) != 1 {
		t.Fatalf("expected one internal message, got %#v", internalList.Items)
	}
	item := internalList.Items[0]
	if item.Scope != "internal" || item.SenderType != "system" || item.SenderID != "router" || item.ReceiverType != "agent" || item.ReceiverID != instanceID || item.SequenceNo != 2 {
		t.Fatalf("unexpected internal item: %#v", item)
	}
	if item.Content == "" {
		t.Fatalf("expected internal dispatch content, got %#v", item)
	}

	listSystemReq := httptest.NewRequest(http.MethodGet, "/api/v1/conversations/"+conversationID+"/messages?scope=system", nil)
	listSystemRec := httptest.NewRecorder()
	handler.ServeHTTP(listSystemRec, listSystemReq)
	if listSystemRec.Code != http.StatusOK {
		t.Fatalf("unexpected system list status: %d body=%s", listSystemRec.Code, listSystemRec.Body.String())
	}
	var systemList struct {
		Items []struct {
			Scope      string `json:"scope"`
			Content    string `json:"content"`
			SequenceNo int    `json:"sequence_no"`
		} `json:"items"`
	}
	if err := json.Unmarshal(listSystemRec.Body.Bytes(), &systemList); err != nil {
		t.Fatal(err)
	}
	if len(systemList.Items) != 1 {
		t.Fatalf("expected one system message, got %#v", systemList.Items)
	}
	if systemList.Items[0].Scope != "system" || systemList.Items[0].SequenceNo != 3 {
		t.Fatalf("unexpected system item: %#v", systemList.Items[0])
	}
}

func TestSingleConversationRejectsInternalMessages(t *testing.T) {
	app := api.NewApp()
	handler := api.NewMux(app)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/conversations", bytes.NewBufferString(`{"mode":"single","title":"demo chat","target_agent_id":"agent-profile-1"}`))
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

	req := httptest.NewRequest(http.MethodPost, "/api/v1/conversations/"+conversationID+"/messages", bytes.NewBufferString(`{"scope":"internal","role":"assistant","sender_type":"agent","sender_id":"lead","receiver_type":"agent","receiver_id":"worker-1","content":"not allowed"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestTeamInternalMessageRequiresParticipants(t *testing.T) {
	app := api.NewApp()
	handler := api.NewMux(app)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/conversations", bytes.NewBufferString(`{"mode":"team","title":"team demo","target_team_id":"team-1"}`))
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

	req := httptest.NewRequest(http.MethodPost, "/api/v1/conversations/"+conversationID+"/messages", bytes.NewBufferString(`{"scope":"internal","role":"assistant","content":"missing participants"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestTeamRoutingWithoutEntryAgentCreatesWarning(t *testing.T) {
	app := api.NewApp()
	handler := api.NewMux(app)

	teamCreateReq := httptest.NewRequest(http.MethodPost, "/api/v1/teams", bytes.NewBufferString(`{"name":"core-team"}`))
	teamCreateRec := httptest.NewRecorder()
	handler.ServeHTTP(teamCreateRec, teamCreateReq)
	if teamCreateRec.Code != http.StatusCreated {
		t.Fatalf("unexpected team create status: %d body=%s", teamCreateRec.Code, teamCreateRec.Body.String())
	}
	var createdTeam map[string]interface{}
	if err := json.Unmarshal(teamCreateRec.Body.Bytes(), &createdTeam); err != nil {
		t.Fatal(err)
	}
	teamID, _ := createdTeam["id"].(string)
	if teamID == "" {
		t.Fatalf("expected team id, got %#v", createdTeam)
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/conversations", bytes.NewBufferString(`{"mode":"team","title":"team demo","target_team_id":"`+teamID+`"}`))
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

	externalReq := httptest.NewRequest(http.MethodPost, "/api/v1/conversations/"+conversationID+"/messages", bytes.NewBufferString(`{"scope":"external","role":"user","content":"请团队协作处理"}`))
	externalRec := httptest.NewRecorder()
	handler.ServeHTTP(externalRec, externalReq)
	if externalRec.Code != http.StatusCreated {
		t.Fatalf("unexpected external message status: %d body=%s", externalRec.Code, externalRec.Body.String())
	}

	listInternalReq := httptest.NewRequest(http.MethodGet, "/api/v1/conversations/"+conversationID+"/messages?scope=internal", nil)
	listInternalRec := httptest.NewRecorder()
	handler.ServeHTTP(listInternalRec, listInternalReq)
	if listInternalRec.Code != http.StatusOK {
		t.Fatalf("unexpected internal list status: %d body=%s", listInternalRec.Code, listInternalRec.Body.String())
	}
	var internalList struct {
		Items []interface{} `json:"items"`
	}
	if err := json.Unmarshal(listInternalRec.Body.Bytes(), &internalList); err != nil {
		t.Fatal(err)
	}
	if len(internalList.Items) != 0 {
		t.Fatalf("expected no internal dispatch when entry agent missing, got %#v", internalList.Items)
	}

	listSystemReq := httptest.NewRequest(http.MethodGet, "/api/v1/conversations/"+conversationID+"/messages?scope=system", nil)
	listSystemRec := httptest.NewRecorder()
	handler.ServeHTTP(listSystemRec, listSystemReq)
	if listSystemRec.Code != http.StatusOK {
		t.Fatalf("unexpected system list status: %d body=%s", listSystemRec.Code, listSystemRec.Body.String())
	}
	var systemList struct {
		Items []struct {
			Content string `json:"content"`
		} `json:"items"`
	}
	if err := json.Unmarshal(listSystemRec.Body.Bytes(), &systemList); err != nil {
		t.Fatal(err)
	}
	if len(systemList.Items) != 1 {
		t.Fatalf("expected one system warning, got %#v", systemList.Items)
	}
	if systemList.Items[0].Content == "" {
		t.Fatalf("expected warning content, got %#v", systemList.Items[0])
	}
}

func TestTeamRoutingWithEntryAgentNotMemberCreatesWarning(t *testing.T) {
	app := api.NewApp()
	handler := api.NewMux(app)

	teamCreateReq := httptest.NewRequest(http.MethodPost, "/api/v1/teams", bytes.NewBufferString(`{"name":"core-team","entry_agent_id":"agent-instance-missing"}`))
	teamCreateRec := httptest.NewRecorder()
	handler.ServeHTTP(teamCreateRec, teamCreateReq)
	if teamCreateRec.Code != http.StatusCreated {
		t.Fatalf("unexpected team create status: %d body=%s", teamCreateRec.Code, teamCreateRec.Body.String())
	}
	var createdTeam map[string]interface{}
	if err := json.Unmarshal(teamCreateRec.Body.Bytes(), &createdTeam); err != nil {
		t.Fatal(err)
	}
	teamID, _ := createdTeam["id"].(string)
	if teamID == "" {
		t.Fatalf("expected team id, got %#v", createdTeam)
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/conversations", bytes.NewBufferString(`{"mode":"team","title":"team demo","target_team_id":"`+teamID+`"}`))
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

	externalReq := httptest.NewRequest(http.MethodPost, "/api/v1/conversations/"+conversationID+"/messages", bytes.NewBufferString(`{"scope":"external","role":"user","content":"请团队协作处理"}`))
	externalRec := httptest.NewRecorder()
	handler.ServeHTTP(externalRec, externalReq)
	if externalRec.Code != http.StatusCreated {
		t.Fatalf("unexpected external message status: %d body=%s", externalRec.Code, externalRec.Body.String())
	}

	listSystemReq := httptest.NewRequest(http.MethodGet, "/api/v1/conversations/"+conversationID+"/messages?scope=system", nil)
	listSystemRec := httptest.NewRecorder()
	handler.ServeHTTP(listSystemRec, listSystemReq)
	if listSystemRec.Code != http.StatusOK {
		t.Fatalf("unexpected system list status: %d body=%s", listSystemRec.Code, listSystemRec.Body.String())
	}
	var systemList struct {
		Items []struct {
			Content string `json:"content"`
		} `json:"items"`
	}
	if err := json.Unmarshal(listSystemRec.Body.Bytes(), &systemList); err != nil {
		t.Fatal(err)
	}
	if len(systemList.Items) != 1 {
		t.Fatalf("expected one system warning, got %#v", systemList.Items)
	}
	if systemList.Items[0].Content == "" {
		t.Fatalf("expected warning content, got %#v", systemList.Items[0])
	}
}

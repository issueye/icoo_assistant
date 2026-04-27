package proxy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"icoo_proxy/internal/catalog"
	"icoo_proxy/internal/config"
	"icoo_proxy/internal/consts"
)

func newLoopbackRequest(method, target string, body *bytes.Buffer) *http.Request {
	req := httptest.NewRequest(method, target, body)
	req.RemoteAddr = "127.0.0.1:34567"
	return req
}

func TestHandleAcceptsConfiguredAuthKeyList(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"chatcmpl_123","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]}`))
	}))
	defer upstream.Close()

	cfg := config.Config{
		AllowUnauthenticatedLocal: false,
		ProxyAPIKeys:              []string{"client-one", "client-two"},
		OpenAIBaseURL:             upstream.URL,
		OpenAIApiKey:              "test-openai-key",
		DefaultChatRoute:          "openai-chat:gpt-test",
	}
	cat, err := catalog.New(cfg)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	service := New(cfg, cat)

	for _, header := range []string{"x-api-key", "Authorization"} {
		req := newLoopbackRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{"model":"gpt-test","messages":[{"role":"user","content":"hello"}]}`))
		if header == "Authorization" {
			req.Header.Set(header, "Bearer client-two")
		} else {
			req.Header.Set(header, "client-two")
		}
		rec := httptest.NewRecorder()

		service.Handle(rec, req, consts.ProtocolOpenAIChat)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s expected status 200, got %d body=%s", header, rec.Code, rec.Body.String())
		}
	}

	req := newLoopbackRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{"model":"gpt-test","messages":[{"role":"user","content":"hello"}]}`))
	req.Header.Set("x-api-key", "bad-key")
	rec := httptest.NewRecorder()

	service.Handle(rec, req, consts.ProtocolOpenAIChat)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("invalid proxy api key")) {
		t.Fatalf("expected invalid key error, got %s", rec.Body.String())
	}
}

func TestHandleAllowsUnauthenticatedLoopbackOnly(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"resp_local","object":"response","status":"completed","output_text":"ok","output":[]}`))
	}))
	defer upstream.Close()

	cfg := config.Config{
		AllowUnauthenticatedLocal: true,
		OpenAIBaseURL:             upstream.URL,
		OpenAIApiKey:              "test-openai-key",
		DefaultResponsesRoute:     "openai-responses:gpt-4.1-mini",
	}
	cat, err := catalog.New(cfg)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	service := New(cfg, cat)

	req := newLoopbackRequest(http.MethodPost, "/v1/responses", bytes.NewBufferString(`{"model":"gpt-4.1-mini","input":"hello"}`))
	req.RemoteAddr = "127.0.0.1:34567"
	rec := httptest.NewRecorder()

	service.Handle(rec, req, consts.ProtocolOpenAIResponses)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected loopback request to be allowed, got %d body=%s", rec.Code, rec.Body.String())
	}

	remoteReq := newLoopbackRequest(http.MethodPost, "/v1/responses", bytes.NewBufferString(`{"model":"gpt-4.1-mini","input":"hello"}`))
	remoteReq.RemoteAddr = "203.0.113.10:4567"
	remoteRec := httptest.NewRecorder()

	service.Handle(remoteRec, remoteReq, consts.ProtocolOpenAIResponses)

	if remoteRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected non-loopback request to require auth, got %d body=%s", remoteRec.Code, remoteRec.Body.String())
	}
	if !bytes.Contains(remoteRec.Body.Bytes(), []byte("proxy api key is required")) {
		t.Fatalf("expected missing auth error, got %s", remoteRec.Body.String())
	}
}

func TestSanitizedHeadersRedactsSecrets(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer secret")
	headers.Set("x-api-key", "secret")
	headers.Set("Content-Type", "application/json")

	got := sanitizedHeaders(headers)
	if got["Authorization"][0] != "<redacted>" {
		t.Fatalf("expected authorization redacted, got %#v", got["Authorization"])
	}
	if got["X-Api-Key"][0] != "<redacted>" {
		t.Fatalf("expected api key redacted, got %#v", got["X-Api-Key"])
	}
	if got["Content-Type"][0] != "application/json" {
		t.Fatalf("expected content type preserved, got %#v", got["Content-Type"])
	}
}

func TestHandleAppliesConfiguredUpstreamUserAgent(t *testing.T) {
	var gotUserAgent string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserAgent = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"resp_ua","object":"response","status":"completed","output_text":"ok","output":[]}`))
	}))
	defer upstream.Close()

	cfg := config.Config{
		AllowUnauthenticatedLocal: true,
		OpenAIResponsesBaseURL:    upstream.URL,
		OpenAIResponsesAPIKey:     "test-openai-key",
		OpenAIResponsesUserAgent:  "SupplierUA/2.0",
		DefaultResponsesRoute:     "openai-responses:gpt-4.1-mini",
	}
	cat, err := catalog.New(cfg)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	service := New(cfg, cat)

	req := newLoopbackRequest(http.MethodPost, "/v1/responses", bytes.NewBufferString(`{"model":"gpt-4.1-mini","input":"hello"}`))
	rec := httptest.NewRecorder()

	service.Handle(rec, req, consts.ProtocolOpenAIResponses)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if gotUserAgent != "SupplierUA/2.0" {
		t.Fatalf("expected configured upstream user agent, got %q", gotUserAgent)
	}
}

func TestHandleUsesSplitOpenAIUpstreams(t *testing.T) {
	var gotChatAuth string
	var gotResponsesAuth string
	var gotChatUserAgent string
	var gotResponsesUserAgent string

	chatUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotChatAuth = r.Header.Get("Authorization")
		gotChatUserAgent = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"chatcmpl_split","choices":[{"index":0,"message":{"role":"assistant","content":"chat ok"},"finish_reason":"stop"}]}`))
	}))
	defer chatUpstream.Close()

	responsesUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotResponsesAuth = r.Header.Get("Authorization")
		gotResponsesUserAgent = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"resp_split","object":"response","status":"completed","output_text":"responses ok","output":[]}`))
	}))
	defer responsesUpstream.Close()

	cfg := config.Config{
		AllowUnauthenticatedLocal: true,
		OpenAIChatBaseURL:         chatUpstream.URL,
		OpenAIChatAPIKey:          "chat-secret",
		OpenAIChatUserAgent:       "ChatUA/1.0",
		OpenAIResponsesBaseURL:    responsesUpstream.URL,
		OpenAIResponsesAPIKey:     "responses-secret",
		OpenAIResponsesUserAgent:  "ResponsesUA/1.0",
		DefaultChatRoute:          "openai-chat:gpt-4o-mini",
		DefaultResponsesRoute:     "openai-responses:gpt-4.1-mini",
	}
	cat, err := catalog.New(cfg)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	service := New(cfg, cat)

	chatReq := newLoopbackRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hello"}]}`))
	chatRec := httptest.NewRecorder()
	service.Handle(chatRec, chatReq, consts.ProtocolOpenAIChat)
	if chatRec.Code != http.StatusOK {
		t.Fatalf("expected chat status 200, got %d body=%s", chatRec.Code, chatRec.Body.String())
	}

	responsesReq := newLoopbackRequest(http.MethodPost, "/v1/responses", bytes.NewBufferString(`{"model":"gpt-4.1-mini","input":"hello"}`))
	responsesRec := httptest.NewRecorder()
	service.Handle(responsesRec, responsesReq, consts.ProtocolOpenAIResponses)
	if responsesRec.Code != http.StatusOK {
		t.Fatalf("expected responses status 200, got %d body=%s", responsesRec.Code, responsesRec.Body.String())
	}

	if gotChatAuth != "Bearer chat-secret" {
		t.Fatalf("expected chat upstream auth header, got %q", gotChatAuth)
	}
	if gotResponsesAuth != "Bearer responses-secret" {
		t.Fatalf("expected responses upstream auth header, got %q", gotResponsesAuth)
	}
	if gotChatUserAgent != "ChatUA/1.0" {
		t.Fatalf("expected chat upstream user agent, got %q", gotChatUserAgent)
	}
	if gotResponsesUserAgent != "ResponsesUA/1.0" {
		t.Fatalf("expected responses upstream user agent, got %q", gotResponsesUserAgent)
	}
}

func TestHandleAnthropicPassthroughRewritesAliasModel(t *testing.T) {
	var gotAuth string
	var gotVersion string
	var gotModel string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("x-api-key")
		gotVersion = r.Header.Get("anthropic-version")
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode upstream body: %v", err)
		}
		gotModel, _ = payload["model"].(string)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"msg_123","type":"message","model":"claude-real"}`))
	}))
	defer upstream.Close()

	cfg := config.Config{
		AllowUnauthenticatedLocal: true,
		AnthropicBaseURL:          upstream.URL,
		AnthropicAPIKey:           "test-anthropic-key",
		AnthropicVersion:          "2023-06-01",
		ModelRoutes:               "claude-sonnet=anthropic:claude-real",
	}
	cat, err := catalog.New(cfg)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	service := New(cfg, cat)

	req := newLoopbackRequest(http.MethodPost, "/v1/messages", bytes.NewBufferString(`{"model":"claude-sonnet","messages":[{"role":"user","content":"hello"}],"max_tokens":64}`))
	rec := httptest.NewRecorder()

	service.Handle(rec, req, consts.ProtocolAnthropic)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if gotAuth != "test-anthropic-key" {
		t.Fatalf("expected upstream auth header, got %q", gotAuth)
	}
	if gotVersion != "2023-06-01" {
		t.Fatalf("expected anthropic version, got %q", gotVersion)
	}
	if gotModel != "claude-real" {
		t.Fatalf("expected aliased model rewritten, got %q", gotModel)
	}
}

func TestHandleAnthropicExplicitModelUsesMappedResponsesSupplier(t *testing.T) {
	var gotPath string
	var gotAuth string
	var gotModel string

	responsesUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode upstream body: %v", err)
		}
		gotModel, _ = payload["model"].(string)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"resp_explicit","object":"response","status":"completed","output_text":"ok","output":[]}`))
	}))
	defer responsesUpstream.Close()

	cfg := config.Config{
		AllowUnauthenticatedLocal: true,
		OpenAIResponsesBaseURL:    responsesUpstream.URL,
		OpenAIResponsesAPIKey:     "responses-secret",
		DefaultAnthropicRoute:     "openai-responses:gpt-5.4",
	}
	cat, err := catalog.New(cfg)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	service := New(cfg, cat)

	req := newLoopbackRequest(http.MethodPost, "/v1/messages", bytes.NewBufferString(`{"model":"claude-3-7-sonnet","messages":[{"role":"user","content":"hello"}],"max_tokens":64}`))
	rec := httptest.NewRecorder()

	service.Handle(rec, req, consts.ProtocolAnthropic)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if gotPath != "/v1/responses" {
		t.Fatalf("expected anthropic request to route to responses upstream path, got %q", gotPath)
	}
	if gotAuth != "Bearer responses-secret" {
		t.Fatalf("expected responses auth header, got %q", gotAuth)
	}
	if gotModel != "claude-3-7-sonnet" {
		t.Fatalf("expected explicit request model to be preserved, got %q", gotModel)
	}
}

func TestHandleAnthropicExplicitUnknownModelReturnsUpstreamError(t *testing.T) {
	var gotPath string
	var gotModel string

	responsesUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode upstream body: %v", err)
		}
		gotModel, _ = payload["model"].(string)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"type":"invalid_request_error","message":"The model 'claude-missing' does not exist"}}`))
	}))
	defer responsesUpstream.Close()

	cfg := config.Config{
		AllowUnauthenticatedLocal: true,
		OpenAIResponsesBaseURL:    responsesUpstream.URL,
		OpenAIResponsesAPIKey:     "responses-secret",
		DefaultAnthropicRoute:     "openai-responses:gpt-5.4",
	}
	cat, err := catalog.New(cfg)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	service := New(cfg, cat)

	req := newLoopbackRequest(http.MethodPost, "/v1/messages", bytes.NewBufferString(`{"model":"claude-missing","messages":[{"role":"user","content":"hello"}],"max_tokens":64}`))
	rec := httptest.NewRecorder()

	service.Handle(rec, req, consts.ProtocolAnthropic)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected upstream bad request to pass through, got %d body=%s", rec.Code, rec.Body.String())
	}
	if gotPath != "/v1/responses" {
		t.Fatalf("expected unknown model request to still hit responses upstream path, got %q", gotPath)
	}
	if gotModel != "claude-missing" {
		t.Fatalf("expected upstream request model to be preserved, got %q", gotModel)
	}
	if !strings.Contains(rec.Body.String(), "claude-missing") {
		t.Fatalf("expected upstream model error to be returned, got %s", rec.Body.String())
	}
}

func TestHandleTranslatesChatToResponses(t *testing.T) {
	var gotModel string
	var gotInstructions string
	var gotInput []interface{}
	var gotTools []interface{}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode upstream body: %v", err)
		}
		gotModel, _ = payload["model"].(string)
		gotInstructions, _ = payload["instructions"].(string)
		gotInput, _ = payload["input"].([]interface{})
		gotTools, _ = payload["tools"].([]interface{})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"resp_123","status":"completed","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"hello from responses"}]},{"type":"function_call","call_id":"call_weather","name":"get_weather","arguments":"{\"city\":\"Shanghai\"}"}],"usage":{"input_tokens":11,"output_tokens":7,"total_tokens":18}}`))
	}))
	defer upstream.Close()

	cfg := config.Config{
		AllowUnauthenticatedLocal: true,
		OpenAIBaseURL:             upstream.URL,
		OpenAIApiKey:              "test-openai-key",
		ModelRoutes:               "assistant-default=openai-responses:gpt-4.1-mini",
	}
	cat, err := catalog.New(cfg)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	service := New(cfg, cat)

	req := newLoopbackRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{"model":"assistant-default","messages":[{"role":"system","content":"You are helpful"},{"role":"user","content":"hi"}],"tools":[{"type":"function","function":{"name":"get_weather","description":"Get weather","parameters":{"type":"object","properties":{"city":{"type":"string"}},"required":["city"]}}}],"max_tokens":32}`))
	rec := httptest.NewRecorder()

	service.Handle(rec, req, consts.ProtocolOpenAIChat)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if gotModel != "gpt-4.1-mini" {
		t.Fatalf("expected translated model, got %q", gotModel)
	}
	if gotInstructions != "You are helpful" {
		t.Fatalf("expected instructions, got %q", gotInstructions)
	}
	if len(gotInput) != 1 {
		t.Fatalf("expected one input message, got %d", len(gotInput))
	}
	if len(gotTools) != 1 {
		t.Fatalf("expected one translated tool, got %d", len(gotTools))
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	choices, _ := payload["choices"].([]interface{})
	if len(choices) != 1 {
		t.Fatalf("expected one choice, got %d", len(choices))
	}
	choice, _ := choices[0].(map[string]interface{})
	message, _ := choice["message"].(map[string]interface{})
	toolCalls, _ := message["tool_calls"].([]interface{})
	if len(toolCalls) != 1 {
		t.Fatalf("expected one translated tool call, got %d", len(toolCalls))
	}
	if choice["finish_reason"] != "tool_calls" {
		t.Fatalf("expected tool_calls finish reason, got %#v", choice["finish_reason"])
	}
}

func TestHandleTranslatesResponsesToChat(t *testing.T) {
	var gotModel string
	var gotMessages []interface{}
	var gotTools []interface{}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode upstream body: %v", err)
		}
		gotModel, _ = payload["model"].(string)
		gotMessages, _ = payload["messages"].([]interface{})
		gotTools, _ = payload["tools"].([]interface{})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"chatcmpl_123","choices":[{"index":0,"message":{"role":"assistant","content":"hello from chat"},"finish_reason":"stop"}],"usage":{"prompt_tokens":9,"completion_tokens":4,"total_tokens":13}}`))
	}))
	defer upstream.Close()

	cfg := config.Config{
		AllowUnauthenticatedLocal: true,
		OpenAIBaseURL:             upstream.URL,
		OpenAIApiKey:              "test-openai-key",
		ModelRoutes:               "chat-default=openai-chat:gpt-4o-mini",
	}
	cat, err := catalog.New(cfg)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	service := New(cfg, cat)

	req := newLoopbackRequest(http.MethodPost, "/v1/responses", bytes.NewBufferString(`{"model":"chat-default","instructions":"Keep it short","input":[{"role":"assistant","type":"function_call","call_id":"call_weather","name":"get_weather","arguments":"{\"city\":\"Shanghai\"}"},{"type":"function_call_output","call_id":"call_weather","output":"sunny"},{"role":"user","content":"hi"}],"tools":[{"type":"function","name":"get_weather","description":"Get weather","parameters":{"type":"object","properties":{"city":{"type":"string"}}}}],"max_output_tokens":64}`))
	rec := httptest.NewRecorder()

	service.Handle(rec, req, consts.ProtocolOpenAIResponses)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if gotModel != "gpt-4o-mini" {
		t.Fatalf("expected translated model, got %q", gotModel)
	}
	if len(gotMessages) != 4 {
		t.Fatalf("expected translated system+assistant-tool-tool-user messages, got %d", len(gotMessages))
	}
	if len(gotTools) != 1 {
		t.Fatalf("expected one translated chat tool, got %d", len(gotTools))
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["object"] != "response" {
		t.Fatalf("expected response object, got %#v", payload["object"])
	}
}

func TestHandleTranslatesAnthropicToResponses(t *testing.T) {
	var gotModel string
	var gotInstructions string
	var gotInput []interface{}
	var gotTools []interface{}
	var gotReasoning map[string]interface{}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode upstream body: %v", err)
		}
		gotModel, _ = payload["model"].(string)
		gotInstructions, _ = payload["instructions"].(string)
		gotInput, _ = payload["input"].([]interface{})
		gotTools, _ = payload["tools"].([]interface{})
		gotReasoning, _ = payload["reasoning"].(map[string]interface{})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"resp_456","status":"completed","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"hello from responses for anthropic"}]},{"type":"function_call","call_id":"tool_123","name":"lookup_docs","arguments":"{\"topic\":\"proxy\"}"}],"usage":{"input_tokens":15,"output_tokens":9,"total_tokens":24}}`))
	}))
	defer upstream.Close()

	cfg := config.Config{
		AllowUnauthenticatedLocal: true,
		OpenAIBaseURL:             upstream.URL,
		OpenAIApiKey:              "test-openai-key",
		ModelRoutes:               "anthropic-default=openai-responses:gpt-4.1-mini",
	}
	cat, err := catalog.New(cfg)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	service := New(cfg, cat)

	req := newLoopbackRequest(http.MethodPost, "/v1/messages", bytes.NewBufferString(`{"model":"anthropic-default","system":"Be concise","tools":[{"name":"lookup_docs","description":"Lookup docs","input_schema":{"type":"object","properties":{"topic":{"type":"string"}}}}],"messages":[{"role":"assistant","content":[{"type":"tool_use","id":"tool_prev","name":"lookup_docs","input":{"topic":"proxy"}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool_prev","content":"done"},{"type":"text","text":"hello"}]}],"max_tokens":64}`))
	rec := httptest.NewRecorder()

	service.Handle(rec, req, consts.ProtocolAnthropic)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if gotModel != "gpt-4.1-mini" {
		t.Fatalf("expected translated model, got %q", gotModel)
	}
	if gotInstructions != "Be concise" {
		t.Fatalf("expected instructions, got %q", gotInstructions)
	}
	if len(gotInput) != 3 {
		t.Fatalf("expected assistant function_call + tool output + user text inputs, got %d", len(gotInput))
	}
	if len(gotTools) != 1 {
		t.Fatalf("expected one translated responses tool, got %d", len(gotTools))
	}
	if gotReasoning["effort"] != defaultResponsesReasoningEffort {
		t.Fatalf("expected default reasoning effort %q, got %#v", defaultResponsesReasoningEffort, gotReasoning)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["type"] != "message" {
		t.Fatalf("expected anthropic message response, got %#v", payload["type"])
	}
	content, _ := payload["content"].([]interface{})
	if len(content) != 2 {
		t.Fatalf("expected text + tool_use anthropic content, got %d", len(content))
	}
}

func TestHandleStreamsAnthropicToResponsesAsAnthropicSSE(t *testing.T) {
	var gotStream bool
	var gotReasoning map[string]interface{}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode upstream body: %v", err)
		}
		gotStream, _ = payload["stream"].(bool)
		gotReasoning, _ = payload["reasoning"].(map[string]interface{})
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(strings.Join([]string{
			`data: {"type":"response.created","response":{"id":"resp_stream_123","object":"response","model":"gpt-4.1-mini","status":"in_progress","usage":{"input_tokens":12,"output_tokens":0}}}`,
			"",
			`data: {"type":"response.output_text.delta","response_id":"resp_stream_123","item_id":"msg_stream_1","output_index":0,"content_index":0,"delta":"你"}`,
			"",
			`data: {"type":"response.output_text.delta","response_id":"resp_stream_123","item_id":"msg_stream_1","output_index":0,"content_index":0,"delta":"好"}`,
			"",
			`data: {"type":"response.completed","response":{"id":"resp_stream_123","object":"response","model":"gpt-4.1-mini","status":"completed","output":[{"type":"message","id":"msg_stream_1","role":"assistant","content":[{"type":"output_text","text":"你好"}]}],"usage":{"input_tokens":12,"output_tokens":2,"total_tokens":14}}}`,
			"",
			`data: [DONE]`,
			"",
		}, "\n")))
	}))
	defer upstream.Close()

	cfg := config.Config{
		AllowUnauthenticatedLocal: true,
		OpenAIBaseURL:             upstream.URL,
		OpenAIApiKey:              "test-openai-key",
		ModelRoutes:               "anthropic-default=openai-responses:gpt-4.1-mini",
	}
	cat, err := catalog.New(cfg)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	service := New(cfg, cat)

	req := newLoopbackRequest(http.MethodPost, "/v1/messages", bytes.NewBufferString(`{"model":"anthropic-default","messages":[{"role":"user","content":[{"type":"text","text":"hello"}]}],"max_tokens":64,"stream":true}`))
	rec := httptest.NewRecorder()

	service.Handle(rec, req, consts.ProtocolAnthropic)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !gotStream {
		t.Fatalf("expected translated stream=true request")
	}
	if gotReasoning["effort"] != defaultResponsesReasoningEffort {
		t.Fatalf("expected default reasoning effort %q, got %#v", defaultResponsesReasoningEffort, gotReasoning)
	}
	if contentType := rec.Header().Get("Content-Type"); !strings.Contains(contentType, "text/event-stream") {
		t.Fatalf("expected text/event-stream content type, got %q", contentType)
	}
	body := rec.Body.String()
	for _, needle := range []string{
		"event: message_start",
		"event: content_block_start",
		`"type":"text_delta"`,
		`"text":"你"`,
		`"text":"好"`,
		`"stop_reason":"end_turn"`,
		`"output_tokens":2`,
		"event: message_stop",
	} {
		if !strings.Contains(body, needle) {
			t.Fatalf("expected stream body to contain %q, got %s", needle, body)
		}
	}
}

func TestHandleStreamsAnthropicToResponsesToolUseAsAnthropicSSE(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(strings.Join([]string{
			`data: {"type":"response.created","response":{"id":"resp_stream_tool","object":"response","model":"gpt-4.1-mini","status":"in_progress","usage":{"input_tokens":10,"output_tokens":0}}}`,
			"",
			`data: {"type":"response.output_item.added","response_id":"resp_stream_tool","output_index":0,"item":{"type":"function_call","id":"fc_1","call_id":"call_1","name":"lookup_docs","arguments":""}}`,
			"",
			`data: {"type":"response.function_call_arguments.delta","response_id":"resp_stream_tool","item_id":"fc_1","output_index":0,"delta":"{\"topic\":\"pro"}`,
			"",
			`data: {"type":"response.function_call_arguments.delta","response_id":"resp_stream_tool","item_id":"fc_1","output_index":0,"delta":"xy\"}"}`,
			"",
			`data: {"type":"response.function_call_arguments.done","response_id":"resp_stream_tool","item_id":"fc_1","output_index":0,"name":"lookup_docs","arguments":"{\"topic\":\"proxy\"}"}`,
			"",
			`data: {"type":"response.completed","response":{"id":"resp_stream_tool","object":"response","model":"gpt-4.1-mini","status":"completed","output":[{"type":"function_call","id":"fc_1","call_id":"call_1","name":"lookup_docs","arguments":"{\"topic\":\"proxy\"}"}],"usage":{"input_tokens":10,"output_tokens":4,"total_tokens":14}}}`,
			"",
			`data: [DONE]`,
			"",
		}, "\n")))
	}))
	defer upstream.Close()

	cfg := config.Config{
		AllowUnauthenticatedLocal: true,
		OpenAIBaseURL:             upstream.URL,
		OpenAIApiKey:              "test-openai-key",
		ModelRoutes:               "anthropic-default=openai-responses:gpt-4.1-mini",
	}
	cat, err := catalog.New(cfg)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	service := New(cfg, cat)

	req := newLoopbackRequest(http.MethodPost, "/v1/messages", bytes.NewBufferString(`{"model":"anthropic-default","messages":[{"role":"user","content":[{"type":"text","text":"call the tool"}]}],"max_tokens":64,"stream":true}`))
	rec := httptest.NewRecorder()

	service.Handle(rec, req, consts.ProtocolAnthropic)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, needle := range []string{
		`"type":"tool_use"`,
		`"name":"lookup_docs"`,
		`"type":"input_json_delta"`,
		`"stop_reason":"tool_use"`,
		"event: message_stop",
	} {
		if !strings.Contains(body, needle) {
			t.Fatalf("expected stream body to contain %q, got %s", needle, body)
		}
	}
}

func TestHandleForcesOnlyStreamResponsesForAnthropicNonStream(t *testing.T) {
	var gotStream bool

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode upstream body: %v", err)
		}
		gotStream, _ = payload["stream"].(bool)
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(strings.Join([]string{
			`data: {"type":"response.created","response":{"id":"resp_force_123","object":"response","model":"gpt-4.1-mini","status":"in_progress"}}`,
			"",
			`data: {"type":"response.output_item.done","output_index":0,"item":{"type":"message","id":"msg_force_1","role":"assistant","content":[{"type":"output_text","text":"hello from forced stream"}]}}`,
			"",
			`data: {"type":"response.output_text.done","output_index":0,"item_id":"msg_force_1","content_index":0,"text":"hello from forced stream"}`,
			"",
			`data: {"type":"response.completed","response":{"id":"resp_force_123","object":"response","model":"gpt-4.1-mini","status":"completed","output":[],"usage":{"input_tokens":10,"output_tokens":4,"total_tokens":14}}}`,
			"",
			`data: [DONE]`,
			"",
		}, "\n")))
	}))
	defer upstream.Close()

	cfg := config.Config{
		AllowUnauthenticatedLocal: true,
		OpenAIResponsesBaseURL:    upstream.URL,
		OpenAIResponsesAPIKey:     "test-openai-key",
		OpenAIResponsesOnlyStream: true,
		ModelRoutes:               "anthropic-default=openai-responses:gpt-4.1-mini",
	}
	cat, err := catalog.New(cfg)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	service := New(cfg, cat)

	req := newLoopbackRequest(http.MethodPost, "/v1/messages", bytes.NewBufferString(`{"model":"anthropic-default","messages":[{"role":"user","content":[{"type":"text","text":"hello"}]}],"max_tokens":64}`))
	rec := httptest.NewRecorder()

	service.Handle(rec, req, consts.ProtocolAnthropic)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !gotStream {
		t.Fatalf("expected upstream request stream=true when only_stream is enabled")
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	content, _ := payload["content"].([]interface{})
	if len(content) != 1 {
		t.Fatalf("expected one anthropic text block, got %d", len(content))
	}
	block, _ := content[0].(map[string]interface{})
	if block["text"] != "hello from forced stream" {
		t.Fatalf("expected forced stream text, got %#v", block["text"])
	}
}

func TestHandleForcesOnlyStreamResponsesForResponsesNonStream(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(strings.Join([]string{
			`data: {"type":"response.created","response":{"id":"resp_force_456","object":"response","model":"gpt-4.1-mini","status":"in_progress"}}`,
			"",
			`data: {"type":"response.output_item.done","output_index":0,"item":{"type":"message","id":"msg_force_2","role":"assistant","content":[{"type":"output_text","text":"hello passthrough"}]}}`,
			"",
			`data: {"type":"response.output_text.done","output_index":0,"item_id":"msg_force_2","content_index":0,"text":"hello passthrough"}`,
			"",
			`data: {"type":"response.completed","response":{"id":"resp_force_456","object":"response","model":"gpt-4.1-mini","status":"completed","output":[],"usage":{"input_tokens":8,"output_tokens":2,"total_tokens":10}}}`,
			"",
			`data: [DONE]`,
			"",
		}, "\n")))
	}))
	defer upstream.Close()

	cfg := config.Config{
		AllowUnauthenticatedLocal: true,
		OpenAIResponsesBaseURL:    upstream.URL,
		OpenAIResponsesAPIKey:     "test-openai-key",
		OpenAIResponsesOnlyStream: true,
		DefaultResponsesRoute:     "openai-responses:gpt-4.1-mini",
	}
	cat, err := catalog.New(cfg)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	service := New(cfg, cat)

	req := newLoopbackRequest(http.MethodPost, "/v1/responses", bytes.NewBufferString(`{"model":"gpt-4.1-mini","input":"hello"}`))
	rec := httptest.NewRecorder()

	service.Handle(rec, req, consts.ProtocolOpenAIResponses)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["output_text"] != "hello passthrough" {
		t.Fatalf("expected output_text fallback, got %#v", payload["output_text"])
	}
	output, _ := payload["output"].([]interface{})
	if len(output) != 1 {
		t.Fatalf("expected one synthesized output item, got %d", len(output))
	}
}

func TestHandleResponsesPassthroughAddsDefaultReasoning(t *testing.T) {
	var gotModel string
	var gotReasoning map[string]interface{}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode upstream body: %v", err)
		}
		gotModel, _ = payload["model"].(string)
		gotReasoning, _ = payload["reasoning"].(map[string]interface{})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"resp_123","object":"response","status":"completed","output_text":"ok","output":[]}`))
	}))
	defer upstream.Close()

	cfg := config.Config{
		AllowUnauthenticatedLocal: true,
		OpenAIBaseURL:             upstream.URL,
		OpenAIApiKey:              "test-openai-key",
		DefaultResponsesRoute:     "openai-responses:gpt-4.1-mini",
	}
	cat, err := catalog.New(cfg)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	service := New(cfg, cat)

	req := newLoopbackRequest(http.MethodPost, "/v1/responses", bytes.NewBufferString(`{"model":"gpt-4.1-mini","input":"hello"}`))
	rec := httptest.NewRecorder()

	service.Handle(rec, req, consts.ProtocolOpenAIResponses)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if gotModel != "gpt-4.1-mini" {
		t.Fatalf("expected model rewrite, got %q", gotModel)
	}
	if gotReasoning["effort"] != defaultResponsesReasoningEffort {
		t.Fatalf("expected default reasoning effort %q, got %#v", defaultResponsesReasoningEffort, gotReasoning)
	}
}

func TestHandleResponsesPassthroughPreservesExplicitReasoning(t *testing.T) {
	var gotReasoning map[string]interface{}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode upstream body: %v", err)
		}
		gotReasoning, _ = payload["reasoning"].(map[string]interface{})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"resp_123","object":"response","status":"completed","output_text":"ok","output":[]}`))
	}))
	defer upstream.Close()

	cfg := config.Config{
		AllowUnauthenticatedLocal: true,
		OpenAIBaseURL:             upstream.URL,
		OpenAIApiKey:              "test-openai-key",
		DefaultResponsesRoute:     "openai-responses:gpt-4.1-mini",
	}
	cat, err := catalog.New(cfg)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	service := New(cfg, cat)

	req := newLoopbackRequest(http.MethodPost, "/v1/responses", bytes.NewBufferString(`{"model":"gpt-4.1-mini","input":"hello","reasoning":{"effort":"high"}}`))
	rec := httptest.NewRecorder()

	service.Handle(rec, req, consts.ProtocolOpenAIResponses)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if gotReasoning["effort"] != "high" {
		t.Fatalf("expected explicit reasoning preserved, got %#v", gotReasoning)
	}
}

func TestHandleTranslatesResponsesOutputTextFallbackToAnthropic(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"resp_fallback","status":"completed","output":[],"output_text":"hello from top-level output text"}`))
	}))
	defer upstream.Close()

	cfg := config.Config{
		AllowUnauthenticatedLocal: true,
		OpenAIBaseURL:             upstream.URL,
		OpenAIApiKey:              "test-openai-key",
		ModelRoutes:               "anthropic-default=openai-responses:gpt-5.4",
	}
	cat, err := catalog.New(cfg)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	service := New(cfg, cat)

	req := newLoopbackRequest(http.MethodPost, "/v1/messages", bytes.NewBufferString(`{"model":"anthropic-default","messages":[{"role":"user","content":"hello"}],"max_tokens":64}`))
	rec := httptest.NewRecorder()

	service.Handle(rec, req, consts.ProtocolAnthropic)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	content, _ := payload["content"].([]interface{})
	if len(content) != 1 {
		t.Fatalf("expected one anthropic text block, got %d", len(content))
	}
	block, _ := content[0].(map[string]interface{})
	if block["text"] != "hello from top-level output text" {
		t.Fatalf("expected fallback output_text, got %#v", block["text"])
	}
}

func TestHandleTranslatesResponsesToAnthropic(t *testing.T) {
	var gotModel string
	var gotSystem string
	var gotMessages []interface{}
	var gotTools []interface{}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode upstream body: %v", err)
		}
		gotModel, _ = payload["model"].(string)
		gotSystem, _ = payload["system"].(string)
		gotMessages, _ = payload["messages"].([]interface{})
		gotTools, _ = payload["tools"].([]interface{})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"msg_789","type":"message","model":"claude-3","role":"assistant","content":[{"type":"text","text":"hello from anthropic"},{"type":"tool_use","id":"tool_lookup","name":"lookup_docs","input":{"topic":"proxy"}}],"stop_reason":"tool_use","usage":{"input_tokens":12,"output_tokens":6}}`))
	}))
	defer upstream.Close()

	cfg := config.Config{
		AllowUnauthenticatedLocal: true,
		AnthropicBaseURL:          upstream.URL,
		AnthropicAPIKey:           "test-anthropic-key",
		AnthropicVersion:          "2023-06-01",
		ModelRoutes:               "response-default=anthropic:claude-sonnet-4",
	}
	cat, err := catalog.New(cfg)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	service := New(cfg, cat)

	req := newLoopbackRequest(http.MethodPost, "/v1/responses", bytes.NewBufferString(`{"model":"response-default","instructions":"Be direct","input":[{"type":"function_call_output","call_id":"tool_prev","output":"done"},{"role":"user","content":"hello"}],"tools":[{"type":"function","name":"lookup_docs","description":"Lookup docs","parameters":{"type":"object","properties":{"topic":{"type":"string"}}}}],"max_output_tokens":32}`))
	rec := httptest.NewRecorder()

	service.Handle(rec, req, consts.ProtocolOpenAIResponses)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if gotModel != "claude-sonnet-4" {
		t.Fatalf("expected translated model, got %q", gotModel)
	}
	if gotSystem != "Be direct" {
		t.Fatalf("expected anthropic system, got %q", gotSystem)
	}
	if len(gotMessages) != 2 {
		t.Fatalf("expected tool_result + user anthropic messages, got %d", len(gotMessages))
	}
	if len(gotTools) != 1 {
		t.Fatalf("expected one translated anthropic tool, got %d", len(gotTools))
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["object"] != "response" {
		t.Fatalf("expected responses object, got %#v", payload["object"])
	}
	output, _ := payload["output"].([]interface{})
	if len(output) != 2 {
		t.Fatalf("expected message + function_call outputs, got %d", len(output))
	}
}

func TestHandleTranslatesAnthropicToChat(t *testing.T) {
	var gotModel string
	var gotMessages []interface{}
	var gotTools []interface{}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode upstream body: %v", err)
		}
		gotModel, _ = payload["model"].(string)
		gotMessages, _ = payload["messages"].([]interface{})
		gotTools, _ = payload["tools"].([]interface{})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"chatcmpl_456","choices":[{"index":0,"message":{"role":"assistant","content":"hello from chat","tool_calls":[{"id":"tool_lookup","type":"function","function":{"name":"lookup_docs","arguments":"{\"topic\":\"proxy\"}"}}]},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}}`))
	}))
	defer upstream.Close()

	cfg := config.Config{
		AllowUnauthenticatedLocal: true,
		OpenAIBaseURL:             upstream.URL,
		OpenAIApiKey:              "test-openai-key",
		ModelRoutes:               "anthropic-chat=openai-chat:gpt-4o-mini",
	}
	cat, err := catalog.New(cfg)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	service := New(cfg, cat)

	req := newLoopbackRequest(http.MethodPost, "/v1/messages", bytes.NewBufferString(`{"model":"anthropic-chat","system":"Be practical","tools":[{"name":"lookup_docs","description":"Lookup docs","input_schema":{"type":"object","properties":{"topic":{"type":"string"}}}}],"messages":[{"role":"user","content":[{"type":"text","text":"hello"}]}],"max_tokens":32}`))
	rec := httptest.NewRecorder()

	service.Handle(rec, req, consts.ProtocolAnthropic)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if gotModel != "gpt-4o-mini" {
		t.Fatalf("expected translated model, got %q", gotModel)
	}
	if len(gotMessages) != 2 {
		t.Fatalf("expected system + user chat messages, got %d", len(gotMessages))
	}
	if len(gotTools) != 1 {
		t.Fatalf("expected one translated chat tool, got %d", len(gotTools))
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["type"] != "message" {
		t.Fatalf("expected anthropic message response, got %#v", payload["type"])
	}
	content, _ := payload["content"].([]interface{})
	if len(content) != 2 {
		t.Fatalf("expected text + tool_use anthropic content, got %d", len(content))
	}
}

func TestHandleTranslatesChatToAnthropic(t *testing.T) {
	var gotModel string
	var gotSystem string
	var gotMessages []interface{}
	var gotTools []interface{}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode upstream body: %v", err)
		}
		gotModel, _ = payload["model"].(string)
		gotSystem, _ = payload["system"].(string)
		gotMessages, _ = payload["messages"].([]interface{})
		gotTools, _ = payload["tools"].([]interface{})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"msg_456","type":"message","model":"claude-3","role":"assistant","content":[{"type":"text","text":"hello from anthropic"},{"type":"tool_use","id":"tool_lookup","name":"lookup_docs","input":{"topic":"proxy"}}],"stop_reason":"tool_use","usage":{"input_tokens":13,"output_tokens":7}}`))
	}))
	defer upstream.Close()

	cfg := config.Config{
		AllowUnauthenticatedLocal: true,
		AnthropicBaseURL:          upstream.URL,
		AnthropicAPIKey:           "test-anthropic-key",
		AnthropicVersion:          "2023-06-01",
		ModelRoutes:               "chat-anthropic=anthropic:claude-sonnet-4",
	}
	cat, err := catalog.New(cfg)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	service := New(cfg, cat)

	req := newLoopbackRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{"model":"chat-anthropic","messages":[{"role":"system","content":"Be practical"},{"role":"assistant","content":null,"tool_calls":[{"id":"tool_prev","type":"function","function":{"name":"lookup_docs","arguments":"{\"topic\":\"proxy\"}"}}]},{"role":"tool","tool_call_id":"tool_prev","content":"done"},{"role":"user","content":"hello"}],"tools":[{"type":"function","function":{"name":"lookup_docs","description":"Lookup docs","parameters":{"type":"object","properties":{"topic":{"type":"string"}}}}}],"max_tokens":32}`))
	rec := httptest.NewRecorder()

	service.Handle(rec, req, consts.ProtocolOpenAIChat)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if gotModel != "claude-sonnet-4" {
		t.Fatalf("expected translated model, got %q", gotModel)
	}
	if gotSystem != "Be practical" {
		t.Fatalf("expected anthropic system, got %q", gotSystem)
	}
	if len(gotMessages) != 3 {
		t.Fatalf("expected assistant tool_use + tool_result + user anthropic messages, got %d", len(gotMessages))
	}
	if len(gotTools) != 1 {
		t.Fatalf("expected one translated anthropic tool, got %d", len(gotTools))
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	choices, _ := payload["choices"].([]interface{})
	if len(choices) != 1 {
		t.Fatalf("expected one choice, got %d", len(choices))
	}
	choice, _ := choices[0].(map[string]interface{})
	if choice["finish_reason"] != "tool_calls" {
		t.Fatalf("expected tool_calls finish reason, got %#v", choice["finish_reason"])
	}
}

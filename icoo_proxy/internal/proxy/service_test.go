package proxy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"icoo_proxy/internal/catalog"
	"icoo_proxy/internal/config"
)

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

	req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewBufferString(`{"model":"claude-sonnet","messages":[{"role":"user","content":"hello"}],"max_tokens":64}`))
	rec := httptest.NewRecorder()

	service.Handle(rec, req, catalog.ProtocolAnthropic)

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

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{"model":"assistant-default","messages":[{"role":"system","content":"You are helpful"},{"role":"user","content":"hi"}],"tools":[{"type":"function","function":{"name":"get_weather","description":"Get weather","parameters":{"type":"object","properties":{"city":{"type":"string"}},"required":["city"]}}}],"max_tokens":32}`))
	rec := httptest.NewRecorder()

	service.Handle(rec, req, catalog.ProtocolOpenAIChat)

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

	req := httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewBufferString(`{"model":"chat-default","instructions":"Keep it short","input":[{"role":"assistant","type":"function_call","call_id":"call_weather","name":"get_weather","arguments":"{\"city\":\"Shanghai\"}"},{"type":"function_call_output","call_id":"call_weather","output":"sunny"},{"role":"user","content":"hi"}],"tools":[{"type":"function","name":"get_weather","description":"Get weather","parameters":{"type":"object","properties":{"city":{"type":"string"}}}}],"max_output_tokens":64}`))
	rec := httptest.NewRecorder()

	service.Handle(rec, req, catalog.ProtocolOpenAIResponse)

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

	req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewBufferString(`{"model":"anthropic-default","system":"Be concise","tools":[{"name":"lookup_docs","description":"Lookup docs","input_schema":{"type":"object","properties":{"topic":{"type":"string"}}}}],"messages":[{"role":"assistant","content":[{"type":"tool_use","id":"tool_prev","name":"lookup_docs","input":{"topic":"proxy"}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool_prev","content":"done"},{"type":"text","text":"hello"}]}],"max_tokens":64}`))
	rec := httptest.NewRecorder()

	service.Handle(rec, req, catalog.ProtocolAnthropic)

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

	req := httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewBufferString(`{"model":"response-default","instructions":"Be direct","input":[{"type":"function_call_output","call_id":"tool_prev","output":"done"},{"role":"user","content":"hello"}],"tools":[{"type":"function","name":"lookup_docs","description":"Lookup docs","parameters":{"type":"object","properties":{"topic":{"type":"string"}}}}],"max_output_tokens":32}`))
	rec := httptest.NewRecorder()

	service.Handle(rec, req, catalog.ProtocolOpenAIResponse)

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

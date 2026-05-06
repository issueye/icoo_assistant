package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"icoo_assistant/internal/agent"
	"icoo_assistant/internal/llm"
	"icoo_assistant/internal/tools"
)

func TestParseServeRequest(t *testing.T) {
	addr, ok := parseServeRequest([]string{"serve", "127.0.0.1:9999"})
	if !ok || addr != "127.0.0.1:9999" {
		t.Fatalf("unexpected serve parse: ok=%v addr=%q", ok, addr)
	}
}

func TestServerHealthz(t *testing.T) {
	srv := newServer(&app{mode: "fake"})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"mode":"fake"`) {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}

func TestServerRun(t *testing.T) {
	client := &llm.FakeClient{Responses: []llm.Response{{StopReason: "end", Text: "done"}}}
	registry, err := tools.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	srv := newServer(&app{
		runner: &agent.Runner{
			Client:   client,
			Registry: registry,
			Config:   agent.Config{SystemPrompt: "test", MaxRounds: 2},
		},
		mode: "anthropic",
	})
	body, _ := json.Marshal(runRequest{Query: "hello"})
	req := httptest.NewRequest(http.MethodPost, "/v1/run", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	srv.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%q", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"output":"done"`) {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}

func TestServerREPLSession(t *testing.T) {
	client := &llm.FakeClient{Responses: []llm.Response{
		{StopReason: "end", Text: "记住了。"},
		{StopReason: "end", Text: "你之前说过。"},
	}}
	registry, err := tools.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	srv := newServer(&app{
		runner: &agent.Runner{
			Client:   client,
			Registry: registry,
			Config:   agent.Config{SystemPrompt: "test", MaxRounds: 2},
		},
		mode: "anthropic",
	})
	call := func(query string) string {
		body, _ := json.Marshal(replRequest{SessionID: "s1", Query: query})
		req := httptest.NewRequest(http.MethodPost, "/v1/repl", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		srv.routes().ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d body=%q", rec.Code, rec.Body.String())
		}
		return rec.Body.String()
	}
	first := call("我叫小明")
	second := call("我刚才说了什么？")
	if !strings.Contains(first, "记住了。") || !strings.Contains(second, "你之前说过。") {
		t.Fatalf("unexpected responses: first=%q second=%q", first, second)
	}
	if len(client.Snapshots) != 2 || !strings.Contains(client.Snapshots[1], "我叫小明") {
		t.Fatalf("expected persisted session history, got %#v", client.Snapshots)
	}
}

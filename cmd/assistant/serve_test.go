package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestServerDocs(t *testing.T) {
	srv := newServer(&app{mode: "fake"})
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()
	srv.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%q", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, needle := range []string{
		`icoo Assistant API Docs`,
		`/v1/openapi.json`,
		`Persistent assistant API`,
		`/v1/repl`,
		`Current mode: <code>fake</code>`,
	} {
		if !strings.Contains(body, needle) {
			t.Fatalf("expected %q in body, got %q", needle, body)
		}
	}
}

func TestServerOpenAPI(t *testing.T) {
	srv := newServer(&app{mode: "fake"})
	req := httptest.NewRequest(http.MethodGet, "/v1/openapi.json", nil)
	rec := httptest.NewRecorder()
	srv.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%q", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, needle := range []string{
		`"openapi":"3.0.3"`,
		`"/healthz"`,
		`"/v1/openapi.json"`,
		`"/v1/run"`,
		`"/v1/repl"`,
		`"/v1/sessions"`,
		`"/v1/sessions/{session_id}"`,
		`"summary":"Run a one-shot query"`,
		`"summary":"Delete a saved session"`,
	} {
		if !strings.Contains(body, needle) {
			t.Fatalf("expected %q in body, got %q", needle, body)
		}
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
	root := t.TempDir()
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
		workdir: root,
		mode:    "anthropic",
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
	if _, err := os.Stat(filepath.Join(root, ".icoo", "sessions", "s1.json")); err != nil {
		t.Fatalf("expected persisted session file, got %v", err)
	}
}

func TestServerREPLSessionReloadsFromDisk(t *testing.T) {
	root := t.TempDir()
	clientA := &llm.FakeClient{Responses: []llm.Response{{StopReason: "end", Text: "first"}}}
	registry, err := tools.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	srvA := newServer(&app{
		runner: &agent.Runner{
			Client:   clientA,
			Registry: registry,
			Config:   agent.Config{SystemPrompt: "test", MaxRounds: 2},
		},
		workdir: root,
		mode:    "anthropic",
	})
	bodyA, _ := json.Marshal(replRequest{SessionID: "persisted", Query: "remember me"})
	reqA := httptest.NewRequest(http.MethodPost, "/v1/repl", bytes.NewReader(bodyA))
	recA := httptest.NewRecorder()
	srvA.routes().ServeHTTP(recA, reqA)
	if recA.Code != http.StatusOK {
		t.Fatalf("unexpected first status: %d body=%q", recA.Code, recA.Body.String())
	}

	clientB := &llm.FakeClient{Responses: []llm.Response{{StopReason: "end", Text: "second"}}}
	srvB := newServer(&app{
		runner: &agent.Runner{
			Client:   clientB,
			Registry: registry,
			Config:   agent.Config{SystemPrompt: "test", MaxRounds: 2},
		},
		workdir: root,
		mode:    "anthropic",
	})
	bodyB, _ := json.Marshal(replRequest{SessionID: "persisted", Query: "what did I say?"})
	reqB := httptest.NewRequest(http.MethodPost, "/v1/repl", bytes.NewReader(bodyB))
	recB := httptest.NewRecorder()
	srvB.routes().ServeHTTP(recB, reqB)
	if recB.Code != http.StatusOK {
		t.Fatalf("unexpected second status: %d body=%q", recB.Code, recB.Body.String())
	}
	if len(clientB.Snapshots) == 0 || !strings.Contains(clientB.Snapshots[0], "remember me") {
		t.Fatalf("expected disk-restored session history, got %#v", clientB.Snapshots)
	}
}

func TestServerListsSessions(t *testing.T) {
	root := t.TempDir()
	store, err := newSessionStore(filepath.Join(root, ".icoo", "sessions"))
	if err != nil {
		t.Fatal(err)
	}
	store.Put("alpha", []llm.Message{{Role: "user", Content: "hello"}})
	store.Put("beta", []llm.Message{{Role: "user", Content: "one"}, {Role: "assistant", Content: "two"}})
	srv := newServerWithSessionDir(&app{mode: "anthropic", workdir: root}, filepath.Join(root, ".icoo", "sessions"))
	req := httptest.NewRequest(http.MethodGet, "/v1/sessions", nil)
	rec := httptest.NewRecorder()
	srv.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%q", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"session_id":"alpha"`) || !strings.Contains(body, `"session_id":"beta"`) {
		t.Fatalf("expected sessions in body, got %q", body)
	}
	if !strings.Contains(body, `"message_count":1`) || !strings.Contains(body, `"message_count":2`) {
		t.Fatalf("expected message counts in body, got %q", body)
	}
}

func TestServerGetsSessionByID(t *testing.T) {
	root := t.TempDir()
	store, err := newSessionStore(filepath.Join(root, ".icoo", "sessions"))
	if err != nil {
		t.Fatal(err)
	}
	store.Put("alpha", []llm.Message{{Role: "user", Content: "hello"}})
	srv := newServerWithSessionDir(&app{mode: "anthropic", workdir: root}, filepath.Join(root, ".icoo", "sessions"))
	req := httptest.NewRequest(http.MethodGet, "/v1/sessions/alpha", nil)
	rec := httptest.NewRecorder()
	srv.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%q", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"session_id":"alpha"`) || !strings.Contains(body, `"message_count":1`) || !strings.Contains(body, `"role":"user"`) {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestServerDeletesSessionByID(t *testing.T) {
	root := t.TempDir()
	store, err := newSessionStore(filepath.Join(root, ".icoo", "sessions"))
	if err != nil {
		t.Fatal(err)
	}
	store.Put("alpha", []llm.Message{{Role: "user", Content: "hello"}})
	srv := newServerWithSessionDir(&app{mode: "anthropic", workdir: root}, filepath.Join(root, ".icoo", "sessions"))
	req := httptest.NewRequest(http.MethodDelete, "/v1/sessions/alpha", nil)
	rec := httptest.NewRecorder()
	srv.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%q", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(root, ".icoo", "sessions", "alpha.json")); err == nil {
		t.Fatal("expected session file to be deleted")
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"icoo_assistant/internal/llm"
)

type sessionStore struct {
	dir string
	mu  sync.Mutex
}

type sessionSummary struct {
	SessionID    string `json:"session_id"`
	MessageCount int    `json:"message_count"`
	UpdatedAt    string `json:"updated_at"`
}

type sessionDetail struct {
	SessionID    string        `json:"session_id"`
	MessageCount int           `json:"message_count"`
	UpdatedAt    string        `json:"updated_at"`
	Messages     []llm.Message `json:"messages,omitempty"`
}

func newSessionStore(dir string) (*sessionStore, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, fmt.Errorf("session dir required")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &sessionStore{dir: dir}, nil
}

func (s *sessionStore) Get(id string) []llm.Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.pathForID(id))
	if err != nil {
		return nil
	}
	var history []llm.Message
	if err := json.Unmarshal(data, &history); err != nil || len(history) == 0 {
		return nil
	}
	copied := make([]llm.Message, len(history))
	copy(copied, history)
	return copied
}

func (s *sessionStore) Put(id string, history []llm.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	path := s.pathForID(id)
	if len(history) == 0 {
		_ = os.Remove(path)
		return
	}
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(path, append(data, '\n'), 0o644)
}

func (s *sessionStore) pathForID(id string) string {
	return filepath.Join(s.dir, sanitizeSessionID(id)+".json")
}

func sanitizeSessionID(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return "default"
	}
	re := regexp.MustCompile(`[^a-zA-Z0-9._-]+`)
	cleaned := re.ReplaceAllString(id, "_")
	cleaned = strings.Trim(cleaned, "._-")
	if cleaned == "" {
		return "default"
	}
	return cleaned
}

type server struct {
	app      *app
	sessions *sessionStore
}

type runRequest struct {
	Query string `json:"query"`
}

type replRequest struct {
	SessionID string `json:"session_id"`
	Query     string `json:"query"`
	Reset     bool   `json:"reset"`
}

type responseEnvelope struct {
	Output    string `json:"output,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	Mode      string `json:"mode,omitempty"`
	Error     string `json:"error,omitempty"`
}

type sessionsEnvelope struct {
	Mode     string           `json:"mode,omitempty"`
	Sessions []sessionSummary `json:"sessions,omitempty"`
	Error    string           `json:"error,omitempty"`
}

type sessionDetailEnvelope struct {
	Mode    string         `json:"mode,omitempty"`
	Session *sessionDetail `json:"session,omitempty"`
	Error   string         `json:"error,omitempty"`
}

func newServer(app *app) *server {
	return newServerWithSessionDir(app, defaultSessionDir(app))
}

func newServerWithSessionDir(app *app, sessionDir string) *server {
	store, err := newSessionStore(sessionDir)
	if err != nil {
		return &server{
			app:      app,
			sessions: &sessionStore{},
		}
	}
	return &server{
		app:      app,
		sessions: store,
	}
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleDocs)
	mux.HandleFunc("/docs", s.handleDocs)
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/v1/openapi.json", s.handleOpenAPI)
	mux.HandleFunc("/v1/run", s.handleRun)
	mux.HandleFunc("/v1/repl", s.handleREPL)
	mux.HandleFunc("/v1/sessions", s.handleSessions)
	mux.HandleFunc("/v1/sessions/", s.handleSessionByID)
	return mux
}

func (s *server) handleDocs(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/docs" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	mode := ""
	if s.app != nil {
		mode = s.app.mode
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>icoo Assistant API Docs</title>
  <style>
    :root { color-scheme: light; }
    body { margin: 0; font-family: "Segoe UI", "PingFang SC", sans-serif; background: linear-gradient(180deg, #f4f1e8 0%%, #ffffff 100%%); color: #18230f; }
    main { max-width: 920px; margin: 0 auto; padding: 48px 24px 64px; }
    h1 { margin: 0 0 12px; font-size: 40px; line-height: 1.1; }
    p { line-height: 1.6; }
    .hero { padding: 28px; border-radius: 20px; background: #dce7c9; box-shadow: 0 10px 30px rgba(24, 35, 15, 0.08); }
    .badge { display: inline-block; margin-bottom: 14px; padding: 6px 10px; border-radius: 999px; background: #18230f; color: #f4f1e8; font-size: 12px; letter-spacing: 0.08em; text-transform: uppercase; }
    .actions { display: flex; gap: 12px; flex-wrap: wrap; margin-top: 20px; }
    a.button { display: inline-block; padding: 10px 14px; border-radius: 10px; background: #255f38; color: #ffffff; text-decoration: none; }
    a.link { color: #255f38; text-decoration: none; }
    section { margin-top: 28px; }
    .card { padding: 20px; border-radius: 16px; background: #ffffff; box-shadow: 0 8px 24px rgba(24, 35, 15, 0.06); }
    table { width: 100%%; border-collapse: collapse; }
    th, td { padding: 12px 10px; text-align: left; border-bottom: 1px solid #dce7c9; vertical-align: top; }
    th { font-size: 13px; text-transform: uppercase; letter-spacing: 0.06em; color: #4f6f52; }
    code { font-family: Consolas, "Courier New", monospace; background: #f4f1e8; padding: 2px 6px; border-radius: 6px; }
  </style>
</head>
<body>
  <main>
    <div class="hero">
      <div class="badge">icoo local api</div>
      <h1>Persistent assistant API</h1>
      <p>Run one-shot tasks, keep session history across requests, and inspect local saved sessions. Current mode: <code>%s</code>.</p>
      <div class="actions">
        <a class="button" href="/v1/openapi.json">OpenAPI JSON</a>
        <a class="link" href="/healthz">Health Check</a>
      </div>
    </div>
    <section class="card">
      <h2>Endpoints</h2>
      <table>
        <thead>
          <tr><th>Method</th><th>Path</th><th>Purpose</th></tr>
        </thead>
        <tbody>
          <tr><td><code>GET</code></td><td><code>/healthz</code></td><td>Returns server health and current runtime mode.</td></tr>
          <tr><td><code>GET</code></td><td><code>/v1/openapi.json</code></td><td>Returns the machine-readable API schema.</td></tr>
          <tr><td><code>POST</code></td><td><code>/v1/run</code></td><td>Runs a one-shot query without session persistence.</td></tr>
          <tr><td><code>POST</code></td><td><code>/v1/repl</code></td><td>Runs a session turn with persisted conversation history.</td></tr>
          <tr><td><code>GET</code></td><td><code>/v1/sessions</code></td><td>Lists saved session summaries.</td></tr>
          <tr><td><code>GET</code></td><td><code>/v1/sessions/{session_id}</code></td><td>Fetches the full saved session detail.</td></tr>
          <tr><td><code>DELETE</code></td><td><code>/v1/sessions/{session_id}</code></td><td>Deletes a saved session from disk.</td></tr>
        </tbody>
      </table>
    </section>
  </main>
</body>
</html>`, mode)
}

func (s *server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, responseEnvelope{Mode: s.app.mode})
}

func (s *server) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeOpenAPIJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"error": "method not allowed",
		})
		return
	}
	writeOpenAPIJSON(w, http.StatusOK, map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "icoo Assistant API",
			"version":     Version,
			"description": "Local persistent API for icoo assistant sessions and one-shot runs.",
		},
		"paths": map[string]any{
			"/healthz": map[string]any{
				"get": map[string]any{
					"summary": "Health check",
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Server is healthy",
						},
					},
				},
			},
			"/v1/openapi.json": map[string]any{
				"get": map[string]any{
					"summary": "Get OpenAPI document",
					"responses": map[string]any{
						"200": map[string]any{
							"description": "OpenAPI schema",
						},
					},
				},
			},
			"/v1/run": map[string]any{
				"post": map[string]any{
					"summary": "Run a one-shot query",
					"requestBody": map[string]any{
						"required": true,
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{
									"type": "object",
									"required": []string{
										"query",
									},
									"properties": map[string]any{
										"query": map[string]any{
											"type": "string",
										},
									},
								},
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Execution result",
						},
					},
				},
			},
			"/v1/repl": map[string]any{
				"post": map[string]any{
					"summary": "Run a stateful session turn",
					"requestBody": map[string]any{
						"required": true,
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{
									"type": "object",
									"required": []string{
										"session_id",
									},
									"properties": map[string]any{
										"session_id": map[string]any{"type": "string"},
										"query":      map[string]any{"type": "string"},
										"reset":      map[string]any{"type": "boolean"},
									},
								},
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Session execution result",
						},
					},
				},
			},
			"/v1/sessions": map[string]any{
				"get": map[string]any{
					"summary": "List saved sessions",
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Saved session summaries",
						},
					},
				},
			},
			"/v1/sessions/{session_id}": map[string]any{
				"get": map[string]any{
					"summary": "Get a saved session",
					"parameters": []map[string]any{
						{
							"name":     "session_id",
							"in":       "path",
							"required": true,
							"schema": map[string]any{
								"type": "string",
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Saved session detail",
						},
					},
				},
				"delete": map[string]any{
					"summary": "Delete a saved session",
					"parameters": []map[string]any{
						{
							"name":     "session_id",
							"in":       "path",
							"required": true,
							"schema": map[string]any{
								"type": "string",
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Session deleted",
						},
					},
				},
			},
		},
	})
}

func (s *server) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, responseEnvelope{Error: "method not allowed"})
		return
	}
	var req runRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, responseEnvelope{Error: "invalid json"})
		return
	}
	query := strings.TrimSpace(req.Query)
	if query == "" {
		writeJSON(w, http.StatusBadRequest, responseEnvelope{Error: "query required"})
		return
	}
	output, err := s.app.execute(query)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, responseEnvelope{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, responseEnvelope{Output: output, Mode: s.app.mode})
}

func (s *server) handleREPL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, responseEnvelope{Error: "method not allowed"})
		return
	}
	var req replRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, responseEnvelope{Error: "invalid json"})
		return
	}
	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		writeJSON(w, http.StatusBadRequest, responseEnvelope{Error: "session_id required"})
		return
	}
	query := strings.TrimSpace(req.Query)
	if query == "" && !req.Reset {
		writeJSON(w, http.StatusBadRequest, responseEnvelope{Error: "query required"})
		return
	}
	if req.Reset {
		s.sessions.Put(sessionID, nil)
		writeJSON(w, http.StatusOK, responseEnvelope{SessionID: sessionID, Mode: s.app.mode})
		return
	}
	history := s.sessions.Get(sessionID)
	nextHistory, output, err := s.app.executeMessages(history, query)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, responseEnvelope{Error: err.Error(), SessionID: sessionID})
		return
	}
	s.sessions.Put(sessionID, nextHistory)
	writeJSON(w, http.StatusOK, responseEnvelope{
		Output:    output,
		SessionID: sessionID,
		Mode:      s.app.mode,
	})
}

func (s *server) handleSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeSessionsJSON(w, http.StatusMethodNotAllowed, sessionsEnvelope{Error: "method not allowed"})
		return
	}
	summaries := s.sessions.List()
	writeSessionsJSON(w, http.StatusOK, sessionsEnvelope{
		Mode:     s.app.mode,
		Sessions: summaries,
	})
}

func (s *server) handleSessionByID(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/v1/sessions/"))
	if sessionID == "" {
		writeSessionDetailJSON(w, http.StatusBadRequest, sessionDetailEnvelope{Error: "session_id required"})
		return
	}
	switch r.Method {
	case http.MethodGet:
		detail, ok := s.sessions.GetWithMeta(sessionID)
		if !ok {
			writeSessionDetailJSON(w, http.StatusNotFound, sessionDetailEnvelope{Error: "session not found"})
			return
		}
		writeSessionDetailJSON(w, http.StatusOK, sessionDetailEnvelope{
			Mode:    s.app.mode,
			Session: &detail,
		})
	case http.MethodDelete:
		if !s.sessions.Delete(sessionID) {
			writeSessionDetailJSON(w, http.StatusNotFound, sessionDetailEnvelope{Error: "session not found"})
			return
		}
		writeSessionDetailJSON(w, http.StatusOK, sessionDetailEnvelope{
			Mode: s.app.mode,
			Session: &sessionDetail{
				SessionID: sessionID,
			},
		})
	default:
		writeSessionDetailJSON(w, http.StatusMethodNotAllowed, sessionDetailEnvelope{Error: "method not allowed"})
	}
}

func writeJSON(w http.ResponseWriter, status int, payload responseEnvelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeSessionsJSON(w http.ResponseWriter, status int, payload sessionsEnvelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeSessionDetailJSON(w http.ResponseWriter, status int, payload sessionDetailEnvelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeOpenAPIJSON(w http.ResponseWriter, status int, payload map[string]any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func runServe(application *app, addr string) error {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		addr = "127.0.0.1:4317"
	}
	srv := newServer(application)
	fmt.Printf("icoo persistent server listening on http://%s\n", addr)
	return http.ListenAndServe(addr, srv.routes())
}

func defaultSessionDir(app *app) string {
	if app == nil || strings.TrimSpace(app.workdir) == "" {
		return filepath.Join(".icoo", "sessions")
	}
	return filepath.Join(app.workdir, ".icoo", "sessions")
}

func (s *sessionStore) List() []sessionSummary {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil
	}
	summaries := make([]sessionSummary, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(s.dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var history []llm.Message
		if err := json.Unmarshal(data, &history); err != nil {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		summaries = append(summaries, sessionSummary{
			SessionID:    strings.TrimSuffix(entry.Name(), ".json"),
			MessageCount: len(history),
			UpdatedAt:    info.ModTime().UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	return summaries
}

func (s *sessionStore) GetWithMeta(id string) (sessionDetail, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	path := s.pathForID(id)
	data, err := os.ReadFile(path)
	if err != nil {
		return sessionDetail{}, false
	}
	var history []llm.Message
	if err := json.Unmarshal(data, &history); err != nil {
		return sessionDetail{}, false
	}
	info, err := os.Stat(path)
	if err != nil {
		return sessionDetail{}, false
	}
	return sessionDetail{
		SessionID:    id,
		MessageCount: len(history),
		UpdatedAt:    info.ModTime().UTC().Format("2006-01-02T15:04:05Z"),
		Messages:     history,
	}, true
}

func (s *sessionStore) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	path := s.pathForID(id)
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return os.Remove(path) == nil
}

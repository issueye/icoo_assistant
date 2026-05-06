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
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/v1/run", s.handleRun)
	mux.HandleFunc("/v1/repl", s.handleREPL)
	mux.HandleFunc("/v1/sessions", s.handleSessions)
	return mux
}

func (s *server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, responseEnvelope{Mode: s.app.mode})
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

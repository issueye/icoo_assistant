package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"icoo_assistant/internal/llm"
)

type sessionStore struct {
	mu       sync.Mutex
	sessions map[string][]llm.Message
}

func newSessionStore() *sessionStore {
	return &sessionStore{sessions: map[string][]llm.Message{}}
}

func (s *sessionStore) Get(id string) []llm.Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	history := s.sessions[id]
	if len(history) == 0 {
		return nil
	}
	copied := make([]llm.Message, len(history))
	copy(copied, history)
	return copied
}

func (s *sessionStore) Put(id string, history []llm.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(history) == 0 {
		delete(s.sessions, id)
		return
	}
	copied := make([]llm.Message, len(history))
	copy(copied, history)
	s.sessions[id] = copied
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

func newServer(app *app) *server {
	return &server{
		app:      app,
		sessions: newSessionStore(),
	}
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/v1/run", s.handleRun)
	mux.HandleFunc("/v1/repl", s.handleREPL)
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

func writeJSON(w http.ResponseWriter, status int, payload responseEnvelope) {
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

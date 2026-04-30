package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	StatusActive   = "active"
	StatusClosed   = "closed"
	StatusArchived = "archived"
)

type Session struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	Status        string     `json:"status"`
	Tags          []string   `json:"tags,omitempty"`
	Summary       string     `json:"summary,omitempty"`
	RoundCount    int        `json:"roundCount"`
	MessageCount  int        `json:"messageCount"`
	MemoryIDs     []string   `json:"memoryIds,omitempty"`
	TranscriptIDs []string   `json:"transcriptIds,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	ClosedAt      *time.Time `json:"closedAt,omitempty"`
}

type CreateInput struct {
	ID    string
	Title string
	Tags  []string
}

type Manager struct {
	Dir string

	mu         sync.Mutex
	activeID   string
	now        func() time.Time
}

func DefaultDir(root string) string {
	return filepath.Join(root, ".sessions")
}

func NewManager(dir string) (*Manager, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil, fmt.Errorf("session dir required")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Manager{
		Dir: dir,
		now: time.Now,
	}, nil
}

func (m *Manager) Create(input CreateInput) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := m.now().UTC()
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = fmt.Sprintf("session-%d", now.Unix())
	}

	if err := validateID(id); err != nil {
		return Session{}, err
	}
	if _, err := os.Stat(m.pathForID(id)); err == nil {
		return Session{}, fmt.Errorf("session %s already exists", id)
	} else if !os.IsNotExist(err) {
		return Session{}, err
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = fmt.Sprintf("Session %s", now.Format("2006-01-02 15:04"))
	}

	tags := make([]string, 0, len(input.Tags))
	for _, tag := range input.Tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	session := Session{
		ID:        id,
		Title:     title,
		Status:    StatusActive,
		Tags:      tags,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := m.writeSessionLocked(session); err != nil {
		return Session{}, err
	}
	m.activeID = id
	return session, nil
}

func (m *Manager) EnsureActive() (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.activeID != "" {
		session, err := m.readSessionLocked(m.activeID)
		if err == nil && session.Status == StatusActive {
			return session, nil
		}
	}

	sessions, err := m.listLocked(StatusActive)
	if err != nil {
		return Session{}, err
	}
	if len(sessions) > 0 {
		m.activeID = sessions[0].ID
		return sessions[0], nil
	}

	now := m.now().UTC()
	session := Session{
		ID:        fmt.Sprintf("session-%d", now.Unix()),
		Title:     fmt.Sprintf("Session %s", now.Format("2006-01-02 15:04")),
		Status:    StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := m.writeSessionLocked(session); err != nil {
		return Session{}, err
	}
	m.activeID = session.ID
	return session, nil
}

func (m *Manager) Close(id string) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, err := m.readSessionLocked(id)
	if err != nil {
		return Session{}, err
	}
	if session.Status != StatusActive {
		return Session{}, fmt.Errorf("session %s is not active (status=%s)", id, session.Status)
	}
	now := m.now().UTC()
	session.Status = StatusClosed
	session.ClosedAt = &now
	session.UpdatedAt = now
	if err := m.writeSessionLocked(session); err != nil {
		return Session{}, err
	}
	if m.activeID == id {
		m.activeID = ""
	}
	return session, nil
}

func (m *Manager) Switch(id string) (Session, Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	target, err := m.readSessionLocked(id)
	if err != nil {
		return Session{}, Session{}, err
	}

	var previous Session
	if m.activeID != "" && m.activeID != id {
		prev, err := m.readSessionLocked(m.activeID)
		if err == nil && prev.Status == StatusActive {
			now := m.now().UTC()
			prev.Status = StatusClosed
			prev.ClosedAt = &now
			prev.UpdatedAt = now
			if err := m.writeSessionLocked(prev); err != nil {
				return Session{}, Session{}, err
			}
			previous = prev
		}
	}

	if target.Status != StatusActive && target.Status != StatusClosed {
		return Session{}, Session{}, fmt.Errorf("cannot switch to session with status %s", target.Status)
	}
	if target.Status == StatusClosed {
		target.Status = StatusActive
		target.ClosedAt = nil
		target.UpdatedAt = m.now().UTC()
	}

	m.activeID = id
	if err := m.writeSessionLocked(target); err != nil {
		return Session{}, Session{}, err
	}
	return target, previous, nil
}

func (m *Manager) GetActive() (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.activeID == "" {
		return Session{}, fmt.Errorf("no active session")
	}
	return m.readSessionLocked(m.activeID)
}

func (m *Manager) Get(id string) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.readSessionLocked(id)
}

func (m *Manager) List(status string) ([]Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.listLocked(status)
}

func (m *Manager) Archive(id string) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, err := m.readSessionLocked(id)
	if err != nil {
		return Session{}, err
	}
	session.Status = StatusArchived
	session.UpdatedAt = m.now().UTC()
	if err := m.writeSessionLocked(session); err != nil {
		return Session{}, err
	}
	if m.activeID == id {
		m.activeID = ""
	}
	return session, nil
}

func (m *Manager) UpdateStats(id string, roundCount, messageCount int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, err := m.readSessionLocked(id)
	if err != nil {
		return err
	}
	session.RoundCount = roundCount
	session.MessageCount = messageCount
	session.UpdatedAt = m.now().UTC()
	return m.writeSessionLocked(session)
}

func (m *Manager) UpdateSummary(id, summary string, memoryIDs []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, err := m.readSessionLocked(id)
	if err != nil {
		return err
	}
	session.Summary = summary
	if memoryIDs != nil {
		session.MemoryIDs = memoryIDs
	}
	session.UpdatedAt = m.now().UTC()
	return m.writeSessionLocked(session)
}

func (m *Manager) History(limit int) ([]Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	all, err := m.listLocked("")
	if err != nil {
		return nil, err
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})
	if limit > 0 && len(all) > limit {
		return all[:limit], nil
	}
	return all, nil
}

func (m *Manager) RecordTranscript(id, runID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, err := m.readSessionLocked(id)
	if err != nil {
		return err
	}
	for _, existing := range session.TranscriptIDs {
		if existing == runID {
			return nil
		}
	}
	session.TranscriptIDs = append(session.TranscriptIDs, runID)
	session.UpdatedAt = m.now().UTC()
	return m.writeSessionLocked(session)
}

func (m *Manager) readSessionLocked(id string) (Session, error) {
	if err := validateID(strings.TrimSpace(id)); err != nil {
		return Session{}, err
	}
	data, err := os.ReadFile(m.pathForID(id))
	if err != nil {
		return Session{}, err
	}
	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return Session{}, err
	}
	return session, nil
}

func (m *Manager) writeSessionLocked(session Session) error {
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.pathForID(session.ID), append(data, '\n'), 0o644)
}

func (m *Manager) listLocked(status string) ([]Session, error) {
	entries, err := os.ReadDir(m.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	sessions := make([]Session, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "session_") || !strings.HasSuffix(name, ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(m.Dir, name))
		if err != nil {
			continue
		}
		var session Session
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}
		if status != "" && session.Status != status {
			continue
		}
		sessions = append(sessions, session)
	}
	sort.Slice(sessions, func(i, j int) bool {
		if sessions[i].CreatedAt.Equal(sessions[j].CreatedAt) {
			return sessions[i].ID < sessions[j].ID
		}
		return sessions[i].CreatedAt.After(sessions[j].CreatedAt)
	})
	return sessions, nil
}

func (m *Manager) pathForID(id string) string {
	return filepath.Join(m.Dir, fmt.Sprintf("session_%s.json", id))
}

func validateID(id string) error {
	if id == "" {
		return fmt.Errorf("id required")
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-', r == '_':
		default:
			return fmt.Errorf("invalid session id %q", id)
		}
	}
	return nil
}

func (m *Manager) GetActiveID() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.activeID
}

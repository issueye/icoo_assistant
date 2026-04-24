package team

import (
	"bufio"
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
	StatusIdle    = "idle"
	StatusBusy    = "busy"
	StatusOffline = "offline"
)

type Config struct {
	LeadID    string    `json:"leadId"`
	Mission   string    `json:"mission,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ConfigUpdateInput struct {
	LeadID  string
	Mission string
}

type Teammate struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	Model     string    `json:"model,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateInput struct {
	ID     string
	Role   string
	Status string
	Model  string
}

type Message struct {
	ID        string    `json:"id"`
	FromID    string    `json:"fromId"`
	ToID      string    `json:"toId"`
	Kind      string    `json:"kind"`
	Body      string    `json:"body"`
	RequestID string    `json:"requestId,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type SendMessageInput struct {
	ID        string
	FromID    string
	ToID      string
	Kind      string
	Body      string
	RequestID string
}

type ReplyInput struct {
	ID        string
	FromID    string
	RequestID string
	Body      string
}

type Manager struct {
	Dir         string
	RegistryDir string
	InboxDir    string
	RequestsDir string

	mu  sync.Mutex
	now func() time.Time
}

func DefaultDir(root string) string {
	return filepath.Join(root, ".team")
}

func NewManager(dir string) (*Manager, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil, fmt.Errorf("team dir required")
	}
	registryDir := filepath.Join(dir, "teammates")
	inboxDir := filepath.Join(dir, "inbox")
	requestsDir := filepath.Join(dir, "requests")
	if err := os.MkdirAll(registryDir, 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(inboxDir, 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(requestsDir, 0o755); err != nil {
		return nil, err
	}
	manager := &Manager{
		Dir:         dir,
		RegistryDir: registryDir,
		InboxDir:    inboxDir,
		RequestsDir: requestsDir,
		now:         time.Now,
	}
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if _, err := manager.readConfigLocked(); err == nil {
		return manager, nil
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if err := manager.writeConfigLocked(defaultConfig(manager.now().UTC())); err != nil {
		return nil, err
	}
	return manager, nil
}

func (m *Manager) GetConfig() (Config, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.readConfigLocked()
}

func (m *Manager) UpdateConfig(input ConfigUpdateInput) (Config, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	current, err := m.readConfigLocked()
	if err != nil {
		return Config{}, err
	}
	updated, err := m.normalizeConfig(Config{
		LeadID:    fallbackTrimmed(input.LeadID, current.LeadID),
		Mission:   strings.TrimSpace(input.Mission),
		CreatedAt: current.CreatedAt,
	})
	if err != nil {
		return Config{}, err
	}
	if err := m.writeConfigLocked(updated); err != nil {
		return Config{}, err
	}
	return updated, nil
}

func (m *Manager) Create(input CreateInput) (Teammate, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, err := m.buildCreateTeammate(input)
	if err != nil {
		return Teammate{}, err
	}
	if _, err := os.Stat(m.pathForID(item.ID)); err == nil {
		return Teammate{}, fmt.Errorf("teammate %s already exists", item.ID)
	} else if !os.IsNotExist(err) {
		return Teammate{}, err
	}
	if err := m.writeTeammateLocked(item); err != nil {
		return Teammate{}, err
	}
	return item, nil
}

func (m *Manager) Get(id string) (Teammate, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.readTeammateLocked(id)
}

func (m *Manager) List() ([]Teammate, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.listLocked()
}

func (m *Manager) Update(item Teammate) (Teammate, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	current, err := m.readTeammateLocked(item.ID)
	if err != nil {
		return Teammate{}, err
	}
	updated, err := m.normalizeTeammate(Teammate{
		ID:        current.ID,
		Role:      fallbackTrimmed(item.Role, current.Role),
		Status:    fallbackTrimmed(item.Status, current.Status),
		Model:     fallbackTrimmed(item.Model, current.Model),
		CreatedAt: current.CreatedAt,
	})
	if err != nil {
		return Teammate{}, err
	}
	if err := m.writeTeammateLocked(updated); err != nil {
		return Teammate{}, err
	}
	return updated, nil
}

func (m *Manager) SendMessage(input SendMessageInput) (Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	msg, err := m.buildMessage(input)
	if err != nil {
		return Message{}, err
	}
	if err := m.ensureParticipantLocked(msg.FromID); err != nil {
		return Message{}, err
	}
	if err := m.ensureParticipantLocked(msg.ToID); err != nil {
		return Message{}, err
	}
	if err := m.appendMessageLocked(msg); err != nil {
		return Message{}, err
	}
	if err := m.syncRequestFromMessageLocked(msg); err != nil {
		return Message{}, err
	}
	return msg, nil
}

func (m *Manager) ListInbox(recipientID string, limit int) ([]Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	recipientID, err := normalizeID(recipientID)
	if err != nil {
		return nil, err
	}
	if err := m.ensureParticipantLocked(recipientID); err != nil {
		return nil, err
	}
	items, err := m.readInboxLocked(recipientID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || len(items) <= limit {
		return items, nil
	}
	return items[len(items)-limit:], nil
}

func (m *Manager) ReplyToRequest(input ReplyInput) (Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	fromID, err := normalizeID(strings.TrimSpace(input.FromID))
	if err != nil {
		return Message{}, fmt.Errorf("invalid from id: %w", err)
	}
	requestID := strings.TrimSpace(input.RequestID)
	if requestID == "" {
		return Message{}, fmt.Errorf("request id required")
	}
	if err := m.ensureParticipantLocked(fromID); err != nil {
		return Message{}, err
	}
	thread, err := m.listThreadLocked(requestID)
	if err != nil {
		return Message{}, err
	}
	if len(thread) == 0 {
		return Message{}, fmt.Errorf("request %s not found", requestID)
	}
	root := thread[0]
	if root.Kind != "request" {
		return Message{}, fmt.Errorf("request %s does not start with a request message", requestID)
	}
	if root.ToID != fromID {
		return Message{}, fmt.Errorf("request %s is not addressed to %s", requestID, fromID)
	}
	msg, err := m.buildMessage(SendMessageInput{
		ID:        input.ID,
		FromID:    fromID,
		ToID:      root.FromID,
		Kind:      "response",
		Body:      input.Body,
		RequestID: requestID,
	})
	if err != nil {
		return Message{}, err
	}
	if err := m.appendMessageLocked(msg); err != nil {
		return Message{}, err
	}
	if err := m.markRequestRespondedLocked(root, msg); err != nil {
		return Message{}, err
	}
	return msg, nil
}

func (m *Manager) ListThread(requestID string, limit int) ([]Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return nil, fmt.Errorf("request id required")
	}
	items, err := m.listThreadLocked(requestID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || len(items) <= limit {
		return items, nil
	}
	return items[len(items)-limit:], nil
}

func defaultConfig(now time.Time) Config {
	return Config{
		LeadID:    "lead",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (m *Manager) buildCreateTeammate(input CreateInput) (Teammate, error) {
	now := m.now().UTC()
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = fmt.Sprintf("mate-%d", now.UnixNano())
	}
	return m.normalizeTeammate(Teammate{
		ID:        id,
		Role:      strings.TrimSpace(input.Role),
		Status:    strings.TrimSpace(input.Status),
		Model:     strings.TrimSpace(input.Model),
		CreatedAt: now,
	})
}

func (m *Manager) normalizeConfig(cfg Config) (Config, error) {
	leadID, err := normalizeID(fallbackTrimmed(cfg.LeadID, "lead"))
	if err != nil {
		return Config{}, err
	}
	now := m.now().UTC()
	createdAt := cfg.CreatedAt.UTC()
	if createdAt.IsZero() {
		createdAt = now
	}
	return Config{
		LeadID:    leadID,
		Mission:   strings.TrimSpace(cfg.Mission),
		CreatedAt: createdAt,
		UpdatedAt: now,
	}, nil
}

func (m *Manager) normalizeTeammate(item Teammate) (Teammate, error) {
	id, err := normalizeID(item.ID)
	if err != nil {
		return Teammate{}, err
	}
	role := strings.TrimSpace(item.Role)
	if role == "" {
		role = "generalist"
	}
	status, err := normalizeStatus(item.Status)
	if err != nil {
		return Teammate{}, err
	}
	now := m.now().UTC()
	createdAt := item.CreatedAt.UTC()
	if createdAt.IsZero() {
		createdAt = now
	}
	return Teammate{
		ID:        id,
		Role:      role,
		Status:    status,
		Model:     strings.TrimSpace(item.Model),
		CreatedAt: createdAt,
		UpdatedAt: now,
	}, nil
}

func normalizeID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("id required")
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-', r == '_':
		default:
			return "", fmt.Errorf("invalid teammate id %q", id)
		}
	}
	return id, nil
}

func normalizeStatus(status string) (string, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "", StatusIdle:
		return StatusIdle, nil
	case StatusBusy, StatusOffline:
		return status, nil
	default:
		return "", fmt.Errorf("invalid teammate status %q", status)
	}
}

func fallbackTrimmed(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return strings.TrimSpace(fallback)
	}
	return value
}

func (m *Manager) buildMessage(input SendMessageInput) (Message, error) {
	now := m.now().UTC()
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = fmt.Sprintf("msg-%d", now.UnixNano())
	}
	id, err := normalizeID(id)
	if err != nil {
		return Message{}, err
	}
	fromID, err := normalizeID(strings.TrimSpace(input.FromID))
	if err != nil {
		return Message{}, fmt.Errorf("invalid from id: %w", err)
	}
	toID, err := normalizeID(strings.TrimSpace(input.ToID))
	if err != nil {
		return Message{}, fmt.Errorf("invalid to id: %w", err)
	}
	kind := strings.ToLower(strings.TrimSpace(input.Kind))
	if kind == "" {
		kind = "note"
	}
	body := strings.TrimSpace(input.Body)
	if body == "" {
		return Message{}, fmt.Errorf("body required")
	}
	return Message{
		ID:        id,
		FromID:    fromID,
		ToID:      toID,
		Kind:      kind,
		Body:      body,
		RequestID: strings.TrimSpace(input.RequestID),
		CreatedAt: now,
	}, nil
}

func (m *Manager) ensureParticipantLocked(id string) error {
	cfg, err := m.readConfigLocked()
	if err != nil {
		return err
	}
	if id == cfg.LeadID {
		return nil
	}
	if _, err := m.readTeammateLocked(id); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("unknown teammate %s", id)
		}
		return err
	}
	return nil
}

func (m *Manager) listLocked() ([]Teammate, error) {
	entries, err := os.ReadDir(m.RegistryDir)
	if err != nil {
		return nil, err
	}
	items := make([]Teammate, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "teammate_") || !strings.HasSuffix(name, ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(m.RegistryDir, name))
		if err != nil {
			return nil, err
		}
		var item Teammate
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].ID < items[j].ID
		}
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})
	return items, nil
}

func (m *Manager) readInboxLocked(recipientID string) ([]Message, error) {
	path := m.inboxPath(recipientID)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	items := make([]Message, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var item Message
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (m *Manager) listThreadLocked(requestID string) ([]Message, error) {
	all, err := m.listAllMessagesLocked()
	if err != nil {
		return nil, err
	}
	filtered := make([]Message, 0)
	for _, item := range all {
		if item.RequestID == requestID {
			filtered = append(filtered, item)
		}
	}
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].CreatedAt.Equal(filtered[j].CreatedAt) {
			return filtered[i].ID < filtered[j].ID
		}
		return filtered[i].CreatedAt.Before(filtered[j].CreatedAt)
	})
	return filtered, nil
}

func (m *Manager) listAllMessagesLocked() ([]Message, error) {
	entries, err := os.ReadDir(m.InboxDir)
	if err != nil {
		return nil, err
	}
	items := make([]Message, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(m.InboxDir, entry.Name()))
		if err != nil {
			return nil, err
		}
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			var item Message
			if err := json.Unmarshal([]byte(line), &item); err != nil {
				return nil, err
			}
			items = append(items, item)
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].ID < items[j].ID
		}
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})
	return items, nil
}

func (m *Manager) readConfigLocked() (Config, error) {
	data, err := os.ReadFile(m.configPath())
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (m *Manager) readTeammateLocked(id string) (Teammate, error) {
	id, err := normalizeID(id)
	if err != nil {
		return Teammate{}, err
	}
	data, err := os.ReadFile(m.pathForID(id))
	if err != nil {
		return Teammate{}, err
	}
	var item Teammate
	if err := json.Unmarshal(data, &item); err != nil {
		return Teammate{}, err
	}
	return item, nil
}

func (m *Manager) appendMessageLocked(item Message) error {
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(m.inboxPath(item.ToID), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write(append(data, '\n')); err != nil {
		return err
	}
	return nil
}

func (m *Manager) writeConfigLocked(cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.configPath(), append(data, '\n'), 0o644)
}

func (m *Manager) writeTeammateLocked(item Teammate) error {
	data, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.pathForID(item.ID), append(data, '\n'), 0o644)
}

func (m *Manager) configPath() string {
	return filepath.Join(m.Dir, "config.json")
}

func (m *Manager) pathForID(id string) string {
	return filepath.Join(m.RegistryDir, fmt.Sprintf("teammate_%s.json", id))
}

func (m *Manager) inboxPath(recipientID string) string {
	return filepath.Join(m.InboxDir, fmt.Sprintf("%s.jsonl", recipientID))
}

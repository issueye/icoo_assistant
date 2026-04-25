package run

import (
	"fmt"
	"icoo_gateway/internal/storage"
	"sort"
	"strings"
	"sync"
	"time"
)

type Run struct {
	ID               string     `json:"id"`
	ConversationID   string     `json:"conversation_id"`
	TriggerType      string     `json:"trigger_type"`
	TriggerMessageID string     `json:"trigger_message_id,omitempty"`
	Status           string     `json:"status"`
	StartedAt        time.Time  `json:"started_at"`
	FinishedAt       *time.Time `json:"finished_at,omitempty"`
	Summary          string     `json:"summary,omitempty"`
	ErrorMessage     string     `json:"error_message,omitempty"`
}

type CreateInput struct {
	ConversationID   string
	TriggerType      string
	TriggerMessageID string
	Status           string
	Summary          string
	ErrorMessage     string
}

type CompleteInput struct {
	Status       string
	Summary      string
	ErrorMessage string
}

type Store interface {
	storage.Creator[Run, CreateInput]
	storage.ListerByParent[Run]
	Completer
}

type Completer interface {
	Complete(id string, input CompleteInput) (Run, error)
}

type Service struct {
	mu      sync.RWMutex
	nextID  int
	records map[string]Run
	byConv  map[string][]string
	now     func() time.Time
}

var _ storage.Creator[Run, CreateInput] = (*Service)(nil)
var _ storage.ListerByParent[Run] = (*Service)(nil)
var _ Completer = (*Service)(nil)
var _ Store = (*Service)(nil)

func NewService() *Service {
	return &Service{
		records: make(map[string]Run),
		byConv:  make(map[string][]string),
		now:     time.Now,
	}
}

func (s *Service) Create(input CreateInput) (Run, error) {
	conversationID := strings.TrimSpace(input.ConversationID)
	if conversationID == "" {
		return Run{}, fmt.Errorf("conversation_id required")
	}
	triggerType := strings.TrimSpace(input.TriggerType)
	if triggerType == "" {
		triggerType = "message"
	}
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = "running"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	now := s.now().UTC()
	record := Run{
		ID:               fmt.Sprintf("run-%d", s.nextID),
		ConversationID:   conversationID,
		TriggerType:      triggerType,
		TriggerMessageID: strings.TrimSpace(input.TriggerMessageID),
		Status:           status,
		StartedAt:        now,
		Summary:          strings.TrimSpace(input.Summary),
		ErrorMessage:     strings.TrimSpace(input.ErrorMessage),
	}
	if record.Status != "running" {
		finished := now
		record.FinishedAt = &finished
	}
	s.records[record.ID] = record
	s.byConv[conversationID] = append(s.byConv[conversationID], record.ID)
	return record, nil
}

func (s *Service) Complete(id string, input CompleteInput) (Run, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[strings.TrimSpace(id)]
	if !ok {
		return Run{}, fmt.Errorf("run not found")
	}
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = "completed"
	}
	now := s.now().UTC()
	record.Status = status
	record.Summary = strings.TrimSpace(input.Summary)
	record.ErrorMessage = strings.TrimSpace(input.ErrorMessage)
	record.FinishedAt = &now
	s.records[record.ID] = record
	return record, nil
}

func (s *Service) ListByConversation(conversationID string) []Run {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := append([]string(nil), s.byConv[strings.TrimSpace(conversationID)]...)
	items := make([]Run, 0, len(ids))
	for _, id := range ids {
		if record, ok := s.records[id]; ok {
			items = append(items, record)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items
}

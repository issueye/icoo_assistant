package audit

import (
	"fmt"
	"icoo_gateway/internal/storage"
	"sort"
	"strings"
	"sync"
	"time"
)

type Event struct {
	ID           string      `json:"id"`
	ResourceType string      `json:"resource_type"`
	ResourceID   string      `json:"resource_id"`
	EventName    string      `json:"event_name"`
	Operator     string      `json:"operator"`
	Payload      interface{} `json:"payload,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
}

type RecordInput struct {
	ResourceType string
	ResourceID   string
	EventName    string
	Operator     string
	Payload      interface{}
}

type Store interface {
	storage.Recorder[Event, RecordInput]
	storage.Reader[Event]
}

type Service struct {
	mu      sync.RWMutex
	nextID  int
	records map[string]Event
	now     func() time.Time
}

var _ storage.Recorder[Event, RecordInput] = (*Service)(nil)
var _ storage.Reader[Event] = (*Service)(nil)
var _ Store = (*Service)(nil)

func NewService() *Service {
	return &Service{
		records: make(map[string]Event),
		now:     time.Now,
	}
}

func (s *Service) Record(input RecordInput) Event {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	now := s.now().UTC()
	record := Event{
		ID:           fmt.Sprintf("audit-%d", s.nextID),
		ResourceType: strings.TrimSpace(input.ResourceType),
		ResourceID:   strings.TrimSpace(input.ResourceID),
		EventName:    strings.TrimSpace(input.EventName),
		Operator:     strings.TrimSpace(input.Operator),
		Payload:      input.Payload,
		CreatedAt:    now,
	}
	if record.Operator == "" {
		record.Operator = "system"
	}
	s.records[record.ID] = record
	return record
}

func (s *Service) Get(id string) (Event, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.records[strings.TrimSpace(id)]
	return record, ok
}

func (s *Service) List() []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Event, 0, len(s.records))
	for _, record := range s.records {
		items = append(items, record)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items
}

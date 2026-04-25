package agentinstance

import (
	"fmt"
	"icoo_gateway/internal/storage"
	"sort"
	"strings"
	"sync"
	"time"
)

type Instance struct {
	ID              string    `json:"id"`
	ProfileID       string    `json:"profile_id,omitempty"`
	DisplayName     string    `json:"display_name"`
	RuntimeType     string    `json:"runtime_type"`
	RuntimeEndpoint string    `json:"runtime_endpoint,omitempty"`
	Status          string    `json:"status"`
	LastHeartbeatAt time.Time `json:"last_heartbeat_at,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateInput struct {
	ProfileID       string `json:"profile_id"`
	DisplayName     string `json:"display_name"`
	RuntimeType     string `json:"runtime_type"`
	RuntimeEndpoint string `json:"runtime_endpoint"`
	Status          string `json:"status"`
}

type Service struct {
	mu      sync.RWMutex
	nextID  int
	records map[string]Instance
	now     func() time.Time
}

var _ storage.Creator[Instance, CreateInput] = (*Service)(nil)
var _ storage.Reader[Instance] = (*Service)(nil)
var _ storage.Heartbeater[Instance] = (*Service)(nil)
var _ storage.Disabler[Instance] = (*Service)(nil)

func NewService() *Service {
	return &Service{
		records: make(map[string]Instance),
		now:     time.Now,
	}
}

func (s *Service) Create(input CreateInput) (Instance, error) {
	displayName := strings.TrimSpace(input.DisplayName)
	if displayName == "" {
		return Instance{}, fmt.Errorf("display_name required")
	}
	runtimeType := strings.TrimSpace(input.RuntimeType)
	if runtimeType == "" {
		runtimeType = "local"
	}
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = "idle"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	now := s.now().UTC()
	record := Instance{
		ID:              fmt.Sprintf("agent-instance-%d", s.nextID),
		ProfileID:       strings.TrimSpace(input.ProfileID),
		DisplayName:     displayName,
		RuntimeType:     runtimeType,
		RuntimeEndpoint: strings.TrimSpace(input.RuntimeEndpoint),
		Status:          status,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	s.records[record.ID] = record
	return record, nil
}

func (s *Service) Get(id string) (Instance, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.records[strings.TrimSpace(id)]
	return record, ok
}

func (s *Service) List() []Instance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Instance, 0, len(s.records))
	for _, record := range s.records {
		items = append(items, record)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items
}

func (s *Service) Heartbeat(id string) (Instance, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Instance{}, fmt.Errorf("agent instance id required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[id]
	if !ok {
		return Instance{}, fmt.Errorf("agent instance not found")
	}
	now := s.now().UTC()
	record.LastHeartbeatAt = now
	if record.Status == "" || record.Status == "offline" || record.Status == "created" {
		record.Status = "idle"
	}
	record.UpdatedAt = now
	s.records[id] = record
	return record, nil
}

func (s *Service) Disable(id string) (Instance, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Instance{}, fmt.Errorf("agent instance id required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[id]
	if !ok {
		return Instance{}, fmt.Errorf("agent instance not found")
	}
	record.Status = "disabled"
	record.UpdatedAt = s.now().UTC()
	s.records[id] = record
	return record, nil
}

package agentprofile

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type Profile struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	ModelProvider string    `json:"model_provider"`
	ModelName     string    `json:"model_name"`
	SystemPrompt  string    `json:"system_prompt"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateInput struct {
	Name          string `json:"name"`
	ModelProvider string `json:"model_provider"`
	ModelName     string `json:"model_name"`
	SystemPrompt  string `json:"system_prompt"`
	Status        string `json:"status"`
}

type Service struct {
	mu      sync.RWMutex
	nextID  int
	records map[string]Profile
	now     func() time.Time
}

func NewService() *Service {
	return &Service{
		records: make(map[string]Profile),
		now:     time.Now,
	}
}

func (s *Service) Create(input CreateInput) (Profile, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Profile{}, fmt.Errorf("name required")
	}
	modelProvider := strings.TrimSpace(input.ModelProvider)
	if modelProvider == "" {
		modelProvider = "anthropic"
	}
	modelName := strings.TrimSpace(input.ModelName)
	if modelName == "" {
		modelName = "claude-opus-4-1"
	}
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = "active"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	now := s.now().UTC()
	record := Profile{
		ID:            fmt.Sprintf("agent-profile-%d", s.nextID),
		Name:          name,
		ModelProvider: modelProvider,
		ModelName:     modelName,
		SystemPrompt:  strings.TrimSpace(input.SystemPrompt),
		Status:        status,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	s.records[record.ID] = record
	return record, nil
}

func (s *Service) Get(id string) (Profile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.records[strings.TrimSpace(id)]
	return record, ok
}

func (s *Service) List() []Profile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Profile, 0, len(s.records))
	for _, record := range s.records {
		items = append(items, record)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items
}

package skill

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type Skill struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateInput struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type Service struct {
	mu      sync.RWMutex
	nextID  int
	records map[string]Skill
	now     func() time.Time
}

func NewService() *Service {
	return &Service{
		records: make(map[string]Skill),
		now:     time.Now,
	}
}

func (s *Service) Create(input CreateInput) (Skill, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Skill{}, fmt.Errorf("name required")
	}
	version := strings.TrimSpace(input.Version)
	if version == "" {
		version = "latest"
	}
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = "active"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	now := s.now().UTC()
	record := Skill{
		ID:          fmt.Sprintf("skill-%d", s.nextID),
		Name:        name,
		Version:     version,
		Description: strings.TrimSpace(input.Description),
		Status:      status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.records[record.ID] = record
	return record, nil
}

func (s *Service) Get(id string) (Skill, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.records[strings.TrimSpace(id)]
	return record, ok
}

func (s *Service) List() []Skill {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Skill, 0, len(s.records))
	for _, record := range s.records {
		items = append(items, record)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items
}

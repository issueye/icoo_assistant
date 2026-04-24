package team

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type Team struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	EntryAgentID string    `json:"entry_agent_id"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Member struct {
	ID             string    `json:"id"`
	TeamID         string    `json:"team_id"`
	AgentID        string    `json:"agent_id"`
	Role           string    `json:"role"`
	SortOrder      int       `json:"sort_order"`
	Status         string    `json:"status"`
	Responsibility string    `json:"responsibility,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateInput struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	EntryAgentID string `json:"entry_agent_id"`
	Status       string `json:"status"`
}

type AddMemberInput struct {
	AgentID        string `json:"agent_id"`
	Role           string `json:"role"`
	SortOrder      int    `json:"sort_order"`
	Status         string `json:"status"`
	Responsibility string `json:"responsibility"`
}

type Service struct {
	mu           sync.RWMutex
	nextID       int
	nextMemberID int
	records      map[string]Team
	members      map[string][]Member
	now          func() time.Time
}

func NewService() *Service {
	return &Service{
		records: make(map[string]Team),
		members: make(map[string][]Member),
		now:     time.Now,
	}
}

func (s *Service) Create(input CreateInput) (Team, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Team{}, fmt.Errorf("name required")
	}
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = "active"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	now := s.now().UTC()
	record := Team{
		ID:           fmt.Sprintf("team-%d", s.nextID),
		Name:         name,
		Description:  strings.TrimSpace(input.Description),
		EntryAgentID: strings.TrimSpace(input.EntryAgentID),
		Status:       status,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.records[record.ID] = record
	return record, nil
}

func (s *Service) Get(id string) (Team, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.records[strings.TrimSpace(id)]
	return record, ok
}

func (s *Service) List() []Team {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Team, 0, len(s.records))
	for _, record := range s.records {
		items = append(items, record)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items
}

func (s *Service) AddMember(teamID string, input AddMemberInput) (Member, error) {
	teamID = strings.TrimSpace(teamID)
	if teamID == "" {
		return Member{}, fmt.Errorf("team id required")
	}
	agentID := strings.TrimSpace(input.AgentID)
	if agentID == "" {
		return Member{}, fmt.Errorf("agent_id required")
	}
	role := strings.TrimSpace(input.Role)
	if role == "" {
		role = "member"
	}
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = "active"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[teamID]
	if !ok {
		return Member{}, fmt.Errorf("team not found")
	}
	for _, member := range s.members[teamID] {
		if member.AgentID == agentID {
			return Member{}, fmt.Errorf("agent already added to team")
		}
	}

	s.nextMemberID++
	now := s.now().UTC()
	member := Member{
		ID:             fmt.Sprintf("team-member-%d", s.nextMemberID),
		TeamID:         teamID,
		AgentID:        agentID,
		Role:           role,
		SortOrder:      input.SortOrder,
		Status:         status,
		Responsibility: strings.TrimSpace(input.Responsibility),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	s.members[teamID] = append(s.members[teamID], member)
	sort.SliceStable(s.members[teamID], func(i, j int) bool {
		left := s.members[teamID][i]
		right := s.members[teamID][j]
		if left.SortOrder == right.SortOrder {
			return left.ID < right.ID
		}
		return left.SortOrder < right.SortOrder
	})
	record.UpdatedAt = now
	s.records[teamID] = record
	return member, nil
}

func (s *Service) ListMembers(teamID string) ([]Member, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	teamID = strings.TrimSpace(teamID)
	if _, ok := s.records[teamID]; !ok {
		return nil, false
	}
	items := append([]Member(nil), s.members[teamID]...)
	return items, true
}

func (s *Service) HasMember(teamID, agentID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	teamID = strings.TrimSpace(teamID)
	agentID = strings.TrimSpace(agentID)
	for _, member := range s.members[teamID] {
		if member.AgentID == agentID {
			return true
		}
	}
	return false
}

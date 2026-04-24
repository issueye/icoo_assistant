package conversation

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type Conversation struct {
	ID            string    `json:"id"`
	Mode          string    `json:"mode"`
	Title         string    `json:"title"`
	TargetAgentID string    `json:"target_agent_id,omitempty"`
	TargetTeamID  string    `json:"target_team_id,omitempty"`
	Status        string    `json:"status"`
	MessageCount  int       `json:"message_count"`
	CreatedBy     string    `json:"created_by,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Message struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Scope          string    `json:"scope"`
	Role           string    `json:"role"`
	SenderType     string    `json:"sender_type,omitempty"`
	SenderID       string    `json:"sender_id,omitempty"`
	ReceiverType   string    `json:"receiver_type,omitempty"`
	ReceiverID     string    `json:"receiver_id,omitempty"`
	Content        string    `json:"content"`
	SequenceNo     int       `json:"sequence_no"`
	CreatedAt      time.Time `json:"created_at"`
}

type CreateInput struct {
	Mode          string `json:"mode"`
	Title         string `json:"title"`
	TargetAgentID string `json:"target_agent_id"`
	TargetTeamID  string `json:"target_team_id"`
	CreatedBy     string `json:"created_by"`
	Status        string `json:"status"`
}

type AddMessageInput struct {
	Scope        string `json:"scope"`
	Role         string `json:"role"`
	SenderType   string `json:"sender_type"`
	SenderID     string `json:"sender_id"`
	ReceiverType string `json:"receiver_type"`
	ReceiverID   string `json:"receiver_id"`
	Content      string `json:"content"`
}

type Service struct {
	mu            sync.RWMutex
	nextConvID    int
	nextMessageID int
	conversations map[string]Conversation
	messages      map[string][]Message
	now           func() time.Time
}

func NewService() *Service {
	return &Service{
		conversations: make(map[string]Conversation),
		messages:      make(map[string][]Message),
		now:           time.Now,
	}
}

func (s *Service) Create(input CreateInput) (Conversation, error) {
	mode := strings.TrimSpace(input.Mode)
	if mode == "" {
		mode = "single"
	}
	if mode != "single" && mode != "team" {
		return Conversation{}, fmt.Errorf("unsupported mode")
	}
	if mode == "single" && strings.TrimSpace(input.TargetAgentID) == "" {
		return Conversation{}, fmt.Errorf("target_agent_id required for single mode")
	}
	if mode == "team" && strings.TrimSpace(input.TargetTeamID) == "" {
		return Conversation{}, fmt.Errorf("target_team_id required for team mode")
	}
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = "created"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextConvID++
	now := s.now().UTC()
	record := Conversation{
		ID:            fmt.Sprintf("conv-%d", s.nextConvID),
		Mode:          mode,
		Title:         strings.TrimSpace(input.Title),
		TargetAgentID: strings.TrimSpace(input.TargetAgentID),
		TargetTeamID:  strings.TrimSpace(input.TargetTeamID),
		Status:        status,
		CreatedBy:     strings.TrimSpace(input.CreatedBy),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	s.conversations[record.ID] = record
	return record, nil
}

func (s *Service) Get(id string) (Conversation, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.conversations[strings.TrimSpace(id)]
	return record, ok
}

func (s *Service) List() []Conversation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Conversation, 0, len(s.conversations))
	for _, record := range s.conversations {
		items = append(items, record)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items
}

func (s *Service) AddMessage(conversationID string, input AddMessageInput) (Message, error) {
	conversationID = strings.TrimSpace(conversationID)
	if conversationID == "" {
		return Message{}, fmt.Errorf("conversation id required")
	}
	scope := strings.TrimSpace(input.Scope)
	if scope == "" {
		scope = "external"
	}
	if scope != "external" && scope != "internal" && scope != "system" {
		return Message{}, fmt.Errorf("unsupported scope")
	}
	role := strings.TrimSpace(input.Role)
	if role == "" {
		return Message{}, fmt.Errorf("role required")
	}
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return Message{}, fmt.Errorf("content required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.conversations[conversationID]
	if !ok {
		return Message{}, fmt.Errorf("conversation not found")
	}
	if record.Mode == "single" && scope != "external" {
		return Message{}, fmt.Errorf("single conversation only supports external messages")
	}
	if record.Mode == "team" && scope == "internal" {
		if strings.TrimSpace(input.SenderType) == "" || strings.TrimSpace(input.SenderID) == "" {
			return Message{}, fmt.Errorf("internal message requires sender_type and sender_id")
		}
		if strings.TrimSpace(input.ReceiverType) == "" || strings.TrimSpace(input.ReceiverID) == "" {
			return Message{}, fmt.Errorf("internal message requires receiver_type and receiver_id")
		}
	}
	s.nextMessageID++
	now := s.now().UTC()
	sequenceNo := len(s.messages[conversationID]) + 1
	message := Message{
		ID:             fmt.Sprintf("msg-%d", s.nextMessageID),
		ConversationID: conversationID,
		Scope:          scope,
		Role:           role,
		SenderType:     strings.TrimSpace(input.SenderType),
		SenderID:       strings.TrimSpace(input.SenderID),
		ReceiverType:   strings.TrimSpace(input.ReceiverType),
		ReceiverID:     strings.TrimSpace(input.ReceiverID),
		Content:        content,
		SequenceNo:     sequenceNo,
		CreatedAt:      now,
	}
	s.messages[conversationID] = append(s.messages[conversationID], message)
	record.MessageCount = len(s.messages[conversationID])
	record.UpdatedAt = now
	if record.Status == "created" {
		record.Status = "running"
	}
	s.conversations[conversationID] = record
	return message, nil
}

func (s *Service) ListMessages(conversationID string) ([]Message, bool) {
	return s.ListMessagesByScope(conversationID, "")
}

func (s *Service) ListMessagesByScope(conversationID, scope string) ([]Message, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conversationID = strings.TrimSpace(conversationID)
	if _, ok := s.conversations[conversationID]; !ok {
		return nil, false
	}
	scope = strings.TrimSpace(scope)
	if scope == "" {
		items := append([]Message(nil), s.messages[conversationID]...)
		return items, true
	}
	items := make([]Message, 0, len(s.messages[conversationID]))
	for _, message := range s.messages[conversationID] {
		if message.Scope == scope {
			items = append(items, message)
		}
	}
	return items, true
}

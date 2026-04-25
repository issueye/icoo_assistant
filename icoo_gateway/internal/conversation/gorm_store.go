package conversation

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormConversationModel struct {
	ID            string `gorm:"primaryKey;size:64"`
	Mode          string `gorm:"size:32;index"`
	Title         string `gorm:"type:text"`
	TargetAgentID string `gorm:"size:128;index"`
	TargetTeamID  string `gorm:"size:128;index"`
	Status        string `gorm:"size:64;index"`
	LastRunID     string `gorm:"size:128;index"`
	MessageCount  int
	CreatedBy     string    `gorm:"size:128"`
	CreatedAt     time.Time `gorm:"index"`
	UpdatedAt     time.Time `gorm:"index"`
}

func (GormConversationModel) TableName() string {
	return "conversations"
}

type GormMessageModel struct {
	ID             string    `gorm:"primaryKey;size:64"`
	ConversationID string    `gorm:"size:64;index"`
	Scope          string    `gorm:"size:32;index"`
	Role           string    `gorm:"size:32"`
	SenderType     string    `gorm:"size:64"`
	SenderID       string    `gorm:"size:128"`
	ReceiverType   string    `gorm:"size:64"`
	ReceiverID     string    `gorm:"size:128"`
	Content        string    `gorm:"type:text"`
	SequenceNo     int       `gorm:"index"`
	CreatedAt      time.Time `gorm:"index"`
}

func (GormMessageModel) TableName() string {
	return "conversation_messages"
}

type GormStore struct {
	db *gorm.DB
}

var _ Store = (*GormStore)(nil)

func NewGormStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db}
}

func (s *GormStore) Create(input CreateInput) (Conversation, error) {
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
	now := time.Now().UTC()
	record := GormConversationModel{
		ID:            fmt.Sprintf("conv-%s", uuid.NewString()),
		Mode:          mode,
		Title:         strings.TrimSpace(input.Title),
		TargetAgentID: strings.TrimSpace(input.TargetAgentID),
		TargetTeamID:  strings.TrimSpace(input.TargetTeamID),
		Status:        status,
		CreatedBy:     strings.TrimSpace(input.CreatedBy),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.db.Create(&record).Error; err != nil {
		return Conversation{}, err
	}
	return modelToConversation(record), nil
}

func (s *GormStore) Get(id string) (Conversation, bool) {
	var record GormConversationModel
	if err := s.db.First(&record, "id = ?", strings.TrimSpace(id)).Error; err != nil {
		return Conversation{}, false
	}
	return modelToConversation(record), true
}

func (s *GormStore) List() []Conversation {
	var records []GormConversationModel
	if err := s.db.Order("created_at asc").Find(&records).Error; err != nil {
		return nil
	}
	items := make([]Conversation, 0, len(records))
	for _, record := range records {
		items = append(items, modelToConversation(record))
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})
	return items
}

func (s *GormStore) AddMessage(conversationID string, input AddMessageInput) (Message, error) {
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

	var created Message
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var record GormConversationModel
		if err := tx.First(&record, "id = ?", conversationID).Error; err != nil {
			return fmt.Errorf("conversation not found")
		}
		if record.Mode == "single" && scope != "external" {
			return fmt.Errorf("single conversation only supports external messages")
		}
		if record.Mode == "team" && scope == "internal" {
			if strings.TrimSpace(input.SenderType) == "" || strings.TrimSpace(input.SenderID) == "" {
				return fmt.Errorf("internal message requires sender_type and sender_id")
			}
			if strings.TrimSpace(input.ReceiverType) == "" || strings.TrimSpace(input.ReceiverID) == "" {
				return fmt.Errorf("internal message requires receiver_type and receiver_id")
			}
		}
		var count int64
		if err := tx.Model(&GormMessageModel{}).Where("conversation_id = ?", conversationID).Count(&count).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		message := GormMessageModel{
			ID:             fmt.Sprintf("msg-%s", uuid.NewString()),
			ConversationID: conversationID,
			Scope:          scope,
			Role:           role,
			SenderType:     strings.TrimSpace(input.SenderType),
			SenderID:       strings.TrimSpace(input.SenderID),
			ReceiverType:   strings.TrimSpace(input.ReceiverType),
			ReceiverID:     strings.TrimSpace(input.ReceiverID),
			Content:        content,
			SequenceNo:     int(count) + 1,
			CreatedAt:      now,
		}
		if err := tx.Create(&message).Error; err != nil {
			return err
		}
		record.MessageCount = message.SequenceNo
		record.UpdatedAt = now
		if record.Status == "created" {
			record.Status = "running"
		}
		if err := tx.Save(&record).Error; err != nil {
			return err
		}
		created = modelToMessage(message)
		return nil
	})
	if err != nil {
		return Message{}, err
	}
	return created, nil
}

func (s *GormStore) ListMessagesByScope(conversationID, scope string) ([]Message, bool) {
	if _, ok := s.Get(conversationID); !ok {
		return nil, false
	}
	query := s.db.Where("conversation_id = ?", strings.TrimSpace(conversationID)).Order("sequence_no asc")
	scope = strings.TrimSpace(scope)
	if scope != "" {
		query = query.Where("scope = ?", scope)
	}
	var records []GormMessageModel
	if err := query.Find(&records).Error; err != nil {
		return nil, true
	}
	items := make([]Message, 0, len(records))
	for _, record := range records {
		items = append(items, modelToMessage(record))
	}
	return items, true
}

func (s *GormStore) SetLastRunID(conversationID, runID string) (Conversation, error) {
	var record GormConversationModel
	if err := s.db.First(&record, "id = ?", strings.TrimSpace(conversationID)).Error; err != nil {
		return Conversation{}, fmt.Errorf("conversation not found")
	}
	record.LastRunID = strings.TrimSpace(runID)
	record.UpdatedAt = time.Now().UTC()
	if err := s.db.Save(&record).Error; err != nil {
		return Conversation{}, err
	}
	return modelToConversation(record), nil
}

func modelToConversation(record GormConversationModel) Conversation {
	return Conversation{
		ID:            record.ID,
		Mode:          record.Mode,
		Title:         record.Title,
		TargetAgentID: record.TargetAgentID,
		TargetTeamID:  record.TargetTeamID,
		Status:        record.Status,
		LastRunID:     record.LastRunID,
		MessageCount:  record.MessageCount,
		CreatedBy:     record.CreatedBy,
		CreatedAt:     record.CreatedAt,
		UpdatedAt:     record.UpdatedAt,
	}
}

func modelToMessage(record GormMessageModel) Message {
	return Message{
		ID:             record.ID,
		ConversationID: record.ConversationID,
		Scope:          record.Scope,
		Role:           record.Role,
		SenderType:     record.SenderType,
		SenderID:       record.SenderID,
		ReceiverType:   record.ReceiverType,
		ReceiverID:     record.ReceiverID,
		Content:        record.Content,
		SequenceNo:     record.SequenceNo,
		CreatedAt:      record.CreatedAt,
	}
}

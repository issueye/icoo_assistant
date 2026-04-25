package run

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormModel struct {
	ID               string    `gorm:"primaryKey;size:64"`
	ConversationID   string    `gorm:"size:128;index"`
	TriggerType      string    `gorm:"size:64"`
	TriggerMessageID string    `gorm:"size:128;index"`
	Status           string    `gorm:"size:64;index"`
	StartedAt        time.Time `gorm:"index"`
	FinishedAt       *time.Time
	Summary          string `gorm:"type:text"`
	ErrorMessage     string `gorm:"type:text"`
}

func (GormModel) TableName() string {
	return "runs"
}

type GormStore struct {
	db *gorm.DB
}

var _ Store = (*GormStore)(nil)

func NewGormStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db}
}

func (s *GormStore) Create(input CreateInput) (Run, error) {
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
	now := time.Now().UTC()
	record := GormModel{
		ID:               fmt.Sprintf("run-%s", uuid.NewString()),
		ConversationID:   conversationID,
		TriggerType:      triggerType,
		TriggerMessageID: strings.TrimSpace(input.TriggerMessageID),
		Status:           status,
		StartedAt:        now,
		Summary:          strings.TrimSpace(input.Summary),
		ErrorMessage:     strings.TrimSpace(input.ErrorMessage),
	}
	if record.Status != "running" {
		record.FinishedAt = &now
	}
	if err := s.db.Create(&record).Error; err != nil {
		return Run{}, err
	}
	return modelToRun(record), nil
}

func (s *GormStore) Complete(id string, input CompleteInput) (Run, error) {
	var record GormModel
	if err := s.db.First(&record, "id = ?", strings.TrimSpace(id)).Error; err != nil {
		return Run{}, fmt.Errorf("run not found")
	}
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = "completed"
	}
	now := time.Now().UTC()
	record.Status = status
	record.Summary = strings.TrimSpace(input.Summary)
	record.ErrorMessage = strings.TrimSpace(input.ErrorMessage)
	record.FinishedAt = &now
	if err := s.db.Save(&record).Error; err != nil {
		return Run{}, err
	}
	return modelToRun(record), nil
}

func (s *GormStore) ListByConversation(conversationID string) []Run {
	var records []GormModel
	if err := s.db.Where("conversation_id = ?", strings.TrimSpace(conversationID)).Order("started_at asc").Find(&records).Error; err != nil {
		return nil
	}
	items := make([]Run, 0, len(records))
	for _, record := range records {
		items = append(items, modelToRun(record))
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].StartedAt.Before(items[j].StartedAt)
	})
	return items
}

func modelToRun(record GormModel) Run {
	return Run{
		ID:               record.ID,
		ConversationID:   record.ConversationID,
		TriggerType:      record.TriggerType,
		TriggerMessageID: record.TriggerMessageID,
		Status:           record.Status,
		StartedAt:        record.StartedAt,
		FinishedAt:       record.FinishedAt,
		Summary:          record.Summary,
		ErrorMessage:     record.ErrorMessage,
	}
}

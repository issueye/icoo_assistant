package audit

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormModel struct {
	ID           string    `gorm:"primaryKey;size:64"`
	ResourceType string    `gorm:"size:64;index"`
	ResourceID   string    `gorm:"size:128;index"`
	EventName    string    `gorm:"size:128;index"`
	Operator     string    `gorm:"size:128"`
	PayloadJSON  string    `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"index"`
}

func (GormModel) TableName() string {
	return "audit_events"
}

type GormStore struct {
	db *gorm.DB
}

var _ Store = (*GormStore)(nil)

func NewGormStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db}
}

func (s *GormStore) Record(input RecordInput) Event {
	now := time.Now().UTC()
	payloadJSON := ""
	if input.Payload != nil {
		if data, err := json.Marshal(input.Payload); err == nil {
			payloadJSON = string(data)
		}
	}
	record := GormModel{
		ID:           fmt.Sprintf("audit-%s", uuid.NewString()),
		ResourceType: strings.TrimSpace(input.ResourceType),
		ResourceID:   strings.TrimSpace(input.ResourceID),
		EventName:    strings.TrimSpace(input.EventName),
		Operator:     strings.TrimSpace(input.Operator),
		PayloadJSON:  payloadJSON,
		CreatedAt:    now,
	}
	if record.Operator == "" {
		record.Operator = "system"
	}
	_ = s.db.Create(&record).Error
	return modelToEvent(record)
}

func (s *GormStore) Get(id string) (Event, bool) {
	var record GormModel
	err := s.db.First(&record, "id = ?", strings.TrimSpace(id)).Error
	if err != nil {
		return Event{}, false
	}
	return modelToEvent(record), true
}

func (s *GormStore) List() []Event {
	var records []GormModel
	if err := s.db.Order("created_at asc").Find(&records).Error; err != nil {
		return nil
	}
	items := make([]Event, 0, len(records))
	for _, record := range records {
		items = append(items, modelToEvent(record))
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})
	return items
}

func modelToEvent(record GormModel) Event {
	var payload interface{}
	if strings.TrimSpace(record.PayloadJSON) != "" {
		var decoded interface{}
		if err := json.Unmarshal([]byte(record.PayloadJSON), &decoded); err == nil {
			payload = decoded
		}
	}
	return Event{
		ID:           record.ID,
		ResourceType: record.ResourceType,
		ResourceID:   record.ResourceID,
		EventName:    record.EventName,
		Operator:     record.Operator,
		Payload:      payload,
		CreatedAt:    record.CreatedAt,
	}
}

package endpoint

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"icoo_proxy/internal/consts"
	"icoo_proxy/internal/storage"
)

type Record struct {
	ID          string          `json:"id"`
	Path        string          `json:"path"`
	Protocol    consts.Protocol `json:"protocol"`
	Description string          `json:"description"`
	Enabled     bool            `json:"enabled"`
	BuiltIn     bool            `json:"built_in"`
	UpdatedAt   string          `json:"updated_at"`
	CreatedAt   string          `json:"created_at"`
}

type endpointModel struct {
	ID          string `gorm:"primaryKey"`
	Path        string `gorm:"uniqueIndex"`
	Protocol    consts.Protocol
	Description string
	Enabled     bool
	BuiltIn     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (endpointModel) TableName() string {
	return "endpoints"
}

type UpsertInput struct {
	ID          string `json:"id"`
	Path        string `json:"path"`
	Protocol    string `json:"protocol"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

type DefaultDefinition struct {
	Path        string
	Protocol    consts.Protocol
	Description string
}

var defaultDefinitions = []DefaultDefinition{
	{Path: "/v1/messages", Protocol: consts.ProtocolAnthropic, Description: "Anthropic Messages official-compatible endpoint."},
	{Path: "/anthropic/v1/messages", Protocol: consts.ProtocolAnthropic, Description: "Anthropic namespaced Messages endpoint."},
	{Path: "/v1/chat/completions", Protocol: consts.ProtocolOpenAIChat, Description: "OpenAI Chat Completions official-compatible endpoint."},
	{Path: "/openai/v1/chat/completions", Protocol: consts.ProtocolOpenAIChat, Description: "OpenAI namespaced Chat Completions endpoint."},
	{Path: "/v1/responses", Protocol: consts.ProtocolOpenAIResponses, Description: "OpenAI Responses official-compatible endpoint."},
	{Path: "/openai/v1/responses", Protocol: consts.ProtocolOpenAIResponses, Description: "OpenAI namespaced Responses endpoint."},
}

func DefaultDefinitions() []DefaultDefinition {
	return append([]DefaultDefinition(nil), defaultDefinitions...)
}

type Service struct {
	db *gorm.DB
}

func NewService(root string) (*Service, error) {
	db, err := storage.Open(root)
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&endpointModel{}); err != nil {
		return nil, err
	}
	svc := &Service{db: db}
	return svc, nil
}

func (s *Service) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (s *Service) List() []Record {
	var rows []endpointModel
	if err := s.db.Order("built_in desc, path asc").Find(&rows).Error; err != nil {
		return nil
	}
	items := make([]Record, 0, len(rows))
	for _, item := range rows {
		items = append(items, toRecord(item))
	}
	return items
}

func (s *Service) Enabled() []Record {
	var rows []endpointModel
	if err := s.db.Where("enabled = ?", true).Order("built_in desc, path asc").Find(&rows).Error; err != nil {
		return nil
	}
	items := make([]Record, 0, len(rows))
	for _, item := range rows {
		items = append(items, toRecord(item))
	}
	return items
}

func (s *Service) Upsert(input UpsertInput) (Record, error) {
	path := normalizePath(input.Path)
	if path == "" {
		return Record{}, fmt.Errorf("endpoint path is required")
	}
	protocol := normalizeProtocol(input.Protocol)
	if protocol == consts.Protocol("") {
		return Record{}, fmt.Errorf("endpoint protocol is required")
	}

	id := strings.TrimSpace(input.ID)
	var existing endpointModel
	found := false
	if id != "" && s.db.Limit(1).Find(&existing, "id = ?", id).RowsAffected > 0 {
		found = true
	} else if s.db.Limit(1).Find(&existing, "path = ?", path).RowsAffected > 0 {
		found = true
	}

	now := time.Now()
	current := endpointModel{
		ID:          buildID(path),
		Path:        path,
		Protocol:    protocol,
		Description: strings.TrimSpace(input.Description),
		Enabled:     input.Enabled,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if found {
		current.ID = existing.ID
		current.BuiltIn = existing.BuiltIn
		current.CreatedAt = existing.CreatedAt
	} else if id != "" {
		current.ID = id
	}
	if err := s.db.Save(&current).Error; err != nil {
		return Record{}, err
	}
	return toRecord(current), nil
}

func (s *Service) Delete(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("endpoint id is required")
	}
	var item endpointModel
	if err := s.db.First(&item, "id = ?", id).Error; err != nil {
		return fmt.Errorf("endpoint not found")
	}
	if item.BuiltIn {
		return fmt.Errorf("built-in endpoint cannot be deleted")
	}
	return s.db.Delete(&endpointModel{}, "id = ?", id).Error
}

func toRecord(item endpointModel) Record {
	return Record{
		ID:          item.ID,
		Path:        item.Path,
		Protocol:    item.Protocol,
		Description: item.Description,
		Enabled:     item.Enabled,
		BuiltIn:     item.BuiltIn,
		UpdatedAt:   item.UpdatedAt.Format(time.RFC3339),
		CreatedAt:   item.CreatedAt.Format(time.RFC3339),
	}
}

func normalizePath(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}
	return value
}

func normalizeProtocol(raw string) consts.Protocol {
	value := consts.Protocol(strings.TrimSpace(raw))
	switch value {
	case consts.ProtocolAnthropic, consts.ProtocolOpenAIChat, consts.ProtocolOpenAIResponses:
		return value
	default:
		return ""
	}
}

func buildID(path string) string {
	id := strings.Trim(strings.ReplaceAll(path, "/", "-"), "-")
	id = strings.ReplaceAll(id, "_", "-")
	if id == "" {
		id = "endpoint"
	}
	return "endpoint-" + id
}

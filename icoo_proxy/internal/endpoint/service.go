package endpoint

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"icoo_proxy/internal/storage"
)

type Record struct {
	ID          string `json:"id"`
	Path        string `json:"path"`
	Protocol    string `json:"protocol"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	BuiltIn     bool   `json:"built_in"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}

type endpointModel struct {
	ID          string `gorm:"primaryKey"`
	Path        string `gorm:"uniqueIndex"`
	Protocol    string
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
	if err := svc.seedDefaults(); err != nil {
		return nil, err
	}
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
	if protocol == "" {
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

func (s *Service) seedDefaults() error {
	for _, item := range defaultEndpoints() {
		var count int64
		if err := s.db.Model(&endpointModel{}).Where("path = ?", item.Path).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := s.db.Create(&item).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func defaultEndpoints() []endpointModel {
	now := time.Now()
	return []endpointModel{
		defaultEndpoint("/v1/messages", "anthropic", "Anthropic Messages official-compatible endpoint.", now),
		defaultEndpoint("/anthropic/v1/messages", "anthropic", "Anthropic namespaced Messages endpoint.", now),
		defaultEndpoint("/v1/chat/completions", "openai-chat", "OpenAI Chat Completions official-compatible endpoint.", now),
		defaultEndpoint("/openai/v1/chat/completions", "openai-chat", "OpenAI namespaced Chat Completions endpoint.", now),
		defaultEndpoint("/v1/responses", "openai-responses", "OpenAI Responses official-compatible endpoint.", now),
		defaultEndpoint("/openai/v1/responses", "openai-responses", "OpenAI namespaced Responses endpoint.", now),
	}
}

func defaultEndpoint(path, protocol, description string, now time.Time) endpointModel {
	return endpointModel{
		ID:          buildID(path),
		Path:        path,
		Protocol:    protocol,
		Description: description,
		Enabled:     true,
		BuiltIn:     true,
		UpdatedAt:   now,
		CreatedAt:   now,
	}
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

func normalizeProtocol(raw string) string {
	value := strings.TrimSpace(strings.ToLower(raw))
	switch value {
	case "anthropic", "openai-chat", "openai-responses":
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

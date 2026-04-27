package supplier

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"gorm.io/gorm"

	"icoo_proxy/internal/consts"
	"icoo_proxy/internal/routepolicy"
	"icoo_proxy/internal/storage"
)

type Record struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Protocol     consts.Protocol `json:"protocol"`
	BaseURL      string          `json:"base_url"`
	APIKeyMasked string          `json:"api_key_masked"`
	OnlyStream   bool            `json:"only_stream"`
	UserAgent    string          `json:"user_agent"`
	Enabled      bool            `json:"enabled"`
	Description  string          `json:"description"`
	Models       []string        `json:"models"`
	Tags         []string        `json:"tags"`
	UpdatedAt    string          `json:"updated_at"`
	CreatedAt    string          `json:"created_at"`
}

type supplierModel struct {
	ID          string `gorm:"primaryKey"`
	Name        string `gorm:"index"`
	Protocol    consts.Protocol
	BaseURL     string
	APIKey      string
	OnlyStream  bool
	UserAgent   string
	Enabled     bool
	Description string
	Models      string
	Tags        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (supplierModel) TableName() string {
	return "suppliers"
}

type UpsertInput struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Protocol    string `json:"protocol"`
	BaseURL     string `json:"base_url"`
	APIKey      string `json:"api_key"`
	OnlyStream  bool   `json:"only_stream"`
	UserAgent   string `json:"user_agent"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
	Models      string `json:"models"`
	Tags        string `json:"tags"`
}

type Service struct {
	db *gorm.DB
}

func NewService(root string) (*Service, error) {
	db, err := storage.Open(root)
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&supplierModel{}); err != nil {
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
	var rows []supplierModel
	if err := s.db.Order("lower(name) asc").Find(&rows).Error; err != nil {
		return nil
	}
	items := make([]Record, 0, len(rows))
	for _, item := range rows {
		items = append(items, toRecord(item))
	}
	return items
}

func (s *Service) Resolve(id string) (routepolicy.SupplierSnapshot, bool) {
	var item supplierModel
	if err := s.db.First(&item, "id = ?", strings.TrimSpace(id)).Error; err != nil {
		return routepolicy.SupplierSnapshot{}, false
	}
	return routepolicy.SupplierSnapshot{
		ID:         item.ID,
		Name:       item.Name,
		Protocol:   item.Protocol,
		BaseURL:    item.BaseURL,
		APIKey:     item.APIKey,
		OnlyStream: item.OnlyStream,
		UserAgent:  item.UserAgent,
		IsEnabled:  item.Enabled,
	}, true
}

func (s *Service) Upsert(input UpsertInput) (Record, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Record{}, fmt.Errorf("supplier name is required")
	}
	protocol := normalizeProtocol(input.Protocol)
	if protocol == consts.Protocol("") {
		return Record{}, fmt.Errorf("supplier protocol is required")
	}
	baseURL := strings.TrimSpace(input.BaseURL)
	if baseURL == "" {
		return Record{}, fmt.Errorf("supplier base_url is required")
	}

	now := time.Now()
	id := strings.TrimSpace(input.ID)
	var existing supplierModel
	found := false
	if id != "" {
		found = s.db.Limit(1).Find(&existing, "id = ?", id).RowsAffected > 0
	}

	current := supplierModel{
		ID:          generateID(name),
		Name:        name,
		Protocol:    protocol,
		BaseURL:     baseURL,
		APIKey:      strings.TrimSpace(input.APIKey),
		OnlyStream:  input.OnlyStream,
		UserAgent:   strings.TrimSpace(input.UserAgent),
		Enabled:     input.Enabled,
		Description: strings.TrimSpace(input.Description),
		Models:      strings.Join(splitCSVLike(input.Models), ","),
		Tags:        strings.Join(splitCSVLike(input.Tags), ","),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if found {
		current.ID = existing.ID
		current.CreatedAt = existing.CreatedAt
		if strings.TrimSpace(input.APIKey) == "" {
			current.APIKey = existing.APIKey
		}
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
		return fmt.Errorf("supplier id is required")
	}
	result := s.db.Delete(&supplierModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("supplier not found")
	}
	return nil
}

func toRecord(item supplierModel) Record {
	return Record{
		ID:           item.ID,
		Name:         item.Name,
		Protocol:     item.Protocol,
		BaseURL:      item.BaseURL,
		APIKeyMasked: maskSecret(item.APIKey),
		OnlyStream:   item.OnlyStream,
		UserAgent:    item.UserAgent,
		Enabled:      item.Enabled,
		Description:  item.Description,
		Models:       slices.Clone(splitCSVLike(item.Models)),
		Tags:         slices.Clone(splitCSVLike(item.Tags)),
		UpdatedAt:    item.UpdatedAt.Format(time.RFC3339),
		CreatedAt:    item.CreatedAt.Format(time.RFC3339),
	}
}

func normalizeProtocol(raw string) consts.Protocol {
	value := consts.Protocol(raw)
	switch value {
	case consts.ProtocolAnthropic, consts.ProtocolOpenAIChat, consts.ProtocolOpenAIResponses:
		return value
	default:
		return value
	}
}

func splitCSVLike(raw string) []string {
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '\n' || r == ';'
	})
	items := make([]string, 0, len(fields))
	for _, field := range fields {
		value := strings.TrimSpace(field)
		if value != "" {
			items = append(items, value)
		}
	}
	return items
}

func maskSecret(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	if len(value) <= 6 {
		return strings.Repeat("*", len(value))
	}
	return value[:3] + strings.Repeat("*", len(value)-6) + value[len(value)-3:]
}

func generateID(name string) string {
	base := strings.ToLower(strings.TrimSpace(name))
	base = strings.ReplaceAll(base, " ", "-")
	base = strings.ReplaceAll(base, "_", "-")
	if base == "" {
		base = "supplier"
	}
	return fmt.Sprintf("%s-%d", base, time.Now().UnixNano())
}

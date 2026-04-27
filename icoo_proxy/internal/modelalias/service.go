package modelalias

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"icoo_proxy/internal/consts"
	"icoo_proxy/internal/storage"
)

type SupplierResolver interface {
	Resolve(id string) (SupplierSnapshot, bool)
}

type SupplierSnapshot struct {
	ID       string
	Name     string
	Protocol consts.Protocol
}

type Record struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	SupplierID       string          `json:"supplier_id"`
	SupplierName     string          `json:"supplier_name"`
	UpstreamProtocol consts.Protocol `json:"upstream_protocol"`
	Model            string          `json:"model"`
	Enabled          bool            `json:"enabled"`
	UpdatedAt        string          `json:"updated_at"`
	CreatedAt        string          `json:"created_at"`
}

type aliasModel struct {
	ID         string `gorm:"primaryKey"`
	Name       string `gorm:"uniqueIndex"`
	SupplierID string
	Model      string
	Enabled    bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (aliasModel) TableName() string {
	return "model_aliases"
}

type UpsertInput struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	SupplierID string `json:"supplier_id"`
	Model      string `json:"model"`
	Enabled    bool   `json:"enabled"`
}

type Service struct {
	db     *gorm.DB
	lookup SupplierResolver
}

func NewService(root string, resolver SupplierResolver) (*Service, error) {
	db, err := storage.Open(root)
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&aliasModel{}); err != nil {
		return nil, err
	}
	return &Service{db: db, lookup: resolver}, nil
}

func (s *Service) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (s *Service) List() []Record {
	var rows []aliasModel
	if err := s.db.Order("lower(name) asc").Find(&rows).Error; err != nil {
		return nil
	}
	items := make([]Record, 0, len(rows))
	for _, item := range rows {
		items = append(items, s.toRecord(item))
	}
	return items
}

func (s *Service) EnabledEntries() []string {
	var rows []aliasModel
	if err := s.db.Where("enabled = ?", true).Order("lower(name) asc").Find(&rows).Error; err != nil {
		return nil
	}
	items := make([]string, 0, len(rows))
	for _, item := range rows {
		supplier, ok := s.lookup.Resolve(item.SupplierID)
		if !ok {
			continue
		}
		items = append(items, fmt.Sprintf("%s=%s:%s",
			item.Name,
			supplier.Protocol.ToString(),
			strings.TrimSpace(item.Model),
		))
	}
	return items
}

func (s *Service) Upsert(input UpsertInput) (Record, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Record{}, fmt.Errorf("model alias name is required")
	}
	supplierID := strings.TrimSpace(input.SupplierID)
	if supplierID == "" {
		return Record{}, fmt.Errorf("model alias supplier is required")
	}
	supplier, ok := s.lookup.Resolve(supplierID)
	if !ok {
		return Record{}, fmt.Errorf("supplier not found")
	}
	model := strings.TrimSpace(input.Model)
	if model == "" {
		return Record{}, fmt.Errorf("model alias target model is required")
	}

	id := strings.TrimSpace(input.ID)
	var existing aliasModel
	found := false
	if id != "" && s.db.Limit(1).Find(&existing, "id = ?", id).RowsAffected > 0 {
		found = true
	} else if s.db.Limit(1).Find(&existing, "name = ?", name).RowsAffected > 0 {
		found = true
	}

	now := time.Now()
	current := aliasModel{
		ID:         buildID(name),
		Name:       name,
		SupplierID: supplier.ID,
		Model:      model,
		Enabled:    input.Enabled,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if found {
		current.ID = existing.ID
		current.CreatedAt = existing.CreatedAt
	} else if id != "" {
		current.ID = id
	}
	if err := s.db.Save(&current).Error; err != nil {
		return Record{}, err
	}
	return s.toRecord(current), nil
}

func (s *Service) Delete(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("model alias id is required")
	}
	result := s.db.Delete(&aliasModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("model alias not found")
	}
	return nil
}

func (s *Service) toRecord(item aliasModel) Record {
	record := Record{
		ID:         item.ID,
		Name:       item.Name,
		SupplierID: item.SupplierID,
		Model:      item.Model,
		Enabled:    item.Enabled,
		UpdatedAt:  item.UpdatedAt.Format(time.RFC3339),
		CreatedAt:  item.CreatedAt.Format(time.RFC3339),
	}
	if supplier, ok := s.lookup.Resolve(item.SupplierID); ok {
		record.SupplierName = supplier.Name
		record.UpstreamProtocol = supplier.Protocol
	}
	return record
}

func buildID(name string) string {
	base := strings.ToLower(strings.TrimSpace(name))
	base = strings.ReplaceAll(base, " ", "-")
	base = strings.ReplaceAll(base, "_", "-")
	if base == "" {
		base = "model-alias"
	}
	return fmt.Sprintf("%s-%d", base, time.Now().UnixNano())
}

func MergeEntries(base string, extra []string) string {
	items := make([]string, 0)
	seen := make(map[string]struct{})
	appendEntry := func(entry string) {
		value := strings.TrimSpace(entry)
		if value == "" {
			return
		}
		alias, _, found := strings.Cut(value, "=")
		alias = strings.TrimSpace(alias)
		if !found || alias == "" {
			return
		}
		if _, ok := seen[alias]; ok {
			return
		}
		seen[alias] = struct{}{}
		items = append(items, value)
	}
	for _, entry := range splitEntries(base) {
		appendEntry(entry)
	}
	for _, entry := range extra {
		value := strings.TrimSpace(entry)
		alias, _, found := strings.Cut(value, "=")
		alias = strings.TrimSpace(alias)
		if !found || alias == "" {
			continue
		}
		if _, ok := seen[alias]; ok {
			for index, item := range items {
				currentAlias, _, _ := strings.Cut(item, "=")
				if strings.TrimSpace(currentAlias) == alias {
					items[index] = value
					break
				}
			}
			continue
		}
		seen[alias] = struct{}{}
		items = append(items, value)
	}
	return strings.Join(items, ",")
}

func splitEntries(raw string) []string {
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

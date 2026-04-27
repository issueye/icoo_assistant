package routepolicy

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"icoo_proxy/internal/consts"
	"icoo_proxy/internal/storage"
)

type Record struct {
	ID                 string          `json:"id"`
	DownstreamProtocol consts.Protocol `json:"downstream_protocol"`
	SupplierID         string          `json:"supplier_id"`
	SupplierName       string          `json:"supplier_name"`
	UpstreamProtocol   consts.Protocol `json:"upstream_protocol"`
	Enabled            bool            `json:"enabled"`
	UpdatedAt          string          `json:"updated_at"`
	CreatedAt          string          `json:"created_at"`
}

type policyModel struct {
	ID                 string          `gorm:"primaryKey"`
	DownstreamProtocol consts.Protocol `gorm:"uniqueIndex"`
	SupplierID         string
	TargetModel        string
	Enabled            bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (policyModel) TableName() string {
	return "route_policies"
}

type SupplierResolver interface {
	Resolve(id string) (SupplierSnapshot, bool)
}

type SupplierSnapshot struct {
	ID           string
	Name         string
	Protocol     consts.Protocol
	BaseURL      string
	APIKey       string
	OnlyStream   bool
	UserAgent    string
	IsEnabled    bool
	DefaultModel string
}

type UpsertInput struct {
	ID                 string          `json:"id"`
	DownstreamProtocol consts.Protocol `json:"downstream_protocol"`
	SupplierID         string          `json:"supplier_id"`
	Enabled            bool            `json:"enabled"`
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
	if err := db.AutoMigrate(&policyModel{}); err != nil {
		return nil, err
	}
	svc := &Service{db: db, lookup: resolver}
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
	var rows []policyModel
	if err := s.db.Order("downstream_protocol asc").Find(&rows).Error; err != nil {
		return nil
	}
	items := make([]Record, 0, len(rows))
	for _, item := range rows {
		items = append(items, s.toRecord(item))
	}
	return items
}

func (s *Service) Enabled() []Record {
	items := s.List()
	filtered := make([]Record, 0, len(items))
	for _, item := range items {
		if item.Enabled {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func (s *Service) FindEnabledBySupplierID(supplierID string) (Record, bool) {
	id := strings.TrimSpace(supplierID)
	if id == "" {
		return Record{}, false
	}
	for _, item := range s.Enabled() {
		if item.SupplierID == id {
			return item, true
		}
	}
	return Record{}, false
}

func (s *Service) Upsert(input UpsertInput) (Record, error) {
	downstream := normalizeProtocol(input.DownstreamProtocol)
	if downstream == "" {
		return Record{}, fmt.Errorf("downstream protocol is required")
	}
	if strings.TrimSpace(input.SupplierID) == "" {
		return Record{}, fmt.Errorf("supplier id is required")
	}
	supplier, ok := s.lookup.Resolve(input.SupplierID)
	if !ok {
		return Record{}, fmt.Errorf("supplier not found")
	}
	if input.Enabled && strings.TrimSpace(supplier.DefaultModel) == "" {
		return Record{}, fmt.Errorf("supplier %q default model is required when enabling a route policy", supplier.Name)
	}

	id := strings.TrimSpace(input.ID)
	var existing policyModel
	found := false
	if id != "" && s.db.Limit(1).Find(&existing, "id = ?", id).RowsAffected > 0 {
		found = true
	} else if s.db.Limit(1).Find(&existing, "downstream_protocol = ?", downstream).RowsAffected > 0 {
		found = true
	}

	now := time.Now()
	current := policyModel{
		ID:                 buildID(downstream),
		DownstreamProtocol: downstream,
		SupplierID:         supplier.ID,
		Enabled:            input.Enabled,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if found {
		current.ID = existing.ID
		current.CreatedAt = existing.CreatedAt
		current.TargetModel = existing.TargetModel
	} else if id != "" {
		current.ID = id
	}
	if err := s.db.Save(&current).Error; err != nil {
		return Record{}, err
	}
	return s.toRecord(current), nil
}

func (s *Service) seedDefaults() error {
	for _, item := range defaultPolicies() {
		var count int64
		if err := s.db.Model(&policyModel{}).Where("downstream_protocol = ?", item.DownstreamProtocol).Count(&count).Error; err != nil {
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

func (s *Service) toRecord(item policyModel) Record {
	record := Record{
		ID:                 item.ID,
		DownstreamProtocol: item.DownstreamProtocol,
		SupplierID:         item.SupplierID,
		Enabled:            item.Enabled,
		UpdatedAt:          item.UpdatedAt.Format(time.RFC3339),
		CreatedAt:          item.CreatedAt.Format(time.RFC3339),
	}
	if supplier, ok := s.lookup.Resolve(item.SupplierID); ok {
		record.SupplierName = supplier.Name
		record.UpstreamProtocol = supplier.Protocol
	}
	return record
}

func normalizeProtocol(raw consts.Protocol) consts.Protocol {
	value := raw
	switch value {
	case consts.ProtocolAnthropic, consts.ProtocolOpenAIChat, consts.ProtocolOpenAIResponses:
		return value
	default:
		return consts.Protocol("")
	}
}

func buildID(downstream consts.Protocol) string {
	return "policy-" + downstream.ToString()
}

func defaultPolicies() []policyModel {
	now := time.Now()
	return []policyModel{
		{
			ID:                 buildID(consts.ProtocolAnthropic),
			DownstreamProtocol: consts.ProtocolAnthropic,
			Enabled:            false,
			UpdatedAt:          now,
			CreatedAt:          now,
		},
		{
			ID:                 buildID(consts.ProtocolOpenAIChat),
			DownstreamProtocol: consts.ProtocolOpenAIChat,
			Enabled:            false,
			UpdatedAt:          now,
			CreatedAt:          now,
		},
		{
			ID:                 buildID(consts.ProtocolOpenAIResponses),
			DownstreamProtocol: consts.ProtocolOpenAIResponses,
			Enabled:            false,
			UpdatedAt:          now,
			CreatedAt:          now,
		},
	}
}

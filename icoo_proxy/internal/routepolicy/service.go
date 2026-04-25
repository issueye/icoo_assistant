package routepolicy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

type Record struct {
	ID                 string `json:"id"`
	DownstreamProtocol string `json:"downstream_protocol"`
	SupplierID         string `json:"supplier_id"`
	SupplierName       string `json:"supplier_name"`
	UpstreamProtocol   string `json:"upstream_protocol"`
	TargetModel        string `json:"target_model"`
	Enabled            bool   `json:"enabled"`
	UpdatedAt          string `json:"updated_at"`
	CreatedAt          string `json:"created_at"`
}

type entry struct {
	ID                 string    `json:"id"`
	DownstreamProtocol string    `json:"downstream_protocol"`
	SupplierID         string    `json:"supplier_id"`
	TargetModel        string    `json:"target_model"`
	Enabled            bool      `json:"enabled"`
	UpdatedAt          time.Time `json:"updated_at"`
	CreatedAt          time.Time `json:"created_at"`
}

type SupplierResolver interface {
	Resolve(id string) (SupplierSnapshot, bool)
}

type SupplierSnapshot struct {
	ID        string
	Name      string
	Protocol  string
	BaseURL   string
	APIKey    string
	IsEnabled bool
}

type UpsertInput struct {
	ID                 string `json:"id"`
	DownstreamProtocol string `json:"downstream_protocol"`
	SupplierID         string `json:"supplier_id"`
	TargetModel        string `json:"target_model"`
	Enabled            bool   `json:"enabled"`
}

type Service struct {
	mu       sync.RWMutex
	path     string
	items    []entry
	lookup   SupplierResolver
	rootPath string
}

func NewService(root string, resolver SupplierResolver) (*Service, error) {
	storeDir := filepath.Join(root, ".route-policies")
	if err := os.MkdirAll(storeDir, 0o755); err != nil {
		return nil, err
	}
	svc := &Service{
		path:     filepath.Join(storeDir, "policies.json"),
		lookup:   resolver,
		rootPath: storeDir,
	}
	if err := svc.load(); err != nil {
		return nil, err
	}
	return svc, nil
}

func (s *Service) List() []Record {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]Record, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, s.toRecord(item))
	}
	slices.SortFunc(items, func(a, b Record) int {
		return strings.Compare(a.DownstreamProtocol, b.DownstreamProtocol)
	})
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

func (s *Service) Upsert(input UpsertInput) (Record, error) {
	downstream := normalizeProtocol(input.DownstreamProtocol)
	if downstream == "" {
		return Record{}, fmt.Errorf("downstream protocol is required")
	}
	if strings.TrimSpace(input.SupplierID) == "" {
		return Record{}, fmt.Errorf("supplier id is required")
	}
	if strings.TrimSpace(input.TargetModel) == "" {
		return Record{}, fmt.Errorf("target model is required")
	}
	supplier, ok := s.lookup.Resolve(input.SupplierID)
	if !ok {
		return Record{}, fmt.Errorf("supplier not found")
	}
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	index := -1
	for i, item := range s.items {
		if item.ID == strings.TrimSpace(input.ID) && input.ID != "" {
			index = i
			break
		}
		if item.DownstreamProtocol == downstream {
			index = i
		}
	}

	current := entry{
		ID:                 buildID(downstream),
		DownstreamProtocol: downstream,
		SupplierID:         supplier.ID,
		TargetModel:        strings.TrimSpace(input.TargetModel),
		Enabled:            input.Enabled,
		UpdatedAt:          now,
		CreatedAt:          now,
	}
	if index >= 0 {
		existing := s.items[index]
		current.ID = existing.ID
		current.CreatedAt = existing.CreatedAt
		s.items[index] = current
	} else {
		if strings.TrimSpace(input.ID) != "" {
			current.ID = strings.TrimSpace(input.ID)
		}
		s.items = append(s.items, current)
	}

	if err := s.saveLocked(); err != nil {
		return Record{}, err
	}
	return s.toRecord(current), nil
}

func (s *Service) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			s.items = defaultPolicies()
			return s.saveLocked()
		}
		return err
	}
	var items []entry
	if len(data) > 0 {
		if err := json.Unmarshal(data, &items); err != nil {
			return err
		}
	}
	if len(items) == 0 {
		items = defaultPolicies()
	}
	s.items = items
	return nil
}

func (s *Service) saveLocked() error {
	data, err := json.MarshalIndent(s.items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func (s *Service) toRecord(item entry) Record {
	record := Record{
		ID:                 item.ID,
		DownstreamProtocol: item.DownstreamProtocol,
		SupplierID:         item.SupplierID,
		TargetModel:        item.TargetModel,
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

func normalizeProtocol(raw string) string {
	value := strings.TrimSpace(strings.ToLower(raw))
	switch value {
	case "anthropic", "openai-chat", "openai-responses":
		return value
	default:
		return ""
	}
}

func buildID(downstream string) string {
	return "policy-" + downstream
}

func defaultPolicies() []entry {
	now := time.Now()
	return []entry{
		{
			ID:                 buildID("anthropic"),
			DownstreamProtocol: "anthropic",
			Enabled:            false,
			UpdatedAt:          now,
			CreatedAt:          now,
		},
		{
			ID:                 buildID("openai-chat"),
			DownstreamProtocol: "openai-chat",
			Enabled:            false,
			UpdatedAt:          now,
			CreatedAt:          now,
		},
		{
			ID:                 buildID("openai-responses"),
			DownstreamProtocol: "openai-responses",
			Enabled:            false,
			UpdatedAt:          now,
			CreatedAt:          now,
		},
	}
}

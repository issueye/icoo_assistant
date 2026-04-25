package supplier

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
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Protocol     string    `json:"protocol"`
	BaseURL      string    `json:"base_url"`
	APIKeyMasked string    `json:"api_key_masked"`
	Enabled      bool      `json:"enabled"`
	Description  string    `json:"description"`
	Models       []string  `json:"models"`
	Tags         []string  `json:"tags"`
	UpdatedAt    time.Time `json:"updated_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type entry struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Protocol    string    `json:"protocol"`
	BaseURL     string    `json:"base_url"`
	APIKey      string    `json:"api_key"`
	Enabled     bool      `json:"enabled"`
	Description string    `json:"description"`
	Models      []string  `json:"models"`
	Tags        []string  `json:"tags"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type UpsertInput struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Protocol    string `json:"protocol"`
	BaseURL     string `json:"base_url"`
	APIKey      string `json:"api_key"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
	Models      string `json:"models"`
	Tags        string `json:"tags"`
}

type Service struct {
	mu      sync.RWMutex
	rootDir string
	path    string
	items   []entry
}

func NewService(root string) (*Service, error) {
	storeDir := filepath.Join(root, ".suppliers")
	if err := os.MkdirAll(storeDir, 0o755); err != nil {
		return nil, err
	}
	svc := &Service{
		rootDir: storeDir,
		path:    filepath.Join(storeDir, "suppliers.json"),
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
		items = append(items, toRecord(item))
	}
	slices.SortFunc(items, func(a, b Record) int {
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})
	return items
}

func (s *Service) Upsert(input UpsertInput) (Record, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Record{}, fmt.Errorf("supplier name is required")
	}
	protocol := normalizeProtocol(input.Protocol)
	if protocol == "" {
		return Record{}, fmt.Errorf("supplier protocol is required")
	}
	baseURL := strings.TrimSpace(input.BaseURL)
	if baseURL == "" {
		return Record{}, fmt.Errorf("supplier base_url is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	id := strings.TrimSpace(input.ID)
	index := -1
	for i, item := range s.items {
		if item.ID == id && id != "" {
			index = i
			break
		}
	}

	current := entry{
		ID:          generateID(name),
		CreatedAt:   now,
		UpdatedAt:   now,
		Name:        name,
		Protocol:    protocol,
		BaseURL:     baseURL,
		APIKey:      strings.TrimSpace(input.APIKey),
		Enabled:     input.Enabled,
		Description: strings.TrimSpace(input.Description),
		Models:      splitCSVLike(input.Models),
		Tags:        splitCSVLike(input.Tags),
	}

	if index >= 0 {
		existing := s.items[index]
		current.ID = existing.ID
		current.CreatedAt = existing.CreatedAt
		if strings.TrimSpace(input.APIKey) == "" {
			current.APIKey = existing.APIKey
		}
		s.items[index] = current
	} else {
		if id != "" {
			current.ID = id
		}
		s.items = append(s.items, current)
	}

	if err := s.saveLocked(); err != nil {
		return Record{}, err
	}
	return toRecord(current), nil
}

func (s *Service) Delete(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("supplier id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	index := -1
	for i, item := range s.items {
		if item.ID == id {
			index = i
			break
		}
	}
	if index < 0 {
		return fmt.Errorf("supplier not found")
	}
	s.items = append(s.items[:index], s.items[index+1:]...)
	return s.saveLocked()
}

func (s *Service) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			s.items = defaultSuppliers()
			return s.saveLocked()
		}
		return err
	}
	if len(data) == 0 {
		s.items = defaultSuppliers()
		return s.saveLocked()
	}
	var items []entry
	if err := json.Unmarshal(data, &items); err != nil {
		return err
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

func toRecord(item entry) Record {
	return Record{
		ID:           item.ID,
		Name:         item.Name,
		Protocol:     item.Protocol,
		BaseURL:      item.BaseURL,
		APIKeyMasked: maskSecret(item.APIKey),
		Enabled:      item.Enabled,
		Description:  item.Description,
		Models:       slices.Clone(item.Models),
		Tags:         slices.Clone(item.Tags),
		UpdatedAt:    item.UpdatedAt,
		CreatedAt:    item.CreatedAt,
	}
}

func normalizeProtocol(raw string) string {
	value := strings.TrimSpace(strings.ToLower(raw))
	switch value {
	case "anthropic", "openai-chat", "openai-responses", "openai":
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

func defaultSuppliers() []entry {
	now := time.Now()
	return []entry{
		{
			ID:          "anthropic-default",
			Name:        "Anthropic Default",
			Protocol:    "anthropic",
			BaseURL:     "https://api.anthropic.com",
			Enabled:     true,
			Description: "Default Anthropic upstream profile for local gateway routing.",
			Models:      []string{"claude-sonnet-4", "claude-opus-4"},
			Tags:        []string{"official", "text"},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "openai-responses-default",
			Name:        "OpenAI Responses Default",
			Protocol:    "openai-responses",
			BaseURL:     "https://api.openai.com",
			Enabled:     true,
			Description: "Default OpenAI Responses profile for cross-protocol routing.",
			Models:      []string{"gpt-4.1", "gpt-4.1-mini"},
			Tags:        []string{"official", "responses"},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
}

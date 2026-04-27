package supplier

import (
	"fmt"
	"icoo_proxy/internal/consts"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type HealthRecord struct {
	SupplierID   string          `json:"supplier_id"`
	Status       string          `json:"status"`
	Message      string          `json:"message"`
	CheckedAt    string          `json:"checked_at"`
	StatusCode   int             `json:"status_code"`
	DurationMS   int64           `json:"duration_ms"`
	Reachable    bool            `json:"reachable"`
	Protocol     consts.Protocol `json:"protocol"`
	BaseURL      string          `json:"base_url"`
	SupplierName string          `json:"supplier_name"`
}

type HealthService struct {
	mu      sync.RWMutex
	client  *http.Client
	store   *Service
	results map[string]HealthRecord
}

func NewHealthService(store *Service) *HealthService {
	return &HealthService{
		client:  &http.Client{Timeout: 8 * time.Second},
		store:   store,
		results: make(map[string]HealthRecord),
	}
}

func (s *HealthService) List() []HealthRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]HealthRecord, 0, len(s.results))
	for _, item := range s.results {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].SupplierName < items[j].SupplierName
	})
	return items
}

func (s *HealthService) Check(id string) (HealthRecord, error) {
	supplier, ok := s.store.Resolve(strings.TrimSpace(id))
	if !ok {
		return HealthRecord{}, fmt.Errorf("supplier not found")
	}
	targetURL := strings.TrimRight(supplier.BaseURL, "/")
	start := time.Now()
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return HealthRecord{}, err
	}
	if userAgent := strings.TrimSpace(supplier.UserAgent); userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		record := HealthRecord{
			SupplierID:   supplier.ID,
			SupplierName: supplier.Name,
			Protocol:     supplier.Protocol,
			BaseURL:      supplier.BaseURL,
			Status:       "unreachable",
			Message:      err.Error(),
			CheckedAt:    time.Now().Format(time.RFC3339),
			DurationMS:   time.Since(start).Milliseconds(),
			Reachable:    false,
		}
		s.save(record)
		return record, nil
	}
	defer resp.Body.Close()

	record := HealthRecord{
		SupplierID:   supplier.ID,
		SupplierName: supplier.Name,
		Protocol:     supplier.Protocol,
		BaseURL:      supplier.BaseURL,
		StatusCode:   resp.StatusCode,
		CheckedAt:    time.Now().Format(time.RFC3339),
		DurationMS:   time.Since(start).Milliseconds(),
		Reachable:    resp.StatusCode > 0,
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		record.Status = "reachable"
		record.Message = "Base endpoint responded"
	} else {
		record.Status = "warning"
		record.Message = fmt.Sprintf("Unexpected status %d", resp.StatusCode)
	}
	s.save(record)
	return record, nil
}

func (s *HealthService) save(record HealthRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results[record.SupplierID] = record
}

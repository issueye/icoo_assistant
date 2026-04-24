package team

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	RequestStatusPending   = "pending"
	RequestStatusResponded = "responded"
)

type RequestRecord struct {
	RequestID         string    `json:"requestId"`
	FromID            string    `json:"fromId"`
	ToID              string    `json:"toId"`
	Kind              string    `json:"kind"`
	Body              string    `json:"body"`
	Status            string    `json:"status"`
	RootMessageID     string    `json:"rootMessageId"`
	ResponseMessageID string    `json:"responseMessageId,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type RequestFilter struct {
	Status string
	FromID string
	ToID   string
}

func (m *Manager) GetRequest(requestID string) (RequestRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.readRequestLocked(requestID)
}

func (m *Manager) ListRequests(filter RequestFilter, limit int) ([]RequestRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	items, err := m.listRequestsLocked(filter)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || len(items) <= limit {
		return items, nil
	}
	return items[len(items)-limit:], nil
}

func normalizeRequestID(requestID string) (string, error) {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return "", fmt.Errorf("request id required")
	}
	return requestID, nil
}

func normalizeRequestStatus(status string) (string, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "", RequestStatusPending:
		return RequestStatusPending, nil
	case RequestStatusResponded:
		return RequestStatusResponded, nil
	default:
		return "", fmt.Errorf("invalid request status %q", status)
	}
}

func (m *Manager) syncRequestFromMessageLocked(msg Message) error {
	if msg.Kind != "request" || strings.TrimSpace(msg.RequestID) == "" {
		return nil
	}
	record, err := m.requestRecordFromRootMessage(msg)
	if err != nil {
		return err
	}
	if existing, err := m.readRequestLocked(record.RequestID); err == nil {
		record.CreatedAt = existing.CreatedAt
		record.ResponseMessageID = existing.ResponseMessageID
		if existing.Status == RequestStatusResponded && existing.ResponseMessageID != "" {
			record.Status = existing.Status
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	return m.writeRequestLocked(record)
}

func (m *Manager) markRequestRespondedLocked(root Message, response Message) error {
	record, err := m.ensureRequestFromRootLocked(root)
	if err != nil {
		return err
	}
	record.Status = RequestStatusResponded
	record.ResponseMessageID = response.ID
	record.UpdatedAt = response.CreatedAt.UTC()
	if record.UpdatedAt.IsZero() {
		record.UpdatedAt = m.now().UTC()
	}
	return m.writeRequestLocked(record)
}

func (m *Manager) ensureRequestFromRootLocked(root Message) (RequestRecord, error) {
	requestID, err := normalizeRequestID(root.RequestID)
	if err != nil {
		return RequestRecord{}, err
	}
	record, err := m.readRequestLocked(requestID)
	if err == nil {
		return record, nil
	}
	if !os.IsNotExist(err) {
		return RequestRecord{}, err
	}
	record, err = m.requestRecordFromRootMessage(root)
	if err != nil {
		return RequestRecord{}, err
	}
	if err := m.writeRequestLocked(record); err != nil {
		return RequestRecord{}, err
	}
	return record, nil
}

func (m *Manager) requestRecordFromRootMessage(root Message) (RequestRecord, error) {
	requestID, err := normalizeRequestID(root.RequestID)
	if err != nil {
		return RequestRecord{}, err
	}
	createdAt := root.CreatedAt.UTC()
	if createdAt.IsZero() {
		createdAt = m.now().UTC()
	}
	return RequestRecord{
		RequestID:     requestID,
		FromID:        root.FromID,
		ToID:          root.ToID,
		Kind:          root.Kind,
		Body:          root.Body,
		Status:        RequestStatusPending,
		RootMessageID: root.ID,
		CreatedAt:     createdAt,
		UpdatedAt:     createdAt,
	}, nil
}

func (m *Manager) listRequestsLocked(filter RequestFilter) ([]RequestRecord, error) {
	status := strings.TrimSpace(filter.Status)
	if status != "" {
		var err error
		status, err = normalizeRequestStatus(status)
		if err != nil {
			return nil, err
		}
	}
	fromID := strings.TrimSpace(filter.FromID)
	if fromID != "" {
		var err error
		fromID, err = normalizeID(fromID)
		if err != nil {
			return nil, fmt.Errorf("invalid from id: %w", err)
		}
	}
	toID := strings.TrimSpace(filter.ToID)
	if toID != "" {
		var err error
		toID, err = normalizeID(toID)
		if err != nil {
			return nil, fmt.Errorf("invalid to id: %w", err)
		}
	}

	entries, err := os.ReadDir(m.RequestsDir)
	if err != nil {
		return nil, err
	}
	items := make([]RequestRecord, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "request_") || !strings.HasSuffix(name, ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(m.RequestsDir, name))
		if err != nil {
			return nil, err
		}
		var item RequestRecord
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, err
		}
		if status != "" && item.Status != status {
			continue
		}
		if fromID != "" && item.FromID != fromID {
			continue
		}
		if toID != "" && item.ToID != toID {
			continue
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].RequestID < items[j].RequestID
		}
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})
	return items, nil
}

func (m *Manager) readRequestLocked(requestID string) (RequestRecord, error) {
	requestID, err := normalizeRequestID(requestID)
	if err != nil {
		return RequestRecord{}, err
	}
	data, err := os.ReadFile(m.requestPath(requestID))
	if err != nil {
		return RequestRecord{}, err
	}
	var item RequestRecord
	if err := json.Unmarshal(data, &item); err != nil {
		return RequestRecord{}, err
	}
	return item, nil
}

func (m *Manager) writeRequestLocked(item RequestRecord) error {
	requestID, err := normalizeRequestID(item.RequestID)
	if err != nil {
		return err
	}
	status, err := normalizeRequestStatus(item.Status)
	if err != nil {
		return err
	}
	item.RequestID = requestID
	item.Status = status
	item.Body = strings.TrimSpace(item.Body)
	item.FromID, err = normalizeID(item.FromID)
	if err != nil {
		return fmt.Errorf("invalid from id: %w", err)
	}
	item.ToID, err = normalizeID(item.ToID)
	if err != nil {
		return fmt.Errorf("invalid to id: %w", err)
	}
	if strings.TrimSpace(item.Kind) == "" {
		item.Kind = "request"
	}
	if strings.TrimSpace(item.RootMessageID) == "" {
		return fmt.Errorf("root message id required")
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = m.now().UTC()
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = item.CreatedAt
	}
	data, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.requestPath(item.RequestID), append(data, '\n'), 0o644)
}

func (m *Manager) requestPath(requestID string) string {
	encoded := base64.RawURLEncoding.EncodeToString([]byte(requestID))
	return filepath.Join(m.RequestsDir, fmt.Sprintf("request_%s.json", encoded))
}

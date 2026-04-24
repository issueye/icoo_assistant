package task

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	StatusPending    = "pending"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
	StatusBlocked    = "blocked"
)

type Task struct {
	ID             string             `json:"id"`
	Title          string             `json:"title"`
	Status         string             `json:"status"`
	BlockedBy      []string           `json:"blockedBy,omitempty"`
	Owner          string             `json:"owner,omitempty"`
	Worktree       string             `json:"worktree,omitempty"`
	LastBackground *BackgroundContext `json:"lastBackground,omitempty"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
}

type BackgroundContext struct {
	JobID     string    `json:"jobId"`
	Status    string    `json:"status"`
	Command   string    `json:"command,omitempty"`
	Error     string    `json:"error,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateInput struct {
	ID        string
	Title     string
	Status    string
	BlockedBy []string
	Owner     string
	Worktree  string
}

type Manager struct {
	Dir string

	mu  sync.Mutex
	now func() time.Time
}

func DefaultDir(root string) string {
	return filepath.Join(root, ".tasks")
}

func NewManager(dir string) (*Manager, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil, fmt.Errorf("task dir required")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Manager{
		Dir: dir,
		now: time.Now,
	}, nil
}

func (m *Manager) Create(input CreateInput) (Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, err := m.buildCreateTask(input)
	if err != nil {
		return Task{}, err
	}
	if _, err := os.Stat(m.pathForID(task.ID)); err == nil {
		return Task{}, fmt.Errorf("task %s already exists", task.ID)
	} else if !os.IsNotExist(err) {
		return Task{}, err
	}
	if err := m.writeTaskLocked(task); err != nil {
		return Task{}, err
	}
	return task, nil
}

func (m *Manager) Get(id string) (Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.readTaskLocked(id)
}

func (m *Manager) List() ([]Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.listLocked()
}

func (m *Manager) Update(task Task) (Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	current, err := m.readTaskLocked(task.ID)
	if err != nil {
		return Task{}, err
	}
	updated, err := m.normalizeTask(task, current.CreatedAt)
	if err != nil {
		return Task{}, err
	}
	if updated.Title == "" {
		return Task{}, fmt.Errorf("title required")
	}
	if err := m.writeTaskLocked(updated); err != nil {
		return Task{}, err
	}
	if updated.Status == StatusCompleted {
		if err := m.unlockDependentsLocked(updated.ID); err != nil {
			return Task{}, err
		}
	}
	return updated, nil
}

func (m *Manager) UpdateStatus(id, status string) (Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, err := m.readTaskLocked(id)
	if err != nil {
		return Task{}, err
	}
	task.Status = status
	task, err = m.normalizeTask(task, task.CreatedAt)
	if err != nil {
		return Task{}, err
	}
	if err := m.writeTaskLocked(task); err != nil {
		return Task{}, err
	}
	if task.Status == StatusCompleted {
		if err := m.unlockDependentsLocked(task.ID); err != nil {
			return Task{}, err
		}
	}
	return task, nil
}

func (m *Manager) RecordBackground(id string, context BackgroundContext) (Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, err := m.readTaskLocked(id)
	if err != nil {
		return Task{}, err
	}
	normalized, err := normalizeBackgroundContext(context)
	if err != nil {
		return Task{}, err
	}
	item.LastBackground = &normalized
	item.UpdatedAt = m.now().UTC()
	if err := m.writeTaskLocked(item); err != nil {
		return Task{}, err
	}
	return item, nil
}

func (m *Manager) buildCreateTask(input CreateInput) (Task, error) {
	now := m.now().UTC()
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = fmt.Sprintf("task-%d", now.UnixNano())
	}
	task := Task{
		ID:        id,
		Title:     strings.TrimSpace(input.Title),
		Status:    strings.TrimSpace(input.Status),
		BlockedBy: input.BlockedBy,
		Owner:     strings.TrimSpace(input.Owner),
		Worktree:  strings.TrimSpace(input.Worktree),
		CreatedAt: now,
		UpdatedAt: now,
	}
	return m.normalizeTask(task, task.CreatedAt)
}

func (m *Manager) normalizeTask(task Task, createdAt time.Time) (Task, error) {
	id, err := normalizeID(task.ID)
	if err != nil {
		return Task{}, err
	}
	title := strings.TrimSpace(task.Title)
	if title == "" {
		return Task{}, fmt.Errorf("title required")
	}
	blockedBy, err := normalizeBlockedBy(task.BlockedBy)
	if err != nil {
		return Task{}, err
	}
	status, err := normalizeStatus(strings.TrimSpace(task.Status), blockedBy)
	if err != nil {
		return Task{}, err
	}
	now := m.now().UTC()
	if createdAt.IsZero() {
		createdAt = now
	}
	return Task{
		ID:             id,
		Title:          title,
		Status:         status,
		BlockedBy:      blockedBy,
		Owner:          strings.TrimSpace(task.Owner),
		Worktree:       strings.TrimSpace(task.Worktree),
		LastBackground: copyBackgroundContext(task.LastBackground),
		CreatedAt:      createdAt.UTC(),
		UpdatedAt:      now,
	}, nil
}

func normalizeBackgroundContext(context BackgroundContext) (BackgroundContext, error) {
	jobID, err := normalizeID(context.JobID)
	if err != nil {
		return BackgroundContext{}, err
	}
	status := strings.ToLower(strings.TrimSpace(context.Status))
	if status == "" {
		return BackgroundContext{}, fmt.Errorf("background status required")
	}
	updatedAt := context.UpdatedAt.UTC()
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}
	return BackgroundContext{
		JobID:     jobID,
		Status:    status,
		Command:   strings.TrimSpace(context.Command),
		Error:     strings.TrimSpace(context.Error),
		UpdatedAt: updatedAt,
	}, nil
}

func copyBackgroundContext(context *BackgroundContext) *BackgroundContext {
	if context == nil {
		return nil
	}
	copied := *context
	return &copied
}

func normalizeID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("id required")
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-', r == '_':
		default:
			return "", fmt.Errorf("invalid task id %q", id)
		}
	}
	return id, nil
}

func normalizeBlockedBy(ids []string) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	seen := map[string]struct{}{}
	result := make([]string, 0, len(ids))
	for _, raw := range ids {
		id, err := normalizeID(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid blockedBy id: %w", err)
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	sort.Strings(result)
	return result, nil
}

func normalizeStatus(status string, blockedBy []string) (string, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if len(blockedBy) > 0 {
		switch status {
		case "", StatusPending, StatusBlocked:
			return StatusBlocked, nil
		case StatusCompleted:
			return StatusCompleted, nil
		default:
			return "", fmt.Errorf("blocked task cannot use status %q", status)
		}
	}
	switch status {
	case "", StatusPending, StatusBlocked:
		return StatusPending, nil
	case StatusInProgress, StatusCompleted:
		return status, nil
	default:
		return "", fmt.Errorf("invalid status %q", status)
	}
}

func (m *Manager) listLocked() ([]Task, error) {
	entries, err := os.ReadDir(m.Dir)
	if err != nil {
		return nil, err
	}
	tasks := make([]Task, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "task_") || !strings.HasSuffix(name, ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(m.Dir, name))
		if err != nil {
			return nil, err
		}
		var task Task
		if err := json.Unmarshal(data, &task); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].CreatedAt.Equal(tasks[j].CreatedAt) {
			return tasks[i].ID < tasks[j].ID
		}
		return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
	})
	return tasks, nil
}

func (m *Manager) readTaskLocked(id string) (Task, error) {
	id, err := normalizeID(id)
	if err != nil {
		return Task{}, err
	}
	data, err := os.ReadFile(m.pathForID(id))
	if err != nil {
		return Task{}, err
	}
	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return Task{}, err
	}
	return task, nil
}

func (m *Manager) writeTaskLocked(task Task) error {
	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.pathForID(task.ID), append(data, '\n'), 0o644)
}

func (m *Manager) unlockDependentsLocked(completedID string) error {
	tasks, err := m.listLocked()
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if task.ID == completedID || len(task.BlockedBy) == 0 {
			continue
		}
		filtered := make([]string, 0, len(task.BlockedBy))
		changed := false
		for _, blockedID := range task.BlockedBy {
			if blockedID == completedID {
				changed = true
				continue
			}
			filtered = append(filtered, blockedID)
		}
		if !changed {
			continue
		}
		task.BlockedBy = filtered
		if len(task.BlockedBy) == 0 && task.Status == StatusBlocked {
			task.Status = StatusPending
		}
		task.UpdatedAt = m.now().UTC()
		if err := m.writeTaskLocked(task); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) pathForID(id string) string {
	return filepath.Join(m.Dir, fmt.Sprintf("task_%s.json", id))
}

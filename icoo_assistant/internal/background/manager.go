package background

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"icoo_assistant/internal/commandutil"
)

const (
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

type Job struct {
	ID         string     `json:"id"`
	Command    string     `json:"command"`
	Status     string     `json:"status"`
	TaskID     string     `json:"taskId,omitempty"`
	Owner      string     `json:"owner,omitempty"`
	Output     string     `json:"output,omitempty"`
	Error      string     `json:"error,omitempty"`
	StartedAt  time.Time  `json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
	NotifiedAt *time.Time `json:"notifiedAt,omitempty"`
}

type StartInput struct {
	ID      string
	Command string
	TaskID  string
	Owner   string
}

type Completion struct {
	JobID   string
	TaskID  string
	Status  string
	Summary string
}

type Manager struct {
	Dir     string
	Workdir string
	Timeout time.Duration

	mu  sync.Mutex
	now func() time.Time
}

func DefaultDir(root string) string {
	return filepath.Join(root, ".background")
}

func NewManager(dir, workdir string, timeout time.Duration) (*Manager, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil, fmt.Errorf("background dir required")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Manager{
		Dir:     dir,
		Workdir: workdir,
		Timeout: timeout,
		now:     time.Now,
	}, nil
}

func (m *Manager) Start(input StartInput) (Job, error) {
	command := strings.TrimSpace(input.Command)
	if command == "" {
		return Job{}, fmt.Errorf("command required")
	}
	if err := commandutil.Validate(command); err != nil {
		return Job{}, err
	}
	m.mu.Lock()
	job, err := m.buildJobLocked(input)
	if err != nil {
		m.mu.Unlock()
		return Job{}, err
	}
	if _, err := os.Stat(m.pathForID(job.ID)); err == nil {
		m.mu.Unlock()
		return Job{}, fmt.Errorf("background job %s already exists", job.ID)
	} else if !os.IsNotExist(err) {
		m.mu.Unlock()
		return Job{}, err
	}
	if err := m.writeJobLocked(job); err != nil {
		m.mu.Unlock()
		return Job{}, err
	}
	m.mu.Unlock()

	go m.run(job.ID, command)
	return job, nil
}

func (m *Manager) Get(id string) (Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.readJobLocked(id)
}

func (m *Manager) List() ([]Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.listLocked()
}

func (m *Manager) PollNotifications() ([]Completion, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	jobs, err := m.listLocked()
	if err != nil {
		return nil, err
	}
	completions := make([]Completion, 0)
	for _, job := range jobs {
		if job.Status == StatusRunning || job.NotifiedAt != nil {
			continue
		}
		now := m.now().UTC()
		job.NotifiedAt = &now
		if err := m.writeJobLocked(job); err != nil {
			return nil, err
		}
		completions = append(completions, Completion{
			JobID:   job.ID,
			TaskID:  job.TaskID,
			Status:  job.Status,
			Summary: renderSummary(job),
		})
	}
	return completions, nil
}

func (m *Manager) buildJobLocked(input StartInput) (Job, error) {
	now := m.now().UTC()
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = fmt.Sprintf("bg-%d", now.UnixNano())
	}
	if err := validateID(id); err != nil {
		return Job{}, err
	}
	return Job{
		ID:        id,
		Command:   strings.TrimSpace(input.Command),
		Status:    StatusRunning,
		TaskID:    strings.TrimSpace(input.TaskID),
		Owner:     strings.TrimSpace(input.Owner),
		StartedAt: now,
	}, nil
}

func validateID(id string) error {
	if id == "" {
		return fmt.Errorf("id required")
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-', r == '_':
		default:
			return fmt.Errorf("invalid background job id %q", id)
		}
	}
	return nil
}

func (m *Manager) run(id, command string) {
	timeout := m.Timeout
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	output, err := commandutil.Execute(ctx, m.Workdir, command)
	status := StatusCompleted
	errorText := ""
	if ctx.Err() == context.DeadlineExceeded {
		status = StatusFailed
		errorText = fmt.Sprintf("timeout after %s", timeout)
	} else if err != nil {
		status = StatusFailed
		errorText = err.Error()
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	job, readErr := m.readJobLocked(id)
	if readErr != nil {
		return
	}
	job.Output = output
	job.Status = status
	job.Error = errorText
	finished := m.now().UTC()
	job.FinishedAt = &finished
	_ = m.writeJobLocked(job)
}

func (m *Manager) listLocked() ([]Job, error) {
	entries, err := os.ReadDir(m.Dir)
	if err != nil {
		return nil, err
	}
	jobs := make([]Job, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "job_") || !strings.HasSuffix(name, ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(m.Dir, name))
		if err != nil {
			return nil, err
		}
		var job Job
		if err := json.Unmarshal(data, &job); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	sort.Slice(jobs, func(i, j int) bool {
		if jobs[i].StartedAt.Equal(jobs[j].StartedAt) {
			return jobs[i].ID < jobs[j].ID
		}
		return jobs[i].StartedAt.Before(jobs[j].StartedAt)
	})
	return jobs, nil
}

func (m *Manager) readJobLocked(id string) (Job, error) {
	if err := validateID(strings.TrimSpace(id)); err != nil {
		return Job{}, err
	}
	data, err := os.ReadFile(m.pathForID(id))
	if err != nil {
		return Job{}, err
	}
	var job Job
	if err := json.Unmarshal(data, &job); err != nil {
		return Job{}, err
	}
	return job, nil
}

func (m *Manager) writeJobLocked(job Job) error {
	data, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.pathForID(job.ID), append(data, '\n'), 0o644)
}

func (m *Manager) pathForID(id string) string {
	return filepath.Join(m.Dir, fmt.Sprintf("job_%s.json", id))
}

func renderSummary(job Job) string {
	var builder strings.Builder
	builder.WriteString("<background_result>")
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("job_id: %s\n", job.ID))
	if job.TaskID != "" {
		builder.WriteString(fmt.Sprintf("task_id: %s\n", job.TaskID))
	}
	builder.WriteString(fmt.Sprintf("status: %s\n", job.Status))
	builder.WriteString(fmt.Sprintf("command: %s\n", job.Command))
	if job.Error != "" {
		builder.WriteString(fmt.Sprintf("error: %s\n", job.Error))
	}
	builder.WriteString("output:\n")
	builder.WriteString(job.Output)
	builder.WriteString("\n</background_result>")
	return builder.String()
}

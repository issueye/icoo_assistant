package task

import (
	"strings"

	"icoo_assistant/internal/background"
)

type BackgroundLifecycleLink struct {
	Manager *Manager
}

func NewBackgroundLifecycleLink(manager *Manager) *BackgroundLifecycleLink {
	return &BackgroundLifecycleLink{Manager: manager}
}

func (l *BackgroundLifecycleLink) BeforeStart(job background.Job) error {
	if l == nil || l.Manager == nil || strings.TrimSpace(job.TaskID) == "" {
		return nil
	}
	item, err := l.Manager.Get(job.TaskID)
	if err != nil {
		return err
	}
	if item.Status == StatusPending {
		if _, err := l.Manager.UpdateStatus(job.TaskID, StatusInProgress); err != nil {
			return err
		}
	}
	_, err = l.Manager.RecordBackground(job.TaskID, BackgroundContext{
		JobID:     job.ID,
		Status:    job.Status,
		Command:   job.Command,
		UpdatedAt: job.StartedAt,
	})
	return err
}

func (l *BackgroundLifecycleLink) AfterFinish(job background.Job) error {
	if l == nil || l.Manager == nil || strings.TrimSpace(job.TaskID) == "" {
		return nil
	}
	if _, err := l.Manager.RecordBackground(job.TaskID, BackgroundContext{
		JobID:     job.ID,
		Status:    job.Status,
		Command:   job.Command,
		Error:     job.Error,
		UpdatedAt: job.FinishedAtValue(),
	}); err != nil {
		return err
	}
	if job.Status == background.StatusFailed {
		item, err := l.Manager.Get(job.TaskID)
		if err != nil {
			return err
		}
		if item.Status == StatusInProgress {
			_, err = l.Manager.UpdateStatus(job.TaskID, StatusPending)
			return err
		}
	}
	return nil
}

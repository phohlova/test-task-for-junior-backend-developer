package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title           string                      `json:"title"`
	Description     string                      `json:"description"`
	Status          string                      `json:"status"`
	RecurringConfig *taskdomain.RecurrenceConfig `json:"recurring_config,omitempty"`
}

type taskDTO struct {
	ID              int64                       `json:"id"`
	Title           string                      `json:"title"`
	Description     string                      `json:"description"`
	Status          string                      `json:"status"`
	CreatedAt       time.Time                   `json:"created_at"`
	UpdatedAt       time.Time                   `json:"updated_at"`
	RecurringConfig *taskdomain.RecurrenceConfig `json:"recurring_config,omitempty"`
	ParentTaskID    *int64                      `json:"parent_task_id,omitempty"`
	ScheduledFor    *time.Time                  `json:"scheduled_for,omitempty"`
}

func newTaskDTO(t *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:              t.ID,
		Title:           t.Title,
		Description:     t.Description,
		Status:          string(t.Status),
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
		RecurringConfig: t.RecurringConfig,
		ParentTaskID:    t.ParentTaskID,
		ScheduledFor:    t.ScheduledFor,
	}
}
package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	RecurringConfig *RecurrenceConfig `json:"recurring_config,omitempty"`
	ParentTaskID    *int64            `json:"parent_task_id,omitempty"`
	ScheduledFor    *time.Time        `json:"scheduled_for,omitempty"`
}

type TaskFilter struct {
	Status *Status `json:"status,omitempty"`
	Limit  int     `json:"limit,omitempty"`
	Offset int     `json:"offset,omitempty"`
}

func DefaultTaskFilter() TaskFilter {
	return TaskFilter{
		Limit:  20,
		Offset: 0,
	}
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

func (t *Task) IsRecurringTemplate() bool {
	return t.RecurringConfig != nil && t.ParentTaskID == nil
}

func (t *Task) IsRecurringInstance() bool {
	return t.ParentTaskID != nil && t.ScheduledFor != nil
}

func (t *Task) CloneAsInstance(instanceID int64, scheduledFor time.Time) *Task {
	normalized := NormalizeToUTCStartOfDay(scheduledFor)
	return &Task{
		Title:        t.Title,
		Description:  t.Description,
		Status:       StatusNew,
		ParentTaskID: &instanceID,
		ScheduledFor: &normalized,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
}

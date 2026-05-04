package scheduler

import (
	"context"
	"fmt"
	"time"

	"example.com/taskservice/internal/domain/task"
)

type RecurringRepository interface {
	GetActiveTemplatesByDate(ctx context.Context, date time.Time, limit, instanceID, clusterSize int) ([]*task.Task, error)
	InsertInstancesIdempotent(ctx context.Context, parentID int64, dates []time.Time) (int, error)
}

type WorkerConfig struct {
	TickInterval time.Duration
	Lookahead    time.Duration 
	BatchLimit   int
	InstanceID   int
	ClusterSize  int
}

type Worker struct {
	repo RecurringRepository
	reg  *task.GeneratorRegistry
	cfg  WorkerConfig
}

func NewWorker(repo RecurringRepository, reg *task.GeneratorRegistry, cfg WorkerConfig) *Worker {
	return &Worker{repo: repo, reg: reg, cfg: cfg}
}

func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.cfg.TickInterval)
	defer ticker.Stop()

	if err := w.processTick(ctx); err != nil {
		fmt.Printf("[scheduler] initial tick error: %v\n", err)
	}

	for {
		select {
		case <-ctx.Done():
			fmt.Println("[scheduler] worker stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := w.processTick(ctx); err != nil {
				fmt.Printf("[scheduler] tick error: %v\n", err)
			}
		}
	}
}

func (w *Worker) processTick(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	now := time.Now().UTC()
	windowEnd := now.Add(w.cfg.Lookahead)

	templates, err := w.repo.GetActiveTemplatesByDate(ctx, now, w.cfg.BatchLimit, w.cfg.InstanceID, w.cfg.ClusterSize)
	if err != nil {
		return fmt.Errorf("fetch templates: %w", err)
	}

	for _, tmpl := range templates {
		if tmpl.RecurringConfig == nil {
			continue
		}

		gen, err := w.reg.Get(tmpl.RecurringConfig.Type)
		if err != nil {
			fmt.Printf("[scheduler] unknown type %q for task %d\n", tmpl.RecurringConfig.Type, tmpl.ID)
			continue
		}

		dates, err := gen.Generate(*tmpl.RecurringConfig)
		if err != nil {
			fmt.Printf("[scheduler] generation failed for task %d: %v\n", tmpl.ID, err)
			continue
		}

		var windowDates []time.Time
		for _, d := range dates {
			if !d.Before(now) && !d.After(windowEnd) {
				windowDates = append(windowDates, d)
			}
		}
		if len(windowDates) == 0 {
			continue
		}

		inserted, err := w.repo.InsertInstancesIdempotent(ctx, tmpl.ID, windowDates)
		if err != nil {
			fmt.Printf("[scheduler] insert failed for task %d: %v\n", tmpl.ID, err)
			continue
		}
		if inserted > 0 {
			fmt.Printf("[scheduler] created %d instances for template %d\n", inserted, tmpl.ID)
		}
	}
	return nil
}
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, title, description, status, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.CreatedAt, task.UpdatedAt)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}
		return nil, err
	}

	return found, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			updated_at = $4
		WHERE id = $5
		RETURNING id, title, description, status, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.UpdatedAt, task.ID)
	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}
		return nil, err
	}

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at
		FROM tasks
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *Repository) GetActiveTemplatesByDate(ctx context.Context, asOf time.Time, limit, instanceID, clusterSize int) ([]*taskdomain.Task, error) {
	query := `
		SELECT id, title, description, status, created_at, updated_at, recurring_config
		FROM tasks
		WHERE parent_task_id IS NULL
		AND recurring_config IS NOT NULL
		AND (recurring_config->>'end_date')::timestamptz IS NULL 
			OR (recurring_config->>'end_date')::timestamptz >= $1
		AND id %% $3 = $4
		ORDER BY id
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, asOf, limit, clusterSize, instanceID%clusterSize)
	if err != nil {
		return nil, fmt.Errorf("query templates: %w", err)
	}
	defer rows.Close()

	var tasks []*taskdomain.Task
	for rows.Next() {
		var t taskdomain.Task
		var statusStr string
		var configJSON []byte

		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &statusStr, &t.CreatedAt, &t.UpdatedAt, &configJSON); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		t.Status = taskdomain.Status(statusStr)

		if len(configJSON) > 0 {
			var cfg taskdomain.RecurrenceConfig
			if err := json.Unmarshal(configJSON, &cfg); err != nil {
				return nil, fmt.Errorf("unmarshal config: %w", err)
			}
			t.RecurringConfig = &cfg
		}
		tasks = append(tasks, &t)
	}
	return tasks, rows.Err()
}

func (r *Repository) InsertInstancesIdempotent(ctx context.Context, parentID int64, dates []time.Time) (int, error) {
	if len(dates) == 0 {
		return 0, nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		WITH tmpl AS (SELECT title FROM tasks WHERE id = $1)
		INSERT INTO tasks (parent_task_id, scheduled_date, title, status, created_at, updated_at)
		SELECT $1, input.d, tmpl.title, 'new', NOW(), NOW()
		FROM tmpl, (SELECT unnest($2::timestamptz[]) AS d) AS input
		ON CONFLICT (parent_task_id, scheduled_date) DO NOTHING
	`

	res, err := tx.Exec(ctx, query, parentID, dates)
	if err != nil {
		return 0, fmt.Errorf("exec insert: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit transaction: %w", err)
	}

	return int(res.RowsAffected()), nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task      taskdomain.Task
		statusStr string
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&statusStr,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(statusStr)
	return &task, nil
}
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type ScheduleRepository struct {
	pool *pgxpool.Pool
}

func NewScheduleRepository(pool *pgxpool.Pool) *ScheduleRepository {
	return &ScheduleRepository{pool: pool}
}

func (r *ScheduleRepository) Create(ctx context.Context, s *taskdomain.Schedule) (*taskdomain.Schedule, error) {
	datesJSON, err := marshalDates(s.Dates)
	if err != nil {
		return nil, err
	}

	const query = `
		INSERT INTO task_schedules
		    (title, description, type, every_n_days, day_of_month, dates, parity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, title, description, type, every_n_days, day_of_month, dates, parity, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query,
		s.Title, s.Description, string(s.Type),
		zeroableInt(s.EveryNDays), zeroableInt(s.DayOfMonth),
		datesJSON, zeroableParity(s.Parity),
		s.CreatedAt, s.UpdatedAt,
	)

	return scanSchedule(row)
}

func (r *ScheduleRepository) GetByID(ctx context.Context, id int64) (*taskdomain.Schedule, error) {
	const query = `
		SELECT id, title, description, type, every_n_days, day_of_month, dates, parity, created_at, updated_at
		FROM task_schedules
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	s, err := scanSchedule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrScheduleNotFound
		}
		return nil, err
	}

	return s, nil
}

func (r *ScheduleRepository) Update(ctx context.Context, s *taskdomain.Schedule) (*taskdomain.Schedule, error) {
	datesJSON, err := marshalDates(s.Dates)
	if err != nil {
		return nil, err
	}

	const query = `
		UPDATE task_schedules
		SET title        = $1,
		    description  = $2,
		    type         = $3,
		    every_n_days = $4,
		    day_of_month = $5,
		    dates        = $6,
		    parity       = $7,
		    updated_at   = $8
		WHERE id = $9
		RETURNING id, title, description, type, every_n_days, day_of_month, dates, parity, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query,
		s.Title, s.Description, string(s.Type),
		zeroableInt(s.EveryNDays), zeroableInt(s.DayOfMonth),
		datesJSON, zeroableParity(s.Parity),
		s.UpdatedAt, s.ID,
	)

	updated, err := scanSchedule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrScheduleNotFound
		}
		return nil, err
	}

	return updated, nil
}

func (r *ScheduleRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM task_schedules WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return taskdomain.ErrScheduleNotFound
	}
	return nil
}

func (r *ScheduleRepository) List(ctx context.Context) ([]taskdomain.Schedule, error) {
	const query = `
		SELECT id, title, description, type, every_n_days, day_of_month, dates, parity, created_at, updated_at
		FROM task_schedules
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []taskdomain.Schedule
	for rows.Next() {
		s, err := scanSchedule(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *s)
	}

	return result, rows.Err()
}

type scheduleScanner interface {
	Scan(dest ...any) error
}

func scanSchedule(sc scheduleScanner) (*taskdomain.Schedule, error) {
	var (
		s          taskdomain.Schedule
		schedType  string
		everyNDays *int
		dayOfMonth *int
		datesRaw   []byte
		parity     *string
	)

	if err := sc.Scan(
		&s.ID, &s.Title, &s.Description, &schedType,
		&everyNDays, &dayOfMonth, &datesRaw, &parity,
		&s.CreatedAt, &s.UpdatedAt,
	); err != nil {
		return nil, err
	}

	s.Type = taskdomain.ScheduleType(schedType)

	if everyNDays != nil {
		s.EveryNDays = *everyNDays
	}
	if dayOfMonth != nil {
		s.DayOfMonth = *dayOfMonth
	}
	if parity != nil {
		s.Parity = taskdomain.Parity(*parity)
	}
	if len(datesRaw) > 0 && string(datesRaw) != "[]" && string(datesRaw) != "null" {
		var strs []string
		if err := json.Unmarshal(datesRaw, &strs); err != nil {
			return nil, err
		}
		for _, str := range strs {
			t, err := time.Parse(time.RFC3339, str)
			if err != nil {
				return nil, err
			}
			s.Dates = append(s.Dates, t)
		}
	}

	return &s, nil
}

func marshalDates(dates []time.Time) ([]byte, error) {
	if len(dates) == 0 {
		return []byte("[]"), nil
	}
	strs := make([]string, len(dates))
	for i, d := range dates {
		strs[i] = d.UTC().Format(time.RFC3339)
	}
	return json.Marshal(strs)
}

func zeroableInt(v int) *int {
	if v == 0 {
		return nil
	}
	return &v
}

func zeroableParity(p taskdomain.Parity) *string {
	if p == "" {
		return nil
	}
	s := string(p)
	return &s
}

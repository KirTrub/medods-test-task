package schedule

import (
	"context"
	"time"

	taskdomain "github.com/KirTrub/medods-test-task/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, s *taskdomain.Schedule) (*taskdomain.Schedule, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Schedule, error)
	Update(ctx context.Context, s *taskdomain.Schedule) (*taskdomain.Schedule, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Schedule, error)
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Schedule, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Schedule, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Schedule, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Schedule, error)
}

type CreateInput struct {
	Title        string
	Description  string
	Type         taskdomain.ScheduleType
	EveryNDays   int
	DayOfMonth   int
	Dates        []time.Time
	DeadlineDays int
	Parity       taskdomain.Parity
}

type UpdateInput struct {
	Title        string
	Description  string
	Type         taskdomain.ScheduleType
	EveryNDays   int
	DayOfMonth   int
	Dates        []time.Time
	DeadlineDays int
	Parity       taskdomain.Parity
}

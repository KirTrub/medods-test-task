package schedule

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "github.com/KirTrub/medods-test-task/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Schedule, error) {
	if err := validateInput(input.Title, input.Type, input.EveryNDays, input.DayOfMonth, input.Dates, input.Parity); err != nil {
		return nil, err
	}

	now := s.now()
	model := &taskdomain.Schedule{
		Title:       strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.Description),
		Type:        input.Type,
		EveryNDays:  input.EveryNDays,
		DayOfMonth:  input.DayOfMonth,
		Dates:       input.Dates,
		Parity:      input.Parity,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return s.repo.Create(ctx, model)
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Schedule, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Schedule, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	if err := validateInput(input.Title, input.Type, input.EveryNDays, input.DayOfMonth, input.Dates, input.Parity); err != nil {
		return nil, err
	}

	model := &taskdomain.Schedule{
		ID:          id,
		Title:       strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.Description),
		Type:        input.Type,
		EveryNDays:  input.EveryNDays,
		DayOfMonth:  input.DayOfMonth,
		Dates:       input.Dates,
		Parity:      input.Parity,
		UpdatedAt:   s.now(),
	}

	return s.repo.Update(ctx, model)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Schedule, error) {
	return s.repo.List(ctx)
}

func validateInput(
	title string,
	schedType taskdomain.ScheduleType,
	everyNDays int,
	dayOfMonth int,
	dates []time.Time,
	parity taskdomain.Parity,
) error {
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !schedType.Valid() {
		return fmt.Errorf("%w: invalid schedule type %q", ErrInvalidInput, schedType)
	}

	switch schedType {
	case taskdomain.ScheduleTypeDaily:
		if everyNDays <= 0 {
			return fmt.Errorf("%w: every_n_days must be a positive integer for daily schedule", ErrInvalidInput)
		}

	case taskdomain.ScheduleTypeMonthly:
		if dayOfMonth < 1 || dayOfMonth > 30 {
			return fmt.Errorf("%w: day_of_month must be between 1 and 30", ErrInvalidInput)
		}

	case taskdomain.ScheduleTypeSpecificDates:
		if len(dates) == 0 {
			return fmt.Errorf("%w: dates must not be empty for specific_dates schedule", ErrInvalidInput)
		}

	case taskdomain.ScheduleTypeParity:
		if !parity.Valid() {
			return fmt.Errorf("%w: parity must be \"even\" or \"odd\"", ErrInvalidInput)
		}
	}

	return nil
}

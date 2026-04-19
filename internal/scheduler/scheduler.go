package scheduler

import (
	"context"
	"log/slog"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type ScheduleRepository interface {
	List(ctx context.Context) ([]taskdomain.Schedule, error)
}

type TaskRepository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
}

type Scheduler struct {
	schedules ScheduleRepository
	tasks     TaskRepository
	logger    *slog.Logger
	now       func() time.Time
}

func New(schedules ScheduleRepository, tasks TaskRepository, logger *slog.Logger) *Scheduler {
	return &Scheduler{
		schedules: schedules,
		tasks:     tasks,
		logger:    logger,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	s.logger.Info("scheduler started")

	s.generateTasks(ctx)

	for {
		timer := time.NewTimer(untilNextMidnight(s.now()))

		select {
		case <-timer.C:
			s.generateTasks(ctx)
		case <-ctx.Done():
			timer.Stop()
			s.logger.Info("scheduler stopped")
			return
		}
	}
}

func (s *Scheduler) generateTasks(ctx context.Context) {
	today := s.now().Truncate(24 * time.Hour)
	s.logger.Info("generating scheduled tasks", "date", today.Format("2006-01-02"))

	schedules, err := s.schedules.List(ctx)
	if err != nil {
		s.logger.Error("failed to list schedules", "error", err)
		return
	}

	created := 0
	for _, sched := range schedules {
		if !sched.ShouldRunOn(today) {
			continue
		}

		schedID := sched.ID
		task := &taskdomain.Task{
			Title:       sched.Title,
			Description: sched.Description,
			Status:      taskdomain.StatusNew,
			ScheduleID:  &schedID,
			CreatedAt:   today,
			UpdatedAt:   today,
		}

		if _, err := s.tasks.Create(ctx, task); err != nil {
			s.logger.Error("failed to create task from schedule",
				"schedule_id", sched.ID,
				"error", err,
			)
			continue
		}

		created++
	}

	s.logger.Info("scheduled tasks generated", "count", created)
}

func untilNextMidnight(now time.Time) time.Duration {
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	return time.Until(tomorrow)
}

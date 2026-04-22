package scheduler

import (
	"context"
	"log/slog"
	"time"

	taskdomain "github.com/KirTrub/medods-test-task/internal/domain/task"
)

type ScheduleRepository interface {
	List(ctx context.Context) ([]taskdomain.Schedule, error)
}

type TaskRepository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	MarkOverdue(ctx context.Context, now time.Time) (int64, error)
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
	now := s.now()
	today := now.Truncate(24 * time.Hour)

	countMarked, err := s.tasks.MarkOverdue(ctx, now)
	if err != nil {
		s.logger.Error("failed to mark overdue tasks", "error", err)
	} else if countMarked > 0 {
		s.logger.Info("tasks marked as overdue", "count", countMarked)
	}

	schedules, _ := s.schedules.List(ctx)
	for _, sched := range schedules {
		if !sched.ShouldRunOn(today) {
			continue
		}

		var deadlineAt *time.Time
		if sched.DeadlineDays > 0 {
			t := today.AddDate(0, 0, sched.DeadlineDays)
			deadlineAt = &t
		}

		task := &taskdomain.Task{
			Title:       sched.Title,
			Description: sched.Description,
			Status:      taskdomain.StatusNew,
			ScheduleID:  &sched.ID,
			DeadlineAt:  deadlineAt,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		s.tasks.Create(ctx, task)
	}
}

func untilNextMidnight(now time.Time) time.Duration {
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	return time.Until(tomorrow)
}

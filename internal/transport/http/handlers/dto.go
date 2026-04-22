package handlers

import (
	"time"

	taskdomain "github.com/KirTrub/medods-test-task/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
}

type taskDTO struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	ScheduleID  *int64            `json:"schedule_id,omitempty"`
	DeadlineAt  *time.Time        `json:"deadline_at,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		DeadlineAt:  task.DeadlineAt,
		ScheduleID:  task.ScheduleID,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

type scheduleMutationDTO struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Type         string   `json:"type"`
	EveryNDays   *int     `json:"every_n_days,omitempty"`
	DayOfMonth   *int     `json:"day_of_month,omitempty"`
	Dates        []string `json:"dates,omitempty"` // format: "2006-01-02"
	Parity       *string  `json:"parity,omitempty"`
	DeadlineDays int      `json:"deadline_days,omitempty"`
}

type scheduleDTO struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Type         string    `json:"type"`
	EveryNDays   *int      `json:"every_n_days,omitempty"`
	DayOfMonth   *int      `json:"day_of_month,omitempty"`
	Dates        []string  `json:"dates,omitempty"`
	Parity       *string   `json:"parity,omitempty"`
	DeadlineDays int       `json:"deadline_days,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func newScheduleDTO(s *taskdomain.Schedule) scheduleDTO {
	dto := scheduleDTO{
		ID:           s.ID,
		Title:        s.Title,
		Description:  s.Description,
		Type:         string(s.Type),
		DeadlineDays: s.DeadlineDays,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}

	if s.EveryNDays != 0 {
		v := s.EveryNDays
		dto.EveryNDays = &v
	}
	if s.DayOfMonth != 0 {
		v := s.DayOfMonth
		dto.DayOfMonth = &v
	}
	if len(s.Dates) > 0 {
		dto.Dates = make([]string, len(s.Dates))
		for i, d := range s.Dates {
			dto.Dates[i] = d.UTC().Format("2006-01-02")
		}
	}
	if s.Parity != "" {
		v := string(s.Parity)
		dto.Parity = &v
	}

	return dto
}

package task

import "time"

type ScheduleType string

const (
	ScheduleTypeDaily         ScheduleType = "daily"
	ScheduleTypeMonthly       ScheduleType = "monthly"
	ScheduleTypeSpecificDates ScheduleType = "specific_dates"
	ScheduleTypeParity        ScheduleType = "parity"
)

func (t ScheduleType) Valid() bool {
	switch t {
	case ScheduleTypeDaily, ScheduleTypeMonthly, ScheduleTypeSpecificDates, ScheduleTypeParity:
		return true
	}
	return false
}

type Parity string

const (
	ParityEven Parity = "even"
	ParityOdd  Parity = "odd"
)

func (p Parity) Valid() bool {
	return p == ParityEven || p == ParityOdd
}

type Schedule struct {
	ID          int64
	Title       string
	Description string
	Type        ScheduleType

	EveryNDays int

	DayOfMonth int

	Dates []time.Time

	Parity Parity

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s *Schedule) ShouldRunOn(date time.Time) bool {
	date = date.UTC().Truncate(24 * time.Hour)

	switch s.Type {
	case ScheduleTypeDaily:
		if s.EveryNDays <= 0 {
			return false
		}
		start := s.CreatedAt.UTC().Truncate(24 * time.Hour)
		days := int(date.Sub(start).Hours() / 24)
		return days >= 0 && days%s.EveryNDays == 0

	case ScheduleTypeMonthly:
		return date.Day() == s.DayOfMonth

	case ScheduleTypeSpecificDates:
		for _, d := range s.Dates {
			if d.UTC().Truncate(24 * time.Hour).Equal(date) {
				return true
			}
		}
		return false

	case ScheduleTypeParity:
		day := date.Day()
		if s.Parity == ParityEven {
			return day%2 == 0
		}
		return day%2 != 0
	}

	return false
}

package recap

import (
	"fmt"
	"time"

	"yld/internal/domain"
)

// PeriodContaining returns the canonical half-open period containing date.
// Entry local dates are represented as UTC midnights, so period math never
// depends on the machine's timezone or on daylight-saving transitions.
func PeriodContaining(kind domain.PeriodKind, date time.Time) (domain.Period, error) {
	date = calendarDate(date)
	var start, end time.Time

	switch kind {
	case domain.PeriodMonth:
		start = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 1, 0)
	case domain.PeriodQuarter:
		startMonth := time.Month(((int(date.Month()) - 1) / 3 * 3) + 1)
		start = time.Date(date.Year(), startMonth, 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 3, 0)
	case domain.PeriodYear:
		start = time.Date(date.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(1, 0, 0)
	default:
		return domain.Period{}, fmt.Errorf("unsupported period kind %q", kind)
	}

	return domain.Period{Kind: kind, StartDate: start, EndDate: end}, nil
}

// PreviousPeriod returns the complete period immediately before date.
func PreviousPeriod(kind domain.PeriodKind, date time.Time) (domain.Period, error) {
	current, err := PeriodContaining(kind, date)
	if err != nil {
		return domain.Period{}, err
	}
	return PeriodContaining(kind, current.StartDate.AddDate(0, 0, -1))
}

// LatestEligiblePeriod returns the newest completed period whose grace window
// has elapsed as of date. A grace of three days makes an April monthly recap
// eligible on May 4: three complete days after April 30.
func LatestEligiblePeriod(kind domain.PeriodKind, date time.Time, graceDays int) (domain.Period, error) {
	if graceDays < 0 {
		return domain.Period{}, fmt.Errorf("grace days cannot be negative")
	}
	date = calendarDate(date)
	candidate, err := PreviousPeriod(kind, date)
	if err != nil {
		return domain.Period{}, err
	}
	eligibleOn := candidate.EndDate.AddDate(0, 0, graceDays)
	if date.Before(eligibleOn) {
		return PreviousPeriod(kind, candidate.StartDate)
	}
	return candidate, nil
}

func calendarDate(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

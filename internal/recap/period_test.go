package recap

import (
	"testing"
	"time"

	"yld/internal/domain"
)

func TestPeriodContaining(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		kind      domain.PeriodKind
		date      time.Time
		wantStart string
		wantEnd   string
	}{
		{"leap-year month", domain.PeriodMonth, date(2024, 2, 29), "2024-02-01", "2024-03-01"},
		{"fourth quarter", domain.PeriodQuarter, date(2026, 12, 31), "2026-10-01", "2027-01-01"},
		{"year", domain.PeriodYear, date(2026, 7, 1), "2026-01-01", "2027-01-01"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := PeriodContaining(test.kind, test.date)
			if err != nil {
				t.Fatal(err)
			}
			if formatDate(got.StartDate) != test.wantStart || formatDate(got.EndDate) != test.wantEnd {
				t.Fatalf("period = [%s, %s), want [%s, %s)", formatDate(got.StartDate), formatDate(got.EndDate), test.wantStart, test.wantEnd)
			}
		})
	}
}

func TestLatestEligiblePeriodHonorsGraceDays(t *testing.T) {
	t.Parallel()

	beforeGrace, err := LatestEligiblePeriod(domain.PeriodMonth, date(2026, 5, 3), 3)
	if err != nil {
		t.Fatal(err)
	}
	if got := formatDate(beforeGrace.StartDate); got != "2026-03-01" {
		t.Fatalf("before grace period = %s, want 2026-03-01", got)
	}

	afterGrace, err := LatestEligiblePeriod(domain.PeriodMonth, date(2026, 5, 4), 3)
	if err != nil {
		t.Fatal(err)
	}
	if got := formatDate(afterGrace.StartDate); got != "2026-04-01" {
		t.Fatalf("after grace period = %s, want 2026-04-01", got)
	}
}

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func formatDate(value time.Time) string { return value.Format(time.DateOnly) }

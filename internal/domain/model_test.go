package domain

import (
	"testing"
	"time"
)

func TestLocalDateUsesEntryTimezone(t *testing.T) {
	t.Parallel()

	occurredAt := time.Date(2026, time.March, 28, 23, 30, 0, 0, time.UTC)
	got, err := LocalDate(occurredAt, "Europe/Stockholm")
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, time.March, 29, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("LocalDate() = %v, want %v", got, want)
	}
}

func TestMetricUpdateLocksHistoricalMeaning(t *testing.T) {
	t.Parallel()

	current := Metric{
		ID:          "metric-1",
		Key:         "walking_distance",
		Name:        "Walking distance",
		Kind:        MetricKindNumber,
		Aggregation: AggregationSum,
		Unit:        "km",
	}
	next := current
	next.Aggregation = AggregationAverage

	if err := ValidateMetricUpdate(current, next, true); err == nil {
		t.Fatal("expected aggregation change to be rejected")
	}

	next = current
	next.Name = "Distance walked"
	if err := ValidateMetricUpdate(current, next, true); err != nil {
		t.Fatalf("expected display-name change to be allowed: %v", err)
	}
}

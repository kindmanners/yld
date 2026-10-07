package recap

import (
	"testing"
	"time"

	"yld/internal/domain"
)

func TestGenerateSnapshotAggregatesAndExcludesBoundaries(t *testing.T) {
	t.Parallel()

	period, err := PeriodContaining(domain.PeriodMonth, date(2026, time.March, 15))
	if err != nil {
		t.Fatal(err)
	}
	metrics := []domain.Metric{
		{ID: "steps", Key: "steps", Name: "Steps", Kind: domain.MetricKindInteger, Aggregation: domain.AggregationSum, Unit: "steps"},
		{ID: "weight", Key: "weight", Name: "Weight", Kind: domain.MetricKindNumber, Aggregation: domain.AggregationLatest, Unit: "kg"},
	}
	entries := []domain.Entry{
		entry("1", "steps", date(2026, time.March, 1), 1000),
		entry("2", "steps", date(2026, time.March, 31), 2500),
		entry("3", "steps", date(2026, time.April, 1), 9999),
		entry("4", "weight", date(2026, time.March, 2), 81.2),
		entry("5", "weight", date(2026, time.March, 30), 80.4),
	}

	got, err := GenerateSnapshot("recap-1", "user-1", "space-1", period, metrics, entries, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 2 {
		t.Fatalf("got %d recap items, want 2", len(got.Items))
	}
	if got.Items[0].MetricKey != "steps" || got.Items[0].Value != 3500 || got.Items[0].EntryCount != 2 {
		t.Fatalf("steps item = %+v", got.Items[0])
	}
	if got.Items[1].MetricKey != "weight" || got.Items[1].Value != 80.4 || got.Items[1].EntryCount != 2 {
		t.Fatalf("weight item = %+v", got.Items[1])
	}
}

func TestGenerateSnapshotRejectsOutOfScopeEntries(t *testing.T) {
	t.Parallel()

	period, _ := PeriodContaining(domain.PeriodMonth, date(2026, time.March, 15))
	metric := domain.Metric{ID: "steps", Key: "steps", Name: "Steps", Kind: domain.MetricKindInteger, Aggregation: domain.AggregationSum}
	outside := entry("1", "steps", date(2026, time.March, 1), 1000)
	outside.UserID = "someone-else"

	if _, err := GenerateSnapshot("recap-1", "user-1", "space-1", period, []domain.Metric{metric}, []domain.Entry{outside}, time.Now()); err == nil {
		t.Fatal("expected out-of-scope entry to be rejected")
	}
}

func entry(id domain.ID, metricID domain.ID, localDate time.Time, value float64) domain.Entry {
	return domain.Entry{
		ID:         id,
		SpaceID:    "space-1",
		UserID:     "user-1",
		MetricID:   metricID,
		OccurredAt: localDate.Add(12 * time.Hour),
		LocalDate:  localDate,
		Timezone:   "UTC",
		Value:      value,
		CreatedAt:  localDate.Add(13 * time.Hour),
	}
}

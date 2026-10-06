package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"yld/internal/domain"
)

func TestRepositoryRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository, err := Open(ctx, filepath.Join(t.TempDir(), "yld.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()

	if err := repository.EnsurePersonalProfile(ctx, "user-1", "space-1", "alex", "Alex", "Europe/Stockholm"); err != nil {
		t.Fatal(err)
	}
	spaceID := domain.ID("space-1")
	metric := domain.Metric{
		ID: "metric-1", OwnerSpaceID: &spaceID, Key: "steps", Name: "Steps",
		Kind: domain.MetricKindInteger, Aggregation: domain.AggregationSum, Unit: "steps",
	}
	if err := repository.CreateMetric(ctx, metric); err != nil {
		t.Fatal(err)
	}
	loadedMetric, err := repository.MetricByKey(ctx, spaceID, "steps")
	if err != nil {
		t.Fatal(err)
	}
	if loadedMetric.ID != metric.ID {
		t.Fatalf("metric id = %q, want %q", loadedMetric.ID, metric.ID)
	}

	localDate := time.Date(2026, time.March, 29, 0, 0, 0, 0, time.UTC)
	entry := domain.Entry{
		ID: "entry-1", UserID: "user-1", SpaceID: spaceID, MetricID: metric.ID,
		OccurredAt: time.Date(2026, time.March, 28, 23, 30, 0, 0, time.UTC),
		LocalDate:  localDate, Timezone: "Europe/Stockholm", Value: 1234, Source: "manual",
	}
	if err := repository.AddEntry(ctx, entry); err != nil {
		t.Fatal(err)
	}
	entries, err := repository.ListEntries(ctx, "user-1", spaceID, localDate, localDate.AddDate(0, 0, 1))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Value != 1234 || entries[0].Timezone != "Europe/Stockholm" {
		t.Fatalf("entries = %+v", entries)
	}
}

func TestOpenAppliesMigrationsOnlyOnce(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "yld.db")
	first, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
}

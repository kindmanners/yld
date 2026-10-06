package recap

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"yld/internal/domain"
)

type accumulator struct {
	metric      domain.Metric
	count       int
	sum         float64
	minimum     float64
	maximum     float64
	latestValue float64
	latestAt    time.Time
	latestID    domain.ID
}

// GenerateSnapshot aggregates already-loaded entries into an immutable recap.
// It rejects out-of-scope entries instead of trusting every caller to filter
// authorization correctly.
func GenerateSnapshot(
	id domain.ID,
	userID domain.ID,
	spaceID domain.ID,
	period domain.Period,
	metrics []domain.Metric,
	entries []domain.Entry,
	generatedAt time.Time,
) (domain.Recap, error) {
	if id == "" || userID == "" || spaceID == "" {
		return domain.Recap{}, errors.New("recap id, user id, and space id are required")
	}
	if generatedAt.IsZero() {
		return domain.Recap{}, errors.New("generation time is required")
	}
	if err := period.Validate(); err != nil {
		return domain.Recap{}, err
	}

	metricByID := make(map[domain.ID]domain.Metric, len(metrics))
	for _, metric := range metrics {
		if err := metric.Validate(); err != nil {
			return domain.Recap{}, fmt.Errorf("validate metric %q: %w", metric.ID, err)
		}
		metricByID[metric.ID] = metric
	}

	accumulators := make(map[domain.ID]*accumulator)
	var sourceUpdatedAt time.Time
	for _, entry := range entries {
		if entry.UserID != userID || entry.SpaceID != spaceID {
			return domain.Recap{}, fmt.Errorf("entry %q is outside recap scope", entry.ID)
		}
		if entry.DeletedAt != nil || entry.LocalDate.Before(period.StartDate) || !entry.LocalDate.Before(period.EndDate) {
			continue
		}
		if err := entry.Validate(); err != nil {
			return domain.Recap{}, fmt.Errorf("validate entry %q: %w", entry.ID, err)
		}
		metric, ok := metricByID[entry.MetricID]
		if !ok {
			return domain.Recap{}, fmt.Errorf("entry %q references unknown metric %q", entry.ID, entry.MetricID)
		}

		acc := accumulators[metric.ID]
		if acc == nil {
			acc = &accumulator{metric: metric}
			accumulators[metric.ID] = acc
		}
		acc.add(entry)

		updatedAt := entry.UpdatedAt
		if updatedAt.IsZero() {
			updatedAt = entry.CreatedAt
		}
		if updatedAt.After(sourceUpdatedAt) {
			sourceUpdatedAt = updatedAt
		}
	}

	items := make([]domain.RecapItem, 0, len(accumulators))
	for _, acc := range accumulators {
		items = append(items, domain.RecapItem{
			MetricID:   acc.metric.ID,
			MetricKey:  acc.metric.Key,
			MetricName: acc.metric.Name,
			Unit:       acc.metric.Unit,
			Value:      acc.value(),
			EntryCount: acc.count,
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].MetricKey < items[j].MetricKey })

	return domain.Recap{
		ID:              id,
		UserID:          userID,
		SpaceID:         spaceID,
		Period:          period,
		GeneratedAt:     generatedAt,
		SourceUpdatedAt: sourceUpdatedAt,
		Items:           items,
	}, nil
}

func (a *accumulator) add(entry domain.Entry) {
	if a.count == 0 {
		a.minimum = entry.Value
		a.maximum = entry.Value
	}
	a.count++
	a.sum += entry.Value
	if entry.Value < a.minimum {
		a.minimum = entry.Value
	}
	if entry.Value > a.maximum {
		a.maximum = entry.Value
	}
	if entry.OccurredAt.After(a.latestAt) || (entry.OccurredAt.Equal(a.latestAt) && entry.ID > a.latestID) {
		a.latestAt = entry.OccurredAt
		a.latestID = entry.ID
		a.latestValue = entry.Value
	}
}

func (a *accumulator) value() float64 {
	switch a.metric.Aggregation {
	case domain.AggregationSum:
		return a.sum
	case domain.AggregationAverage:
		return a.sum / float64(a.count)
	case domain.AggregationMinimum:
		return a.minimum
	case domain.AggregationMaximum:
		return a.maximum
	case domain.AggregationLatest:
		return a.latestValue
	case domain.AggregationCount:
		return float64(a.count)
	default:
		panic("metric aggregation was validated")
	}
}

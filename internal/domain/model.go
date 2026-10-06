package domain

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

// ID is an opaque identifier. The storage layer is responsible for generating
// globally unique values.
type ID string

type MetricKind string

const (
	MetricKindNumber   MetricKind = "number"
	MetricKindInteger  MetricKind = "integer"
	MetricKindDuration MetricKind = "duration"
)

type Aggregation string

const (
	AggregationSum     Aggregation = "sum"
	AggregationAverage Aggregation = "average"
	AggregationMinimum Aggregation = "minimum"
	AggregationMaximum Aggregation = "maximum"
	AggregationLatest  Aggregation = "latest"
	AggregationCount   Aggregation = "count"
)

var metricKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,63}$`)

// Metric defines how entries are recorded and summarized. A nil OwnerSpaceID
// marks a built-in metric; custom metrics belong to a space.
type Metric struct {
	ID           ID
	OwnerSpaceID *ID
	Key          string
	Name         string
	Category     string
	Kind         MetricKind
	Aggregation  Aggregation
	Unit         string
	Archived     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (m Metric) Validate() error {
	if m.ID == "" {
		return errors.New("metric id is required")
	}
	if !metricKeyPattern.MatchString(m.Key) {
		return errors.New("metric key must start with a letter and contain only lowercase letters, digits, or underscores")
	}
	if strings.TrimSpace(m.Name) == "" {
		return errors.New("metric name is required")
	}
	switch m.Kind {
	case MetricKindNumber, MetricKindInteger, MetricKindDuration:
	default:
		return fmt.Errorf("unsupported metric kind %q", m.Kind)
	}
	switch m.Aggregation {
	case AggregationSum, AggregationAverage, AggregationMinimum, AggregationMaximum, AggregationLatest, AggregationCount:
	default:
		return fmt.Errorf("unsupported aggregation %q", m.Aggregation)
	}
	return nil
}

// ValidateMetricUpdate protects the parts of a metric definition that would
// reinterpret existing history. Display metadata may still change.
func ValidateMetricUpdate(current, next Metric, hasEntries bool) error {
	if current.ID != next.ID {
		return errors.New("metric id cannot change")
	}
	if err := next.Validate(); err != nil {
		return err
	}
	if !hasEntries {
		return nil
	}
	if current.Kind != next.Kind {
		return errors.New("metric kind cannot change after entries exist")
	}
	if current.Aggregation != next.Aggregation {
		return errors.New("metric aggregation cannot change after entries exist")
	}
	if current.Unit != next.Unit {
		return errors.New("metric unit cannot change after entries exist; create a new metric for conversions")
	}
	return nil
}

// Entry is one observation. LocalDate and Timezone are persisted at write time
// so later timezone changes cannot rewrite historical recap boundaries.
type Entry struct {
	ID         ID
	SpaceID    ID
	UserID     ID
	MetricID   ID
	ActivityID *ID
	OccurredAt time.Time
	LocalDate  time.Time
	Timezone   string
	Value      float64
	Note       string
	Source     string
	ExternalID string
	CreatedAt  time.Time
	DeletedAt  *time.Time
}

func (e Entry) Validate() error {
	if e.ID == "" || e.SpaceID == "" || e.UserID == "" || e.MetricID == "" {
		return errors.New("entry id, space id, user id, and metric id are required")
	}
	if e.OccurredAt.IsZero() || e.LocalDate.IsZero() {
		return errors.New("occurred time and local date are required")
	}
	if strings.TrimSpace(e.Timezone) == "" {
		return errors.New("entry timezone is required")
	}
	if math.IsNaN(e.Value) || math.IsInf(e.Value, 0) {
		return errors.New("entry value must be finite")
	}
	if e.LocalDate.Hour() != 0 || e.LocalDate.Minute() != 0 || e.LocalDate.Second() != 0 || e.LocalDate.Nanosecond() != 0 {
		return errors.New("local date must be normalized to midnight")
	}
	return nil
}

func LocalDate(occurredAt time.Time, timezone string) (time.Time, error) {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("load timezone %q: %w", timezone, err)
	}
	local := occurredAt.In(location)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC), nil
}

type PeriodKind string

const (
	PeriodMonth   PeriodKind = "month"
	PeriodQuarter PeriodKind = "quarter"
	PeriodYear    PeriodKind = "year"
)

type Period struct {
	Kind      PeriodKind
	StartDate time.Time
	EndDate   time.Time
}

func (p Period) Validate() error {
	if p.Kind != PeriodMonth && p.Kind != PeriodQuarter && p.Kind != PeriodYear {
		return fmt.Errorf("unsupported period kind %q", p.Kind)
	}
	if p.StartDate.IsZero() || p.EndDate.IsZero() || !p.StartDate.Before(p.EndDate) {
		return errors.New("period must have a non-empty half-open date range")
	}
	return nil
}

type RecapItem struct {
	MetricID   ID
	MetricKey  string
	MetricName string
	Unit       string
	Value      float64
	EntryCount int
}

// Recap is an immutable snapshot once it has been sent. SourceUpdatedAt is the
// data watermark used to show whether entries changed after generation.
type Recap struct {
	ID              ID
	UserID          ID
	SpaceID         ID
	Period          Period
	GeneratedAt     time.Time
	SentAt          *time.Time
	SourceUpdatedAt time.Time
	Items           []RecapItem
}

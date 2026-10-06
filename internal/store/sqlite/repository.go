package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"yld/internal/domain"
	"yld/internal/store"
)

type Repository struct {
	db *sql.DB
}

func Open(ctx context.Context, path string) (*Repository, error) {
	if path == "" {
		return nil, errors.New("database path is required")
	}

	var dsn string
	if path == ":memory:" {
		dsn = "file:yld-memory?mode=memory&cache=shared&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"
	} else {
		absolute, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("resolve database path: %w", err)
		}
		u := url.URL{Scheme: "file", Path: absolute}
		query := u.Query()
		query.Add("_pragma", "foreign_keys(1)")
		query.Add("_pragma", "busy_timeout(5000)")
		u.RawQuery = query.Encode()
		dsn = u.String()
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite database: %w", err)
	}
	if err := store.ApplyMigrations(ctx, db); err != nil {
		db.Close()
		return nil, err
	}
	return &Repository{db: db}, nil
}

func (r *Repository) Close() error { return r.db.Close() }

func (r *Repository) EnsurePersonalProfile(
	ctx context.Context,
	userID domain.ID,
	spaceID domain.ID,
	handle string,
	displayName string,
	timezone string,
) error {
	if userID == "" || spaceID == "" || handle == "" || displayName == "" || timezone == "" {
		return errors.New("profile ids, handle, display name, and timezone are required")
	}
	if _, err := time.LoadLocation(timezone); err != nil {
		return fmt.Errorf("load timezone %q: %w", timezone, err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := formatTime(time.Now().UTC())
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO users(id, handle, display_name, timezone, password_hash, created_at)
		VALUES (?, ?, ?, ?, '!', ?)
		ON CONFLICT(id) DO UPDATE SET
			handle = excluded.handle,
			display_name = excluded.display_name,
			timezone = excluded.timezone`, userID, handle, displayName, timezone, now); err != nil {
		return fmt.Errorf("upsert user: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO spaces(id, name, is_personal, created_at)
		VALUES (?, ?, 1, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name`, spaceID, displayName+"'s life", now); err != nil {
		return fmt.Errorf("upsert personal space: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO space_members(space_id, user_id, role, joined_at)
		VALUES (?, ?, 'owner', ?)
		ON CONFLICT(space_id, user_id) DO UPDATE SET role = 'owner'`, spaceID, userID, now); err != nil {
		return fmt.Errorf("upsert personal membership: %w", err)
	}
	return tx.Commit()
}

func (r *Repository) UserTimezone(ctx context.Context, userID domain.ID) (string, error) {
	var timezone string
	if err := r.db.QueryRowContext(ctx, "SELECT timezone FROM users WHERE id = ?", userID).Scan(&timezone); err != nil {
		return "", fmt.Errorf("load user timezone: %w", err)
	}
	return timezone, nil
}

func (r *Repository) CreateMetric(ctx context.Context, metric domain.Metric) error {
	if err := metric.Validate(); err != nil {
		return err
	}
	now := time.Now().UTC()
	if metric.CreatedAt.IsZero() {
		metric.CreatedAt = now
	}
	if metric.UpdatedAt.IsZero() {
		metric.UpdatedAt = now
	}
	var owner any
	if metric.OwnerSpaceID != nil {
		owner = *metric.OwnerSpaceID
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO metrics(id, owner_space_id, key, name, category, kind, aggregation, unit, archived, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		metric.ID, owner, metric.Key, metric.Name, metric.Category, metric.Kind, metric.Aggregation, metric.Unit,
		boolInt(metric.Archived), formatTime(metric.CreatedAt), formatTime(metric.UpdatedAt))
	if err != nil {
		return fmt.Errorf("create metric: %w", err)
	}
	return nil
}

func (r *Repository) MetricByKey(ctx context.Context, spaceID domain.ID, key string) (domain.Metric, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, owner_space_id, key, name, category, kind, aggregation, unit, archived, created_at, updated_at
		FROM metrics
		WHERE key = ? AND (owner_space_id = ? OR owner_space_id IS NULL)
		ORDER BY owner_space_id IS NULL
		LIMIT 1`, key, spaceID)
	metric, err := scanMetric(row)
	if err != nil {
		return domain.Metric{}, fmt.Errorf("load metric %q: %w", key, err)
	}
	return metric, nil
}

func (r *Repository) ListMetrics(ctx context.Context, spaceID domain.ID) ([]domain.Metric, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, owner_space_id, key, name, category, kind, aggregation, unit, archived, created_at, updated_at
		FROM metrics
		WHERE owner_space_id = ? OR owner_space_id IS NULL
		ORDER BY key`, spaceID)
	if err != nil {
		return nil, fmt.Errorf("list metrics: %w", err)
	}
	defer rows.Close()

	var metrics []domain.Metric
	for rows.Next() {
		metric, err := scanMetric(rows)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}
	return metrics, rows.Err()
}

func (r *Repository) AddEntry(ctx context.Context, entry domain.Entry) error {
	if err := entry.Validate(); err != nil {
		return err
	}
	now := time.Now().UTC()
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = now
	}
	if entry.UpdatedAt.IsZero() {
		entry.UpdatedAt = entry.CreatedAt
	}
	var activity any
	if entry.ActivityID != nil {
		activity = *entry.ActivityID
	}
	var external any
	if entry.ExternalID != "" {
		external = entry.ExternalID
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO entries(
			id, space_id, user_id, metric_id, activity_id, occurred_at, local_date, timezone,
			value, note, source, external_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.ID, entry.SpaceID, entry.UserID, entry.MetricID, activity,
		formatTime(entry.OccurredAt), entry.LocalDate.Format(time.DateOnly), entry.Timezone,
		entry.Value, entry.Note, entry.Source, external, formatTime(entry.CreatedAt), formatTime(entry.UpdatedAt))
	if err != nil {
		return fmt.Errorf("add entry: %w", err)
	}
	return nil
}

func (r *Repository) ListEntries(
	ctx context.Context,
	userID domain.ID,
	spaceID domain.ID,
	startDate time.Time,
	endDate time.Time,
) ([]domain.Entry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, space_id, user_id, metric_id, activity_id, occurred_at, local_date, timezone,
		       value, note, source, external_id, created_at, updated_at, deleted_at
		FROM entries
		WHERE user_id = ? AND space_id = ? AND local_date >= ? AND local_date < ? AND deleted_at IS NULL
		ORDER BY local_date, occurred_at, id`,
		userID, spaceID, startDate.Format(time.DateOnly), endDate.Format(time.DateOnly))
	if err != nil {
		return nil, fmt.Errorf("list entries: %w", err)
	}
	defer rows.Close()

	var entries []domain.Entry
	for rows.Next() {
		entry, err := scanEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanMetric(row scanner) (domain.Metric, error) {
	var metric domain.Metric
	var owner sql.NullString
	var archived int
	var createdAt, updatedAt string
	if err := row.Scan(
		&metric.ID, &owner, &metric.Key, &metric.Name, &metric.Category, &metric.Kind,
		&metric.Aggregation, &metric.Unit, &archived, &createdAt, &updatedAt,
	); err != nil {
		return domain.Metric{}, err
	}
	if owner.Valid {
		ownerID := domain.ID(owner.String)
		metric.OwnerSpaceID = &ownerID
	}
	metric.Archived = archived != 0
	var err error
	if metric.CreatedAt, err = parseTime(createdAt); err != nil {
		return domain.Metric{}, err
	}
	if metric.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return domain.Metric{}, err
	}
	return metric, nil
}

func scanEntry(row scanner) (domain.Entry, error) {
	var entry domain.Entry
	var activity, external, deleted sql.NullString
	var occurredAt, localDate, createdAt, updatedAt string
	if err := row.Scan(
		&entry.ID, &entry.SpaceID, &entry.UserID, &entry.MetricID, &activity,
		&occurredAt, &localDate, &entry.Timezone, &entry.Value, &entry.Note, &entry.Source,
		&external, &createdAt, &updatedAt, &deleted,
	); err != nil {
		return domain.Entry{}, err
	}
	if activity.Valid {
		id := domain.ID(activity.String)
		entry.ActivityID = &id
	}
	if external.Valid {
		entry.ExternalID = external.String
	}
	var err error
	if entry.OccurredAt, err = parseTime(occurredAt); err != nil {
		return domain.Entry{}, err
	}
	if entry.LocalDate, err = time.Parse(time.DateOnly, localDate); err != nil {
		return domain.Entry{}, err
	}
	if entry.CreatedAt, err = parseTime(createdAt); err != nil {
		return domain.Entry{}, err
	}
	if entry.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return domain.Entry{}, err
	}
	if deleted.Valid {
		value, err := parseTime(deleted.String)
		if err != nil {
			return domain.Entry{}, err
		}
		entry.DeletedAt = &value
	}
	return entry, nil
}

func formatTime(value time.Time) string { return value.UTC().Format(time.RFC3339Nano) }

func parseTime(value string) (time.Time, error) { return time.Parse(time.RFC3339Nano, value) }

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

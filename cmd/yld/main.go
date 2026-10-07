// Command yld is the command-line interface for Your Life's Database.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"yld/internal/domain"
	"yld/internal/recap"
	"yld/internal/store/sqlite"
)

const (
	defaultUserID  domain.ID = "user-personal"
	defaultSpaceID domain.ID = "space-personal"
)

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	global := flag.NewFlagSet("yld", flag.ContinueOnError)
	global.SetOutput(stderr)
	databasePath := global.String("db", "yld.db", "path to the SQLite database")
	global.Usage = func() { printUsage(stderr) }
	if err := global.Parse(args); err != nil {
		return err
	}
	remaining := global.Args()
	if len(remaining) == 0 {
		printUsage(stderr)
		return errors.New("a command is required")
	}

	repository, err := sqlite.Open(ctx, *databasePath)
	if err != nil {
		return err
	}
	defer repository.Close()

	switch remaining[0] {
	case "init":
		return runInit(ctx, repository, remaining[1:], stdout, stderr)
	case "metric":
		return runMetric(ctx, repository, remaining[1:], stdout, stderr)
	case "entry":
		return runEntry(ctx, repository, remaining[1:], stdout, stderr)
	case "recap":
		return runRecap(ctx, repository, remaining[1:], stdout, stderr)
	case "help", "-h", "--help":
		printUsage(stdout)
		return nil
	default:
		return fmt.Errorf("unknown command %q", remaining[0])
	}
}

func runInit(ctx context.Context, repository *sqlite.Repository, args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	flags.SetOutput(stderr)
	handle := flags.String("handle", "me", "short account handle")
	name := flags.String("name", "Me", "display name")
	timezone := flags.String("timezone", "UTC", "IANA timezone, for example Europe/Stockholm")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("init does not accept positional arguments")
	}
	if err := repository.EnsurePersonalProfile(ctx, defaultUserID, defaultSpaceID, *handle, *name, *timezone); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Initialized personal database for %s in %s.\n", *name, *timezone)
	return nil
}

func runMetric(ctx context.Context, repository *sqlite.Repository, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 || args[0] != "add" {
		return errors.New("usage: yld metric add --key KEY --name NAME --kind KIND --aggregation AGGREGATION [--unit UNIT]")
	}
	flags := flag.NewFlagSet("metric add", flag.ContinueOnError)
	flags.SetOutput(stderr)
	key := flags.String("key", "", "stable lowercase metric key")
	name := flags.String("name", "", "display name")
	category := flags.String("category", "", "optional category")
	kind := flags.String("kind", string(domain.MetricKindNumber), "number, integer, or duration")
	aggregation := flags.String("aggregation", string(domain.AggregationSum), "sum, average, minimum, maximum, latest, or count")
	unit := flags.String("unit", "", "display unit")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("metric add does not accept positional arguments")
	}
	spaceID := defaultSpaceID
	metric := domain.Metric{
		ID:           newID("metric"),
		OwnerSpaceID: &spaceID,
		Key:          *key,
		Name:         *name,
		Category:     *category,
		Kind:         domain.MetricKind(*kind),
		Aggregation:  domain.Aggregation(*aggregation),
		Unit:         *unit,
	}
	if err := repository.CreateMetric(ctx, metric); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Added metric %s (%s).\n", metric.Name, metric.Key)
	return nil
}

func runEntry(ctx context.Context, repository *sqlite.Repository, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 || args[0] != "add" {
		return errors.New("usage: yld entry add --metric KEY --value NUMBER [--at RFC3339|YYYY-MM-DD]")
	}
	flags := flag.NewFlagSet("entry add", flag.ContinueOnError)
	flags.SetOutput(stderr)
	metricKey := flags.String("metric", "", "metric key")
	valueText := flags.String("value", "", "numeric value")
	occurredText := flags.String("at", "", "RFC3339 timestamp or YYYY-MM-DD; defaults to now")
	note := flags.String("note", "", "optional note")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("entry add does not accept positional arguments")
	}
	if *metricKey == "" || *valueText == "" {
		return errors.New("metric and value are required")
	}
	value, err := strconv.ParseFloat(*valueText, 64)
	if err != nil {
		return fmt.Errorf("parse value: %w", err)
	}
	metric, err := repository.MetricByKey(ctx, defaultSpaceID, *metricKey)
	if err != nil {
		return err
	}
	timezone, err := repository.UserTimezone(ctx, defaultUserID)
	if err != nil {
		return err
	}
	occurredAt, err := parseOccurredAt(*occurredText, timezone)
	if err != nil {
		return err
	}
	localDate, err := domain.LocalDate(occurredAt, timezone)
	if err != nil {
		return err
	}
	entry := domain.Entry{
		ID:         newID("entry"),
		SpaceID:    defaultSpaceID,
		UserID:     defaultUserID,
		MetricID:   metric.ID,
		OccurredAt: occurredAt,
		LocalDate:  localDate,
		Timezone:   timezone,
		Value:      value,
		Note:       *note,
		Source:     "manual",
	}
	if err := repository.AddEntry(ctx, entry); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Logged %s %s for %s.\n", formatNumber(value), metric.Unit, localDate.Format(time.DateOnly))
	return nil
}

func runRecap(ctx context.Context, repository *sqlite.Repository, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 || args[0] != "show" {
		return errors.New("usage: yld recap show [--period month|quarter|year] [--date YYYY-MM-DD]")
	}
	flags := flag.NewFlagSet("recap show", flag.ContinueOnError)
	flags.SetOutput(stderr)
	periodKind := flags.String("period", string(domain.PeriodMonth), "month, quarter, or year")
	dateText := flags.String("date", time.Now().Format(time.DateOnly), "a date within the recap period")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("recap show does not accept positional arguments")
	}
	date, err := time.Parse(time.DateOnly, *dateText)
	if err != nil {
		return fmt.Errorf("parse recap date: %w", err)
	}
	period, err := recap.PeriodContaining(domain.PeriodKind(*periodKind), date)
	if err != nil {
		return err
	}
	metrics, err := repository.ListMetrics(ctx, defaultSpaceID)
	if err != nil {
		return err
	}
	entries, err := repository.ListEntries(ctx, defaultUserID, defaultSpaceID, period.StartDate, period.EndDate)
	if err != nil {
		return err
	}
	snapshot, err := recap.GenerateSnapshot(newID("recap"), defaultUserID, defaultSpaceID, period, metrics, entries, time.Now().UTC())
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "%s recap: %s to %s\n", strings.ToUpper(string(period.Kind[:1]))+string(period.Kind[1:]), period.StartDate.Format(time.DateOnly), period.EndDate.AddDate(0, 0, -1).Format(time.DateOnly))
	if len(snapshot.Items) == 0 {
		fmt.Fprintln(stdout, "No entries in this period.")
		return nil
	}
	for _, item := range snapshot.Items {
		unit := ""
		if item.Unit != "" {
			unit = " " + item.Unit
		}
		entryLabel := "entries"
		if item.EntryCount == 1 {
			entryLabel = "entry"
		}
		fmt.Fprintf(stdout, "%s: %s%s (%d %s)\n", item.MetricName, formatNumber(item.Value), unit, item.EntryCount, entryLabel)
	}
	return nil
}

func parseOccurredAt(value, timezone string) (time.Time, error) {
	if value == "" {
		return time.Now(), nil
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}
	date, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return time.Time{}, errors.New("at must be an RFC3339 timestamp or YYYY-MM-DD")
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, location), nil
}

func newID(prefix string) domain.ID {
	var random [12]byte
	if _, err := rand.Read(random[:]); err != nil {
		panic(fmt.Sprintf("read random id: %v", err))
	}
	return domain.ID(prefix + "-" + hex.EncodeToString(random[:]))
}

func formatNumber(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func printUsage(output io.Writer) {
	fmt.Fprintln(output, `Usage: yld [--db PATH] COMMAND

Commands:
  init       Create or update the personal profile
  metric add Define a custom metric
  entry add  Log one observation
  recap show Calculate and print a recap`)
}

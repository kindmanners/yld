package main

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestCoreCLILoop(t *testing.T) {
	t.Parallel()

	databasePath := filepath.Join(t.TempDir(), "yld.db")
	commands := [][]string{
		{"--db", databasePath, "init", "--handle", "alex", "--name", "Alex", "--timezone", "Europe/Stockholm"},
		{"--db", databasePath, "metric", "add", "--key", "steps", "--name", "Steps", "--kind", "integer", "--aggregation", "sum", "--unit", "steps"},
		{"--db", databasePath, "entry", "add", "--metric", "steps", "--value", "1234", "--at", "2026-03-29"},
	}
	for _, command := range commands {
		if err := run(context.Background(), command, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
			t.Fatalf("run(%v): %v", command, err)
		}
	}

	var output bytes.Buffer
	if err := run(context.Background(), []string{"--db", databasePath, "recap", "show", "--period", "month", "--date", "2026-03-15"}, &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Steps: 1234 steps (1 entry)") {
		t.Fatalf("recap output:\n%s", output.String())
	}
}

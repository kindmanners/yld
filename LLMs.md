# AI Coding Assistants

This document guides people and AI tools contributing to YLD. Read
[`CONTRIBUTING.md`](CONTRIBUTING.md), [`SECURITY.md`](SECURITY.md), and the
relevant files under [`docs/`](docs/) before making changes. Human contributors
remain responsible for every submitted change.

## Project boundaries

YLD's implemented v1 is deliberately small: a single-user CLI, one personal
space, custom metrics, manual entries, recap calculation, and local SQLite
persistence. Do not document or assume that deferred package placeholders are
implemented features.

Preserve these responsibilities unless a requested architectural change says
otherwise:

- `internal/domain`: data types and invariants.
- `internal/recap`: period boundaries and aggregation.
- `internal/store`: migration infrastructure and persistence contracts.
- `internal/store/sqlite`: SQLite implementation.
- `cmd/yld`: CLI parsing, composition, and presentation.
- `internal/api`, `importer`, `notify`, `scheduler`, and `service`: reserved
  boundaries whose current `doc.go` files do not imply complete features.

Read [`docs/architecture.md`](docs/architecture.md) for the data rules and v2
decisions that must survive v1 changes.

## Data correctness and privacy

- Never put real life-tracking data, database contents, passwords, tokens, or
  other sensitive information in source, tests, documentation, logs, commits,
  or issue reports. Use clearly synthetic examples.
- Preserve `occurred_at`, derived `local_date`, and the original timezone.
- Keep recap ranges half-open: start inclusive, end exclusive.
- Maintain user and space scoping in repository queries and recap generation.
- Preserve soft-delete behavior and importer idempotency fields.
- Do not claim that SQLite data is encrypted or isolated from other processes
  running as the same OS user.
- Avoid placing note contents or other personal values in diagnostic errors.

## SQLite and migrations

- Keep released migrations immutable and add a new numbered migration for
  schema changes.
- Apply migrations transactionally and preserve foreign-key enforcement.
- Parameterize values in SQL; never concatenate untrusted input into queries.
- Consider rollback, partial failure, uniqueness, nullability, and existing
  databases when changing persistence.
- Add repository tests for persistence changes, preferably using an isolated
  temporary or in-memory database with synthetic records.

## Procedure

1. Inspect the relevant documentation, code, tests, and working-tree state.
2. Confirm the behavior or problem against the current implementation.
3. Make the smallest focused change that addresses the request.
4. Add or update tests for behavioral changes and regressions.
5. Run `gofmt` on changed Go files, then `go test ./...` and `go build ./...`
   when the environment permits.
6. Review the diff for unrelated changes, secrets, personal data, license
   issues, and inaccurate documentation.
7. State exactly what was verified and what was not.

Do not weaken validation to make a test pass, invent commands or features,
rewrite unrelated code, destroy uncommitted work, or claim checks were run
when they were not.

## Dependencies and licensing

YLD is licensed under the GNU Affero General Public License v3.0. New code and
dependencies must be license-compatible and properly attributed. Prefer the Go
standard library or an existing dependency when suitable, and do not introduce
code of unknown provenance.

Substantial AI assistance should be disclosed in the pull request. An optional
commit trailer is:

```text
Assisted-by: LLM <tool name>
```

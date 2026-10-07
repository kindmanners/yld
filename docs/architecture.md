# Your Life's Database v1 architecture

## Product loop

Version 1 proves one loop: record something meaningful, receive a useful recap, and enjoy seeing the accumulated history. It supports one user, one personal space, manual and CLI logging, one importer, and immutable monthly, quarterly, and yearly recap snapshots.

## Data rules

- Every entry stores `occurred_at`, the derived `local_date`, and the timezone used for that derivation. Changing the user's timezone never rewrites history.
- Built-in metrics have no owning space. Custom metrics belong to a space.
- A metric's kind, aggregation, and unit cannot change after it has entries. Its name, category, and archive state may change.
- Recap periods use half-open date ranges: the start is included and the end is excluded.
- Recaps are generated after a grace period and freeze when sent. `source_updated_at` is retained so the UI can indicate that newer data exists.
- Source and external IDs make importer writes idempotent.

## Deliberately deferred to v2

The schema contains users, spaces, memberships, activities, and ownership columns so v2 can add shared spaces without replacing the v1 data model. Shared-space leave flows, change requests, and pseudonymous detached copies are not part of v1.

The settled leave-flow rule is: originals move to the leaver's personal space; an explicit opt-in may first create stripped, irreversible copies owned by a per-space pseudonym. Custom metric definitions referenced by moved entries must be copied into the personal space in the same transaction.

## Package boundaries

- `internal/domain`: pure types and invariants
- `internal/recap`: period boundaries and aggregation
- `internal/store`: repositories and SQLite persistence
- `internal/service`: application use cases
- `internal/importer`: source adapters and import contracts
- `internal/scheduler`: grace-period scheduling and catch-up
- `internal/notify`: recap delivery contracts
- `internal/api`: optional HTTP transport
- `cmd/yld`: CLI composition root

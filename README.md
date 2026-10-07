# Your Life's Database

Your Life's Database is a self-hosted record of the things you do, with snapshot recaps for each month, quarter, and year. The current v1 slice is deliberately small: one person, one personal space, custom metrics, manual CLI logging, and local SQLite storage.

## Run the core loop

This project requires Go 1.27 or newer.

```sh
go build -o yld ./cmd/yld

./yld --db ./my-life.db init \
  --handle me \
  --name "My Name" \
  --timezone Europe/Stockholm

./yld --db ./my-life.db metric add \
  --key steps \
  --name Steps \
  --kind integer \
  --aggregation sum \
  --unit steps

./yld --db ./my-life.db entry add \
  --metric steps \
  --value 8432 \
  --at 2026-10-06

./yld --db ./my-life.db recap show \
  --period month \
  --date 2026-10-06
```

Supported metric kinds are `number`, `integer`, and `duration`. Supported aggregations are `sum`, `average`, `minimum`, `maximum`, `latest`, and `count`.

## Development

```sh
go test ./...
```

The [documentation index](docs/README.md) links the architecture, contributor,
security, community, and AI-assistance guides. Product and data-model decisions
are recorded in [`docs/architecture.md`](docs/architecture.md).

Before contributing, read [`CONTRIBUTING.md`](CONTRIBUTING.md). Report security
issues privately as described in [`SECURITY.md`](SECURITY.md); do not include
real life-tracking data in public issues, examples, or test fixtures.

## License

YLD is licensed under the [GNU Affero General Public License v3.0](LICENSE.md).

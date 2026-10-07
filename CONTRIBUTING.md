# Contributing to YLD

Thanks for helping improve Your Life's Database (YLD). Contributions to code,
tests, documentation, and bug reports are welcome.

## Before you start

- Search the [existing issues](https://github.com/kindmanners/yld/issues) before
  opening a new one.
- For a bug, include the YLD revision, Go version, operating system, expected
  and actual behavior, and concise reproduction steps.
- Discuss substantial features in an issue before implementation so their fit
  with the deliberately small v1 scope can be agreed first.
- Do not put vulnerabilities or real life-tracking data in a public issue.
  Follow [the security policy](SECURITY.md) instead.
- Follow the [Code of Conduct](CODE_OF_CONDUCT.md) in all project spaces.

## Development setup

YLD is a Go project. The required toolchain is recorded in [`go.mod`](go.mod);
it currently requires Go 1.27. Install Git and the required Go version, then:

```bash
git clone https://github.com/kindmanners/yld.git
cd yld
git checkout -b fix/short-description
go mod download
```

Use a focused branch name such as `fix/...`, `feat/...`, `docs/...`, or
`refactor/...`. Do not commit local SQLite databases or build output.

## Build and test

Run these checks from the repository root before opening a pull request:

```bash
# Git Bash, macOS, or Linux
gofmt -w $(git ls-files '*.go')
go test ./...
go build ./...
```

On PowerShell, format tracked Go files with:

```powershell
gofmt -w (git ls-files '*.go')
```

To build the CLI as a named binary:

```bash
go build -o bin/yld ./cmd/yld
```

Run `go test -race ./...` for changes involving concurrency or scheduling.
There is no repository-managed linter configuration; the commands above are
the portable project checks.

## Code and documentation

- Follow standard Go conventions and let `gofmt` format Go code.
- Add or update tests when behavior changes. Tests live beside the relevant
  packages under `internal/` and `cmd/`.
- Preserve the package boundaries and data rules in
  [`docs/architecture.md`](docs/architecture.md).
- Treat dates and timezones carefully. Entries preserve the timezone and local
  date used when they were recorded; recap periods are half-open ranges.
- Keep SQLite migrations append-only after release. Do not edit an applied
  migration when a new migration can express the change.
- Keep changes focused and avoid unrelated formatting, dependency upgrades,
  or generated files.
- Update the README and documentation when CLI flags, persistence behavior,
  recap semantics, security assumptions, or architecture change.
- Use synthetic data in tests and examples. Never commit a contributor's real
  health, activity, habit, or other personal tracking data.

## Pull requests

1. Bring your branch up to date with the target branch and resolve conflicts.
2. Summarize the problem and solution, including compatibility or migration
   effects.
3. List the checks you ran and link the related issue when there is one.
4. Include example CLI output when user-visible behavior changes, with all
   personal data replaced by synthetic values.
5. Respond to review feedback with follow-up commits or a clear explanation of
   the trade-off.

## Commit messages

The project uses [Conventional Commits](https://www.conventionalcommits.org/).
Write a concise, imperative subject with an optional scope:

```text
fix(recap): preserve local-date boundaries
docs: explain SQLite backup safety
```

Useful types include `feat`, `fix`, `docs`, `refactor`, `test`, `build`, `ci`,
and `chore`. Mark breaking changes with `!` or a `BREAKING CHANGE:` footer.

## AI coding assistants

Contributors remain responsible for AI-assisted work. Before submitting it:

- Review and understand every generated or modified file.
- Verify correctness, security, licensing, and third-party provenance.
- Run the relevant tests and inspect the final diff.
- Remove secrets and personal data from prompts, logs, patches, and examples.
- Manually review changes involving SQLite queries or migrations, path
  handling, importers, authentication, authorization, or personal data.
- Disclose substantial AI assistance in the pull request.

See [`LLMs.md`](LLMs.md) for repository-specific guidance.

## License

By contributing, you confirm that you have the right to submit the work and
agree to license it under the repository's
[GNU Affero General Public License v3.0](LICENSE.md).

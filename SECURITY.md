# Security Policy

## Supported versions

YLD is under active development. Security fixes are provided for the latest
release and the current `main` branch.

| Version | Supported |
| --- | --- |
| Latest release | Yes |
| `main` development branch | Yes |
| Older releases | No |

Update to the newest available version before reporting an issue that may
already have been fixed.

## Current trust and data boundaries

YLD's current v1 is a local command-line application backed by SQLite. It has
no network server, remote synchronization, or multi-user authentication. It is
not a security boundary between users or processes on the same machine.

The database can contain sensitive personal history. YLD currently does not
encrypt the database or manage operating-system permissions. Users are
responsible for restricting access to the database file, its directory,
backups, exported copies, terminal history, and command output. Use full-disk
encryption where appropriate and do not place a database in a publicly shared
or untrusted synchronized folder.

SQLite migrations run automatically when a database is opened. Back up
important databases before upgrading. Treat database paths, future importer
inputs, and all stored notes as untrusted data when extending the project.

## Reporting a vulnerability

Do **not** report vulnerabilities through public GitHub issues, discussions,
or pull requests. Use GitHub Private Vulnerability Reporting for the
[YLD repository](https://github.com/kindmanners/yld/security/advisories/new).
If that is unavailable, email **maxiefeseymen@gmail.com**.

Include, where possible:

- A description of the vulnerability and its impact.
- The affected YLD version, commit, command, or package.
- Reproduction steps and a minimal proof of concept.
- Relevant logs or sample data with all personal information, database
  contents, credentials, and secrets removed.
- Known mitigations or a suggested fix.

We aim to acknowledge reports within seven days. Accepted issues are
investigated privately, and details should remain private until a fix or
mitigation is available. A security advisory or release note may be published
after resolution.

## Scope

Report issues privately when they could cause meaningful security impact,
including:

- Unauthorized reading, alteration, or deletion of YLD data.
- SQL injection or bypass of repository scoping rules.
- Path traversal or unsafe file creation through a YLD-controlled path.
- Exposure of personal data through logs, errors, recap output, imports, or
  exports beyond the behavior explicitly requested by the user.
- Malicious database or importer input causing code execution.
- Dependency vulnerabilities that are reachable through YLD.
- Future authentication, authorization, API, scheduler, notification, or
  importer flaws that cross a documented trust boundary.

Ordinary bugs, expected access by an authorized local account, performance
problems, and feature requests without a security impact belong in the public
issue tracker. Vulnerabilities entirely in an upstream dependency should
normally be reported upstream unless YLD causes or worsens the exposure.

## Responsible testing

- Test only systems and data you own or have explicit permission to test.
- Use synthetic databases; do not access or publish another person's records.
- Avoid destructive testing against a user's only database copy.
- Stop and report the issue if you gain unintended access to sensitive data.
- Do not retain or publicly disclose personal data or credentials obtained
  during testing.

Thank you for helping keep YLD and its users safe.

# Development Guide

Conventions for working on AppCMS day to day.

## Repository setup

1. Clone AppCMS with sibling repos `@Platform`, `@Framework`, `@Templ` (paths in `go.mod` replace directives).
2. `bun install` in AppCMS root.
3. `bun verify` before pushing.

## Commands

See [Commands and verify](commands-and-verify.md).

## Code conventions

| Rule | Detail |
|------|--------|
| Language | Comments and docs in **English** |
| UI | Use `github.com/fastygo/templ` only |
| Imports | `internal/` is private to AppCMS |
| Capabilities | Namespace under module IDs (`content.write`, not generic `write`) |
| URLs | Do not rename root GoCMS paths |
| Platform | Never import product apps from Platform |

## Layer discipline

- Domain: no SQL, HTTP, templ
- Application: no HTTP handlers
- Delivery: thin—parse HTTP, call services inside `WithinTx`
- `pkg/module`: descriptors only, no SQLite

## Adding dependencies

Prefer existing stack modules. New shared abstractions belong in Platform only when **two products** need the same contract.

## Pull request checklist

- [ ] `bun verify` passes
- [ ] Integration tests for new routes or auth behavior
- [ ] No `ui8kit` imports (see `TestAppCMSDoesNotImportUI8Kit`)
- [ ] Codex URLs unchanged unless intentionally documented
- [ ] Docs updated under `docs/` for user-visible behavior

## Further reading

- [Testing](testing.md)
- [Architecture](../architecture/README.md)
- [Codex schema alignment](../appcms/codex-schema-alignment.md)
- [GoCMS migration mapping](../compat/gocms/migration-mapping.md)

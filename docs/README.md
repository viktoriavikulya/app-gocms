# AppCMS Documentation

Welcome to the AppCMS 101 guide. This documentation explains how **AppCMS** (`github.com/fastygo/app-gocms`) fits into the FastyGo product stack, which packages it uses, and how the repository is organized.

AppCMS is the runnable **GoCMS compatibility target** on the shared stack:

- **Framework** — HTTP lifecycle, middleware, sessions, security, health probes
- **Platform** — module contracts, toolset, panel, profiles, render models
- **Templ** — shared UI primitives for admin and public views

Legacy **GoCMS** (`@GoCMS`) remains the **behavior oracle** via `go-codex/en`. AppCMS ports observable behavior deliberately; it does not reintroduce UI8Kit.

## Who this guide is for

- Developers new to AppCMS or the FastyGo product stack
- Contributors porting features from GoCMS
- Teams embedding AppCMS inside **AppSuite** or building sibling products (**AppCRM**)

## Documentation map

| Section | Purpose |
|--------|---------|
| [Getting started](getting-started/README.md) | Run the server, verify the build, first URLs |
| [Architecture](architecture/README.md) | Layers, boundaries, dependency direction |
| [Ecosystem](ecosystem/README.md) | Framework, Platform, Templ, GoCMS oracle, related apps |
| [AppCMS packages](appcms/README.md) | This repo: `pkg/`, `internal/`, delivery surfaces |
| [Development](development/README.md) | Commands, tests, conventions, guardrails |

## Quick facts

| Topic | Value |
|-------|--------|
| Module path | `github.com/fastygo/app-gocms` |
| Default bind | `127.0.0.1:8080` (via Framework config / `ADDR`) |
| Root admin | `/go-admin` |
| Root API | `/go-json`, `/go-json/go/v2/` |
| Auth surfaces | `/go-login`, `/go-logout` |
| Public site | `/`, `/posts/{slug}`, `/{slug}` (when not headless/admin-only) |
| GraphQL extension | `/go-graphql` (compiled-in plugin) |
| Primary storage (today) | SQLite via `internal/storage/sqlite` |
| UI kit | `github.com/fastygo/templ` only — **no** `ui8kit` |

## Preserved GoCMS URL contracts

These root paths are compatibility contracts. Do not rename them when adding features; add **spaces overlays** beside them (for example `/go-admin/spaces/*`), not instead of them.

- `/go-admin`
- `/go-login`
- `/go-logout`
- `/go-json`
- `/go-json/go/v2/`

See [URL contracts](appcms/url-contracts.md) for route tables and runtime profile gating.

## Related documents

- [Migration from GoCMS](migration-from-gocms.md) — mapping table from legacy monolith areas to AppCMS owners
- [Platform current progress](../../@Platform/.project/current-progress.md) — slice checklist for the whole product stack (sibling repo)

## Suggested reading order

1. [Getting started](getting-started/README.md)
2. [Architecture → Layers and repos](architecture/layers-and-repos.md)
3. [Ecosystem overview](ecosystem/README.md)
4. [AppCMS package layout](appcms/package-layout.md)
5. [Internal layers](appcms/internal-layers.md)
6. Topic guides: [storage](appcms/storage.md), [auth](appcms/auth-and-capabilities.md), [public site](appcms/public-site-and-themes.md), [extensions](appcms/extensions-and-operations.md)
7. [Development](development/README.md)

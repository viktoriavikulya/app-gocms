# Getting Started

This page gets AppCMS running locally and confirms that codex-shaped routes respond.

## Prerequisites

- **Go** 1.25+ (see `go.mod`)
- **Bun** for CSS generation and the verify script
- Local sibling repos (via `replace` in `go.mod`):
  - `@Framework` → `github.com/fastygo/framework`
  - `@Platform` → `github.com/fastygo/platform`
  - `@Templ` → `github.com/fastygo/templ`

## Install and verify

From the AppCMS repository root:

```bash
bun install
bun verify
```

`bun verify` runs:

1. `templ generate` for product and dependency templates
2. Tailwind build (`web/static/css/app.css`)
3. `go test ./...`

## Run the server

```bash
bun go
```

Or directly:

```bash
go run ./cmd/server
```

Default URLs (when bind is `127.0.0.1:8080`):

| URL | What you should see |
|-----|---------------------|
| `/` | Themed public home (full / conformance runtime) |
| `/go-login` | Login form with signed `action_token` |
| `/go-admin/` | Admin dashboard (after login) |
| `/go-json/` | API root discovery JSON |
| `/go-json/go/v2/` | v2 discovery with route list |
| `/healthz` | Framework liveness probe |
| `/readyz` | Framework readiness probe |

### Seeded credentials (development)

The default in-memory auth store seeds:

| User | Password | Typical role |
|------|----------|--------------|
| `admin` | `admin` | Full admin |
| `editor` | `editor` | Content write |
| `viewer` | `viewer` | Read-only |

Use these only in local development. Replace the auth store and session secret in production.

## Runtime modes (host options)

AppCMS supports host-level **runtime modes** (distinct from Platform deployment **profiles**):

| Mode | Public HTML | Admin UI | REST / extensions |
|------|-------------|----------|-------------------|
| `full` (default) | Yes | Yes | Yes |
| `headless` | No | No | Yes |
| `admin` | No | Yes | Yes |
| `conformance` | Yes (fixtures) | Yes | Yes |

Legacy flag: `Options.Headless: true` maps to `headless` mode.

Example in tests or custom wiring:

```go
gocmsapp.NewApp(gocmsapp.Options{
    RuntimeMode: gocmsapp.RuntimeModeHeadless,
    Seed:        true,
})
```

See [Runtime profiles](../appcms/runtime-profiles.md).

## Next steps

- Understand [package layout](../appcms/package-layout.md)
- Read [URL contracts](../appcms/url-contracts.md)
- Explore the [FastyGo ecosystem](../ecosystem/README.md)

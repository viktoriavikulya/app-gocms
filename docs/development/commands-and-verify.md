# Commands and Verify

## Daily commands

| Command | Action |
|---------|--------|
| `bun install` | Install JS tooling for Tailwind |
| `bun verify` | templ generate + CSS build + `go test ./...` |
| `bun go` | Run development server |
| `go test ./...` | Go tests only |
| `go tool templ generate ./...` | Regenerate templ Go files |
| `bun run build:css` | Compile Tailwind to `web/static/css/app.css` |

## Run server

```bash
bun go
# or
go run ./cmd/server
```

Environment:

| Variable | Effect |
|----------|--------|
| `ADDR` | Override bind address (e.g. `127.0.0.1:8080`) |
| `APP_STATIC_DIR` | Static files directory |

## Verify pipeline

`bun verify` script (from `package.json`):

1. **Templ** — `go tool templ generate ./...`
2. **CSS** — Tailwind from `web/static/css/input.css`
3. **Tests** — full module test suite

Always run before commits that touch templates, CSS, or Go handlers.

## Formatting

```bash
gofmt -w path/to/file.go
```

Keep generated `*_templ.go` in sync via templ generate, not manual edits.

## Working on sibling modules

When changing Platform contracts:

```bash
cd ../@Platform && go test ./...
cd ../@AppCMS && bun verify
```

Framework auth or middleware changes:

```bash
cd ../@Framework && go test ./...
cd ../@AppCMS && go test ./cmd/server/...
```

## Local SQLite file

Default dev DSN can point to a file:

```
file:appcms.db
```

Pass via `Options.StorageDSN` in custom mains or future CLI flags.

## Debugging routes

Use integration tests in `cmd/server/main_test.go` as executable documentation:

```bash
go test ./cmd/server -run TestCodexRouteSurfaces -v
```

Discovery payload:

```bash
curl -s http://127.0.0.1:8080/go-json/go/v2/ | jq .
```

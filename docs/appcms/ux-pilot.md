# UX Pilot Branch (`ux-pilot`)

The `ux-pilot` branch is the **Stage 1** admin-only baseline for AppCMS BFF work.
It was created from `stable` to remove sprint drift and focus on schema-driven admin UX
before Platform BFF extraction.

## Branch layout

| Branch | Purpose |
|--------|---------|
| `stable` | Tagged baseline (`08b29e2`) |
| `canary` | Archive of post-stable commits + full infra sprint (`9984b8d`) |
| `ux-pilot` | Active pilot: thin admin shell, Templ-first |
| `main` | Unchanged until explicit merge from `ux-pilot` |

Cherry-pick from `canary` only into the layers listed in the Platform doc
`@Platform/.project/roadmap-bff-model.md` (section **Deferred work map**).

## Default runtime

`pkg/app` defaults to `RuntimeModeAdmin`:

- `/go-login`, `/go-logout`
- `GET /go-admin/*` → `appschema.Registry.Screen` → `pkg/templ.Page`
- REST `/go-json/go/v2/*`

**Not mounted in default pilot:**

- Public site (`/`, `/posts/{slug}`, `/{slug}`)
- GraphQL (`/go-graphql`)
- Ops plugin routes (`/go-json/go/v2/ops/*`)

Use `RuntimeModeFull` or `RuntimeModeConformance` in tests or explicit `Options` to exercise deferred surfaces.

## Pilot screens (UX-first order)

1. `/go-login`
2. `/go-admin/` (dashboard)
3. `/go-admin/posts` (table shell from `ScreenModel`)
4. `/go-admin/posts/new` or edit (form shell; save deferred until Platform BFF actions)

## Kept packages

- `pkg/module` — records, panels, capabilities
- `internal/appschema` — URL → `render.ScreenModel`
- `internal/application`, `internal/storage` — kernel (not imported by renderer)
- `internal/delivery/rest` — codex REST
- `pkg/app` — host wiring
- `pkg/templ` — Templ renderer (props-only)

## Deferred (on disk, unwired or on `canary`)

- `internal/delivery/admin` — on `canary` only; returns via Platform generic BFF actions
- `internal/delivery/publicsite`, `internal/themes` — public delivery profile
- `internal/plugins/graphql`, `internal/plugins/ops` — extension slice
- Sprint: hydration, hooks, audit, transitions, Level0 conformance

## Local run

```bash
go run ./cmd/server
```

Open `http://127.0.0.1:8080/go-login`, then admin after sign-in.

## Renderer rule

Templ is a renderer over the BFF model, not a shortcut into services.
Enforced by `internal/ui/architecture_test.go`.

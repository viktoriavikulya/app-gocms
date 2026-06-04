# Request Flow

How HTTP requests move through AppCMS for the three main surfaces: **public**, **admin**, and **REST**.

## Public page (full runtime)

```mermaid
sequenceDiagram
    participant Browser
    participant Framework
    participant PublicHandler as publicsite.Handler
    participant Permalinks
    participant Assembler
    participant Storage
    participant Themes

    Browser->>Framework: GET /posts/hello-world
    Framework->>PublicHandler: ServeHTTP
    PublicHandler->>Permalinks: Resolve path
    PublicHandler->>Storage: WithinTx (repos)
    Storage->>Assembler: Assemble Page DTO
    Assembler-->>PublicHandler: publicrender.Page
    PublicHandler->>Themes: Resolve active theme
    Themes-->>Browser: HTML (templ)
```

Key packages:

- `internal/delivery/publicsite/handler.go` — gate headless/admin-only modes, system path exclusions
- `internal/permalinks/resolver.go` — home, page slug, post slug patterns
- `internal/delivery/publicsite/assembler.go` — loads content, settings, menus
- `internal/publicrender/model.go` — DTO shared with themes
- `internal/themes` — `blank`, `gocms-default` manifests and views

System paths (`/go-admin`, `/go-json`, `/static`, etc.) return 404 from the public handler so they are not shadowed by slug catch-alls.

## Admin screen

```mermaid
sequenceDiagram
    participant Browser
    participant Auth as authBoundary
    participant Registry as appschema.Registry
    participant Views as pkg/templ

    Browser->>Auth: GET /go-admin/posts
    Auth->>Auth: Read session cookie
    Auth->>Auth: Require admin.access
    Auth->>Registry: Screen(path)
    Registry-->>Auth: render.ScreenModel
    Auth->>Views: Page(title, screen)
    Views-->>Browser: HTML
```

- Login: `/go-login` issues Framework `CookieSession` with `SessionClaims`
- Forms: metadata `action_token` for CSRF-style protection on admin writes
- Screen definitions come from `pkg/module` panel descriptors via `internal/appschema`

## REST mutation

```mermaid
sequenceDiagram
    participant Client
    participant REST as delivery/rest
    participant AuthZ as Authorizer
    participant Services as application/*
    participant Storage

    Client->>REST: PATCH /go-json/go/v2/posts/{id}
    REST->>AuthZ: capability content.write
    alt not authenticated
        AuthZ-->>Client: 401
    else forbidden
        AuthZ-->>Client: 403
    else allowed
        REST->>Storage: WithinTx
        Storage->>Services: domain operations
        Services-->>Client: JSON envelope
    end
```

Authorizer is wired from `pkg/app` `authBoundary.Authorize` (session cookie or `Bearer` app token).

## Extension routes (plugins)

After core routes register, `extensionRuntime` activates compiled-in plugins and mounts their routes:

| Plugin | Routes | Auth |
|--------|--------|------|
| `graphql` | `GET/POST /go-graphql` | Public reads filter drafts/private; mutations disabled in baseline |
| `operations` | `/go-json/go/v2/ops/*`, `/go-admin/import-export` | Capability-gated (`admin.access`, `settings.manage`) |

Registration: `internal/extensions/runtime.go` → `RegisterRoutes` with per-route capability checks.

## Framework wrapping

All product routes register on a `http.ServeMux` inside a Framework `Feature`:

```go
frameworkapp.New(cfg).
    WithSecurity(...).
    WithHealthEndpoints(...).
    WithFeature(feature{...}).
    Build()
```

Framework adds middleware (recover, request ID, logging), static file hosting for `/static/`, and process-level health endpoints.

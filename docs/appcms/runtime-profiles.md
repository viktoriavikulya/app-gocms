# Runtime Profiles

AppCMS distinguishes two related concepts:

| Concept | Where | Purpose |
|---------|-------|---------|
| **Platform profile** | `profile.Profile` in `pkg/app/bundle.go` | Deployment bases, workspaces, panel mounts |
| **Runtime mode** | `pkg/app.RuntimeMode` | Which HTTP surfaces are registered |

This document covers **runtime mode** (host gating). Platform profiles are described in [Platform ecosystem](../ecosystem/platform.md).

## Runtime modes

```go
type RuntimeMode string

const (
    RuntimeModeFull        RuntimeMode = "full"
    RuntimeModeHeadless    RuntimeMode = "headless"
    RuntimeModeAdmin       RuntimeMode = "admin"
    RuntimeModeConformance RuntimeMode = "conformance"
)
```

Set via `gocmsapp.Options`:

```go
Options{
    RuntimeMode: gocmsapp.RuntimeModeAdmin,
}
```

Legacy compatibility: `Options.Headless: true` forces `headless` mode.

## Surface matrix

| Surface | full | headless | admin | conformance |
|---------|------|----------|-------|-------------|
| Public `/`, slugs | ✓ | ✗ | ✗ | ✓ |
| Admin `/go-admin/*` | ✓ | ✗ | ✓ | ✓ |
| REST `/go-json/*` | ✓ | ✓ | ✓ | ✓ |
| Login/logout | ✓ | ✓ | ✓ | ✓ |
| Plugins (GraphQL, ops) | ✓ | ✓ | ✓ | ✓ |
| Framework `/healthz` | ✓ | ✓ | ✓ | ✓ |

## Use cases

### full (default)

Standalone CMS with admin, API, and public site. Local development and typical content management.

### headless

Content API and extensions without public HTML or admin UI. Decoupled frontends consume REST/GraphQL; operators use another channel for admin (or a separate profile instance).

### admin

Back-office only: admin + API, no public theme routes. Useful for split deployments where public traffic hits a different service.

### conformance

Deterministic routes and seeded fixtures for integration tests. Public and admin enabled; pairs with `Seed: true` in tests.

## Implementation notes

`registerRoutesWithOptions` in `pkg/app/app.go`:

- Skips `mux.Handle` for public routes when mode is `headless` or `admin`
- Skips admin handlers when mode is `headless`
- Sets `publicsite.Handler.Headless` for defense in depth

Extension runtime receives mode string for health reporting (`plugin_runtime` check message).

## Platform profile vs runtime mode

Example standalone bundle:

```go
profile.Profile{
    ID:         "gocms-admin",
    AdminBase:  "/go-admin",
    APIBase:    "/go-json",
    PublicBase: "/",
    Workspaces: []profile.Workspace{{ID: "root", ...}},
}
```

Runtime mode does not change these bases—it only controls which route groups are mounted on the HTTP mux for this process.

AppSuite may mount the same bundle under different workspace paths while choosing runtime mode per deployment target.

## Testing

`TestRuntimeProfileModesGateSurfaces` in `cmd/server/main_test.go` asserts:

- `admin` → public 404, REST 200
- `conformance` → public home renders seeded content

Add tests when introducing new surfaces (e.g. new plugin routes) to verify headless gating.

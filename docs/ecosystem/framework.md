# Framework (`github.com/fastygo/framework`)

Framework is the **production HTTP shell** for FastyGo product applications. AppCMS uses it for process lifecycle, middleware, security configuration, session cookies, and health probes—not for CMS domain logic.

## What AppCMS uses

| Area | Package | Role in AppCMS |
|------|---------|----------------|
| Application host | `pkg/app` | `NewApp`, `Run`, `Feature` interface for route registration |
| Configuration | `pkg/app` `Config` | Bind address, static dir, session secret, health paths |
| Sessions | `pkg/auth` `CookieSession` | HMAC-signed `appcms_session` cookie with `contracts.SessionClaims` |
| Action tokens | `pkg/auth` `SignedEncode` / `SignedDecode` | Login/logout and admin form CSRF-style tokens |
| Security | `pkg/web/security` | Loaded via `WithSecurity(security.LoadConfig())` |
| Health | `pkg/web/health` | `/healthz` (live), `/readyz` (ready) — process-level, not CMS diagnostics |

## AppBuilder pattern

AppCMS `pkg/app.NewApp` constructs:

```go
frameworkapp.New(cfg).
    WithSecurity(security.LoadConfig()).
    WithHealthEndpoints(cfg.HealthLivePath, cfg.HealthReadyPath).
    WithFeature(feature{registry, provider, auth, mode}).
    Build()
```

The `feature` type implements `frameworkapp.Feature`:

- `ID()` → `"app-gocms"`
- `Routes(mux)` → registers GoCMS codex routes, public site, REST, plugins

## Middleware chain (typical)

Framework's builder attaches middleware such as:

- Recovery from panics
- Request ID / correlation
- Structured request logging
- Security headers and method guards (from security config)

Product code should not reimplement these on individual handlers unless there is a special case.

## Sessions vs product auth

Framework provides **session transport** (cookie name, TTL, signing).

AppCMS owns **authentication semantics** in `internal/application/authn`:

- Argon2id password hashing
- Role → capability maps (`pkg/module` capability IDs)
- Login lockout
- App tokens (`Authorization: Bearer`)

The auth boundary in `pkg/app` bridges Framework sessions to `authn.Service` and REST `Authorizer`.

## Health: two levels

| Endpoint | Owner | Meaning |
|----------|-------|---------|
| `/healthz`, `/readyz` | Framework | Process up / ready to serve |
| `/go-json/go/v2/ops/health` | AppCMS operations plugin | CMS runtime checks (storage, plugins, audit store) |

Do not conflate them in monitoring dashboards.

## Configuration environment variables

Common Framework-related env vars (see Framework `pkg/app/config.go` for the full set):

| Variable | Effect |
|----------|--------|
| `ADDR` | HTTP bind address |
| `APP_STATIC_DIR` | Static files root (defaults to `web/static` under repo) |
| Session secret | From config file or `Options.SessionKey` in tests |

## When to change Framework vs AppCMS

| Change | Where |
|--------|-------|
| New global middleware, shutdown behavior | Framework |
| GoCMS route, REST handler, content service | AppCMS |
| Shared capability or storage port shape | Platform `contracts` |

## Further reading

- Framework examples under `@Framework` repository
- AppCMS [auth and capabilities](../appcms/auth-and-capabilities.md)

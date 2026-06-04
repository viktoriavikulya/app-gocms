# Auth and Capabilities

AppCMS authentication is **product-owned** (`internal/application/authn`) with **Framework session transport** (`pkg/auth`). Authorization uses capability IDs registered by the CMS module.

## Surfaces

| Surface | Mechanism |
|---------|-----------|
| Admin HTML | Cookie session `appcms_session` |
| REST mutations | Cookie session or `Authorization: Bearer` app token |
| REST public reads | Unauthenticated where content is public |
| GraphQL (baseline) | Public reads with visibility filter; mutations disabled |
| Operations API | Session + `admin.access` or `settings.manage` |

## Session flow

1. `GET /go-login` — page includes signed `action_token` (`action=login`)
2. `POST /go-login` — validates token, `authn.AuthenticatePassword`, issues cookie
3. Admin routes — `requireAdmin` checks `admin.access` capability
4. `POST /go-logout` — optional logout token; clears session

Session payload: `contracts.SessionClaims` with `PrincipalID` and `ProfileID` (`gocms-admin`).

Signing secret: Framework config / `Options.SessionKey` (development default exists; override in production).

## Principals and roles

Built-in users (seeded `MemoryStore` for dev/tests):

| User | Role | Typical capabilities |
|------|------|----------------------|
| admin | administrator | All CMS capabilities |
| editor | editor | Content write, media, taxonomies |
| viewer | viewer | Read-only |

Roles map to capability sets in `authn.BuiltInRoles()` using IDs from `pkg/module/capabilities.go`.

## Capability reference

| Capability ID | Used for |
|---------------|----------|
| `admin.access` | Admin UI entry |
| `content.read` | Read content via API |
| `content.write` | Create/update/delete content |
| `content.read_private` | Draft/private visiblity in API |
| `media.upload` | Upload media |
| `media.edit` | Edit media |
| `taxonomies.manage` | Taxonomy CRUD |
| `taxonomies.assign` | Assign terms to content |
| `users.manage` | Authors |
| `settings.manage` | Settings and snapshot import/export |

REST handlers call `authorize(r, capability)` before mutations. Failures:

- **401** — not authenticated
- **403** — authenticated but missing capability

## App tokens

`authn.Service.CreateAppToken` issues scoped bearer tokens for automation.

Example: editor token with only `content.write` can create posts but cannot access admin-only ops routes.

Tokens are stored hashed in SQLite (`auth_app_tokens` table) when using persistent auth schema; dev memory store keeps tokens in process.

## Login lockout

Failed password attempts per identifier are tracked. After repeated failures, login returns **429 Too Many Requests** until the lockout window expires (see `TestLoginLockout`).

## Action tokens (CSRF-style)

Admin forms receive `action_token` in screen metadata for write screens. Tokens are short-lived HMAC payloads (`action=admin-write`, login, logout).

## Runtime context

Authenticated admin requests attach:

```go
contracts.RuntimeContext{
    ProfileID:   "gocms-admin",
    WorkspaceID: "root",
    ModuleID:    "cms",
    PrincipalID: principal.ID(),
}
```

Downstream code may read this from `context` for auditing or future workspace policies.

## AppSuite note

Standalone AppCMS uses workspace `root`. Composed AppSuite profiles add workspace-aware **policy** evaluation in the launcher; AppCMS module code should remain workspace-agnostic where possible and rely on `StoragePort` workspace IDs.

## Security checklist for production

- [ ] Replace development session secret
- [ ] Use persistent auth store (SQLite auth tables migrated)
- [ ] Disable default `admin`/`admin` credentials
- [ ] Terminate TLS at the edge
- [ ] Scope app tokens to minimum capabilities

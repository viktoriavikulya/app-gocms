# Testing

AppCMS uses Go unit tests per package and HTTP integration tests at `cmd/server`.

## Test pyramid

| Level | Location | Examples |
|-------|----------|----------|
| Domain/application | `internal/application/*_test.go` | Service rules, repositories fakes |
| Storage | `internal/storage/*_test.go` | Provider, SQLite round-trip |
| Extensions | `internal/extensions/runtime_test.go` | Plugin lifecycle, hooks |
| HTTP integration | `cmd/server/main_test.go` | Full mux, auth, codex routes |

## Integration test helpers

| Helper | Purpose |
|--------|---------|
| `testHandler(t)` | Default app with seed + auth |
| `testHandlerOptions(t, opts)` | Custom `RuntimeMode`, headless |
| `testHandlerWithAuthStore` | Access auth memory store for tokens |
| `loginCookie(t, handler, user, pass)` | Session cookie for admin/REST |
| `postLogin` | Login flow with action token |

Apps built with `gocmsapp.NewApp(Options{Seed: true, ...})` and Framework middleware enabled.

## Key test cases (reference)

| Test | Guards |
|------|--------|
| `TestCodexRouteSurfaces` | Root URLs respond |
| `TestRESTContentCRUDRoundTrip` | REST CRUD + auth |
| `TestAuthSessionAndRESTAuthorization` | 401/403/token scope |
| `TestPublicSiteRendersThemeAndSlugs` | Themes + permalinks |
| `TestHeadlessDisablesPublicButKeepsREST` | Runtime gating |
| `TestGraphQLExtensionRouteAndVisibility` | GraphQL plugin |
| `TestOperationsHealthAuditAndSnapshot` | Ops plugin + snapshot |
| `TestRuntimeProfileModesGateSurfaces` | admin/conformance modes |
| `TestAppCMSDoesNotImportUI8Kit` | Import guard |

## Running subsets

```bash
go test ./cmd/server -run TestREST -v
go test ./internal/storage/sqlite -v
go test ./internal/extensions -v
```

## Auth testing patterns

- Unauthenticated mutations expect **401**
- Viewer creating content expects **403**
- App token with `content.write` only can create posts without admin access

## Storage testing

SQLite in-memory DSN avoids file locking:

```
file:appcms?mode=memory&cache=shared
```

`store_test.go` validates migrations and seed counts.

## Writing new tests

1. Prefer HTTP tests for route contracts
2. Use seeded data IDs from `SeedMinimalSite` (`hello-world`, `about`)
3. Encode JSON envelope shape assertions (`pagination`, error `code`)
4. For admin forms, assert `action_token` hidden input exists

## CI expectation

Same as local: `bun verify` in AppCMS repository root.

Platform changes may require `go test ./...` in Platform before merge.

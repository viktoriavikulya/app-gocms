# AppCMS

AppCMS is the runnable GoCMS replacement assembly for the Platform developer
preview.

It is a separate application repository. ModuleCMS owns the CMS domain schemas
and Panel descriptors. Platform owns renderer contracts and module hosting.
AppCMS owns the product shell, theme tokens, Tailwind pipeline, static files,
codex route surfaces, and the executable server.

## Package Layout

- `pkg/module`: CMS business module, records, relations, panel descriptors, codex helpers, and migration anchors.
- `pkg/templ`: reusable CMS admin templ views that can render under the AppCMS theme or inside AppSuite.
- `pkg/app`: standalone runtime plus reusable app bundle exported to launchers.
- `cmd/server`: thin executable entry point.

## Commands

```bash
bun install
bun verify
bun go
```

`bun go` serves:

- `http://127.0.0.1:8080/`
- `http://127.0.0.1:8080/go-admin/`
- `http://127.0.0.1:8080/go-json/`
- `http://127.0.0.1:8080/go-json/go/v2/`

## Current Admin Routes

- `/go-admin/posts`
- `/go-admin/pages`
- `/go-admin/content-types`
- `/go-admin/taxonomies`
- `/go-admin/terms`
- `/go-admin/meta`
- `/go-admin/media`
- `/go-admin/authors`

## Compatibility

AppCMS preserves the root GoCMS paths:

- `/go-admin`
- `/go-login`
- `/go-logout`
- `/go-json`
- `/go-json/go/v2/`

See [`docs/migration-from-gocms.md`](docs/migration-from-gocms.md) for the
current migration mapping and compatibility notes.

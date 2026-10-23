# Development and operations

Install the exact pnpm version from packageManager. Use the frozen lockfile in CI and for reproducibility. Dependency updates must update pnpm-lock.yaml and go.sum together with compatibility checks. Go tools use pinned `tool` directives in go.mod; their transitive dependencies may include alternate SQL drivers, but application persistence supports PostgreSQL only.

The latest npm TypeScript release exceeded both ESLint and OpenAPI generator peer ranges; the common supported release is pinned instead. Reka is used directly instead of adding redundant shadcn wrappers. Lucide Vue Next is retained as requested despite upstream deprecation; migrate to its replacement in a separately reviewed dependency change. Redis is diagnostic-only because no current feature justifies a cache or worker system.

## Database changes

Create sequential Goose SQL migrations under apps/core/migrations with Up and Down sections. Add only parameterized queries under apps/core/queries. Run `pnpm --filter @velora/core codegen`, migration integration tests and codegen freshness. Existing deployed migrations must never be edited. `pnpm velora migrate up` explicitly applies embedded migrations. Back up production data first. The initial down migration drops the foundation tables and is exercised only inside an isolated test schema.

## API changes

Edit docs/openapi.json, update handlers and tests, then run `pnpm --filter @velora/api-client codegen`. Do not edit generated TypeScript or sqlc output. Request cancellation uses AbortSignal through TanStack Query and openapi-fetch. Backend database calls inherit bounded contexts; Fiber does not guarantee automatic client-disconnect cancellation.

## Troubleshooting

- Startup database failure: verify PostgreSQL is ready, DATABASE_URL is correct and reachable, and migrations have run. Credentials are deliberately omitted from logs.
- Login blocked: check PUBLIC_URL exactly matches the browser origin. Do not interchange localhost and 127.0.0.1. Wait for the one-minute rate limit window after repeated failed attempts.
- Cookies missing: HTTPS and COOKIE_SECURE must agree. Production requires HTTPS; use explicit development mode only for local HTTP.
- Unknown API table: run the migration CLI. Readiness checks connectivity, not schema version; migration execution is a deployment prerequisite.
- Direct SPA route returns 404: configure index.html fallback except for /api. See docker/nginx.conf. The container configuration uses Docker DNS (127.0.0.11) to re-resolve Core after recreation; replace that resolver when adapting it outside Docker.
- Ports in use: change the local port and PUBLIC_URL together. Browser tests reserve 15173 and 18080.
- Native Go command ignores .env: export variables in your shell or use the pnpm Core wrappers.

## Releases and automation

All workspace packages are private. Changesets prepares future release notes but nothing publishes automatically. GitHub Actions uses immutable action commits; update them by resolving the upstream release tag and reviewing the change. Container base images use readable release tags; pin deployment-specific digests after your registry/security review. CI runs the same checks as local commands, with an isolated PostgreSQL service and Chromium.

The root `pnpm velora` command replaces package-specific migration/admin scripts. Run `pnpm velora migrate up`, then `pnpm velora admin create`; `pnpm dev` starts the CLI's serve command and Vite. Core is a library module; build the native executable with `pnpm --filter @velora/cli build`. Production can run it without JavaScript.

shadcn-vue was initialized with the supported CLI (Reka base, reka-nova style, Tailwind v4). `apps/console/components.json` records aliases and configuration. Add controls with `cd apps/console && pnpm dlx shadcn-vue@2.8.2 add <component>`. Preserve strict optional-property handling through the tested defined helper, and map colors to the semantic variables in src/styles/main.css.

The generator is invoked on demand rather than installed as an application dependency. Its MIT-licensed Tailwind support stylesheet is copied from shadcn-vue 2.8.2 under src/styles/vendor, alongside the upstream license. This avoids retaining the generator's vulnerable transitive glob/URI tooling in the application install. Review upstream advisories before running future generators, and refresh the vendored stylesheet when deliberately upgrading generated components.

The retained FormField composition has a specific purpose: it binds generated Label/Input primitives to existing form validation and consistent aria-invalid/aria-describedby error associations. It is not a parallel input implementation. Generated generic Button/Input/Card/Badge replacements are imported directly elsewhere.

# Velora

An original, open-source foundation for self-hosted game server and container hosting management. **Console** is a static Vue SPA. **Core** is a Go modular monolith. **Beacon**, the future external node daemon, is deliberately absent.

Implemented: explicit administrator bootstrap, Argon2id passwords, PostgreSQL-backed opaque sessions, login/logout/current-user APIs, administrator account management, profile/password changes, privileged password resets, health checks, and persisted Node metadata creation/list/detail. The Console uses real Core requests, generated shadcn-vue components backed by Reka UI, a responsive dark shell, validated forms and honest loading/error/empty states.

This foundation does not run containers or game servers. Instances and Blueprints have reserved foundational tables and clearly labeled empty pages. No live metrics, daemon connectivity, remote execution or public registration is simulated.

## Quick start

Prerequisites: Go 1.27, Node 24.19+, pnpm 11.21.0, and PostgreSQL 18. Docker is optional. The checked-in lockfiles determine exact dependencies.

```sh
cp .env.example .env
# Edit .env: choose a random POSTGRES_PASSWORD and use its URL-encoded value in DATABASE_URL.
pnpm install --frozen-lockfile
docker compose -f compose.dev.yaml up -d --wait postgres
pnpm velora migrate up
pnpm velora admin create
pnpm dev
```

Open http://localhost:5173. Enter the administrator credentials you chose, open Nodes, select Add Node, create a metadata record, inspect it, then use Account → Sign out. An administrator password must contain 12–128 characters. Interactive input hides the password; automation must explicitly use `--password-stdin` with credential JSON (see [CLI](docs/cli.md)). There is no default administrator.

The root CLI wrapper loads the root `.env` without overriding existing environment variables. For direct native Go execution, export your configuration first, then run the compiled `apps/cli/bin/velora serve`. No migration runs implicitly on startup. PostgreSQL connection failures stop startup within a bounded attempt.

Without Docker, provision PostgreSQL yourself and set DATABASE_URL. Redis is optional and disabled by default. `docker compose -f compose.dev.yaml --profile redis up -d redis` starts optional local Redis; enable it in `.env` only if desired. It currently receives a bounded connectivity check and provides no required application capability.

## Commands

| Command                               | Purpose                                                                   |
| ------------------------------------- | ------------------------------------------------------------------------- |
| `pnpm dev`                            | Core and Vite through Turborepo                                           |
| `pnpm build`                          | Static Console and the unified native velora executable                   |
| `pnpm check`                          | Format, lint, types, strict unit coverage, builds and generated freshness |
| `pnpm test`                           | Vue unit tests and race-enabled Go tests                                  |
| `pnpm test:coverage`                  | Hard 100% coverage gates                                                  |
| `pnpm test:integration`               | Isolated PostgreSQL repository/migration tests                            |
| `pnpm test:cli`                       | Linux PTY prompts, native CLI and isolated schema bootstrap               |
| `pnpm test:e2e`                       | Real Chromium authentication and Node workflow                            |
| `pnpm lint` / `pnpm lint:fix`         | ESLint, Go vet and golangci-lint                                          |
| `pnpm format` / `pnpm format:check`   | Prettier and goimports                                                    |
| `pnpm typecheck`                      | vue-tsc, shared TypeScript and Go vet                                     |
| `pnpm codegen` / `pnpm codegen:check` | sqlc and OpenAPI artifacts / freshness                                    |
| `pnpm clean`                          | Delete build and coverage outputs                                         |

See [exact dependency versions](docs/versions.md), [repository tree](docs/repository.md), [testing](docs/testing.md), [development](docs/development.md), [configuration](docs/configuration.md), [CLI](docs/cli.md), [migrations](docs/migrations.md), [authentication](docs/authentication.md), [architecture](docs/architecture.md) and [validation record](docs/validation.md).

## Docker operation

Choose a URL-safe random database password. For production set APP_ENV=production, PUBLIC_URL to your HTTPS origin, and COOKIE_SECURE=true. Place a TLS reverse proxy in front of the loopback-only Console port. For a local Docker trial use APP_ENV=development, PUBLIC_URL=http://localhost:8088 and COOKIE_SECURE=false.

```sh
docker compose build
docker compose up -d --wait postgres
docker compose run --rm core migrate up
docker compose run --rm core admin create
docker compose up -d --wait core console
```

Production Compose publishes only Console on `127.0.0.1:8088`. PostgreSQL and Redis are private. Core and nginx run as non-root users. Database storage persists in a named volume. Do not use `down -v` unless intentionally deleting the database. Optional Redis: `docker compose --profile redis up -d redis`, then set REDIS_ENABLED=true and recreate Core.

Console builds to `apps/console/dist`. A Node server is not needed in production. The included nginx configuration proxies `/api` to Core and falls back to `index.html` for history-mode routes. Only this build directory is public; `assets_temp` and root design references never enter image contexts or public roots.

## Next milestones

Define the external Beacon trust and enrollment contract, add audit history, verified reset-email delivery and MFA, introduce bounded pagination, and only then build workload/Blueprint orchestration. Before exposing a deployment, review TLS, backups, secret rotation, resource limits, dependency updates and your threat model. Passing scaffold checks is not a production-readiness certification.

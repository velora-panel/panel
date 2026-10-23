# Testing and coverage policy

`pnpm test:coverage` enforces exactly 100% statements, branches, functions and lines for authored Console executable source, and 100% native Go statement coverage for designated testable Core packages. These are hard failures. Go does not measure branch coverage.

Console scope: all src TypeScript and Vue components, including route guard, API error handling, query hooks, preference store, forms, shell and feature pages. Exclusions: tests/test harness, main.ts plugin/mount wiring, App.vue RouterView-only root and router/index.ts declarative route registration. Browser tests exercise excluded wiring. Generated API declarations, static assets and tool configuration contain no application logic and are outside this scope. The shared API package only re-exports generated types and the third-party transport.

Go scope: config, database repository adapter, auth service/password functions, Station service and HTTP transport. Explicit exemptions: cmd/velora process exit entry point; internal/app infrastructure wiring, OS signals, terminal input and lifecycle (validated by CLI/browser/container checks); generated sqlc. Domain DTOs and embedded migrations have no executable statements. Do not place business logic into exempt wiring to avoid testing it. Coverage checker rejects empty scope and any uncovered statement. Repository integration tests are separate and do not inflate unit coverage.

## Isolated service tests

```sh
docker run -d --name velora-tests \
  -e POSTGRES_USER=velora -e POSTGRES_PASSWORD=test-only \
  -e POSTGRES_DB=velora_test -p 127.0.0.1:55432:5432 postgres:18-alpine
export TEST_DATABASE_URL='postgres://velora:test-only@127.0.0.1:55432/velora_test'
pnpm test:integration
pnpm --filter @velora/console exec playwright install chromium
pnpm test:e2e
docker rm -f velora-tests
```

Wait for `docker exec velora-tests pg_isready -U velora` before tests. Both suites reject database names not ending in `_test`. Never use a normal database renamed with this suffix. Integration tests create a random schema, migrate it, test constraints, sessions, queries and transaction rollback, then drop only that schema. E2E tests migrate the dedicated database and create random test-only accounts/Stations; discard the disposable database after testing. No test fixture is presented as operational infrastructure.

Unit tests make no external network calls. Fake repositories exercise service and HTTP failure branches. Vue tests use real forms, query state and component interactions; only API calls and portal placement are controlled. Playwright uses real Core HTTP requests, real PostgreSQL and a randomly generated account, checks protected-route redirect, login, navigation, creation, detail reload and logout invalidation.

`pnpm check` runs non-service quality gates. Integration and browser tests remain explicit because they need PostgreSQL. `pnpm audit --audit-level high` and `pnpm --filter @velora/core vuln` need network access to advisory databases. CI also builds both Docker images. Check docs/validation.md for actual results rather than assuming configuration implies success.

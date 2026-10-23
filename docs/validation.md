# Validation record

Verified locally on October 21–22, 2026. These results describe this scaffold, not a certification for production hosting.

| Check                          | Actual result                                                                                                                                                                                                    |
| ------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Frozen pnpm install            | Passed locally and from a clean install inside the Console Docker build                                                                                                                                          |
| Dependency compatibility       | pnpm peer checks: no issues; TypeScript 5.9.3 is the supported common version                                                                                                                                    |
| Go dependency resolution       | Passed; go.mod/go.sum and tool directives present in the workspace                                                                                                                                               |
| Root `pnpm check`              | Passed: Prettier/goimports, ESLint with zero warnings, vue-tsc/tsc, Go vet, golangci-lint, coverage and both builds                                                                                              |
| Frontend unit tests            | 13 tests in six files passed                                                                                                                                                                                     |
| Frontend coverage              | 100% statements (159/159), branches (162/162), functions (64/64) and lines (138/138)                                                                                                                             |
| Backend unit tests             | Passed with race detector; 100% statements (221/221) in designated packages                                                                                                                                      |
| PostgreSQL integration         | Passed against PostgreSQL 18: schema isolation, migration up/down/up, CRUD, uniqueness, foreign keys, expiry, revocation and failed-transaction rollback                                                         |
| Chromium smoke test            | Passed against real Core/PostgreSQL; login, protected routes, navigation, Station creation/detail/reload and logout revocation                                                                                   |
| Mobile check                   | Passed at 390×844 with no horizontal overflow; screenshots inspected                                                                                                                                             |
| Generated freshness            | Both sqlc and OpenAPI regeneration produced identical tracked artifacts                                                                                                                                          |
| Root development startup       | `pnpm dev` started Core and Vite; readiness returned 200                                                                                                                                                         |
| Core Docker build              | Passed, non-root runtime                                                                                                                                                                                         |
| Console Docker build           | Passed, frozen install, static build and non-root nginx                                                                                                                                                          |
| Compose configuration          | Production and development definitions validated with official checksum-verified Compose v2.39.4                                                                                                                 |
| Design reference isolation     | Vite filesystem request to assets_temp returned 403; references excluded from image contexts                                                                                                                     |
| Container runtime workflow     | Passed: explicit migration/bootstrap, login, persisted Station creation, logout, SPA fallback, 404 for design references, Redis-disabled and Redis-enabled readiness, and Core recreation with dynamic proxy DNS |
| JavaScript vulnerability audit | No known vulnerabilities reported                                                                                                                                                                                |
| Go vulnerability scan          | No affected symbols or imported packages; module-only advisory GO-2026-5932 concerns unused x/crypto/openpgp, not the imported Argon2 implementation                                                             |

Coverage scope and explicit exemptions are recorded in [testing.md](testing.md). Go percentages are statement coverage only. CI is configured but was not run on GitHub from this workspace. The workspace does not contain accessible Git repository metadata, so no commit, push, PR or release was created. Lockfiles and generated files are present for review and commit.

The host lacks the Docker Compose plugin; verification used a temporary verified official binary. Contributors should install Docker Compose v2 and use the documented `docker compose` commands. Playwright used its Ubuntu fallback Chromium build successfully on this host.

See [versions](versions.md) for exact resolved dependencies, [repository structure](repository.md), and README.md for native and Docker startup instructions.

Temporary validation containers/volumes and development servers were stopped after verification. Built images and local build/test artifacts remain available. The Compose recreation test initially found stale nginx DNS resolution; the supplied configuration now resolves Docker service addresses dynamically, and the repeated test passed.

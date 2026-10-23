# Configuration

| Variable                    | Meaning / default                                                                          |
| --------------------------- | ------------------------------------------------------------------------------------------ |
| APP_ENV                     | Only `development` permits insecure HTTP cookies. Other values enforce production rules.   |
| DATABASE_URL                | Required PostgreSQL URL; never logged. Use verified TLS for remote databases.              |
| PUBLIC_URL                  | Required exact Console origin, no path/trailing slash/query/userinfo; HTTPS in production. |
| COOKIE_SECURE               | Boolean, false by default; must be true outside development.                               |
| HTTP_HOST                   | Bind address, defaults to 127.0.0.1; containers explicitly use 0.0.0.0.                    |
| HTTP_PORT                   | 1–65535, defaults to 8080.                                                                 |
| LOG_LEVEL                   | DEBUG, INFO, WARN or ERROR; default INFO.                                                  |
| REDIS_ENABLED               | Boolean, default false.                                                                    |
| REDIS_URL                   | Required redis:// or rediss:// URL when enabled.                                           |
| POSTGRES_USER / POSTGRES_DB | Compose initialization only, default velora.                                               |
| POSTGRES_PASSWORD           | Required by Compose; choose a random URL-safe value.                                       |
| TEST_DATABASE_URL           | Integration/E2E only; database name must end in `_test`.                                   |
| CORE_PROXY                  | Vite tooling only; backend proxy target, defaults to http://127.0.0.1:8080.                |

No SESSION_SECRET is required: sessions use random opaque tokens whose hashes are stored in PostgreSQL, not signed client-side state. No server secret belongs in a VITE_* variable. No wildcard CORS is enabled; same-origin deployment is intentional. Environment files are ignored and excluded from Docker contexts.

Core uses ten maximum PostgreSQL connections, one minimum, a one-hour connection lifetime and a five-second connection timeout. Requests receive a five-second context deadline; readiness uses two seconds. HTTP read/write/idle limits are 10/15/60 seconds. Body limit is 16 KiB. Shutdown has ten seconds to drain.

`MIGRATION_DATABASE_URL` optionally selects a separate owner/migration role for `velora migrate` commands. It defaults to DATABASE_URL when absent. Runtime commands always use DATABASE_URL. Keep both in ignored local environment files or a secret manager. Native binaries consume exported environment; only the repository pnpm wrapper loads root .env. See migrations.md for schema permissions and grants.

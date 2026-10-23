# 001 — Modular Core, static Console

Core is a native Go module orchestrated through thin pnpm scripts. Console is a Vue history-mode SPA served by a reverse proxy. PostgreSQL owns users, expiring opaque sessions and Node metadata. Beacon belongs to a separate repository and no remote execution is implemented here.

The OpenAPI contract is the source of truth for public TypeScript types. SQL migrations and sqlc queries own persistence types. Handlers map between these boundaries. Authentication uses Argon2id, hashed opaque session tokens, same-origin mutation checks and HttpOnly SameSite cookies. Only administrators create Node metadata.

Redis is optional operational connectivity, not an authentication dependency. Login limits initially apply to a single Core process; deploy one replica until a distributed limiter is implemented.

Only the selected velora.svg reference is copied into Console source assets. assets_temp is never served or bundled wholesale.

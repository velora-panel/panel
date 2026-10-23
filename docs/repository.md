# Repository layout

- apps/cli: single native executable, command parsing and terminal input.
- apps/core: modular Go monolith; application facade, domain/services, PostgreSQL repositories, Fiber HTTP transport, embedded migrations and sqlc queries.
- apps/console: Vue SPA; generated shadcn-vue primitives, account/auth/Nodes features, TanStack Query and local navigation preferences.
- packages/api-client: OpenAPI-generated types and typed fetch client.
- packages/eslint-config and packages/typescript-config: shared quality configuration.
- tooling: native CLI launcher, formatting/coverage/codegen checks and Linux PTY regression test.
- docker and Compose files: non-root deployment and local dependencies.
- docs: architecture, configuration, authentication, CLI, migrations and verification.

Root assets and assets_temp are private reference directories and never public roots. go.work joins Core and CLI; pnpm-workspace.yaml joins frontend and tooling packages.

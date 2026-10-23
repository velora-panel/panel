# Repository structure

Build artifacts, dependency installations and tool state are omitted.

```text
velora-panel/
├── .changeset/
│   ├── README.md
│   └── config.json
├── .github/
│   └── workflows/
│       └── ci.yaml
├── apps/
│   ├── console/
│   │   ├── e2e/
│   │   │   ├── setup.ts
│   │   │   └── smoke.spec.ts
│   │   ├── src/
│   │   │   ├── app/
│   │   │   │   └── App.vue
│   │   │   ├── assets/
│   │   │   │   └── velora-mark.svg
│   │   │   ├── components/
│   │   │   │   ├── Badge.vue
│   │   │   │   ├── Button.vue
│   │   │   │   ├── Card.vue
│   │   │   │   ├── FormField.vue
│   │   │   │   ├── Input.vue
│   │   │   │   ├── PageHeader.vue
│   │   │   │   ├── Placeholder.vue
│   │   │   │   ├── State.vue
│   │   │   │   └── components.test.ts
│   │   │   ├── features/
│   │   │   │   ├── auth/
│   │   │   │   │   ├── Login.vue
│   │   │   │   │   ├── Settings.vue
│   │   │   │   │   └── auth.test.ts
│   │   │   │   ├── dashboard/
│   │   │   │   │   ├── Dashboard.vue
│   │   │   │   │   └── dashboard.test.ts
│   │   │   │   └── stations/
│   │   │   │       ├── StationDetail.vue
│   │   │   │       ├── Stations.vue
│   │   │   │       ├── queries.ts
│   │   │   │       └── stations.test.ts
│   │   │   ├── layouts/
│   │   │   │   ├── Shell.vue
│   │   │   │   └── shell.test.ts
│   │   │   ├── lib/
│   │   │   │   ├── api.test.ts
│   │   │   │   └── api.ts
│   │   │   ├── router/
│   │   │   │   ├── guard.ts
│   │   │   │   └── index.ts
│   │   │   ├── stores/
│   │   │   │   └── preferences.ts
│   │   │   ├── styles/
│   │   │   │   └── main.css
│   │   │   ├── test/
│   │   │   │   ├── helpers.ts
│   │   │   │   └── setup.ts
│   │   │   └── main.ts
│   │   ├── eslint.config.mjs
│   │   ├── index.html
│   │   ├── package.json
│   │   ├── playwright.config.ts
│   │   ├── tsconfig.json
│   │   ├── vite.config.ts
│   │   └── vitest.config.ts
│   └── core/
│       ├── cmd/
│       │   └── velora/
│       │       └── main.go
│       ├── internal/
│       │   ├── app/
│       │   │   └── run.go
│       │   ├── config/
│       │   │   ├── config.go
│       │   │   └── config_test.go
│       │   ├── database/
│       │   │   ├── generated/
│       │   │   │   ├── db.go
│       │   │   │   ├── models.go
│       │   │   │   ├── querier.go
│       │   │   │   └── queries.sql.go
│       │   │   ├── integration_test.go
│       │   │   ├── repository.go
│       │   │   └── repository_test.go
│       │   ├── domain/
│       │   │   └── models.go
│       │   ├── modules/
│       │   │   ├── auth/
│       │   │   │   ├── password.go
│       │   │   │   ├── service.go
│       │   │   │   └── service_test.go
│       │   │   └── stations/
│       │   │       ├── service.go
│       │   │       └── service_test.go
│       │   └── transport/
│       │       └── http/
│       │           ├── server.go
│       │           └── server_test.go
│       ├── migrations/
│       │   ├── 00001_foundation.sql
│       │   └── embed.go
│       ├── queries/
│       │   └── queries.sql
│       ├── go.mod
│       ├── go.sum
│       ├── package.json
│       └── sqlc.yaml
├── assets_temp/
│   ├── color_scheme.css
│   ├── velora.png
│   ├── velora.svg
│   ├── velora_inverse.svg
│   └── velora_monochrome.svg
├── docker/
│   ├── console.Dockerfile
│   ├── core.Dockerfile
│   └── nginx.conf
├── docs/
│   ├── adr/
│   │   └── 001-foundation.md
│   ├── architecture.md
│   ├── configuration.md
│   ├── development.md
│   ├── openapi.json
│   ├── repository.md
│   ├── testing.md
│   ├── validation.md
│   └── versions.md
├── packages/
│   ├── api-client/
│   │   ├── src/
│   │   │   ├── index.ts
│   │   │   └── schema.d.ts
│   │   ├── package.json
│   │   └── tsconfig.json
│   ├── eslint-config/
│   │   ├── index.mjs
│   │   └── package.json
│   └── typescript-config/
│       ├── base.json
│       └── package.json
├── tooling/
│   ├── codegen-check.mjs
│   ├── core.mjs
│   ├── go-coverage.mjs
│   └── go-format.mjs
├── .dockerignore
├── .editorconfig
├── .env.example
├── .gitignore
├── .golangci.yml
├── .node-version
├── .prettierignore
├── .prettierrc.json
├── AGENTS.md
├── CONTRIBUTING.md
├── LICENSE
├── README.md
├── SECURITY.md
├── compose.dev.yaml
├── compose.yaml
├── package.json
├── pnpm-lock.yaml
├── pnpm-workspace.yaml
└── turbo.json
```

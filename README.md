# TreeBlog implementation

This repository is the application workspace for the supplied [TreeBlog blueprint](./BN_TREEBLOG_MASTER_BLUEPRINT/START_HERE.md). The specification describes a multi-phase platform; this repository currently contains the initial monorepo and infrastructure foundation.

## Current state

Measured 2026-11-24. Details live in [IMPLEMENTATION_STATUS.md](./IMPLEMENTATION_STATUS.md),
[docs/BLUEPRINT_ERRATA.md](./docs/BLUEPRINT_ERRATA.md), and [docs/COMMERCIAL_GAP_ANALYSIS.md](./docs/COMMERCIAL_GAP_ANALYSIS.md).

- Go API: 16 domain modules with 207 handler methods, migrations `0001`~`0013`
  (78 `CREATE TABLE` statements), transactional outbox, Asynq worker, and 63 test files.
  `go vet ./...` is clean and `go test ./...` passes in all 28 packages,
  including DB-backed integration tests against a disposable PostgreSQL container.
- OpenAPI contract: 128 paths / 142 operations in `04_API/openapi.yaml`.
  `packages/api-client` generates typed clients from it with a CI drift gate;
  frontends still use hand-written clients pending migration.
- Frontends build and typecheck: Nuxt user web (19 pages), Vue blog console
  (15 feature routes), Vue admin console (7 feature routes). They are wired
  beyond scaffolds but still partial against the 46-screen inventory. The shared
  `editor-core` package now ships AST types, document validation, and text
  helpers with tests, and the console editor adds table blocks, pre-save
  validation, word counts, and IndexedDB offline drafts.
- Docker Compose describes local PostgreSQL, Redis, OpenSearch, MinIO, Mailpit,
  migration, and API services. Full Compose execution, E2E tests, and the
  independent audit are still pending, so no module is marked `VERIFIED`.

## Local development

1. Install Go 1.23+, Node 24+, pnpm 12, and a Docker-compatible container runtime.
2. Copy `.env.example` to `.env` and replace every placeholder credential. If the PostgreSQL password contains URL-special characters, percent-encode it for the `TREEBLOG_DATABASE_URL` values in Compose.
3. Generate local-only secrets with `./infra/dev/generate-keys.sh` (or `infra/dev/generate-keys.ps1` on Windows) and paste the printed `TREEBLOG_CODE_PEPPER` into `.env`.
4. Run `docker compose up --build` to start local services and apply the first two migrations.
5. Run `pnpm install`, `pnpm typecheck`, and `pnpm build` for the frontend workspaces.
6. Run `go vet ./...`, `go test ./...`, and `go build ./...` for Go.

The Go migration command can also be run directly with `TREEBLOG_DATABASE_URL` set: `go run ./cmd/migrate up`. `down-one` reverts only the latest applied migration; use it only against a disposable development database.

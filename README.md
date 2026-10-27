# TreeBlog implementation

This repository is the application workspace for the supplied [TreeBlog blueprint](./BN_TREEBLOG_MASTER_BLUEPRINT/START_HERE.md). The specification describes a multi-phase platform; this repository currently contains the initial monorepo and infrastructure foundation.

## Current state

- Go API bootstrapping, standard JSON health/error envelopes, configuration, graceful shutdown, and module boundaries are implemented.
- The first two PostgreSQL migrations are available in [`migrations/`](./migrations/), with a migration runner and a CI integration test.
- Nuxt user web, Vue blog console, Vue admin console, and shared TypeScript package workspaces build and typecheck. Their pages are scaffolds, not completed product features.
- Docker Compose describes local PostgreSQL, Redis, OpenSearch, MinIO, Mailpit, migration, and API services. It has not been run on this machine because Docker is unavailable.

See [IMPLEMENTATION_STATUS.md](./IMPLEMENTATION_STATUS.md) for work-order status and known verification gaps.

## Local development

1. Install Go 1.23+, Node 24+, pnpm 12, and a Docker-compatible container runtime.
2. Copy `.env.example` to `.env` and replace every placeholder credential. If the PostgreSQL password contains URL-special characters, percent-encode it for the `TREEBLOG_DATABASE_URL` values in Compose.
3. Run `docker compose up --build` to start local services and apply the first two migrations.
4. Run `pnpm install`, `pnpm typecheck`, and `pnpm build` for the frontend workspaces.
5. Run `go vet ./...`, `go test ./...`, and `go build ./...` for Go.

The Go migration command can also be run directly with `TREEBLOG_DATABASE_URL` set: `go run ./cmd/migrate up`. `down-one` reverts only the latest applied migration; use it only against a disposable development database.

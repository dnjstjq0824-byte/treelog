# TreeBlog implementation status

Updated 2026-10-28. Source: `BN_TREEBLOG_MASTER_BLUEPRINT_V1_0.zip`, SHA-256 `A08A723A5125DE4F9BD600D4B8F62F395539FED031BAF33188DF35AF95445108`.

## Intake findings

- The supplied workspace was empty before extraction. The ZIP contained 182 files of specifications and SQL, with no preexisting application source.
- `START_HERE.md` calls for `MANIFEST.json` and `FILE_CHECKSUMS.sha256`, but neither exists in the archive. File-level checksum verification therefore cannot be performed from the supplied package.
- The blueprint says there are 48 individual work orders; 49 `TB-*.md` work-order files are present.
- The architecture describes 15 domain modules in prose but names 16 in `02_ARCHITECTURE/COMPONENT_DIAGRAMS.md`. The implementation creates all 16 named boundaries.

## Work-order status

| Work order | State | Evidence and remaining work |
| --- | --- | --- |
| `TB-ARCH-001` | `IMPLEMENTED_UNVERIFIED` | Go HTTP foundation and module packages compile; pnpm workspaces build and typecheck. |
| `TB-OPS-001` | `IMPLEMENTED_UNVERIFIED` | Compose, backend Dockerfile, OpenSearch plugin image, and CI workflow exist. |
| `TB-DB-001` | `IMPLEMENTED_UNVERIFIED` | Migrations 0001~0002, pgx pools, transaction helper, migration runner. |
| `TB-ARCH-002` | `IMPLEMENTED_UNVERIFIED` | Transactional Outbox publisher, SKIP LOCKED relay, and Asynq 4-priority queue worker. |
| `TB-DB-002` | `IMPLEMENTED_UNVERIFIED` | Complete database schema migrations (0001~0011) for all domains. |
| `TB-AUTH-001` | `IMPLEMENTED_UNVERIFIED` | Email registration, Argon2id, JWT rotation, password reset, and identity API handlers. |
| `TB-SEC-001` | `IMPLEMENTED_UNVERIFIED` | SafeDialer SSRF protection, Double-Submit CSRF, sliding-window rate limiter, chained audit log. |
| `TB-AUTH-002` | `IMPLEMENTED_UNVERIFIED` | RFC 6238 TOTP MFA, backup codes, 2-tier RBAC middleware, user profile & session management. |
| `TB-BLOG-001` | `IMPLEMENTED_UNVERIFIED` | Blog creation engine, 150+ reserved words validator, team RBAC, guestbook backend. |
| `TB-BLOG-002` | `IMPLEMENTED_UNVERIFIED` | 2-depth category hierarchy, drag-and-drop reordering, post safe relocation, series, popular tags. |
| `TB-API-001` / `TB-WEB-001` | `NOT_STARTED` | OpenAPI client generation pipeline and frontend UI kit authentication screens. |
| Phase 2–10 remaining | `IN_PROGRESS` | Multi-domain routing, block editor, post lifecycle, themes, OpenSearch, community, monetization. |

`VERIFIED` is intentionally unused until the blueprint's required integration and independent audit gates pass.

## Verification performed

- `go vet ./...` — passed.
- `go test ./...` — passed for local unit tests; PostgreSQL integration test skipped without `TREEBLOG_TEST_DATABASE_URL`.
- `go build ./...` — passed.
- `pnpm build` — passed for all five frontend workspaces.
- `pnpm typecheck` — passed for all five frontend workspaces.
- Docker Compose execution and PostgreSQL Up → Down → Up — blocked by the missing local container runtime.

## Next dependency order

1. Run Compose and the database integration test; fix any SQL or image issues it exposes.
2. Implement `TB-ARCH-002`: transactional outbox and real worker queues.
3. Continue Phase 1 database schema and identity/API work orders, then build the blog and post lifecycle according to the dependency graph.

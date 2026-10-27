# CI_CD.md — GitHub Actions 기반 자동화 빌드·검증·보안스캔·배포 파이프라인

## 1. PR 및 Main 브랜치 필수 검증 게이트 (`.github/workflows/ci.yml`)
모든 Pull Request는 아래 6개 자동화 스테이지를 **모두 통과해야만** Merge가 허용된다.
1. **Stage 1 — Architecture & Ownership Lint**: 에이전트 수정 허용 디렉터리 준수 검사, Go 모듈 순환 참조 검사, 하드코딩 시크릿 스캔(`gitleaks`).
2. **Stage 2 — Static Analysis & Typecheck**: `golangci-lint run`, `govulncheck ./...`, `pnpm lint`, `pnpm typecheck` (`vue-tsc --noEmit`), OpenAPI 3.1.0 스키마 유효성 검사.
3. **Stage 3 — Database Migration Test**: PostgreSQL 16 서비스 컨테이너에서 `0001`~`0010` Up → Down → Up 전수 실행 검증.
4. **Stage 4 — Backend Unit & Real-Container Integration Tests**: Postgres + Redis + OpenSearch + MinIO 컨테이너 환경에서 `go test -race -cover ./...` 실행 (Mock 대체 불가).
5. **Stage 5 — Security Penetration Suite**: XSS 50종, SSRF 차단, IDOR 권한 매트릭스, EXIF 제거 검증 테스트 실행.
6. **Stage 6 — Playwright E2E & Accessibility Audit**: 8대 핵심 사용자 여정 E2E 및 `axe-core` 웹 접근성, 320px 모바일 반응형 검사.

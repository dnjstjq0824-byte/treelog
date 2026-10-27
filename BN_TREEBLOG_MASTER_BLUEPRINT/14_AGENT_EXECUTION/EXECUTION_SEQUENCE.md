# EXECUTION_SEQUENCE.md — Phase 0 ~ Phase 10 단계별 병렬 실행 시퀀스

## 1. Wave 기반 병렬 실행 스케줄
- **Wave 1 (Phase 0 — 기반 구축)**: `TB-ARCH-001` → `TB-OPS-001` → `TB-DB-001` → `TB-ARCH-002` (순차 실행, 전 에이전트 공통 기반 확립)
- **Wave 2 (Phase 1 — 인증 및 전체 스키마)**: `TB-DB-002` 완료 후 → `[TB-AUTH-001, TB-SEC-001]` 병렬 실행 → `[TB-AUTH-002, TB-API-001]` 병렬 실행 → `TB-WEB-001`
- **Wave 3 (Phase 2 — 블로그 및 도메인 엔진)**: `TB-BLOG-001` 완료 후 → `[TB-BLOG-002, TB-BLOG-003, TB-DB-003]` 3-Way 병렬 실행 → `TB-WEB-002`
- **Wave 4 (Phase 3 & 4 — 에디터·미디어·발행·테마·사용자웹 대규모 병렬 구간)**:
  - Track A (Media & Editor): `TB-MEDIA-001` → `[TB-MEDIA-002, TB-EDITOR-001]` → `TB-EDITOR-002` → `TB-EDITOR-003`
  - Track B (Publishing): `TB-EDITOR-001` 완료 즉시 `TB-POST-001` → `TB-POST-002`
  - Track C (Theme & User Web): `TB-THEME-001` → `[TB-THEME-002, TB-WEB-003]` → `TB-WEB-004`
- **Wave 5 (Phase 5 & 6 — 검색·SEO·커뮤니티·이동성 병렬 구간)**:
  - Track D (Search & SEO): `[TB-SEARCH-001, TB-SEO-001]` → `[TB-SEARCH-002, TB-SEO-002]`
  - Track E (Community & Portability): `TB-COMMUNITY-001` → `TB-COMMUNITY-002` → `TB-COMMUNITY-003` / `TB-POST-003`
- **Wave 6 (Phase 7 & 8 — 모더레이션·어드민·통계·수익화 병렬 구간)**:
  - Track F (Moderation & Admin): `TB-SEC-002` → `TB-ADMIN-001` → `TB-ADMIN-002` → `TB-ADMIN-003`
  - Track G (Analytics & Monetization): `[TB-ANL-001, TB-BILLING-001]` → `TB-ADS-001` → `TB-BILLING-002`
- **Wave 7 (Phase 9 & 10 — 상용화 최적화·TreeScan 준비·최종 독립 감사)**:
  - `[TB-OPS-002, TB-OPS-003, TB-TREESCAN-001]` → `[TB-QA-001, TB-TREESCAN-002]` → `TB-QA-002` → **`TB-QA-003` (Final Independent Audit)**

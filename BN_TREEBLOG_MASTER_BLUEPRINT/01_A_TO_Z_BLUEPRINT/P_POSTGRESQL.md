# P_POSTGRESQL — PostgreSQL & Data Architecture (데이터베이스 전체 아키텍처)
> **부제**: 전체 ERD, 실제 DDL, 인덱스, 제약조건, 마이그레이션, 파티셔닝 필요성, 트랜잭션, 데이터 보존
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-P_POSTGRESQL`
> **적용 조항**: 제3조 (P_POSTGRESQL 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- TreeBlog의 모든 도메인 데이터를 유실이나 정합성 깨짐 없이 영구 보존하는 **단일 진실 공급원(Single Source of Truth)**으로서 PostgreSQL 16+ 기반의 58개 핵심 테이블 스키마, 외래키/체크 제약조건, 고성능 인덱스, 월별 파티셔닝, 순차 마이그레이션 및 롤백 체계를 정의한다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **PostgreSQL Agent / Backend Developer Agent**: `03_DATABASE/migrations/`의 SQL 스크립트를 순서대로 실행하여 스키마를 구성하고, `sqlc` 또는 `pgx`를 통해 타입 안전한 쿼리와 트랜잭션을 작성한다.
- **DBA / DevOps Agent**: 대용량 시계열 테이블 파티션을 자동 생성(`pg_partman` 또는 워커 배치)하고 WAL 아카이빙 및 Read Replica 복제 지연을 모니터링한다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-DB-001 (제14조 10대 도메인 전체 핵심 엔터티 및 확장 테이블 총 58개 구현)**:
  1) **Identity (8)**: `users`, `user_profiles`, `identities`, `sessions`, `roles`, `permissions`, `role_permissions`, `user_roles`, `user_consents` (9개)
  2) **Blog (8)**: `blogs`, `blog_members`, `blog_settings`, `blog_domains`, `themes`, `blog_themes`, `blog_navigation`, `blog_categories`
  3) **Content (11)**: `posts`, `post_revisions`, `autosave_drafts`, `post_categories`, `tags`, `post_tags`, `post_media`, `post_visibility_rules`, `scheduled_publications`, `series`, `post_slug_history`
  4) **Media (4)**: `upload_sessions`, `media_assets`, `media_variants`, `media_processing_jobs`
  5) **Community (8)**: `neighbor_groups`, `follows`, `subscriptions`, `reactions`, `comments`, `bookmarks`, `guestbook_entries`, `user_blocks`
  6) **Search (4)**: `search_index_jobs`, `search_queries`, `search_click_events`, `content_quality_signals`
  7) **Moderation (5)**: `user_reports`, `moderation_cases`, `moderation_actions`, `appeals`, `spam_detection_results`
  8) **Monetization (10)**: `ad_accounts`, `ad_campaigns`, `ad_creatives`, `ad_placements`, `ad_events`, `memberships`, `user_subscriptions`, `orders`, `payments`, `payouts`, `ledger_entries` (11개)
  9) **Operations & Integration (7)**: `audit_logs`, `feature_flags`, `system_settings`, `outbox_events`, `background_jobs`, `migration_jobs`, `treescan_sync_state`
  *(총 58개 테이블 완비 — 상세 DDL은 `03_DATABASE/migrations/`에 수록)*
- **REQ-DB-002 (엄격한 무결성 제약조건 및 삭제 정책)**:
  - 모든 테이블에 명시적 `PRIMARY KEY`, `FOREIGN KEY ... ON DELETE {CASCADE|RESTRICT|SET NULL}`, `UNIQUE`, `CHECK` 제약조건(예: 상태값 열거형 검사, 음수 금액 방지 `amount_minor >= 0`)을 부여한다.
  - 금전/감사 테이블(`orders`, `payments`, `payouts`, `ledger_entries`, `audit_logs`)은 `ON DELETE RESTRICT`를 적용하여 부모 레코드 삭제로 인해 거래 기록이 연쇄 삭제되는 것을 원천 금지한다.
- **REQ-DB-003 (대용량 테이블 파티셔닝 전략)**:
  - 빠르게 증가하는 이벤트/로그성 테이블 4종(`ad_events`, `search_click_events`, `audit_logs`, `notifications`)은 `created_at` 기준 **Monthly Range Partitioning**을 적용하여 오래된 파티션 아카이빙 및 인덱스 크기 최적화를 달성한다.

## 4. 데이터 구조 (Data Structures & Domain Models)
```sql
-- 핵심 posts 테이블 DDL 발췌 (전체 DDL은 03_DATABASE/migrations/0003_content.up.sql 참조)
CREATE TABLE posts (
    id UUID PRIMARY KEY,
    blog_id UUID NOT NULL REFERENCES blogs(id) ON DELETE CASCADE,
    author_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    category_id UUID REFERENCES blog_categories(id) ON DELETE SET NULL,
    series_id UUID REFERENCES series(id) ON DELETE SET NULL,
    slug VARCHAR(220) NOT NULL,
    title VARCHAR(300) NOT NULL,
    subtitle VARCHAR(500),
    content_json JSONB NOT NULL,
    content_html TEXT NOT NULL,
    content_plain TEXT NOT NULL,
    excerpt VARCHAR(600),
    cover_media_id UUID,
    language_code VARCHAR(16) NOT NULL DEFAULT 'ko',
    translation_group_id UUID,
    status VARCHAR(24) NOT NULL DEFAULT 'DRAFT'
        CHECK (status IN ('DRAFT','IN_REVIEW','SCHEDULED','PUBLISHED','UNPUBLISHED','QUARANTINED','TRASHED','PURGED')),
    visibility VARCHAR(24) NOT NULL DEFAULT 'PUBLIC'
        CHECK (visibility IN ('PUBLIC','UNLISTED','NEIGHBOR_ONLY','MUTUAL_ONLY','PAID_MEMBERS_ONLY','PRIVATE')),
    allow_comments VARCHAR(24) NOT NULL DEFAULT 'ALL_ALLOWED'
        CHECK (allow_comments IN ('ALL_ALLOWED','NEIGHBORS_ONLY','MUTUAL_ONLY','REQUIRE_APPROVAL','DISABLED')),
    allow_reactions BOOLEAN NOT NULL DEFAULT TRUE,
    is_notice BOOLEAN NOT NULL DEFAULT FALSE,
    is_pinned BOOLEAN NOT NULL DEFAULT FALSE,
    seo_title VARCHAR(200),
    seo_description VARCHAR(320),
    canonical_url VARCHAR(1024),
    license_type VARCHAR(32) NOT NULL DEFAULT 'ALL_RIGHTS_RESERVED',
    ai_disclosure_level VARCHAR(24) NOT NULL DEFAULT 'NONE',
    word_count INTEGER NOT NULL DEFAULT 0 CHECK (word_count >= 0),
    reading_time_sec INTEGER NOT NULL DEFAULT 0 CHECK (reading_time_sec >= 0),
    view_count BIGINT NOT NULL DEFAULT 0 CHECK (view_count >= 0),
    reaction_count INTEGER NOT NULL DEFAULT 0 CHECK (reaction_count >= 0),
    comment_count INTEGER NOT NULL DEFAULT 0 CHECK (comment_count >= 0),
    bookmark_count INTEGER NOT NULL DEFAULT 0 CHECK (bookmark_count >= 0),
    revision_no INTEGER NOT NULL DEFAULT 1 CHECK (revision_no >= 1),
    published_at TIMESTAMPTZ,
    trashed_at TIMESTAMPTZ,
    purge_after_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (blog_id, slug)
);
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- N/A (데이터베이스 아키텍처는 백엔드 및 워커 계층을 통해 캡슐화되며 프론트엔드가 직접 DB에 접근하지 않음).

## 6. 백엔드 로직 (Backend Logic & Processing)
- `internal/db`: `golang-migrate` 호환 마이그레이션 러너, `pgx/v5` 트랜잭션 헬퍼(`WithTx(ctx, func(tx pgx.Tx) error)`), Read/Write 커넥션 풀 라우터.

## 7. API 계약 (API Contracts)
- N/A (DB는 내부 영속성 계층이며 모든 외부 인터페이스는 `04_API/openapi.yaml`을 통해 노출됨).

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- `03_DATABASE/ERD.md`, `TABLE_DICTIONARY.md`, `INDEX_STRATEGY.md`, 그리고 `03_DATABASE/migrations/0001_identity.up.sql` ~ `0010_operations_and_treescan.up.sql` (및 대응 `.down.sql`)에 전체 58개 테이블 DDL이 완벽히 정의됨.

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **데드락(Deadlock) 및 직렬화 오류 대응**: 다중 행 업데이트 시 항상 PK 오름차순으로 락을 획득하며, `SQLSTATE 40P01` (deadlock_detected) 또는 `40001` (serialization_failure) 발생 시 트랜잭션 헬퍼가 최대 3회 지터 백오프 재시도를 수행한다.

## 10. UX/UI 구조 (UX/UI Structure)
- N/A (데이터베이스 계층). 단, 커서 페이지네이션(`(published_at, id) < ($1, $2)`) 인덱스를 통해 수십만 페이지 탐색 시에도 UI 응답 속도가 일정하게 유지되도록 뒷받침한다.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- 최소 권한 DB 계정 분리:
  - `treeblog_migrator`: DDL 실행 권한 보유 (배포 시에만 단기 실행).
  - `treeblog_app`: 일반 DML(`SELECT`, `INSERT`, `UPDATE`, `DELETE`)만 보유하며 `DROP`/`TRUNCATE` 권한 없음. 또한 `audit_logs`와 `ledger_entries` 테이블에 대해서는 `UPDATE`/`DELETE` 권한을 회수하여 불변성을 DB 엔진 레벨에서 강제한다.

## 12. 성능 및 확장성 (Performance & Scalability)
- Keyset(Cursor) Pagination 복합 인덱스, JSONB GIN 인덱스, Partial Index(`WHERE status = 'PUBLISHED' AND trashed_at IS NULL`)를 결합하여 모든 OLTP 쿼리가 P95 < 10ms 이내에 실행되도록 설계.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- `github.com/jackc/pgx/v5`, `github.com/golang-migrate/migrate/v4`.

## 14. 개발 우선순위 (Development Priority)
- **Phase 0 ~ Phase 1 (최우선 기반 작업)** (`TB-DB-001` ~ `TB-DB-003`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 선행: `A_ARCHITECTURE`.
- 후행: 모든 백엔드 API 및 워커 모듈.

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-DB-001`: `0001`~`0010` 전체 Up 마이그레이션 실행 → 스키마 검증 → 전체 Down 롤백 → 재실행이 에러 0건으로 완료되는지 검증.
- `TEST-DB-002`: `ledger_entries`에 음수 금액 삽입 시도 또는 허용되지 않은 `posts.visibility` 문자열 삽입 시도 시 PostgreSQL `CHECK` 제약조건이 즉시 에러를 발생시키는지 검증.

## 17. Acceptance Criteria (인수 기준)
1. 제14조의 모든 필수 엔터티를 포함한 58개 테이블의 실제 실행 가능한 SQL DDL과 인덱스, 제약조건이 한 치의 문법 오류도 없이 준비되어야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- SQL 파서/구문 검증 스크립트 통과 로그, `03_DATABASE/migrations/` 전체 파일 및 체크섬.

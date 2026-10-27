# INDEX_STRATEGY.md — PostgreSQL 인덱스, 커서 페이지네이션 및 파티셔닝 전략

## 1. 인덱스 설계 5대 원칙
1. **OFFSET 페이지네이션 금지, Keyset(Cursor) 복합 인덱스 표준화**:
   - 블로그 글 목록이나 피드에서 `OFFSET 50000 LIMIT 20`을 실행하면 앞선 50,000건을 풀스캔 후 버리므로 DB CPU가 고갈된다.
   - 모든 목록 조회는 `(published_at DESC, id DESC)` 복합 인덱스와 `WHERE (published_at, id) < ($1, $2) ORDER BY published_at DESC, id DESC LIMIT $3` 커서 쿼리를 사용하여 1페이지와 100,000페이지의 조회 속도를 동일하게 O(log N)으로 유지한다.
2. **Partial Index (조건부 부분 인덱스) 적극 활용**:
   - 전체 행의 극히 일부만 폴링하는 작업 큐(`scheduled_publications`, `outbox_events`, `search_index_jobs`, `media_assets` 고아 파일)에는 `WHERE status = 'PENDING'` 조건부 인덱스를 생성하여 인덱스 크기를 99% 이상 축소하고 캐시 적중률(Buffer Cache Hit Ratio ≥ 99.5%)을 극대화한다.
3. **CITEXT 및 대소문자 무시 고유 인덱스**:
   - 이메일(`users.email`), 사용자명(`user_profiles.username`), 블로그 핸들(`blogs.handle`), 커스텀 도메인(`blog_domains.domain_name`), 태그(`tags.slug`)는 `CITEXT` 확장을 사용하여 `TreeBlog`와 `treeblog`의 중복 가입/개설을 DB 레벨에서 원천 차단한다.
4. **외래키(FK) 인덱스 누락 방지**:
   - 부모 레코드 삭제나 조인 시 테이블 락(Table Scan)이 발생하지 않도록 모든 주요 FK 컬럼(`blog_id`, `post_id`, `user_id`, `case_id`, `order_id`)에 B-Tree 인덱스를 구성한다.
5. **시계열 대용량 테이블 월별 Range Partitioning**:
   - `notifications`, `search_click_events`, `ad_events`, `audit_logs` 4개 테이블은 `PARTITION BY RANGE (created_at)`으로 구성하여 월 단위 파티션 프루닝(Partition Pruning)과 오래된 파티션의 무중단 `DETACH PARTITION`(콜드 스토리지 아카이빙)을 지원한다.

## 2. 핵심 쿼리 패턴별 인덱스 매핑표

| 도메인 쿼리 시나리오 | 대상 테이블 | 적용 인덱스 정의 | 예상 실행 계획 및 목표 지연 |
| :--- | :--- | :--- | :--- |
| 개별 블로그 공개 글 목록 조회 | `posts` | `idx_posts_blog_published ON posts(blog_id, published_at DESC, id DESC) WHERE status = 'PUBLISHED' AND trashed_at IS NULL` | Index Scan Forward, **< 1.5ms** |
| 플랫폼 전체 최신 공개 글 탐색 | `posts` | `idx_posts_public_discover ON posts(published_at DESC, id DESC) WHERE status = 'PUBLISHED' AND visibility = 'PUBLIC' AND trashed_at IS NULL` | Index Scan, **< 2.0ms** |
| 예약 발행 도래 글 5초 주기 폴링 | `scheduled_publications` | `idx_scheduled_pubs_pending ON scheduled_publications(scheduled_for) WHERE status = 'PENDING'` | Partial Index Scan + `SKIP LOCKED`, **< 0.5ms** |
| Transactional Outbox 이벤트 폴링 | `outbox_events` | `idx_outbox_pending ON outbox_events(next_attempt_at ASC) WHERE status = 'PENDING'` | Partial Index Scan, **< 0.5ms** |
| 미확인 알림 목록 및 배지 조회 | `notifications` | `idx_notifications_recipient_unread ON notifications(recipient_user_id, created_at DESC) WHERE is_read = FALSE` | Partitioned Partial Index Scan, **< 1.2ms** |
| 크리에이터 정산 가능 잔액 조회 | `ledger_entries` | `idx_ledger_owner_account ON ledger_entries(owner_user_id, account_type, created_at DESC)` | Index Only Scan / Aggregate, **< 3.0ms** |
| TreeScan 증분 색인 피드 조회 | `treescan_sync_state` | `idx_treescan_sync_cursor ON treescan_sync_state(sequence_cursor ASC) WHERE sync_status = 'PENDING'` | Partial Index Scan, **< 1.0ms** |

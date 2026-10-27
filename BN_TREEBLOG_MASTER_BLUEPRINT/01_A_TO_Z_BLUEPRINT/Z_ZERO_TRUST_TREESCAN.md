# Z_ZERO_TRUST_TREESCAN — Zero-Trust Verification & TreeScan Readiness (독립 검증 및 TreeScan 통합 준비)
> **부제**: 독립 플랫폼 검증, 보안 감사, 기능 완료 증빙, TreeScan API·SSO·검색 통합 사전 준비
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-Z_ZERO_TRUST_TREESCAN`
> **적용 조항**: 제3조 (Z_ZERO_TRUST_TREESCAN 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- TreeBlog가 **제1조(독립 개발 우선 및 미래 TreeScan 통합)**와 **제18조(Booking Network 생태계 연동 준비)**를 완벽히 충족하도록, 1) TreeScan 없이 100% 단독 운영됨을 증명하는 Zero-Trust 검증 기준과, 2) 향후 TreeScan 코드 수정 없이 즉시 연결 가능한 **9대 통합 인터페이스(Search API Adapter, Index Sync, OIDC SSO, Unified Profile, Discovery API, Blog Category, Cross-platform Linking, Sitemap/Crawling, Global Nav)**를 설계한다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **Independent Audit Agent**: 네트워크에서 외부 연동 엔드포인트를 모두 차단한 상태에서 TreeBlog 단독으로 전 기능이 작동하는지 검증한다.
- **TreeScan Integration Engineer (향후 통합 단계)**: PO 승인 후 `FF_TREESCAN_INTEGRATION_ENABLED=true` 플래그와 OIDC/Webhook 자격증명을 주입하여 코드 변경 없이 TreeBlog 공개 게시글을 TreeScan 검색 색인에 실시간 공급하고 통합 로그인을 활성화한다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-ZTR-001 (제1.1조 완전 독립 운영 검증)**:
  - `TREESCAN_ENABLED=false` (기본값) 상태에서 회원가입, 로그인, 블로그 개설, 글쓰기, 자체 OpenSearch 검색, 미디어 서빙, 커뮤니티, 광고/정산, 어드민이 외부 호출 에러나 타임아웃 없이 100% 동작해야 한다.
- **REQ-ZTR-002 (제1.2조 및 제18조 미래 TreeScan 9대 통합 어댑터 설계)**:
  1) **Search API Adapter (`internal/modules/treescan/search_adapter.go`)**: TreeScan 통합 검색이 TreeBlog의 블로그/게시글 결과를 조회하거나, 반대로 TreeBlog에서 TreeScan 웹 검색 결과를 참조할 수 있는 표준 REST 인터페이스.
  2) **Search Index Synchronization (`content.index.upsert.v1`, `content.index.delete.v1`)**: `outbox_events`를 통해 **오직 `status = 'PUBLISHED' AND visibility = 'PUBLIC' AND blog.status = 'ACTIVE'`인 공개 게시글만** TreeScan 색인 수집기로 푸시(Webhook) 또는 커서 기반 폴링(`GET /api/v1/treescan/sync/feed`)으로 전달한다.
  3) **SSO / OIDC Integration**: Booking Network / TreeScan 통합 인증 서버(IdP)와 표준 OIDC Authorization Code + PKCE로 연동하며, `identities` 테이블을 통해 기존 로컬 계정과 안전하게 병합(Account Linking)한다.
  4) **Unified User Profile**: 사용자의 명시적 동의(`user_consents`) 하에서만 프로필 닉네임/아바타를 동기화한다.
  5) **Content Discovery API**: Treewiki, App Market, getforTrip, TripScanner, OLMOA, Global Video 등 자매 플랫폼에서 주제별/장소별/제품별 우수 블로그 리뷰 글을 조회할 수 있는 큐레이션 API.
  6) **Blog Search Category & Structured Metadata**: 여행 일정/장소 좌표(getforTrip 연동용), 앱 메타데이터(App Market 연동용), 상품 리뷰 스키마(OLMOA 연동용), 위키 개념 참조 태그(Treewiki 연동용)를 게시글 구조화 블록 메타데이터로 포함 가능.
  7) **Cross-platform Content Linking**: 자매 서비스 링크 삽입 시 풍부한 공식 프리뷰 카드 렌더링.
  8) **Sitemap / Crawling Interface**: `TreeScanBot` 전용 고속 증분 Sitemap(`sitemap-treescan-delta.xml`) 및 IndexNow 스타일 즉시 색인 통지 지원.
  9) **Global Navigation Integration**: 설정 플러그인 방식으로 상단 Booking Network 패밀리 앱 스위처(App Launcher) 노출 여부 제어.
- **REQ-ZTR-003 (검색 중립성 및 프라이버시 절대 원칙 — 제9조, 제18조)**:
  - TreeScan 연동 페이로드에는 객관적인 문서 메타데이터와 품질 시그널만 전달하며, 타 웹 문서보다 무조건 상위에 고정 노출시키는 불공정 부스팅 플래그(`force_top_rank` 등)를 절대 설계하지 않는다.
  - 비공개(`PRIVATE`), 이웃공개(`NEIGHBOR_ONLY`), 상호이웃공개(`MUTUAL_ONLY`), 유료전용(`PAID_MEMBERS_ONLY` 본문), 삭제/제재된 글은 TreeScan 색인 피드에 **단 1바이트도 전송되지 않도록** 어댑터 게이트에서 이중 필터링한다.

## 4. 데이터 구조 (Data Structures & Domain Models)
```json
{
  "specversion": "1.0",
  "id": "evt-018f2a10-9911-7000-8000-000000099123",
  "source": "urn:bookingnetwork:treeblog",
  "type": "com.bookingnetwork.treeblog.post.upsert.v1",
  "time": "2026-10-27T14:30:00Z",
  "data": {
    "document_id": "treeblog:post:018f2a10-7b11-7000-8000-000000000101",
    "canonical_url": "https://tech-pioneer.treeblog.example.com/posts/go-opensearch-architecture",
    "title": "Go와 OpenSearch로 구축하는 글로벌 블로그 검색 아키텍처",
    "summary": "PostgreSQL Transactional Outbox와 OpenSearch를 결합한 실전 설계...",
    "content_plain": "본문 전체 평문 텍스트 (공개 게시글만 포함)...",
    "language": "ko",
    "author_name": "김아키",
    "blog_title": "Tech Pioneer Blog",
    "tags": ["golang", "opensearch"],
    "category_topic": "it-tech",
    "published_at": "2026-10-27T12:00:00Z",
    "updated_at": "2026-10-27T14:30:00Z",
    "quality_signals": {
      "word_count": 2450,
      "quality_score": 88.5,
      "is_original": true
    },
    "visibility_verified": "PUBLIC"
  }
}
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- 계정 설정 내 `Booking Network / TreeScan 통합 계정 연결` UI (Feature Flag 활성 시 노출) 및 앱 스위처 컴포넌트.

## 6. 백엔드 로직 (Backend Logic & Processing)
- `internal/modules/treescan`:
  - `TreeScanAdapter` 인터페이스 및 `NoopAdapter`(독립 운영 기본값) / `HTTPWebhookAdapter`(연동 활성 시) 구현체.
  - HMAC-SHA256 요청 서명(`X-TreeBlog-Signature`) 및 mTLS 클라이언트 인증서 지원.

## 7. API 계약 (API Contracts)
- `GET /api/v1/treescan/sync/feed?cursor={cursor}&limit=100` (TreeScan 색인 크롤러용 증분 변경 피드 — 공개 콘텐츠 및 삭제 툼스톤만 반환)
- `GET /api/v1/treescan/discovery/posts?topic={slug}&entity_ref={id}` (Booking Network 자매 플랫폼 연동용 공개 글 조회)
- `POST /api/v1/treescan/sso/link` (OIDC 토큰 검증 후 기존 TreeBlog 계정과 안전 매핑)
- `GET /api/v1/treescan/health` (연동 어댑터 상태 및 동기화 오프셋 확인)

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- 주요 테이블: `treescan_sync_state`, `identities` (`provider = 'TREESCAN_OIDC'`), `user_consents`, `outbox_events`.

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **TreeScan 연동 서버 다운타임 시 격리**: 향후 연동 활성화 상태에서 TreeScan 측 수신 서버가 응답하지 않더라도, TreeBlog 내부 트랜잭션에는 영향이 0%이며 `treescan_sync_state`의 커서 오프셋만 유지되어 TreeScan 복구 시 누락된 시점부터 자동 재동기화(Catch-up)된다.

## 10. UX/UI 구조 (UX/UI Structure)
- 플랫폼 간 프로필이나 활동 정보를 연동할 때는 반드시 명확한 권한 범위(Scope) 설명과 함께 사용자의 개별 체크 동의를 받는다.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- **비공개 게시글 유출 Zero-Tolerance**: `GET /api/v1/treescan/sync/feed` 쿼리는 DB 뷰 수준에서 `WHERE status = 'PUBLISHED' AND visibility = 'PUBLIC'` 조건을 강제하며, 공개였다가 비공개/삭제로 바뀐 글은 본문 없이 `{"event": "DELETE", "document_id": "..."}` 툼스톤(Tombstone)만 전송하여 TreeScan 색인에서도 즉시 삭제되도록 한다.

## 12. 성능 및 확장성 (Performance & Scalability)
- 커서 기반 증분 피드(`(updated_at, id)`)를 통해 일일 수백만 건의 게시글 변경사항도 부하 없이 스트리밍 동기화 가능.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- `github.com/coreos/go-oidc/v3/oidc`, 표준 `crypto/hmac`, `crypto/sha256`.

## 14. 개발 우선순위 (Development Priority)
- **Phase 10 (TreeScan 통합 준비 및 최종 감사)** (`TB-TREESCAN-001`, `TB-TREESCAN-002`, `TB-QA-003`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 선행: Phase 0 ~ Phase 9 전체 모듈.
- 주의: **Phase 10은 TreeBlog 측 어댑터 및 계약 검증까지만 수행하며, TreeScan 기존 저장소 코드는 절대 수정하지 않는다(제1조, 제22조).**

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-ZTR-001`: `TREESCAN_ENABLED=false` 및 외부 네트워크 차단 상태에서 전체 E2E 시나리오가 100% 성공하는지 독립성 검증.
- `TEST-ZTR-002`: `PUBLIC`, `NEIGHBOR_ONLY`, `MUTUAL_ONLY`, `PRIVATE` 게시글 4건을 생성한 뒤 `/api/v1/treescan/sync/feed` 조회 시 오직 `PUBLIC` 게시글 1건만 응답에 포함되고 나머지 3건은 절대 노출되지 않는지 검증.
- `TEST-ZTR-003`: `PUBLIC` 게시글을 `PRIVATE`으로 전환한 직후 `/api/v1/treescan/sync/feed` 조회 시 해당 문서 ID에 대해 `DELETE` 툼스톤 이벤트가 정확히 생성되는지 검증.

## 17. Acceptance Criteria (인수 기준)
1. TreeBlog 단독 운영 완전성이 증명되어야 하며, TreeScan 연동 어댑터가 비공개 데이터 유출 없이 공개 콘텐츠 색인 동기화 및 OIDC 계정 매핑 계약을 100% 충족해야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- `17_TREESCAN_INTEGRATION/*` 5개 상세 명세서, 독립 구동 테스트 및 Sync Feed 프라이버시 필터링 테스트 로그.

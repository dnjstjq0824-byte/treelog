# L_LIFECYCLE — Lifecycle Management (블로그 및 게시글 생명주기 관리)
> **부제**: 블로그 생성·이전·폐쇄, 게시글 작성·수정·예약·공개·비공개·삭제·복원 및 보관 정책
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-L_LIFECYCLE`
> **적용 조항**: 제3조 (L_LIFECYCLE 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- 블로그와 게시글이 생성되어 수정, 예약, 공개, 비공개 전환, 휴지통 삭제, 복원, 영구 파기에 이르는 전 과정의 상태 머신(State Machine)과 데이터 보존 정책(Retention Policy)을 빈틈없이 정의한다.
- 실수로 인한 데이터 유실 사고를 방지하면서도 개인정보보호법 및 삭제권(Right to Erasure) 요구사항을 충족한다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **블로거**: 실수로 삭제한 게시글을 30일 이내에 '휴지통(Trash)'에서 원클릭으로 복원하거나, 블로그 전체 백업 ZIP을 다운로드한 뒤 블로그 폐쇄를 신청한다.
- **시스템 배치 워커**: 휴지통 보관 기간(30일)이 만료된 게시글과 연결된 고아 미디어 파일을 스토리지에서 영구 파기(Hard Purge)한다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-LFC-001 (블로그 생명주기 6단계)**:
  - `ACTIVE` (정상 운영) → `READ_ONLY` (점검/이전 중 쓰기 잠금) → `SUSPENDED` (운영정책 위반 제재) → `CLOSING` (소유자 폐쇄 신청 후 **30일 유예 기간** — 이 기간 동안 외부 접근은 차단되나 소유자는 언제든 폐쇄 취소 및 데이터 백업 가능) → `CLOSED` (논리적 폐쇄 완료) → `PURGED` (보관 만료 후 영구 삭제).
- **REQ-LFC-002 (게시글 생명주기 8단계)**:
  - `DRAFT` (초안) ↔ `IN_REVIEW` (팀 블로그 검토 대기) ↔ `SCHEDULED` (예약 발행 대기) → `PUBLISHED` (발행됨: 세부 공개 범위는 `visibility` 컬럼으로 제어) ↔ `UNPUBLISHED` (발행 취소/비공개 보관) ↔ `QUARANTINED` (신고/법적 임시조치에 의한 게시 중단) ↔ `TRASHED` (작성자 삭제 — 휴지통 **30일 보관**) → `PURGED` (영구 파기).
- **REQ-LFC-003 (상태 변경 시 부수 효과 동기화)**:
  - `PUBLISHED (PUBLIC)` 상태에서 `UNPUBLISHED`, `TRASHED`, `QUARANTINED` 또는 `visibility != 'PUBLIC'`으로 변경되는 즉시:
    1) OpenSearch 공개 색인에서 문서를 즉시 삭제(`DELETE_POST`)한다.
    2) CDN Edge 캐시 및 Redis SSR 캐시를 즉시 무효화(Purge)한다.
    3) Sitemap 및 RSS 피드에서 즉시 제외한다.
    4) 외부 크롤러가 해당 URL 재방문 시 상태에 따라 `404 Not Found`(휴지통/비공개) 또는 `410 Gone`(영구 삭제) 또는 `451 Unavailable For Legal Reasons`(법적 차단)를 명확히 반환한다.

## 4. 데이터 구조 (Data Structures & Domain Models)
```go
type PostLifecycleTransition struct {
    PostID       uuid.UUID `json:"post_id"`
    FromStatus   string    `json:"from_status"`
    ToStatus     string    `json:"to_status"`
    FromVis      string    `json:"from_visibility"`
    ToVis        string    `json:"to_visibility"`
    ActorUserID  uuid.UUID `json:"actor_user_id"`
    Reason       string    `json:"reason,omitempty"`
    TrashedAt    *time.Time `json:"trashed_at,omitempty"`
    PurgeAfterAt *time.Time `json:"purge_after_at,omitempty"` // TrashedAt + 30 days
}
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- 블로그 관리 센터 내 게시글 상태 탭(`전체`, `발행됨`, `임시저장/검토중`, `예약됨`, `비공개`, `휴지통`) 및 일괄 상태 변경/복원/영구삭제 UI.
- 블로그 소유권 이전(Transfer Ownership) 및 블로그 폐쇄 신청(2FA 및 핸들 직접 타이핑 확인) 모달.

## 6. 백엔드 로직 (Backend Logic & Processing)
- `internal/modules/post/lifecycle.go` & `internal/modules/blog/lifecycle.go`:
  - 허용되지 않는 상태 전이(예: `QUARANTINED` 상태의 글을 작성자가 임의로 `PUBLISHED`로 변경 시도)를 차단하는 유한 상태 기계(FSM) 가드.
  - 일일 배치 워커(`PurgeExpiredTrashJob`): `purge_after_at <= NOW()`인 게시글과 블로그를 찾아 DB 레코드 삭제 및 S3 미디어 오브젝트 삭제 수행.

## 7. API 계약 (API Contracts)
- `POST /api/v1/blogs/{blogId}/posts/{postId}/unpublish` (발행 취소/비공개 전환)
- `DELETE /api/v1/blogs/{blogId}/posts/{postId}` (휴지통 이동 — Soft Delete)
- `POST /api/v1/blogs/{blogId}/posts/{postId}/restore` (휴지통에서 복원)
- `DELETE /api/v1/blogs/{blogId}/posts/{postId}/permanent` (휴지통에서 즉시 영구 삭제)
- `POST /api/v1/blogs/{blogId}/transfer-ownership` (블로그 소유권 이전 요청)
- `POST /api/v1/blogs/{blogId}/close` / `POST /api/v1/blogs/{blogId}/cancel-close` (블로그 폐쇄 신청 및 취소)

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- `posts` 테이블의 `status`, `visibility`, `published_at`, `trashed_at`, `purge_after_at` 컬럼 및 `blogs` 테이블의 `status`, `closing_requested_at` 컬럼.
- 법적 분쟁이나 결제 정산과 관련된 원장(`ledger_entries`), 주문(`orders`), 감사 로그(`audit_logs`)는 전자상거래법 등 법정 보존 기간(5년) 동안 익명화하여 분리 보존(`RESTRICT` 정책).

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **URL 슬러그 변경 시 리다이렉트 보존**: 이미 공개 발행된 게시글의 `slug`를 수정할 경우, 기존 슬러그를 `post_slug_history`에 기록하여 기존 링크로 접속한 방문자와 검색엔진을 새 슬러그로 `301 Moved Permanently` 리다이렉트한다.

## 10. UX/UI 구조 (UX/UI Structure)
- 휴지통에 있는 게시글 옆에는 `남은 복원 가능 기간: D-28` 배지를 표시하여 영구 삭제 시점을 명확히 인지시킨다.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- 블로그 폐쇄 및 소유권 이전은 탈취된 세션 악용을 막기 위해 반드시 **비밀번호 재확인 및 MFA 인증**을 통과해야만 실행된다.

## 12. 성능 및 확장성 (Performance & Scalability)
- 일반 목록 조회 쿼리는 부분 인덱스(`WHERE status = 'PUBLISHED' AND trashed_at IS NULL`)를 타도록 설계하여 휴지통/초안 데이터가 조회 성능을 저하시키지 않도록 한다.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- Go 표준 패키지 및 `pgx/v5` 트랜잭션.

## 14. 개발 우선순위 (Development Priority)
- **Phase 3 (게시글 생명주기)** 및 **Phase 7 (제재/보관 배치)** (`TB-POST-001`, `TB-POST-003`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 선행: `B_BLOGGING_EXPERIENCE`, `C_CONTENT_EDITOR`.
- 후행: `O_OPENSEARCH`, `V_VIDEO_MEDIA`, `W_WORKFLOW_ADMIN`.

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-LFC-001`: 공개 게시글을 휴지통(`TRASHED`)으로 이동한 직후 공개 URL 조회 시 `404`가 반환되고 OpenSearch 검색 결과에서 즉시 사라지는지 검증.
- `TEST-LFC-002`: 휴지통에서 복원(`restore`) 실행 시 기존 슬러그, 댓글, 공감 수가 손실 없이 원상 복구되는지 검증.

## 17. Acceptance Criteria (인수 기준)
1. 게시글과 블로그의 모든 상태 전이 규칙이 FSM 명세대로 동작하고, 삭제/비공개 시 캐시와 검색 색인이 즉시 제거되어야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- 생명주기 상태 전이 통합 테스트 로그, 슬러그 변경 후 `301` 리다이렉트 응답 캡처.

# E_ENGAGEMENT_COMMUNITY — Engagement & Community (이웃·구독·커뮤니티 엔진)
> **부제**: 이웃, 구독, 팔로우, 댓글, 답글, 공감, 좋아요, 북마크, 알림, 차단, 신고, 커뮤니티 관계
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-E_ENGAGEMENT_COMMUNITY`
> **적용 조항**: 제3조 (E_ENGAGEMENT_COMMUNITY 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- 네이버 블로그의 높은 사용자 고착도(Retention)를 만든 핵심 동력인 **이웃·상호이웃(서로이웃)·공감·댓글·방명록** 문화와 글로벌 플랫폼의 **팔로우·이메일 뉴스레터 구독·북마크** 모델을 하나의 정교한 소셜 관계 엔진으로 통합한다.
- 관계 상태에 따른 게시글 열람 권한(`PUBLIC`, `NEIGHBOR_ONLY`, `MUTUAL_ONLY`, `PRIVATE`)을 엄격히 제어하여 사생활 보호와 친밀한 커뮤니티 소통을 동시에 달성한다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **독자 / 블로거**: 관심 있는 블로그를 가볍게 '팔로우'하거나 '이웃 추가'하고, 더 깊은 교류를 원할 경우 신청 메시지와 함께 '상호이웃(서로이웃) 신청'을 보낸다.
- **블로그 소유자**: 수신된 상호이웃 신청을 확인하여 수락/거절하거나, 스팸 계정을 차단(`user_blocks`)하고 댓글 승인제/이웃 전용 댓글 허용 정책을 설정한다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-COM-001 (팔로우·이웃·상호이웃·구독의 명확한 정의 및 권한 매트릭스 — 제8조 핵심)**:
  1) **팔로우 (`FOLLOW`)**: 사용자 대 블로그(또는 사용자)의 가벼운 단방향 관심 등록. 승인 불필요. 추천 피드 가중치 부여. `PUBLIC` 게시글만 열람 가능.
  2) **이웃 (`NEIGHBOR`)**: 특정 블로그를 내 '이웃 그룹(폴더)'에 분류하여 모아보는 단방향 관계. 승인 불필요. 블로그 소유자가 `NEIGHBOR_ONLY`(이웃 공개)로 설정한 게시글까지 열람 가능.
  3) **상호이웃 (`MUTUAL_NEIGHBOR`, 서로이웃)**: 신청자(`requester`)가 신청 메시지를 보내고 대상 블로그 소유자(`target`)가 수락(`ACCEPTED`)해야 성립하는 양방향 친밀 관계. **`MUTUAL_ONLY`(상호이웃 공개) 게시글 및 상호이웃 전용 댓글/방명록에 접근할 수 있는 유일한 관계**.
  4) **구독 (`SUBSCRIPTION`)**: 새 글 발행 시 알림 수신 채널(인앱 알림, 웹 푸시, 이메일 뉴스레터) 및 유료 멤버십(`FREE` / `PAID`)을 관리하는 전달 계약.
- **REQ-COM-002 (공감 및 좋아요 시스템)**:
  - 게시글에는 6종 공감 리액션(`LIKE` 좋아요, `LOVE` 최고예요, `HELPFUL` 유익해요, `FUNNY` 재밌어요, `TOUCHING` 감동이에요, `INSIGHTFUL` 깊이있어요)을 지원하며, 댓글에는 `LIKE`를 지원한다.
  - 블로거 설정에 따라 공감한 사용자 목록 공개/비공개를 선택할 수 있다.
- **REQ-COM-003 (계층형 댓글·대댓글 및 모더레이션)**:
  - 최대 2뎁스(댓글 + 대댓글) 구조, 비밀 댓글(`is_secret`), 작성자 고정 댓글(`is_pinned`), 멘션(`@username`), 스티커/이모지 지원.
  - 게시글별/블로그별 댓글 정책: `ALL_ALLOWED`(전체 허용), `NEIGHBORS_ONLY`(이웃 이상만), `MUTUAL_ONLY`(상호이웃만), `REQUIRE_APPROVAL`(작성자 승인 후 노출), `DISABLED`(댓글 비허용).
- **REQ-COM-004 (북마크 및 컬렉션)**: 사용자는 게시글을 개인 북마크 폴더별로 저장하고 읽은 위치를 기록할 수 있다.
- **REQ-COM-005 (차단 및 스팸 방어)**: 특정 사용자를 차단(`user_blocks`)하면 해당 사용자는 내 모든 블로그에 댓글/방명록/공감/상호이웃 신청을 할 수 없으며, 내 이웃/상호이웃 관계에서 즉시 강제 해제된다.

## 4. 데이터 구조 (Data Structures & Domain Models)
```go
// 관계 상태 및 게시글 가시성 평가 규칙
// Visibility: PUBLIC | UNLISTED | NEIGHBOR_ONLY | MUTUAL_ONLY | PAID_MEMBERS_ONLY | PRIVATE
func CanReadPost(viewerID *uuid.UUID, post Post, blog Blog, rel RelationState, hasPaidAccess bool) bool {
    if viewerID != nil && (*viewerID == blog.OwnerUserID || blog.HasEditorRole(*viewerID)) {
        return true
    }
    if rel.IsBlockedByBlogOwner || blog.Status != "ACTIVE" || post.Status != "PUBLISHED" {
        return false
    }
    switch post.Visibility {
    case "PUBLIC", "UNLISTED":
        return true
    case "NEIGHBOR_ONLY":
        return rel.Type == "NEIGHBOR" || rel.Type == "MUTUAL_NEIGHBOR"
    case "MUTUAL_ONLY":
        return rel.Type == "MUTUAL_NEIGHBOR" && rel.Status == "ACCEPTED"
    case "PAID_MEMBERS_ONLY":
        return hasPaidAccess
    case "PRIVATE":
        return false
    default:
        return false
    }
}
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- `apps/user-web`: 게시글 하단 공감 리액션 바, 실시간 댓글/대댓글 컴포넌트(`SCR-MEM-008`), 이웃/상호이웃 신청 모달, 북마크 토글 버튼.
- `apps/blog-console`: 이웃 그룹 관리, 수신된 상호이웃 신청 일괄 수락/거절함(`SCR-MEM-012`), 댓글 통합 관리(승인 대기/스팸함/휴지통), 차단 사용자 관리(`SCR-MEM-014`).

## 6. 백엔드 로직 (Backend Logic & Processing)
- `internal/modules/community`:
  - 상호이웃 수락 시 트랜잭션으로 양방향 `follows` 레코드를 `MUTUAL_NEIGHBOR` + `ACCEPTED`로 갱신하고 신청자에게 알림 이벤트 발행.
  - 공감/댓글 증감 시 PostgreSQL 원자적 카운터 업데이트(`reaction_count = reaction_count + 1`) 및 Redis 디바운싱 적용.

## 7. API 계약 (API Contracts)
- `POST /api/v1/blogs/{blogId}/follow` / `DELETE /api/v1/blogs/{blogId}/follow` (팔로우/이웃/상호이웃 신청 및 해제)
- `GET /api/v1/me/mutual-requests` / `POST /api/v1/me/mutual-requests/{requestId}/respond` (상호이웃 신청 수락/거절)
- `POST /api/v1/blogs/{blogId}/subscribe` / `PATCH /api/v1/blogs/{blogId}/subscribe` (알림/뉴스레터 구독 설정)
- `GET /api/v1/posts/{postId}/comments` / `POST /api/v1/posts/{postId}/comments` / `PATCH /api/v1/comments/{commentId}` / `DELETE /api/v1/comments/{commentId}` (댓글 CRUD)
- `POST /api/v1/posts/{postId}/reactions` (6종 공감 리액션 토글)
- `POST /api/v1/posts/{postId}/bookmarks` / `GET /api/v1/me/bookmarks` (북마크 관리)
- `POST /api/v1/users/{userId}/block` / `DELETE /api/v1/users/{userId}/block` (사용자 차단/해제)

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- 주요 테이블: `follows`, `neighbor_groups`, `subscriptions`, `reactions`, `comments`, `bookmarks`, `guestbook_entries`, `user_blocks`.
- 제약조건: `reactions` 테이블에 `UNIQUE (target_type, target_id, user_id)` 제약으로 중복 공감 원천 차단.

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **상호이웃 상태 전이**: `NONE` → `PENDING` (신청) → `ACCEPTED` (상호이웃 성립) 또는 `REJECTED` (거절). 어느 한쪽이 상호이웃을 끊거나 차단하면 양쪽 모두 즉시 `TERMINATED` 상태로 전이되어 `MUTUAL_ONLY` 게시글 접근 권한이 박탈된다.
- **대댓글 있는 댓글 삭제**: 하위 대댓글이 존재하는 상위 댓글 삭제 시 물리 삭제하지 않고 `status = 'DELETED_BY_USER'`로 마스킹(`삭제된 댓글입니다`)하여 대댓글 문맥을 보존한다.

## 10. UX/UI 구조 (UX/UI Structure)
- 비밀 댓글은 자물쇠 아이콘과 함께 작성자·원댓글 작성자·블로그 주인에게만 본문이 보이며, 제3자에게는 `비밀 댓글입니다.` 문구만 노출된다.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- **비공개/상호이웃 게시글 및 비밀 댓글 IDOR 방지**: 댓글 조회 API에서도 반드시 상위 게시글의 열람 권한(`CanReadPost`)과 개별 비밀 댓글 열람 권한을 서버 측에서 먼저 검증한다. 본문을 빈 문자열로만 가리고 JSON 필드에 원문을 내려보내는 실수를 엄격히 금지한다.

## 12. 성능 및 확장성 (Performance & Scalability)
- 내 이웃/구독 블로그의 새 글 피드(`SCR-EXP-010`)는 Fan-out-on-Read(일반 블로거)와 Redis ZSET 캐시(인기 블로거) 하이브리드 방식으로 P95 < 100ms 이내에 조회한다.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- Backend: 표준 `pgx/v5` 트랜잭션 및 Redis Pipelining.
- Frontend: Optimistic UI 업데이트를 위한 `@tanstack/vue-query` 또는 Pinia 액션 스토어.

## 14. 개발 우선순위 (Development Priority)
- **Phase 6 (커뮤니티 및 소셜 엔진)**: 블로그/에디터/검색 구축 후 구현 (`TB-COMMUNITY-001` ~ `TB-COMMUNITY-003`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 선행: `I_IDENTITY_AUTHORIZATION`, `B_BLOGGING_EXPERIENCE`, `C_CONTENT_EDITOR`.
- 후행: `N_NOTIFICATIONS`, `R_RECOMMENDATIONS`.

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-COM-001`: 일반 이웃(`NEIGHBOR`) 또는 비로그인 사용자가 `MUTUAL_ONLY` 게시글 URL에 직접 접근 시 `403 FORBIDDEN`이 반환되고 본문이 전혀 노출되지 않는지 검증.
- `TEST-COM-002`: 상호이웃 수락 즉시 `MUTUAL_ONLY` 게시글이 정상 열람되고, 이후 차단(`block`) 실행 시 즉시 열람 권한이 회수되는지 검증.
- `TEST-COM-003`: 동일 사용자가 동일 게시글에 동시 10회 공감 요청(Race Condition)을 보낼 때 DB UNIQUE 제약으로 정확히 1건만 반영되는지 검증.

## 17. Acceptance Criteria (인수 기준)
1. 팔로우, 이웃, 상호이웃, 구독의 상태 전이와 게시글 가시성 제어 규칙이 100% 정확하게 동작해야 한다.
2. 댓글, 대댓글, 비밀댓글, 승인제 댓글, 공감, 북마크, 사용자 차단이 UI와 DB에서 모두 일관되게 동작해야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- 권한별(`ANON`, `NEIGHBOR`, `MUTUAL`, `BLOCKED`) 게시글/비밀댓글 접근 제어 통합 테스트 로그 및 E2E 실행 결과.

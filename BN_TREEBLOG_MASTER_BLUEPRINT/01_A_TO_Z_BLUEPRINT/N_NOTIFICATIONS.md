# N_NOTIFICATIONS — Notification Engine (멀티채널 알림 및 중복 방지 엔진)
> **부제**: 댓글, 답글, 구독, 공감, 운영 공지, 이메일, 푸시, 웹 알림, 알림 설정 및 중복 방지
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-N_NOTIFICATIONS`
> **적용 조항**: 제3조 (N_NOTIFICATIONS 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- 댓글, 대댓글, 멘션, 공감, 이웃/상호이웃 신청 및 수락, 구독 블로그의 새 글 발행, 정산/보안/운영 공지 이벤트를 **인앱 웹 알림(In-App SSE/Web)**, **브라우저/모바일 웹 푸시(W3C Web Push)**, **이메일(Email)** 3대 채널로 적시에 전달한다.
- 단시간 내 집중되는 리액션(예: 인기글 공감 500건)으로 인한 알림 폭탄을 방지하기 위해 **스마트 병합(Rollup Aggregation) 및 중복 제거(Deduplication)** 엔진을 제공한다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **블로거**: 자신의 글에 댓글이 달리면 실시간 토스트/배지 알림을 받고, 공감 알림은 `홍길동님 외 24명이 회원님의 게시글에 공감했습니다` 형태로 깔끔하게 묶여서 수신한다.
- **독자**: 구독 중인 블로그마다 새 글 알림을 이메일 뉴스레터로 받을지, 웹 푸시로만 받을지 세밀하게 제어하고 야간 방해금지 시간대(Quiet Hours)를 설정한다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-NOT-001 (알림 이벤트 유형 및 채널 라우팅)**:
  - `COMMENT_CREATED`, `COMMENT_REPLY`, `MENTION`, `REACTION_ROLLUP`, `FOLLOW_NEW`, `MUTUAL_REQUEST_RECEIVED`, `MUTUAL_REQUEST_ACCEPTED`, `SUBSCRIBED_POST_PUBLISHED`, `MODERATION_NOTICE`, `PAYOUT_STATUS`, `SECURITY_ALERT`, `SYSTEM_ANNOUNCEMENT`.
  - 보안 알림(`SECURITY_ALERT`) 및 법적/제재 통지(`MODERATION_NOTICE`)는 마케팅 수신 거부 여부와 무관하게 필수 발송(Transactional Mandatory)한다.
- **REQ-NOT-002 (중복 방지 및 Rollup 병합)**:
  - 고유 `dedup_key` (예: `reaction:{postId}:{actorId}`)를 통해 동일 사용자가 공감 버튼을 눌렀다 취소했다 다시 누르는 행위로 인한 반복 알림 발송을 차단한다(5분 쿨다운 윈도우).
  - 동일 대상(`target_id`)에 대한 `REACTION` 이벤트는 `group_key = 'react_post_' || post_id`로 묶어 읽지 않은 기존 알림 레코드의 `actor_count`와 `actors_preview` JSON을 갱신한다.
- **REQ-NOT-003 (수신자별 채널 설정 및 방해금지 시간)**:
  - `notification_preferences` 테이블에서 카테고리별(`comments`, `reactions`, `neighbors`, `subscriptions`, `marketing`) × 채널별(`in_app`, `web_push`, `email`) ON/OFF 및 사용자 현지 타임존 기준 방해금지 시작/종료 시각(`quiet_hours_start`, `quiet_hours_end`)을 지원한다.

## 4. 데이터 구조 (Data Structures & Domain Models)
```go
type Notification struct {
    ID            uuid.UUID       `json:"id"`
    RecipientID   uuid.UUID       `json:"recipient_id"`
    Category      string          `json:"category"`       // COMMENT, REACTION, NEIGHBOR, SUBSCRIPTION, SYSTEM, SECURITY
    GroupKey      string          `json:"group_key"`      // Rollup 병합 키
    DedupKey      string          `json:"dedup_key"`      // 중복 발송 방지 키
    ActorCount    int             `json:"actor_count"`
    ActorsPreview json.RawMessage `json:"actors_preview"` // 최근 액터 최대 3명 프로필 요약
    Title         string          `json:"title"`
    Body          string          `json:"body"`
    ActionURL     string          `json:"action_url"`
    IsRead        bool            `json:"is_read"`
    ReadAt        *time.Time      `json:"read_at,omitempty"`
    CreatedAt     time.Time       `json:"created_at"`
}
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- 글로벌 헤더 종 모양 아이콘(미확인 카운트 배지, SSE 실시간 스트림 연동) + 드롭다운 프리뷰 + 전체 알림 센터 페이지(`SCR-MEM-013`).
- Service Worker(`sw.js`) 기반 VAPID Web Push 구독 등록 및 클릭 시 해당 게시글/댓글 앵커로 즉시 이동.

## 6. 백엔드 로직 (Backend Logic & Processing)
- `internal/modules/notification`:
  - Outbox 이벤트를 소비하여 수신자의 차단 목록(`user_blocks`), 블로그 접근 권한, 수신 설정(`notification_preferences`)을 검사한 뒤 인앱 DB 저장, Redis Pub/Sub(SSE 푸시), Web Push 발송, SES 이메일 발송 큐로 분기한다.
  - 구독자가 10만 명 이상인 대형 블로거의 새 글 발행 시, 청크 단위(1,000명씩) 배치 팬아웃 잡으로 분할하여 워커 메모리를 보호한다.

## 7. API 계약 (API Contracts)
- `GET /api/v1/me/notifications` (카테고리/읽음여부 필터 및 커서 페이지네이션)
- `GET /api/v1/me/notifications/unread-count` / `GET /api/v1/me/notifications/stream` (SSE 실시간 스트림)
- `POST /api/v1/me/notifications/{notificationId}/read` / `POST /api/v1/me/notifications/read-all`
- `GET /api/v1/me/notification-preferences` / `PUT /api/v1/me/notification-preferences`
- `POST /api/v1/me/push-subscriptions` (VAPID Web Push 엔드포인트 등록)

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- 주요 테이블: `notifications` (생성일 기준 월별 파티셔닝 및 180일 보존 정책), `notification_preferences`, `push_subscriptions`.

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **Web Push 만료(`410 Gone`) 처리**: 푸시 서버(FCM/Mozilla/Apple)로부터 `410 Gone` 또는 `404` 응답 수신 시 해당 `push_subscriptions` 레코드를 즉시 비활성화/삭제하여 불필요한 재시도를 방지한다.

## 10. UX/UI 구조 (UX/UI Structure)
- 알림 목록에서 '모두 읽음 표시', '읽은 알림 삭제', 카테고리 탭(`전체`, `댓글/멘션`, `공감`, `이웃/구독`, `시스템/정산`) 필터를 제공한다.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- 모든 마케팅/구독 이메일 하단에는 RFC 8058 원클릭 수신거부 헤더(`List-Unsubscribe`, `List-Unsubscribe-Post`)와 서명된 수신거부 링크(로그인 없이 즉시 처리)를 포함한다.
- 차단한 사용자(`user_blocks`)가 발생시킨 이벤트는 알림 생성 단계에서 조용히 드롭(Silent Drop)한다.

## 12. 성능 및 확장성 (Performance & Scalability)
- 미확인 알림 카운트는 Redis `notif:unread:{userId}` 카운터로 캐싱하여 매 페이지 로드 시 DB `COUNT(*)` 쿼리를 생략한다.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- `github.com/SherClockHolmes/webpush-go` (VAPID Web Push), `github.com/aws/aws-sdk-go-v2/service/sesv2` 또는 표준 `net/smtp`.

## 14. 개발 우선순위 (Development Priority)
- **Phase 6 (커뮤니티 및 알림 연동)** (`TB-COMMUNITY-003`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 선행: `I_IDENTITY_AUTHORIZATION`, `E_ENGAGEMENT_COMMUNITY`, `J_JOBS_SCHEDULING`.
- 후행: 없음.

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-NOT-001`: 동일 게시글에 서로 다른 사용자 10명이 연속으로 공감 클릭 시 수신자에게 10개의 개별 알림이 아닌 `actor_count = 10`으로 병합된 단일 롤업 알림 1건만 유지되는지 검증.
- `TEST-NOT-002`: 사용자가 공감 클릭 후 1초 뒤 취소하고 다시 클릭할 때 중복 푸시/이메일이 발송되지 않는지 검증.

## 17. Acceptance Criteria (인수 기준)
1. 댓글, 답글, 상호이웃, 새 글 발행, 시스템 공지 알림이 인앱/웹푸시/이메일 설정에 따라 정확히 발송되고 중복 방지 및 일괄 읽음 처리가 동작해야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- 알림 병합(Rollup) 통합 테스트 로그, Mailpit 이메일 수신 캡처, SSE 실시간 이벤트 수신 로그.

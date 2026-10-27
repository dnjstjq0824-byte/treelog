# J_JOBS_SCHEDULING — Jobs & Scheduling (백그라운드 작업 및 스케줄링 엔진)
> **부제**: 예약 발행, 자동저장 복구, 색인 갱신, 미디어 처리, 알림 발송, 배치 작업, 실패 재시도 및 작업 큐
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-J_JOBS_SCHEDULING`
> **적용 조항**: 제3조 (J_JOBS_SCHEDULING 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- 사용자의 동기 API 응답 속도를 저해하지 않으면서도 **분 단위 정확한 예약 발행**, **실시간 검색 색인 동기화**, **고해상도 미디어 변환**, **대량 구독 알림 발송**, **정산 원장 집계**를 신뢰성 있게 처리하는 분산 작업 큐 및 스케줄링 아키텍처를 제공한다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **블로거**: 게시글을 내일 오전 8시 00분에 예약 발행(`SCHEDULED`)으로 걸어두면 정확히 해당 시각(오차 < 5초 이내)에 공개 전환되고 구독자 알림과 검색 색인이 자동으로 실행된다.
- **플랫폼 운영자**: 관리자 콘솔에서 실패한 백그라운드 작업(Dead Letter Queue)의 에러 원인을 확인하고 수동 재실행(Retry)을 트리거한다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-JOB-001 (우선순위 분리 멀티 큐)**:
  - `critical` (가중치 8): 예약 발행 트리거, 비공개/삭제 게시글 색인 즉시 제거, 보안/계정 알림.
  - `default` (가중치 4): 신규/수정 게시글 OpenSearch 색인, 이미지 WebP/AVIF 변환, 일반 댓글/이웃 알림.
  - `low` / `bulk` (가중치 1): 외부 블로그 가져오기(Import), 콘텐츠 전체 내보내기(Export ZIP), 일일 통계 집계, 고아 미디어 파일 정리.
- **REQ-JOB-002 (정밀 예약 발행 스케줄러)**:
  - `scheduled_publications` 테이블(`status = 'PENDING' AND scheduled_for <= NOW()`)을 매 5초마다 `FOR UPDATE SKIP LOCKED`로 폴링하여 다중 워커 환경에서도 중복 발행 없이 정확히 1회만 발행 트랜잭션을 실행한다.
- **REQ-JOB-003 (멱등성 및 지수 백오프 재시도)**:
  - 모든 잡 핸들러는 `job_id` 또는 `idempotency_key`를 기반으로 멱등성(Idempotent) 있게 동작해야 한다.
  - 일시적 네트워크/외부 API 오류 시 `retry_delay = base * (2 ^ attempt) + jitter` 공식으로 최대 8회 재시도 후 실패 시 `background_jobs.status = 'DEAD_LETTER'`로 기록한다.

## 4. 데이터 구조 (Data Structures & Domain Models)
```go
type ScheduledPublicationJob struct {
    ID           uuid.UUID `json:"id"`
    PostID       uuid.UUID `json:"post_id"`
    BlogID       uuid.UUID `json:"blog_id"`
    ScheduledFor time.Time `json:"scheduled_for"` // UTC
    Status       string    `json:"status"`        // PENDING, EXECUTING, COMPLETED, CANCELED, FAILED
}
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- 블로거용: 에디터 내 예약 발행 일시 선택기 및 '예약 대기 글 목록' 관리 UI. Publish/Export/Import 진행률 폴링 또는 SSE 상태 표시기.
- 관리자용: `SCR-ADM-008` 내 백그라운드 잡 큐 상태 및 Dead Letter 재처리 UI.

## 6. 백엔드 로직 (Backend Logic & Processing)
- `internal/worker`:
  - Outbox Relay Worker, Scheduled Post Poller, Media Processor Worker, Search Indexer Worker, Notification Worker, Settlement Batch Worker, Orphan Media GC Worker.

## 7. API 계약 (API Contracts)
- `GET /api/v1/blogs/{blogId}/scheduled-posts` (예약 발행 대기 목록)
- `DELETE /api/v1/blogs/{blogId}/scheduled-posts/{postId}` (예약 취소 후 초안 `DRAFT` 전환)
- `GET /api/v1/admin/jobs` / `POST /api/v1/admin/jobs/{jobId}/retry` (관리자 잡 모니터링 및 재시도)

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- 주요 테이블: `scheduled_publications`, `outbox_events`, `background_jobs`, `search_index_jobs`, `media_processing_jobs`, `migration_jobs`.

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **예약 시점에 블로그가 정지(`SUSPENDED`)된 경우**: 예약 발행 워커가 실행 시점에 블로그 상태를 재검사하여 `SUSPENDED`이면 발행하지 않고 `FAILED_BLOG_SUSPENDED`로 기록하며 작성자에게 알림을 보낸다.
- **워커 프로세스 비정상 종료(Crash) 예외**: `EXECUTING` 상태로 5분 이상 하트비트가 갱신되지 않은 잡(Stale Lock)은 Reaper 워커가 자동으로 복구하여 재큐잉한다.

## 10. UX/UI 구조 (UX/UI Structure)
- 예약된 게시글 상세 편집 화면 상단에 `이 글은 2026-10-28 09:00 (Asia/Seoul)에 자동 발행될 예정입니다 [일정 변경] [예약 취소]` 배너를 명확히 표시한다.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- 잡 페이로드에는 비밀번호나 전체 세션 토큰을 절대 저장하지 않으며, 필요한 최소 식별자(`post_id`, `user_id`)만 전달하여 워커가 실행 시점의 최신 DB 권한을 다시 조회하도록 한다.

## 12. 성능 및 확장성 (Performance & Scalability)
- `FOR UPDATE SKIP LOCKED` 쿼리와 부분 인덱스(`CREATE INDEX idx_sched_pub_pending ON scheduled_publications (scheduled_for) WHERE status = 'PENDING'`)를 통해 수백만 건의 이력이 쌓여도 폴링 쿼리 실행 시간을 1ms 미만으로 유지한다.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- `github.com/hibiken/asynq` (Redis 기반 분산 작업 큐), `github.com/robfig/cron/v3`.

## 14. 개발 우선순위 (Development Priority)
- **Phase 3 (예약 발행/미디어 큐)** 및 **Phase 5 (색인 큐)** (`TB-POST-002`, `TB-OPS-002`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 선행: `A_ARCHITECTURE`, `P_POSTGRESQL`.
- 후행: `C_CONTENT_EDITOR`, `O_OPENSEARCH`, `V_VIDEO_MEDIA`, `N_NOTIFICATIONS`.

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-JOB-001`: 3대의 워커 인스턴스가 동시에 동일 시각 예약 게시글 50건을 폴링할 때 중복 발행이 0건이고 정확히 50건 모두 `PUBLISHED`로 전환되는지 동시성 테스트.
- `TEST-JOB-002`: 예약 발행 대기 중인 글(`SCHEDULED`)이 예약 시각 도래 전에는 공개 API 및 검색 결과에서 `404`로 완벽히 숨겨져 있는지 검증.

## 17. Acceptance Criteria (인수 기준)
1. 예약 발행, 색인 갱신, 이미지 처리, 알림 발송 잡이 모두 비동기 워커에서 실행되고 실패 시 지수 백오프 재시도 및 DLQ 격리가 동작해야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- 예약 발행 동시성 통합 테스트 로그, `background_jobs` 및 `scheduled_publications` 상태 전이 DB 로그.

# A_ARCHITECTURE — Architecture (전체 시스템 아키텍처)
> **부제**: 독립형 글로벌 블로그 플랫폼의 전체 시스템 아키텍처, 모듈 경계, 서비스 간 데이터 흐름, 배포 구조, 장애 격리 및 확장 전략
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-A_ARCHITECTURE`
> **적용 조항**: 제3조 (A_ARCHITECTURE 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- TreeScan 서버·DB·인증 시스템 없이 **100% 단독 구동 가능한 글로벌 독립 블로그 플랫폼**의 기술 기반을 확립한다.
- 초기 과도한 마이크로서비스(MSA) 분할로 인한 분산 트랜잭션 장애와 운영 오버헤드를 방지하기 위해 **Go 기반 모듈형 모놀리스(Modular Monolith) + 독립 비동기 워커(Go Workers)** 구조를 정의한다.
- 트래픽 성장에 따라 검색, 미디어 처리, 광고 서빙, TreeScan 어댑터를 무중단으로 분리·확장할 수 있는 명확한 모듈 경계와 데이터 흐름을 수립한다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **System Architect / DevOps Engineer**: 단일 코드베이스에서 `treeblog-api`와 `treeblog-worker`를 빌드하고, 컨테이너 오케스트레이션을 통해 무중단 배포 및 오토스케일링을 운영한다.
- **Backend / Frontend Developer Agent**: 15개 도메인 모듈 경계(`internal/modules/*`)와 OpenAPI 3.1.0 계약을 준수하며 타 모듈과 충돌 없이 병렬 개발을 수행한다.
- **시나리오 A-1 (장애 격리)**: OpenSearch 클러스터나 이미지 변환 워커에 일시적 과부하가 발생하더라도, PostgreSQL 기반의 게시글 작성·임시저장·기본 열람 API는 Circuit Breaker와 Outbox 큐를 통해 정상 응답(Graceful Degradation)을 유지한다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-ARCH-001 (독립 구동 보장)**: 외부 TreeScan 의존성 없이 자체 인증, 자체 DB(PostgreSQL 16), 자체 검색(OpenSearch 2.x), 자체 오브젝트 스토리지(S3 호환)만으로 전체 기능이 동작해야 한다.
- **REQ-ARCH-002 (모듈 경계 강제)**: Go 백엔드는 `identity`, `user`, `blog`, `domain`, `theme`, `editor`, `post`, `media`, `community`, `notification`, `search`, `seo`, `moderation`, `monetization`, `admin`, `treescan` 모듈로 분리하며, 모듈 간 직접 테이블 쓰기(Cross-module Write)를 금지하고 Service Interface 또는 `outbox_events`로만 통신한다.
- **REQ-ARCH-003 (Transactional Outbox)**: DB 상태 변경과 외부 부수 효과(검색 색인, 캐시 무효화, 알림, 웹훅)는 반드시 `outbox_events` 테이블을 통해 원자성(Atomicity)과 최소 1회 전달(At-least-once delivery)을 보장한다.
- **REQ-ARCH-004 (멀티 테넌트 호스트 라우팅)**: 단일 프론트엔드(`apps/user-web`) 및 API 클러스터가 기본 도메인, 서브도메인(`{handle}.domain`), 개인 커스텀 도메인(`Host` 헤더 기반)을 동적으로 식별하고 격리된 블로그 컨텍스트를 로드해야 한다.

## 4. 데이터 구조 (Data Structures & Domain Models)
```go
// 도메인 이벤트 봉투 (Transactional Outbox 표준 구조)
type OutboxEvent struct {
    ID            uuid.UUID       `json:"id"`
    AggregateType string          `json:"aggregate_type"` // "post", "blog", "comment", "member"
    AggregateID   uuid.UUID       `json:"aggregate_id"`
    EventType     string          `json:"event_type"`     // "post.published.v1", "post.visibility_changed.v1"
    Payload       json.RawMessage `json:"payload"`
    IdempotencyKey string         `json:"idempotency_key"`
    Status        string          `json:"status"`         // "PENDING", "PROCESSING", "PUBLISHED", "DEAD_LETTER"
    RetryCount    int             `json:"retry_count"`
    CreatedAt     time.Time       `json:"created_at"`
}
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- **3-App 모노레포 아키텍처**:
  1. `apps/user-web` (Nuxt 3 + Vue 3 + TypeScript): SSR/ISR 기반 블로그 탐색 및 개별 블로그 렌더링.
  2. `apps/blog-console` (Vue 3 + Vite + TypeScript): 블로거 전용 스튜디오, Tiptap 에디터, 통계/수익 대시보드.
  3. `apps/admin-console` (Vue 3 + Vite + TypeScript): 플랫폼 운영자 통합 관리 콘솔.
- **공통 패키지**: `packages/ui-kit`, `packages/editor-core`, `packages/api-client`, `packages/theme-renderer`.

## 6. 백엔드 로직 (Backend Logic & Processing)
- **HTTP Server (`cmd/api/main.go`)**: Chi 라우터 기반 REST API 제공. 미들웨어 체인: `Recoverer` → `RequestID` → `OTelTracer` → `RealIP` → `HostResolver` → `RateLimiter` → `AuthVerifier` → `RBACGuard` → `CSRFProtector`.
- **Worker Server (`cmd/worker/main.go`)**: Outbox Poller(100ms 주기) + Asynq/Redis Streams Consumer + Cron Scheduler.
- **Graceful Shutdown**: `SIGTERM`/`SIGINT` 수신 시 신규 요청 수락 중단 후 진행 중인 HTTP 요청 및 워커 잡을 최대 30초간 안전하게 마무리한 뒤 종료.

## 7. API 계약 (API Contracts)
- `GET /api/v1/health/live` (Liveness Probe — 프로세스 생존 확인)
- `GET /api/v1/health/ready` (Readiness Probe — PostgreSQL, Redis, OpenSearch 연결 상태 확인)
- `GET /internal/v1/domains/resolve?host={hostname}` (Ingress 및 SSR 서버용 호스트→블로그 매핑 API)

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- 핵심 인프라 테이블: `outbox_events`, `background_jobs`, `system_settings`, `feature_flags`, `audit_logs`.
- 모든 비즈니스 테이블(총 58개)은 UUID v7(시간순 정렬 가능 UUID)을 기본 PK로 사용하여 B-Tree 인덱스 단편화를 방지하고 분산 확장성을 확보한다.

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **Outbox 이벤트 상태 전이**: `PENDING` → `PROCESSING` → `PUBLISHED` (성공 시) 또는 `FAILED` (지수 백오프 재시도 최대 10회 후 `DEAD_LETTER` 전이 및 관리자 알림 발생).
- **OpenSearch 장애 시 예외 처리**: 검색 API 호출 시 Circuit Breaker가 Open되면 즉시 `503 SEARCH_TEMPORARILY_UNAVAILABLE`을 반환하거나 인기 게시글 캐시 폴백을 제공하며, 쓰기 트랜잭션(글 발행)은 중단되지 않고 `search_index_jobs`에 안전하게 적재된다.

## 10. UX/UI 구조 (UX/UI Structure)
- 아키텍처 레벨에서는 사용자에게 시스템 내부 경계가 드러나지 않도록 단일 도메인/서브도메인 아래 매끄러운 전환(App Shell, SSO 세션 쿠키 공유)을 제공한다.
- 시스템 점검 또는 부분 장애 시 `SCR-SYS-503` 배너를 통해 읽기 전용 모드 전환 여부를 투명하게 안내한다.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- Zero-Trust 내부 통신: 관리자 콘솔 및 내부 엔드포인트(`/internal/*`)는 별도 네트워크 정책 및 서비스 토큰으로 보호.
- 모든 시크릿(DB 패스워드, JWT 서명키, S3 키)은 코드 저장소 커밋을 절대 금지하며 환경변수 또는 Secret Manager를 통해 런타임 주입한다.

## 12. 성능 및 확장성 (Performance & Scalability)
- **목표 성능**: 공개 블로그 SSR TTFB P95 < 150ms (Edge/Redis 캐시 히트 시 < 35ms), API 읽기 응답 P95 < 80ms, 쓰기 응답 P95 < 200ms.
- **확장 전략**: API 및 Nuxt SSR 노드는 무상태(Stateless)로 설계하여 CPU/RPS 기반 수평 오토스케일링(HPA)을 즉시 적용한다.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- **Go**: `github.com/go-chi/chi/v5`, `github.com/jackc/pgx/v5`, `github.com/hibiken/asynq`, `github.com/redis/go-redis/v9`, `go.opentelemetry.io/otel`.
- **Frontend**: `nuxt@^3.11`, `vue@^3.4`, `typescript@^5.4`, `pinia@^2.1`, `tailwindcss@^3.4`.

## 14. 개발 우선순위 (Development Priority)
- **Phase 0 (최우선 선행 작업)**: 모든 도메인 개발의 기반이므로 가장 먼저 구축해야 한다 (`TB-ARCH-001`, `TB-ARCH-002`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 선행 의존성 없음 (최상위 루트 모듈). 모든 후속 모듈(`B`~`Z`)이 본 아키텍처 규약을 의존한다.

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-ARCH-001`: 모듈 간 순환 참조 및 금지된 Cross-module DB 직접 참조 정적 분석(`go vet` + 아키텍처 린터) 통과 확인.
- `TEST-ARCH-002`: PostgreSQL 트랜잭션 롤백 시 `outbox_events`도 함께 롤백되며, 커밋 시 워커가 1초 이내에 이벤트를 소비하는지 검증.
- `TEST-ARCH-003`: OpenSearch 컨테이너 강제 중지 상태에서 게시글 작성/발행 API가 `201 Created`로 정상 성공하고 복구 후 자동 색인되는지 검증.

## 17. Acceptance Criteria (인수 기준)
1. Docker Compose 단일 명령(`docker compose up -d`)으로 전체 스택이 기동되고 `/api/v1/health/ready`가 `200 OK`를 반환해야 한다.
2. TreeScan 관련 환경변수가 전혀 설정되지 않은 상태에서도 모든 API 및 웹 화면이 에러 없이 구동되어야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- `docker compose ps` 상태 출력 로그, `/api/v1/health/ready` JSON 응답 캡처, 아키텍처 의존성 그래프 검증 스크립트 실행 로그.

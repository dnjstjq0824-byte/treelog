# SYSTEM_OVERVIEW.md — TreeBlog 전체 시스템 아키텍처 개요

## 1. 아키텍처 핵심 철학
TreeBlog는 **독립성(Autonomy)**, **고성능 SEO(SSR/Edge Caching)**, **데이터 무결성(PostgreSQL Single Source of Truth)**, **단계적 확장성(Modular Monolith to Selective Microservices)**을 4대 기둥으로 설계되었다.

```
[글로벌 사용자 / 검색엔진 봇]
       │ (HTTPS / HTTP3)
       ▼
[Global CDN & Edge WAF / DDoS Protection]
       │
       ├──► Static Assets & Public Media Variants ──► [S3 Public/CDN Bucket]
       │
       ▼ (Dynamic Routing & Custom Domain Ingress / Auto-TLS)
[Ingress Controller / Reverse Proxy (Caddy / Envoy)]
       │
       ├──────────────────────────┬──────────────────────────┐
       ▼                          ▼                          ▼
[User Web (Nuxt 3 SSR)]    [Blog & Admin Console]     [Go Backend API Server]
 - Public Blog Rendering    - Vue 3 + TS SPA/SSR       - Modular Monolith (15 Modules)
 - Global Explore & Search  - Tiptap Rich Editor       - REST API (/api/v1/*)
 - Dynamic Sitemap/RSS      - Analytics & Ad Admin     - OIDC / JWT / RBAC Middleware
       │                          │                          │
       └──────────────────────────┴──────────────────────────┤
                                                             │
         ┌────────────────────────┬──────────────────────────┼────────────────────────┐
         ▼                        ▼                          ▼                        ▼
[PostgreSQL 16 Primary]    [Redis 7 Cluster]        [OpenSearch 2.x Cluster]   [S3 Private Bucket]
 - 58 Core Tables           - Session & Rate Limit   - Posts & Blogs Index      - Raw Uploads
 - ACID & Outbox Events     - Hot Feed & SSR Cache   - Nori/ICU Analyzers       - Backups & Exports
 - Read Replicas            - Asynq Job Queues       - Autocomplete & Ranking
         │                        │                          ▲
         └───────────┬────────────┘                          │
                     ▼                                       │
          [Go Background Workers (treeblog-worker)] ─────────┘
           - Transactional Outbox Relay
           - Scheduled Post Publisher
           - Media Processor (libvips WebP/AVIF, EXIF strip)
           - Notification Dispatcher (Web, Push, SES Email)
           - Revenue & Ledger Settlement Batch
           - TreeScan Integration Adapter (Future Sync)
```

## 2. 계층별 아키텍처 구성 요소

### 2.1 Edge & Ingress Layer
- **Global CDN & WAF**: 정적 에셋(`/_nuxt/*`), 변환 완료된 공개 이미지(`media.treeblog.example.com`), 공개 블로그 게시글의 Stale-While-Revalidate(SWR) 엣지 캐싱을 수행하며, SQLi/XSS/Bot Flood를 1차 방어한다.
- **Custom Domain & Auto-TLS Router**:
  - 플랫폼 기본 도메인(`treeblog.example.com`), 서브도메인(`{handle}.treeblog.example.com`), 그리고 사용자가 연결한 개인 도메인(`blog.customdomain.com`) 트래픽을 모두 수신한다.
  - 요청의 `Host` 헤더를 분석하여 내부 Go API(`GET /internal/v1/domains/resolve?host=...`, Redis 캐싱 적용)를 통해 `blog_id`로 매핑하고, 검증된 도메인에 대해 ACME(Let's Encrypt) TLS 인증서를 자동 발급·갱신한다.

### 2.2 Frontend Layer (Vue 3 + Nuxt 3 + TypeScript)
- **`apps/user-web` (Nuxt 3 SSR/ISR)**:
  - 플랫폼 탐색 메인, 카테고리/태그 탐색, 통합 검색 결과, 그리고 모든 개별 블로그(`/@{handle}`, `{handle}.domain`, 커스텀 도메인)의 화면을 서버 사이드 렌더링(SSR)한다.
  - 비공개 게시글이나 이웃 전용 게시글 접근 시 SSR 단계에서 즉시 쿠키/세션 권한을 검사하여 비인가 요청에는 본문 데이터를 내려주지 않고 `403`/`404` 및 `X-Robots-Tag: noindex` 헤더를 반환한다.
- **`apps/blog-console` (Vue 3 + TypeScript)**:
  - 블로거가 사용하는 '내 블로그 관리 센터' 및 'Tiptap 글쓰기 에디터' 웹 애플리케이션이다.
  - IndexedDB 기반 로컬 자동저장과 서버 자동저장을 결합하여 네트워크 단절 시에도 작성 중인 원고 유실을 0%로 방지한다.
- **`apps/admin-console` (Vue 3 + TypeScript)**:
  - 플랫폼 운영자 전용 통합 관리 콘솔이다. 사내 VPN/Zero-Trust 네트워크 및 MFA 강제 정책 하에서 접근하며, 회원·블로그·게시글·댓글·검색·광고·정산·시스템 설정을 제어한다.

### 2.3 Backend Core Layer (Go Modular Monolith)
- 단일 코드베이스 내 `internal/modules/` 아래 15개 도메인 모듈이 명확한 인터페이스 경계를 가지고 동작한다.
- **Clean/Hexagonal 계층 구조**: 각 모듈은 `handler` (HTTP/REST 입출력 및 입력 검증) → `service` (도메인 비즈니스 규칙 및 트랜잭션 경계) → `repository` (`pgx`/`sqlc` 기반 타입 안전 SQL 실행) → `domain` (엔터티, 상태 머신, 도메인 이벤트)으로 구성된다.

### 2.4 Asynchronous Worker Layer (`treeblog-worker`)
- API 서버와 동일한 도메인 모델을 공유하되 별도 프로세스로 구동된다.
- **Transactional Outbox Relay**: 100ms 주기로 `outbox_events` 테이블의 미처리 이벤트를 읽어 Redis Streams/Asynq 작업 큐로 발행한다.
- **전담 워커 풀**:
  - `queue:critical` (우선순위 100): 비공개/삭제/제재 게시글의 검색 색인 즉시 삭제, 보안 알림, 예약 발행 정시 트리거.
  - `queue:default` (우선순위 50): 일반 게시글 색인 갱신, 이미지 리사이징 및 EXIF 제거, 댓글/구독 알림 발송.
  - `queue:bulk` (우선순위 10): 외부 블로그 가져오기(WXR Import), 블로그 전체 백업 ZIP 생성, 일일 통계 집계 및 정산 원장 배치.

### 2.5 Data & Storage Layer
- **PostgreSQL 16**: 모든 핵심 데이터의 원본. Primary-Replica 구조로 읽기 부하를 분산하며, 대용량 테이블(`ad_events`, `search_click_events`, `audit_logs`, `notifications`)은 월별 Range Partitioning을 적용한다.
- **OpenSearch 2.15**: 게시글(`posts_v1`), 블로그(`blogs_v1`), 태그/자동완성(`suggest_v1`) 인덱스 운영.
- **Redis 7.2**: 세션 관리, 슬라이딩 윈도우 Rate Limiter, 도메인 라우팅 캐시, 조회수 어뷰징 방지 HyperLogLog/Set, 분산 락(Redlock).
- **S3 호환 오브젝트 스토리지**:
  - `treeblog-raw-private`: 업로드 원본 파일 및 내보내기/가져오기 아카이브 저장 (외부 직접 접근 전면 차단).
  - `treeblog-media-public`: EXIF가 제거되고 WebP/AVIF로 최적화된 공개용 미디어 파생본 저장 (CDN 연동).

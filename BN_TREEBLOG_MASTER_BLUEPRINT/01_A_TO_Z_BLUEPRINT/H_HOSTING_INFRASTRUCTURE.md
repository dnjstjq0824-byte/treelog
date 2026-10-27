# H_HOSTING_INFRASTRUCTURE — Hosting & Infrastructure (호스팅 및 인프라스트럭처)
> **부제**: 서버, DB, 캐시, 스토리지, CDN, 도메인 라우팅, 프록시, WAF, 모니터링, 백업 및 장애 복구
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-H_HOSTING_INFRASTRUCTURE`
> **적용 조항**: 제3조 (H_HOSTING_INFRASTRUCTURE 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- 상용 글로벌 블로그 서비스를 99.95% 이상의 가용성(SLA)으로 안정적으로 운영하기 위한 컨테이너 인프라, 데이터베이스 고가용성, 글로벌 CDN 캐싱, OpenTelemetry 관측성, 그리고 자동화된 백업 및 재해 복구(DR) 체계를 정의한다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **DevOps / SRE Agent**: CI/CD 파이프라인을 통해 무중단 카나리/롤링 배포를 수행하고, OpenTelemetry 대시보드에서 지연 시간 급증이나 에러율 상승 알람을 수신하여 런북(`OPERATIONS_RUNBOOK.md`)에 따라 대응한다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-INF-001 (컨테이너 표준화)**: 모든 애플리케이션 컴포넌트(`api`, `worker`, `user-web`, `blog-console`, `admin-console`, `caddy-ingress`)는 멀티스테이지 `Dockerfile`로 빌드되며 Non-root 유저(UID 65532)로 실행된다.
- **REQ-INF-002 (CDN & Edge WAF)**:
  - 정적 에셋(`/_nuxt/*`) 및 변환된 미디어 파일은 `Cache-Control: public, max-age=31536000, immutable`로 CDN에 영구 캐싱한다.
  - 공개 블로그 게시글 HTML은 `Cache-Control: public, s-maxage=60, stale-while-revalidate=600` 및 `Surrogate-Key: blog-{blogId} post-{postId}`를 부여하여 글 수정/비공개 전환 시 즉시 퍼지(Purge)한다.
- **REQ-INF-003 (OpenTelemetry 통합 관측성)**:
  - **Traces**: 프론트엔드 SSR → Go API → PostgreSQL/Redis/OpenSearch/Worker 전 구간에 W3C `traceparent` 컨텍스트를 전파한다.
  - **Metrics**: RED 메트릭(Rate, Errors, Duration), Go 런타임 메모리/고루틴, DB 커넥션 풀, Asynq 큐 적체량을 Prometheus 포맷으로 노출한다.
  - **Logs**: 민감정보(비밀번호, 토큰, 이메일 일부)가 마스킹된 구조화 JSON 로그(`trace_id`, `span_id`, `request_id`, `user_id` 포함)를 출력한다.
- **REQ-INF-004 (백업 및 재해 복구 목표)**:
  - **PostgreSQL**: 일 1회 전체 스냅샷 + WAL(Write-Ahead Log) 실시간 아카이빙을 통한 PITR(Point-In-Time Recovery) 구성. **목표 RPO < 5분, 목표 RTO < 30분**.
  - **S3 미디어**: 버전 관리(Versioning) 및 리전 간 복제(CRR) 또는 일일 증분 백업.
  - **OpenSearch**: 일 1회 스냅샷 백업 + 최악의 경우 PostgreSQL 원본으로부터 전체 재색인 스크립트(`cmd/reindex`) 상시 구비.

## 4. 데이터 구조 (Data Structures & Domain Models)
```yaml
# OpenTelemetry Collector 및 서비스 헬스체크 표준 응답 구조
status: "healthy"
version: "1.0.0"
commit_sha: "e4a91bc..."
dependencies:
  postgres: { status: "up", latency_ms: 1.4 }
  redis:    { status: "up", latency_ms: 0.6 }
  opensearch: { status: "up", latency_ms: 4.2 }
  object_storage: { status: "up", latency_ms: 8.1 }
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- Nuxt 3 서버의 `/api/health` 엔드포인트 및 클라이언트 측 Core Web Vitals 비콘 수집기 구성.

## 6. 백엔드 로직 (Backend Logic & Processing)
- PgBouncer 또는 `pgxpool` 커넥션 풀링(MaxConns, MinConns, MaxConnLifetime) 튜닝.
- 읽기 전용 쿼리(공개 게시글 조회, 통계 조회)를 Read Replica로 자동 라우팅하는 `DBRouter` 래퍼 구현.

## 7. API 계약 (API Contracts)
- `GET /api/v1/health/live`, `GET /api/v1/health/ready`, `GET /metrics` (내부 네트워크 전용 Prometheus 스크래핑 엔드포인트).

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- 인프라 상태 추적: `background_jobs`, `outbox_events`, `system_settings` (점검 모드 플래그 등).

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **Read Replica 지연(Replication Lag) 예외**: 작성자가 글을 막 발행하거나 수정한 직후 본인이 조회할 때는 세션에 5초간 `force_primary_read = true` 플래그를 설정하여 Read-Your-Own-Writes 일관성을 보장한다.

## 10. UX/UI 구조 (UX/UI Structure)
- 관리자 콘솔 내 `SCR-ADM-008` (시스템 관리 — 백업 상태, 워커 큐 적체 현황, API 응답 시간, 에러율 모니터링 대시보드) 제공.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- 네트워크 3-Tier 분리: Public Subnet(Load Balancer/Ingress) → Private App Subnet(API/SSR/Workers) → Isolated Data Subnet(PostgreSQL/Redis/OpenSearch). 데이터 서브넷은 외부 인터넷 아웃바운드/인바운드를 전면 차단한다.

## 12. 성능 및 확장성 (Performance & Scalability)
- 단일 API 노드(4 vCPU, 8GB RAM) 기준 초당 2,500+ RPS(캐시 결합 시 15,000+ RPS) 처리 가능하도록 설계.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- `go.opentelemetry.io/otel`, `github.com/prometheus/client_golang`, `rs/zerolog` 또는 표준 `log/slog`.

## 14. 개발 우선순위 (Development Priority)
- **Phase 0 (Docker/CI 기초)** 및 **Phase 9 (고가용성/모니터링/백업 복구 리허설)** (`TB-OPS-001`, `TB-OPS-002`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 선행: `A_ARCHITECTURE`.
- 후행: 전체 시스템 배포 및 운영.

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-INF-001`: PostgreSQL 샘플 데이터 삽입 → 백업 수행 → 인위적 테이블 DROP → 백업 복구 스크립트 실행 후 데이터 100% 일치 검증.
- `TEST-INF-002`: API 요청 시 응답 헤더의 `X-Request-Id`와 서버 JSON 로그의 `request_id`, OpenTelemetry Trace ID가 일치하는지 검증.

## 17. Acceptance Criteria (인수 기준)
1. `16_INFRASTRUCTURE/`에 정의된 배포, CI/CD, 모니터링, 백업/복구, 운영 런북 절차에 따라 로컬 및 스테이징 환경이 재현 가능하게 구축되어야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- Docker 이미지 빌드 성공 로그, 백업/복원 리허설 실행 로그, Prometheus/OTel 메트릭 출력 샘플.

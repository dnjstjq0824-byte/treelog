# DEPLOYMENT_ARCHITECTURE.md — 컨테이너 및 멀티 환경 배포 아키텍처

## 1. 환경 분리 원칙 (Environment Isolation)
TreeBlog는 코드 변경부터 상용 배포까지의 안정성을 보장하기 위해 4단계 격리 환경을 운영한다.

| 환경 (Environment) | 목적 | 인프라 구성 | TreeScan 연동 상태 |
| :--- | :--- | :--- | :--- |
| **`local` (Docker Compose)** | 개발자 및 에이전트 로컬 구현·통합·E2E 테스트 | 단일 호스트 Docker Compose (`postgres`, `redis`, `opensearch`, `minio`, `mailpit`, `api`, `worker`, `web`) | Mock Adapter 내장 |
| **`ci` (Ephemeral Test)** | GitHub Actions PR 검증 및 독립 감사(Independent Audit) | 매 실행 시 클린 컨테이너 기동 → 마이그레이션 → 통합/Playwright E2E 테스트 후 즉시 파기 | Contract Test 실행 |
| **`staging` (Pre-Prod)** | 상용 동일 아키텍처 성능·보안·복구 리허설 | Kubernetes (또는 ECS/Docker Swarm) 멀티 노드, Managed DB/OpenSearch, Staging CDN | Sandbox Adapter 연동 |
| **`production` (Live)** | 실제 글로벌 사용자 서비스 운영 (PO 승인 필수) | Multi-AZ 고가용성 클러스터, Primary + Read Replica DB, 3-Node OpenSearch, Global CDN + WAF | Feature Flag로 제어 |

## 2. 프로덕션 고가용성(HA) 토폴로지

```
                    [Route53 / Global DNS + Anycast CDN (Cloudflare / CloudFront)]
                                                  │
                              [WAF & DDoS Protection + Rate Limiting]
                                                  │
                 ┌────────────────────────────────┴────────────────────────────────┐
                 ▼ (AZ-a)                                                          ▼ (AZ-b)
      [Ingress / Caddy Auto-TLS]                                        [Ingress / Caddy Auto-TLS]
                 │                                                                 │
     ┌───────────┼───────────┐                                         ┌───────────┼───────────┐
     ▼           ▼           ▼                                         ▼           ▼           ▼
[user-web]  [api-server] [consoles]                               [user-web]  [api-server] [consoles]
(Nuxt SSR)   (Go API)    (Static/CDN)                             (Nuxt SSR)   (Go API)    (Static/CDN)
     │           │                                                     │           │
     └───────────┼─────────────────────────────────────────────────────┼───────────┘
                 │                                                     │
     ┌───────────┴───────────────┬──────────────────────┬──────────────┴───────────┐
     ▼                           ▼                      ▼                          ▼
[PostgreSQL 16 HA]       [Redis 7 Sentinel/Cluster]  [OpenSearch 3+ Nodes]   [Go Worker Pool]
- Primary (AZ-a)         - Master (AZ-a)             - Master/Data Across    - Critical Workers
- Sync Standby (AZ-b)    - Replica (AZ-b)              AZ-a, AZ-b, AZ-c      - Default Workers
- Read Replicas          - AOF Persistence           - Automated Snapshots   - Media Workers
```

## 3. 컨테이너 구성 명세 (`docker-compose.yml` 표준 기준)

1. **`treeblog-api`**: Go 1.22 멀티스테이지 빌드(`distroless/static-debian12:nonroot`), 포트 `8080` (REST API), `9090` (Prometheus Metrics/Health).
2. **`treeblog-worker`**: 동일 바이너리의 워커 모드 실행(`--mode=worker --queues=critical,default,media,bulk`), 무상태 수평 확장(HPA) 지원.
3. **`treeblog-user-web`**: Node 20 Alpine 기반 Nuxt 3 Nitro SSR 서버, 포트 `3000`.
4. **`treeblog-blog-console` & `treeblog-admin-console`**: 정적 빌드 후 CDN/Nginx를 통해 서빙하거나 Nuxt 하이브리드 모드로 서빙.
5. **`treeblog-caddy`**: 커스텀 도메인 동적 TLS 발급(On-Demand TLS) 및 리버스 프록시 담당. `/internal/v1/domains/ask` 엔드포인트와 통신하여 검증된 도메인만 인증서 발급 허용.

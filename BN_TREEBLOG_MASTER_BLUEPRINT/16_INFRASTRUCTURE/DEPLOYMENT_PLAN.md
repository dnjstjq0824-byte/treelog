# DEPLOYMENT_PLAN.md — 컨테이너 패키징 및 무중단 배포 계획서

## 1. 배포 아키텍처 및 무중단 롤아웃 절차
1. **컨테이너 이미지 구성**:
   - `treeblog-api`: Go 1.22 정적 컴파일 바이너리 (`CGO_ENABLED=1` with `libvips` for media worker, Distroless/Debian-slim non-root 이미지).
   - `treeblog-worker`: 백그라운드 큐(`critical`, `default`, `media`, `bulk`) 및 스케줄러 전용 워커 컨테이너.
   - `treeblog-user-web`: Nuxt 3 SSR 서버 컨테이너.
   - `treeblog-consoles`: 블로그 관리 센터 및 어드민 콘솔 에셋 서빙 컨테이너.
2. **Zero-Downtime 배포 순서**:
   - Step 1: 하위 호환성이 보장된(Expand-and-Contract 패턴) PostgreSQL 마이그레이션 Job 실행 (`TB-DB-001`~`002`).
   - Step 2: `treeblog-worker` 롤링 업데이트 (진행 중인 작업은 Graceful Shutdown 최대 30초 대기 후 종료).
   - Step 3: `treeblog-api` 및 `treeblog-user-web` 롤링 업데이트 (`Readiness Probe: /api/v1/health/ready` 통과 확인 후 구 파드 종료).
   - Step 4: 배포 직후 자동 Smoke Test 실행 (실패 시 즉시 이전 이미지 태그로 자동 롤백).

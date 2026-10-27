# TB-ARCH-002 — Transactional Outbox 이벤트 릴레이 및 Asynq 멀티 큐 워커 엔진 구축

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 0`

---

## 1. Work Order ID
- `TB-ARCH-002`

## 2. 작업명 (Task Title)
- **Transactional Outbox 이벤트 릴레이 및 Asynq 멀티 큐 워커 엔진 구축**

## 3. 목적 (Purpose)
- DB 트랜잭션과 비동기 작업(검색 색인, 알림, 이미지 처리, TreeScan 동기화) 간의 원자성을 보장하는 Outbox 패턴과 4단계 우선순위 큐를 구축한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Lead Architect Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-DB-001`

## 6. 구현 범위 (Implementation Scope)
- Outbox Publisher, 100ms 주기 Poller(SKIP LOCKED), Asynq critical/default/media/bulk 큐 및 지수 백오프 재시도/DLQ 구현

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/platform/outbox/, internal/worker/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `outbox_events, background_jobs` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `GET /api/v1/admin/jobs` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-ADM-008` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 동일 idempotency_key 중복 발행 차단. 2) 최대 8회 실패 시 DEAD_LETTER 상태 전이 및 에러 기록.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/platform/outbox/publisher.go, internal/platform/outbox/relay.go, internal/worker/server.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-ARCH-002, TEST-JOB-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- DB 트랜잭션 롤백 시 Outbox 이벤트가 발행되지 않고, 커밋 시 500ms 내에 워커가 수신해야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 다중 워커 중복 폴링 -> FOR UPDATE SKIP LOCKED로 완벽 해결

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-ARCH-002/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

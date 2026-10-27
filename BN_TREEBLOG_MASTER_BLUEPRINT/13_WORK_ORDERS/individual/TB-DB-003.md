# TB-DB-003 — 데이터베이스 백업·시점복구(PITR) 스크립트 및 파티션 자동 관리 워커 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 2`

---

## 1. Work Order ID
- `TB-DB-003`

## 2. 작업명 (Task Title)
- **데이터베이스 백업·시점복구(PITR) 스크립트 및 파티션 자동 관리 워커 구현**

## 3. 목적 (Purpose)
- 월별 파티션 테이블의 차월 파티션 사전 자동 생성 워커와 pg_dump / WAL 기반 백업 및 무결성 복원 검증 스크립트를 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **PostgreSQL Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-DB-002, TB-ARCH-002`

## 6. 구현 범위 (Implementation Scope)
- 매월 20일 익월/익익월 파티션 자동 DDL 생성 잡, 백업/복원 자동화 쉘 스크립트

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `03_DATABASE/, infra/scripts/backup/, internal/db/partition_manager.go`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `notifications, search_click_events, ad_events, audit_logs` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `N/A` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-ADM-008` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 파티션 누락으로 인한 INSERT 실패를 막기 위해 미래 3개월치 파티션 상시 유지 및 _default 파티션 모니터링.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/db/partition_manager.go, infra/scripts/backup/pg_backup_verify.sh`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-INF-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 백업 생성 후 임시 DB에 복원했을 때 57개 테이블 스키마와 레코드 수가 100% 일치해야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 파티션 생성 중 락 경합 -> 트래픽 최저 시간대 실행 및 IF NOT EXISTS 사용

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-DB-003/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

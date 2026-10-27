# TB-QA-001 — 전체 도메인 단위·실제 DB/OpenSearch 통합 테스트 및 보안 침투(XSS/SSRF/CSRF/IDOR) 테스트 스위트 구축

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 9`

---

## 1. Work Order ID
- `TB-QA-001`

## 2. 작업명 (Task Title)
- **전체 도메인 단위·실제 DB/OpenSearch 통합 테스트 및 보안 침투(XSS/SSRF/CSRF/IDOR) 테스트 스위트 구축**

## 3. 목적 (Purpose)
- Mock을 배제하고 실제 PostgreSQL, Redis, OpenSearch, MinIO 컨테이너를 대상으로 전 API 계약, DB 불변식, 20대 보안 위협 방어를 자동 검증하는 테스트 스위트를 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **QA / E2E Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-ADMIN-003, TB-BILLING-002`

## 6. 구현 범위 (Implementation Scope)
- 24개 API 그룹 통합 테스트, OWASP 보안 퍼징 테스트, 복식부기 원장 및 가시성 불변식 검증기

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `tests/integration/, tests/security/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `All 57 Tables` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `All 106 Endpoints` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `N/A` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 테스트 종료 시 AssertDatabaseInvariants(고아 레코드 0건, 차변==대변, 비공개 글 색인 0건)를 강제 실행.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `tests/integration/api_suite_test.go, tests/security/penetration_test.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-SEC-001, TEST-SEC-002, TEST-SEC-003` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 전체 통합 및 보안 테스트 케이스가 실제 컨테이너 위에서 100% 통과해야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 비동기 워커 타이밍 플래키(Flaky) 테스트 -> 결정론적 Outbox Flush 헬퍼 제공

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-QA-001/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

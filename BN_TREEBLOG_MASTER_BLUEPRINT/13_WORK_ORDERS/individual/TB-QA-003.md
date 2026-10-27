# TB-QA-003 — Independent Audit Agent 전수 독립 감사, TreeScan 무의존 단독 구동 검증 및 최종 납품 승인

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 10`

---

## 1. Work Order ID
- `TB-QA-003`

## 2. 작업명 (Task Title)
- **Independent Audit Agent 전수 독립 감사, TreeScan 무의존 단독 구동 검증 및 최종 납품 승인**

## 3. 목적 (Purpose)
- 구현 에이전트와 분리된 독립 감사 에이전트가 제25조 및 제27조 금지사항 위반 여부, TreeScan 서버 없는 100% 단독 구동, 전 테스트 진위 여부를 감사한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Independent Audit Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-QA-002, TB-TREESCAN-001, TB-TREESCAN-002`

## 6. 구현 범위 (Implementation Scope)
- 코드베이스 정적/동적 감사, 시크릿 스캔, Mock 위장 여부 검사, E2E 재실행, 최종 감사 리포트 발행

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `15_TESTING/, 18_FINAL_DELIVERY/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `cmd/, internal/, apps/, packages/, 03_DATABASE/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `All 57 Tables` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `All 106 Endpoints` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `All 45 Screens` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 코드 수정 권한 없이 오직 검증만 수행하며, 단 하나의 실패나 가짜 테스트 발견 시 즉시 반려(FAILED) 처리.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `15_TESTING/INDEPENDENT_AUDIT.md, 18_FINAL_DELIVERY/FINAL_BLUEPRINT_REPORT.md`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-ZTR-001, TEST-ZTR-002, TEST-ZTR-003` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 모든 Work Order가 VERIFIED 상태이고 제27조의 14대 금지사항 위반이 0건이어야 최종 승인 가능
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 감사 누락 -> 18개 자동 검증 체크리스트 스크립트 강제 실행

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-QA-003/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

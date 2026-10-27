# TB-ADMIN-003 — 통합 관리자 콘솔 검색 관리·광고/정산 심사·시스템 설정·감사 로그 UI(SCR-ADM-006~008) 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 7`

---

## 1. Work Order ID
- `TB-ADMIN-003`

## 2. 작업명 (Task Title)
- **통합 관리자 콘솔 검색 관리·광고/정산 심사·시스템 설정·감사 로그 UI(SCR-ADM-006~008) 구현**

## 3. 목적 (Purpose)
- 검색 색인 상태/오류 재처리, 스팸 검색어 관리, 광고 소재 심사, 크리에이터 정산 승인, Feature Flag 제어 및 감사 로그 Diff 뷰어 화면을 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Admin Console Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-ADMIN-002`

## 6. 구현 범위 (Implementation Scope)
- SCR-ADM-006 ~ SCR-ADM-008 3개 관리자 화면 완성

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `apps/admin-console/src/pages/search/, monetization/, system/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `internal/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `search_index_jobs, ad_campaigns, payouts, feature_flags, audit_logs` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-ADM-006, EP-ADM-007, EP-ADM-008` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-ADM-006, SCR-ADM-007, SCR-ADM-008` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 감사 로그 뷰어에서 변경 전(before_state_json)과 변경 후(after_state_json) 하이라이트 비교 제공.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `apps/admin-console/src/pages/search/index.vue, monetization/index.vue, system/audit-logs.vue`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-ADM-003, TEST-ADM-004` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- Feature Flag 토글, 실패 색인 잡 재시도, 감사 로그 조회가 실제 백엔드와 연동되어 동작해야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 민감 설정 노출 -> SUPER_ADMIN 외 시스템 설정 수정 버튼 비활성화 및 API 차단

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-ADMIN-003/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

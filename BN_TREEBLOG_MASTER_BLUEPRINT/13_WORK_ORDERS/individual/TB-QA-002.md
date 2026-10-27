# TB-QA-002 — Playwright 멀티 브라우저 E2E(가입->블로그개설->글쓰기->발행->검색->이웃/댓글->어드민) 및 a11y/성능 검증

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 9`

---

## 1. Work Order ID
- `TB-QA-002`

## 2. 작업명 (Task Title)
- **Playwright 멀티 브라우저 E2E(가입->블로그개설->글쓰기->발행->검색->이웃/댓글->어드민) 및 a11y/성능 검증**

## 3. 목적 (Purpose)
- 실제 헤드리스 브라우저(Playwright)로 제27조의 핵심 사용자 여정 전체와 320px~데스크톱 반응형, axe-core 웹 접근성, k6 부하 성능을 검증한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **QA / E2E Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-QA-001, TB-OPS-003`

## 6. 구현 범위 (Implementation Scope)
- Playwright E2E 시나리오, axe-core 접근성 전수 스캔, k6 부하 테스트 스크립트

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `tests/e2e/, tests/load/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `internal/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `All 57 Tables` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `All 106 Endpoints` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `All 45 Screens` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 회원가입, 블로그 생성, 이미지 첨부 글 작성/발행, 검색, 상호이웃, 댓글/공감, 관리자 제재까지 E2E 100% 통과.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `tests/e2e/full_journey.spec.ts, tests/e2e/accessibility.spec.ts, tests/load/k6_blog_read.js`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-FE-001, TEST-UX-001, TEST-UX-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- Playwright 전 시나리오 통과 및 axe-core Critical/Serious 접근성 위반 0건 달성
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- CI 환경 렌더링 속도 차이 -> 명시적 data-testid 및 네트워크 idle 대기 규약 적용

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-QA-002/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

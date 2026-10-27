# TB-WEB-001 — 디자인 시스템(packages/ui-kit) 및 회원 인증·프로필·보안 설정 UI 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 1`

---

## 1. Work Order ID
- `TB-WEB-001`

## 2. 작업명 (Task Title)
- **디자인 시스템(packages/ui-kit) 및 회원 인증·프로필·보안 설정 UI 구현**

## 3. 목적 (Purpose)
- Forest & Paper 디자인 토큰 기반 공통 UI 컴포넌트와 회원가입, 로그인, 비밀번호 복구, 프로필/MFA/세션 관리, 계정 탈퇴 화면을 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Frontend Developer Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-API-001, TB-AUTH-002`

## 6. 구현 범위 (Implementation Scope)
- 32개 공통 UI 컴포넌트(Loading/Empty/Error Boundary 포함) 및 5개 회원 핵심 화면 반응형 구현

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `packages/ui-kit/, apps/user-web/src/pages/auth/, apps/user-web/src/pages/settings/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `internal/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `users, user_profiles, sessions` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-AUTH-001~009, EP-USER-001~006` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-MEM-001, SCR-MEM-002, SCR-MEM-003, SCR-MEM-004, SCR-MEM-015` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 320px~2560px 반응형 대응. 2) WCAG 2.1 AA 키보드 포커스 및 ARIA 속성 완비.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `packages/ui-kit/src/components/*, apps/user-web/src/pages/auth/login.vue, register.vue`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-USR-001, TEST-UX-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 회원가입부터 로그인, MFA 설정, 타 기기 로그아웃까지 브라우저에서 동작해야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- SSR 하이드레이션 불일치 -> 클라이언트 전용 상태는 ClientOnly 또는 쿠키 동기화 사용

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-WEB-001/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

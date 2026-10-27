# TB-THEME-001 — 블로그 테마·스킨 엔진, 디자인 토큰 주입기 및 커스텀 CSS 보안 파서/샌드박스 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 4`

---

## 1. Work Order ID
- `TB-THEME-001`

## 2. 작업명 (Task Title)
- **블로그 테마·스킨 엔진, 디자인 토큰 주입기 및 커스텀 CSS 보안 파서/샌드박스 구현**

## 3. 목적 (Purpose)
- 6종 공식 반응형 베이스 테마와 제7조의 19대 디자인 커스터마이징 설정, 그리고 악성 스크립트/외부 임포트를 차단하는 커스텀 CSS 샌드박스를 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Frontend Developer Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-BLOG-001, TB-SEC-001`

## 6. 구현 범위 (Implementation Scope)
- 테마 카탈로그 API, 블로그 테마 토큰 오버라이드 저장, 기본값 복원, CSS AST Sanitizer 및 .tb-blog-theme-root 스코핑

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/theme/, packages/theme-renderer/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/admin-console/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `themes, blog_themes` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-THM-001, EP-THM-002, EP-THM-003, EP-THM-004` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-BLOG-001, SCR-MEM-006` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) SSR 시점에 CSS 변수를 인라인 주입하여 다크모드/폰트 깜빡임(FOUC) 0% 달성. 2) javascript: url 및 플랫폼 헤더 은닉 CSS 차단.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/theme/service.go, internal/modules/theme/css_sanitizer.go, packages/theme-renderer/src/ThemeProvider.vue`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-DSN-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 테마 변경 및 커스텀 CSS 저장 시 악성 속성이 제거되고 기본값 원클릭 복원이 정상 동작해야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 사용자 CSS가 신고/광고 배지를 숨김 -> !important 보호 클래스 및 셀렉터 필터로 차단

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-THEME-001/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

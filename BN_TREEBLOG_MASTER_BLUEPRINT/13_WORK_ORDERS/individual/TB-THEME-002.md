# TB-THEME-002 — 블로그 관리 센터 내 실시간 비주얼 테마 스튜디오(모바일/데스크톱 프리뷰어) UI 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 4`

---

## 1. Work Order ID
- `TB-THEME-002`

## 2. 작업명 (Task Title)
- **블로그 관리 센터 내 실시간 비주얼 테마 스튜디오(모바일/데스크톱 프리뷰어) UI 구현**

## 3. 목적 (Purpose)
- 좌측 설정 패널(스킨/다크모드/컬러/폰트/프로필/카드/레이아웃/위젯/메뉴/푸터)과 우측 실시간 반응형 프리뷰어(Mobile/Tablet/Desktop)를 갖춘 테마 스튜디오를 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Frontend Developer Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-THEME-001, TB-WEB-002`

## 6. 구현 범위 (Implementation Scope)
- 비주얼 테마 커스터마이저 UI, 사이드바 위젯 순서 편집기, 상단 메뉴/푸터 편집기

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `apps/blog-console/src/pages/studio/[blogId]/theme/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `internal/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `themes, blog_themes, blog_navigation` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-THM-001~004, EP-BMGT-006` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-MEM-006` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 저장 전 실시간 postMessage 브리지로 우측 프리뷰 즉시 반영. 2) 기본값 복원 확인 모달 제공.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `apps/blog-console/src/pages/studio/[blogId]/theme/index.vue`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-DSN-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 프리뷰어에서 375px/768px/1440px 해상도 전환 및 테마 저장이 1초 내 반영되어야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 프리뷰 iframe 메모리 누수 -> 컴포넌트 언마운트 시 리스너 정리

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-THEME-002/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

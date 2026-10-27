# TB-WEB-004 — 플랫폼 탐색 메인, 인기 블로그/게시글, 카테고리/주제/태그 탐색 화면(SCR-EXP-001~010) 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 4`

---

## 1. Work Order ID
- `TB-WEB-004`

## 2. 작업명 (Task Title)
- **플랫폼 탐색 메인, 인기 블로그/게시글, 카테고리/주제/태그 탐색 화면(SCR-EXP-001~010) 구현**

## 3. 목적 (Purpose)
- 플랫폼 홈, 실시간/주간 인기글, 인기 블로그, 최신글, 카테고리별/주제별 큐레이션, 통합 검색 결과, 태그 검색, 추천 피드, 구독/이웃 피드 화면을 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Frontend Developer Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-WEB-003`

## 6. 구현 범위 (Implementation Scope)
- 블로그 탐색 10개 전체 화면 구현, 무한 스크롤 커서 페이지네이션, 검색 자동완성 드롭다운

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `apps/user-web/src/pages/index.vue, apps/user-web/src/pages/explore/, apps/user-web/src/pages/tags/, apps/user-web/src/pages/feed/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `internal/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `posts, blogs, tags` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-REC-001~005, EP-SRC-001~004` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-EXP-001, SCR-EXP-002, SCR-EXP-003, SCR-EXP-004, SCR-EXP-005, SCR-EXP-006, SCR-EXP-007, SCR-EXP-008, SCR-EXP-009, SCR-EXP-010` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 구독 피드 0건일 때 인기 블로그 추천 Empty State 제공. 2) 검색어 하이라이팅 및 다중 패싯 필터 바 구현.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `apps/user-web/src/pages/index.vue, explore/trending.vue, search/index.vue, feed/following.vue`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-UX-001, TEST-UX-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 탐색 10개 화면 모두에서 Loading/Empty/Error 상태와 반응형 그리드가 완벽히 동작해야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 무한 스크롤 뒤로가기 시 스크롤 위치 유실 -> sessionStorage 커서/스크롤 복원 적용

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-WEB-004/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

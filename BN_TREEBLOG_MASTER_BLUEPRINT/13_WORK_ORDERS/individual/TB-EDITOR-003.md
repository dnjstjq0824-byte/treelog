# TB-EDITOR-003 — 블로그 글쓰기 에디터 스튜디오 UI(SCR-BLOG-013) 데스크톱 및 모바일 완벽 대응 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 3`

---

## 1. Work Order ID
- `TB-EDITOR-003`

## 2. 작업명 (Task Title)
- **블로그 글쓰기 에디터 스튜디오 UI(SCR-BLOG-013) 데스크톱 및 모바일 완벽 대응 구현**

## 3. 목적 (Purpose)
- 슬래시(/) 커맨드 메뉴, 플로팅 버블 툴바, 이미지 드래그앤드롭/크롭/alt 입력, 디바이스별 미리보기, 모바일 소프트 키보드 고정 툴바를 갖춘 에디터 화면을 완성한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Editor Engine Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-EDITOR-001, TB-EDITOR-002, TB-MEDIA-002`

## 6. 구현 범위 (Implementation Scope)
- SCR-BLOG-013 전체 컴포넌트, 단축키 시스템, 발행 설정 드로어(카테고리/태그/슬러그/SEO/공개범위/예약시간/CCL), 리비전 비교 드로어

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `apps/blog-console/src/pages/editor/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `internal/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `posts, post_revisions, autosave_drafts, media_assets` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-EDIT-001~006, EP-MED-001~005` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-BLOG-013` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 320px 모바일 뷰포트에서 가상 키보드 위 고정 툴바 동작. 2) 10,000단어 장문에서도 60fps 타이핑 반응성 유지.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `apps/blog-console/src/pages/editor/[postId].vue, components/SlashMenu.vue, PublishDrawer.vue`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-EDIT-001, TEST-FE-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 데스크톱과 모바일 뷰포트에서 제5조 35개 에디터 기능이 모두 정상 조작되어야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 모바일 Safari visualViewport 스크롤 점프 -> 고정 포지셔닝 보정 훅 적용

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-EDITOR-003/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

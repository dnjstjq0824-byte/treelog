# TB-WEB-003 — 개인 블로그 전체 사용자 화면(SCR-BLOG-001~012) Nuxt 3 SSR 및 반응형 UI 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 4`

---

## 1. Work Order ID
- `TB-WEB-003`

## 2. 작업명 (Task Title)
- **개인 블로그 전체 사용자 화면(SCR-BLOG-001~012) Nuxt 3 SSR 및 반응형 UI 구현**

## 3. 목적 (Purpose)
- 개별 블로그의 메인, 프로필, 전체글, 카테고리글, 게시글 상세(TOC/코드블록/라이트박스), 미디어뷰, 공지, 방명록, 소개, 이웃목록, 블로그내검색, 신고 화면을 SSR로 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Frontend Developer Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-POST-001, TB-THEME-001`

## 6. 구현 범위 (Implementation Scope)
- 개인 블로그 12개 화면 전체 구현, 호스트(서브도메인/개인도메인/경로) 기반 동적 블로그 컨텍스트 미들웨어

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `apps/user-web/src/pages/[blogHandle]/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `internal/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `blogs, posts, blog_categories, series, guestbook_entries` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-PUB-006, EP-PUB-007, EP-CAT-001, EP-CMT-005` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-BLOG-001, SCR-BLOG-002, SCR-BLOG-003, SCR-BLOG-004, SCR-BLOG-005, SCR-BLOG-006, SCR-BLOG-007, SCR-BLOG-008, SCR-BLOG-009, SCR-BLOG-010, SCR-BLOG-011, SCR-BLOG-012` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 비공개/상호이웃 전용 글 접근 시 권한별 안내 화면 및 신청 CTA 렌더링. 2) 본문 목차(TOC) 스크롤 스파이 구현.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `apps/user-web/src/pages/@[handle]/index.vue, posts/[slug].vue, guestbook.vue, about.vue`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-FE-001, TEST-FE-002, TEST-UX-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 12개 개인 블로그 화면이 320px~데스크톱에서 깨짐 없이 SSR 렌더링되어야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 커스텀 도메인 접속 시 라우트 매칭 혼선 -> Nuxt 서버 미들웨어에서 Host 헤더로 블로그 컨텍스트 선행 주입

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-WEB-003/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

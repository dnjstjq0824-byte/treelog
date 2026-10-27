# TB-BLOG-002 — 2뎁스 계층형 카테고리, 시리즈(연재물) 및 태그 사전 관리 API 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 2`

---

## 1. Work Order ID
- `TB-BLOG-002`

## 2. 작업명 (Task Title)
- **2뎁스 계층형 카테고리, 시리즈(연재물) 및 태그 사전 관리 API 구현**

## 3. 목적 (Purpose)
- 블로그별 2단계 카테고리 트리 생성·순서변경·삭제 시 게시글 안전 이동, 연재 시리즈 묶음, 정규화 해시태그 관리를 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Backend Developer Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-BLOG-001`

## 6. 구현 범위 (Implementation Scope)
- 카테고리 CRUD 및 드래그 순서 재정렬, 카테고리 삭제 시 미분류 자동 이전, 시리즈 CRUD

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/blog/, internal/modules/post/tags.go`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `blog_categories, series, tags, post_categories, post_tags` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-CAT-001, EP-CAT-002, EP-CAT-003, EP-CAT-004, EP-CAT-005` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-BLOG-004, SCR-MEM-006` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 3뎁스 이상 중첩 생성 시도 차단. 2) 카테고리별 공개 범위 및 목록 스타일(CARD/LIST/GALLERY/COMPACT) 저장.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/blog/category_service.go, internal/modules/blog/series_service.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-BLOG-003` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 글이 있는 카테고리 삭제 시 글이 삭제되지 않고 미분류 또는 지정 카테고리로 100% 이동되어야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 카테고리 순환 부모 참조 -> parent_id가 자기 자신이거나 하위 카테고리인지 검증

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-BLOG-002/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

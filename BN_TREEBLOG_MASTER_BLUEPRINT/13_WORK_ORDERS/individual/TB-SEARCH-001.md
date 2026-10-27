# TB-SEARCH-001 — OpenSearch 다국어 인덱스(Nori/ICU), Outbox 비동기 인덱서 워커 및 통합 검색 API 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 5`

---

## 1. Work Order ID
- `TB-SEARCH-001`

## 2. 작업명 (Task Title)
- **OpenSearch 다국어 인덱스(Nori/ICU), Outbox 비동기 인덱서 워커 및 통합 검색 API 구현**

## 3. 목적 (Purpose)
- OpenSearch 게시글/블로그/자동완성 인덱스를 구성하고, PostgreSQL Outbox 기반 실시간 색인/삭제 워커와 다중 필터 검색 API를 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Search & SEO Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-POST-001, TB-ARCH-002`

## 6. 구현 범위 (Implementation Scope)
- 인덱스 매핑/Alias 관리, 비동기 인덱서(비공개 전환 시 즉시 삭제), 자동완성/오타교정, 기간/언어/카테고리/블로그내 검색

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/search/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `search_index_jobs, search_queries, search_click_events, posts, blogs` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-SRC-001, EP-SRC-002, EP-SRC-003, EP-SRC-004, EP-SRC-005` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-EXP-007, SCR-EXP-008, SCR-BLOG-011` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 공개(PUBLIC) + 발행(PUBLISHED) + 정상 블로그(ACTIVE) 글만 색인 허용. 2) 무중단 재색인(Alias Switching) 커맨드 포함.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/search/indexer.go, internal/modules/search/query_builder.go, internal/modules/search/handler.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-SRC-001, TEST-SRC-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 발행 후 2초 내 검색 노출 및 비공개/삭제 전환 후 1초 내 검색 결과에서 100% 제거되어야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- OpenSearch 일시 장애 시 글 발행 실패 위험 -> 비동기 큐 분리로 발행 트랜잭션 100% 보호

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-SEARCH-001/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

# TB-TREESCAN-001 — TreeScan 검색 색인 동기화 어댑터(공개 콘텐츠 전용 증분 피드 & 삭제 툼스톤) 및 디스커버리 API 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 10`

---

## 1. Work Order ID
- `TB-TREESCAN-001`

## 2. 작업명 (Task Title)
- **TreeScan 검색 색인 동기화 어댑터(공개 콘텐츠 전용 증분 피드 & 삭제 툼스톤) 및 디스커버리 API 구현**

## 3. 목적 (Purpose)
- 제1조 및 제18조에 따라 TreeScan 기존 코드를 전혀 수정하지 않고, 향후 TreeScan이 TreeBlog의 공개 게시글을 실시간 색인·연동할 수 있는 어댑터와 증분 피드를 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Search & SEO Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-SEARCH-001, TB-SEC-001`

## 6. 구현 범위 (Implementation Scope)
- treescan_sync_state 커서 스트림, 비공개/삭제 전환 시 DELETE_TOMBSTONE 발행, 자매 플랫폼(getforTrip/Treewiki 등) 디스커버리 API

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/treescan/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `treescan_sync_state, posts, blogs, outbox_events` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-TSC-001, EP-TSC-002, EP-TSC-004` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `N/A` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) visibility == 'PUBLIC' 게시글만 피드에 포함하며 비공개/이웃공개/상호이웃공개 글은 단 1바이트도 외부로 노출하지 않음. 2) 검색 순위 인위적 우대 조작 필드 배제.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/treescan/sync_feed.go, internal/modules/treescan/discovery.go, internal/modules/treescan/adapter.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-ZTR-001, TEST-ZTR-002, TEST-ZTR-003, TEST-ZTR-004` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- PUBLIC 외 게시글 노출 0건 보장 및 공개->비공개 전환 시 DELETE_TOMBSTONE이 즉시 출력되어야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 연동 비활성 상태에서 불필요한 리소스 소모 -> FF_TREESCAN_INTEGRATION_ENABLED=false 기본값으로 경량 동작

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-TREESCAN-001/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

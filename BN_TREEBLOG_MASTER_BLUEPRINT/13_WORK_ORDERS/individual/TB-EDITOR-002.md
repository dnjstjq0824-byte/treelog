# TB-EDITOR-002 — 듀얼 레이어 자동저장(IndexedDB + 서버), 버전 충돌 해결 및 리비전 Diff 복원 엔진 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 3`

---

## 1. Work Order ID
- `TB-EDITOR-002`

## 2. 작업명 (Task Title)
- **듀얼 레이어 자동저장(IndexedDB + 서버), 버전 충돌 해결 및 리비전 Diff 복원 엔진 구현**

## 3. 목적 (Purpose)
- 3초 로컬 IndexedDB 스냅샷과 15초 서버 자동저장, 오프라인 작성 복구, 멀티 디바이스 버전 충돌 감지 모달 및 비파괴 리비전 복원을 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Editor Engine Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-EDITOR-001, TB-DB-002`

## 6. 구현 범위 (Implementation Scope)
- useAutosave 훅, Optimistic Lock(base_revision_no) 검증, 문단 Diff 계산기, 과거 리비전 복원 API

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `packages/editor-core/src/autosave/, internal/modules/editor/autosave.go, revisions.go`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/admin-console/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `posts, post_revisions, autosave_drafts` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-EDIT-002, EP-EDIT-003, EP-EDIT-004` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-BLOG-013` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 과거 버전 복원 시 기존 이력을 지우지 않고 새 최신 revision_no로 스냅샷을 추가 생성.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `packages/editor-core/src/autosave/useAutosave.ts, internal/modules/editor/revisions.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-EDIT-002, TEST-EDIT-003` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 네트워크 단절 상태에서 작성 후 새로고침 시 로컬 스냅샷으로 100% 복구되고, 버전 충돌 시 409 응답 및 비교 모달이 동작해야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- IndexedDB 용량 초과 -> 블로그당 최근 50개 로컬 스냅샷 LRU 정리

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-EDITOR-002/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

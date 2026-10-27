# TB-ADMIN-002 — 통합 관리자 콘솔 프론트엔드(apps/admin-console) 회원·블로그·게시글·커뮤니티·이의신청 UI 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 7`

---

## 1. Work Order ID
- `TB-ADMIN-002`

## 2. 작업명 (Task Title)
- **통합 관리자 콘솔 프론트엔드(apps/admin-console) 회원·블로그·게시글·커뮤니티·이의신청 UI 구현**

## 3. 목적 (Purpose)
- 디자인만 있는 더미 대시보드가 아닌, 실제 백엔드 API와 연결되어 회원·블로그·게시글 제재 및 신고/이의신청 심사를 수행하는 관리자 콘솔 화면을 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Admin Console Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-ADMIN-001, TB-WEB-001`

## 6. 구현 범위 (Implementation Scope)
- SCR-ADM-001 ~ SCR-ADM-005 5개 관리자 화면, 원문/신고 스냅샷 비교 뷰어, 일괄 스팸 댓글 삭제 그리드

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `apps/admin-console/src/pages/dashboard/, users/, blogs/, posts/, community/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `internal/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `users, blogs, posts, moderation_cases, appeals` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-ADM-001~005` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-ADM-001, SCR-ADM-002, SCR-ADM-003, SCR-ADM-004, SCR-ADM-005` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 제재 처분 시 정책 위반 조항 선택 및 사유 입력 필수 검증.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `apps/admin-console/src/pages/index.vue, users/index.vue, moderation/cases.vue, moderation/appeals.vue`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-ADM-001, TEST-ADM-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 관리자 콘솔 UI에서 신고 케이스 심사 및 이의신청 복원 처리를 클릭하여 실제 게시글 상태가 변경되어야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 실수 클릭으로 인한 대량 정지 -> 2단계 확인 모달 및 대상 수 명시

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-ADMIN-002/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

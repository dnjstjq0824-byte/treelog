# TB-BLOG-001 — 블로그 개설 엔진, 핸들 예약어 검증, 팀 멤버 RBAC 및 운영 설정 API 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 2`

---

## 1. Work Order ID
- `TB-BLOG-001`

## 2. 작업명 (Task Title)
- **블로그 개설 엔진, 핸들 예약어 검증, 팀 멤버 RBAC 및 운영 설정 API 구현**

## 3. 목적 (Purpose)
- 개인·전문·기업·커뮤니티 블로그 개설, 150개 예약어 차단, 5단계 팀 블로그 멤버 초대/권한 관리, 방명록 및 블로그 폐쇄(30일 유예)를 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Backend Developer Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-AUTH-002, TB-DB-002`

## 6. 구현 범위 (Implementation Scope)
- 원자적 블로그 개설 트랜잭션, 핸들 검증기, 팀 멤버 초대 토큰, 설정 캐싱(Redis), 방명록 CRUD

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/blog/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `blogs, blog_members, blog_settings, blog_navigation, guestbook_entries` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-BCRT-001, EP-BCRT-002, EP-BCRT-003, EP-BMGT-001, EP-BMGT-002, EP-BMGT-003, EP-BMGT-004, EP-BMGT-005, EP-BMGT-006, EP-BMGT-007, EP-CMT-005` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-MEM-005, SCR-MEM-006, SCR-BLOG-008` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 블로그 개설 시 기본 설정, 기본 서브도메인, 기본 테마, 미분류 카테고리를 단일 트랜잭션으로 생성. 2) 방명록 비밀글 서버 마스킹 구현.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/blog/service.go, internal/modules/blog/handle_validator.go, internal/modules/blog/members.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-BLOG-001, TEST-BLOG-002, TEST-BLOG-005` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 예약어 차단 및 정상 블로그 개설이 1초 내 완료되고 팀 멤버 역할별 권한 제어가 동작해야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 동시 동일 핸들 개설 요청 -> CITEXT UNIQUE 제약으로 409 안전 반환

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-BLOG-001/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

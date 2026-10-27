# TB-ADMIN-001 — 플랫폼 관리자 백엔드 API(회원·블로그·게시글·댓글·검색·시스템·감사로그) 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 7`

---

## 1. Work Order ID
- `TB-ADMIN-001`

## 2. 작업명 (Task Title)
- **플랫폼 관리자 백엔드 API(회원·블로그·게시글·댓글·검색·시스템·감사로그) 구현**

## 3. 목적 (Purpose)
- 제11조의 관리자 통합 콘솔을 구동하는 백엔드 API 전체를 실제 DB/OpenSearch/Redis와 직결하여 구현하고 모든 관리 행위를 불변 감사 로그에 기록한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Admin Console Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-SEC-002, TB-SEARCH-001`

## 6. 구현 범위 (Implementation Scope)
- 회원 정지/권한관리, 블로그 정지/도메인해제, 검색 재색인/스팸어차단, Feature Flag 실시간 토글, 감사로그 해시체인 검증 조회

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/admin/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/user-web/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `users, blogs, posts, comments, search_index_jobs, search_queries, feature_flags, system_settings, audit_logs, background_jobs` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-ADM-001, EP-ADM-002, EP-ADM-003, EP-ADM-006, EP-ADM-008` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-ADM-001, SCR-ADM-002, SCR-ADM-003, SCR-ADM-006, SCR-ADM-008` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 모든 관리자 변경 API는 MFA 인증된 세션 및 세부 역할(SUPPORT_AGENT, TRUST_SAFETY_MOD, SUPER_ADMIN)을 검증하고 before/after JSON Diff를 audit_logs에 기록.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/admin/user_admin.go, blog_admin.go, search_admin.go, system_admin.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-ADM-001, TEST-ADM-003` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 비관리자 접근 시 100% 403 차단 및 관리자 조치 실행 시 DB 상태와 audit_logs가 즉시 갱신되어야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 무거운 어드민 통계 쿼리로 사용자 DB 지연 -> Read Replica 전용 커넥션 사용

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-ADMIN-001/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

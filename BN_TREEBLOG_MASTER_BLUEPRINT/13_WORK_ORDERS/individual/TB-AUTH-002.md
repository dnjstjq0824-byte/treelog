# TB-AUTH-002 — 소셜 OAuth/OIDC 어댑터, TOTP MFA 및 2계층 RBAC 인가 미들웨어 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 1`

---

## 1. Work Order ID
- `TB-AUTH-002`

## 2. 작업명 (Task Title)
- **소셜 OAuth/OIDC 어댑터, TOTP MFA 및 2계층 RBAC 인가 미들웨어 구현**

## 3. 목적 (Purpose)
- 소셜 로그인(OIDC PKCE), TOTP 2단계 인증, 기기 세션 관리, 계정 탈퇴 및 플랫폼/블로그 2계층 RBAC 미들웨어를 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Backend Developer Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-AUTH-001`

## 6. 구현 범위 (Implementation Scope)
- OAuth2/OIDC 클라이언트, RFC 6238 TOTP 및 백업코드, RequirePlatformRole / RequireBlogRole 미들웨어

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/identity/, internal/modules/user/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `users, identities, roles, permissions, role_permissions, user_roles, blog_members` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-AUTH-008, EP-AUTH-009, EP-USER-001, EP-USER-002, EP-USER-003, EP-USER-004, EP-USER-005, EP-USER-006` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-MEM-004, SCR-MEM-015` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) identities 테이블을 분리하여 향후 TreeScan OIDC 연동 시 기존 블로그 소유권이 100% 유지되도록 구현. 2) 단독 소유 블로그 보유 시 탈퇴 전 양도/폐쇄 선택 강제.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/identity/mfa.go, internal/modules/identity/rbac.go, internal/modules/user/service.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-USR-004, TEST-USR-005, TEST-USR-006, TEST-IAM-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- RBAC 권한 부족 시 100% 403 반환 및 MFA 미인증 관리자/출금 요청 차단 검증
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 탈퇴 후 고아 블로그 발생 -> 탈퇴 전 소유 블로그 체크 트랜잭션 강제

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-AUTH-002/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

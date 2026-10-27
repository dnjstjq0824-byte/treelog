# TB-AUTH-001 — 독립 이메일 회원가입·로그인·Argon2id·JWT Refresh Token Rotation 엔진 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 1`

---

## 1. Work Order ID
- `TB-AUTH-001`

## 2. 작업명 (Task Title)
- **독립 이메일 회원가입·로그인·Argon2id·JWT Refresh Token Rotation 엔진 구현**

## 3. 목적 (Purpose)
- TreeScan 없이 단독으로 동작하는 안전한 회원 인증 및 세션 회전(Refresh Token Rotation) 백엔드를 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Backend Developer Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-DB-002`

## 6. 구현 범위 (Implementation Scope)
- 가입/이메일인증, Argon2id 해싱, RS256 JWT 발급, Refresh Token 재사용 탈취 감지 시 패밀리 전체 폐기, 비밀번호 재설정

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/identity/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `users, user_profiles, identities, sessions, user_consents` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-AUTH-001, EP-AUTH-002, EP-AUTH-003, EP-AUTH-004, EP-AUTH-005, EP-AUTH-006, EP-AUTH-007` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-MEM-001, SCR-MEM-002, SCR-MEM-003` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 로그인 5회 실패 시 지연, 10회 실패 시 15분 계정 잠금(423). 2) 약관 동의 버전 기록.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/identity/service.go, internal/modules/identity/handler.go, internal/modules/identity/jwt.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-USR-001, TEST-USR-002, TEST-USR-003, TEST-IAM-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 가입->인증->로그인->토큰갱신->재사용차단->로그아웃 전 흐름이 통합 테스트를 통과해야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 토큰 동시 갱신 레이스 컨디션 -> 5초 Grace Window 허용으로 해결

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-AUTH-001/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

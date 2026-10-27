# TB-TREESCAN-002 — Booking Network / TreeScan OIDC 통합 SSO 계정 연결 및 데이터 소유권 무손실 병합 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 10`

---

## 1. Work Order ID
- `TB-TREESCAN-002`

## 2. 작업명 (Task Title)
- **Booking Network / TreeScan OIDC 통합 SSO 계정 연결 및 데이터 소유권 무손실 병합 구현**

## 3. 목적 (Purpose)
- 향후 Booking Network / TreeScan 통합 인증(OIDC) 연동 시 기존 TreeBlog 사용자의 블로그·게시글·이웃·정산 원장 소유권이 손실 없이 100% 보존되는 계정 연결 모듈을 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Backend Developer Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-AUTH-002, TB-TREESCAN-001`

## 6. 구현 범위 (Implementation Scope)
- TREESCAN_OIDC 프로바이더 토큰 검증, 기존 users.id에 identities 레코드 안전 매핑, 프로필 동기화 명시적 동의(TREESCAN_PROFILE_SYNC) 확인

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/treescan/sso.go, apps/user-web/src/pages/settings/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/admin-console/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `users, identities, user_consents, blogs` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-TSC-003` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-MEM-004` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 계정 연동 전후 blogs.owner_user_id 및 ledger_entries.owner_user_id 불변 보장.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/treescan/sso_linker.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-IAM-002, TEST-ZTR-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- TreeScan OIDC 연결 후에도 기존 개설 블로그와 작성 글, 정산 잔액이 100% 동일하게 유지되어야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 이메일 충돌 계정 병합 악용 -> 반드시 양쪽 인증(기존 세션 로그인 + OIDC 검증) 통과 시에만 병합 허용

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-TREESCAN-002/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

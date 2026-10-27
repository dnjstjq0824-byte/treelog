# TB-API-001 — OpenAPI 3.1.0 스키마 검증기 및 프론트엔드 타입 안전 API 클라이언트 생성 파이프라인

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 1`

---

## 1. Work Order ID
- `TB-API-001`

## 2. 작업명 (Task Title)
- **OpenAPI 3.1.0 스키마 검증기 및 프론트엔드 타입 안전 API 클라이언트 생성 파이프라인**

## 3. 목적 (Purpose)
- 04_API/openapi.yaml로부터 TypeScript 클라이언트(packages/api-client)를 자동 생성하고 백엔드 요청/응답 스키마 검증기를 연결한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Lead Architect Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-ARCH-001, TB-AUTH-001`

## 6. 구현 범위 (Implementation Scope)
- OpenAPI TypeScript 코드젠, 401 자동 토큰 리프레시 인터셉터, 요청 ID/CSRF 헤더 자동 주입

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `04_API/, packages/api-client/, internal/platform/validator/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `03_DATABASE/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `N/A` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `All 24 API Groups` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `N/A` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 모든 API 오류를 StandardErrorResponse 타입으로 파싱. 2) Idempotency-Key 자동 UUID 생성 헬퍼 내장.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `packages/api-client/src/index.ts, packages/api-client/src/client.ts`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-FE-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- openapi.yaml 변경 시 단일 명령으로 타입 에러 없는 클라이언트가 생성되어야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 수동 수정으로 인한 스키마 드리프트 -> CI에서 코드젠 diff 검사 강제

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-API-001/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

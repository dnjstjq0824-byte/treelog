# TB-ARCH-001 — 모노레포 디렉터리 스캐폴딩 및 Go 모듈형 모놀리스 코어 프레임워크 구축

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 0`

---

## 1. Work Order ID
- `TB-ARCH-001`

## 2. 작업명 (Task Title)
- **모노레포 디렉터리 스캐폴딩 및 Go 모듈형 모놀리스 코어 프레임워크 구축**

## 3. 목적 (Purpose)
- TreeBlog 전체 백엔드(Go 1.22+) 및 프론트엔드(Nuxt 3/Vue 3 모노레포)의 표준 디렉터리 구조와 공통 서버 부트스트랩을 생성한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Lead Architect Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `NONE`

## 6. 구현 범위 (Implementation Scope)
- Chi 라우터, Graceful Shutdown, 환경변수 로더, 공통 응답/에러 봉투, 모듈 인터페이스 레지스트리 구축

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `cmd/, internal/platform/, packages/, pnpm-workspace.yaml, go.mod`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `03_DATABASE/migrations/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `system_settings, feature_flags` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `GET /api/v1/health/live, GET /api/v1/health/ready` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `N/A` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) TreeScan 의존성 없이 단독 실행 가능한 설정 구조체 정의. 2) 15개 도메인 모듈 디렉터리 생성 및 순환 참조 방지 린트 규칙 설정.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `cmd/api/main.go, cmd/worker/main.go, internal/platform/config/config.go, internal/platform/httpserver/server.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-ARCH-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- go build ./... 및 pnpm build가 에러 없이 성공하고 /api/v1/health/live가 200 OK를 반환해야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 초기 패키지 경로 불일치 위험 -> go.mod 모듈명 고정으로 완화

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-ARCH-001/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

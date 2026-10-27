# TB-BLOG-003 — 멀티 도메인 라우팅(서브도메인/경로/개인 도메인), DNS 검증 및 자동 ACME TLS 엔진 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 2`

---

## 1. Work Order ID
- `TB-BLOG-003`

## 2. 작업명 (Task Title)
- **멀티 도메인 라우팅(서브도메인/경로/개인 도메인), DNS 검증 및 자동 ACME TLS 엔진 구현**

## 3. 목적 (Purpose)
- 서브도메인, 경로형, 사용자 개인 커스텀 도메인 연결 및 DNS TXT 소유권 검증, Let's Encrypt 자동 TLS 발급, 대표 도메인 301 리다이렉트를 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Backend Developer Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-BLOG-001, TB-SEC-001`

## 6. 구현 범위 (Implementation Scope)
- DNS TXT/CNAME/A 검증기, Caddy On-Demand TLS 내부 검증 엔드포인트(/internal/v1/domains/ask), 호스트->blogId 고속 캐시 리졸버

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/domain/, infra/caddy/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `blog_domains, blogs` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-DOM-001, EP-DOM-002, EP-DOM-003, EP-DOM-004, EP-DOM-005` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-MEM-006` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) _treeblog-verify TXT 레코드 일치 확인 전까지 TLS 발급 및 라우팅 절대 비활성화(도메인 탈취 방지). 2) 6시간 주기 백그라운드 DNS 재검증 워커 구현.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/domain/service.go, internal/modules/domain/dns_verifier.go, infra/caddy/Caddyfile`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-DOM-001, TEST-DOM-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 검증된 개인 도메인 접속 시 200 OK 렌더링 및 비대표 주소 접속 시 대표 주소로 301 리다이렉트 동작 확인
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 미검증 도메인 대량 TLS 요청 공격 -> /internal/v1/domains/ask에서 DB VERIFIED 상태만 허용

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-BLOG-003/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

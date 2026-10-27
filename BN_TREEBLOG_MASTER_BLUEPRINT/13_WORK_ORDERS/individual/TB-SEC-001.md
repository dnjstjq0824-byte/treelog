# TB-SEC-001 — 공통 보안 미들웨어(CSRF, Rate Limiter, SafeDialer SSRF 방어, 보안 헤더) 구축

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 1`

---

## 1. Work Order ID
- `TB-SEC-001`

## 2. 작업명 (Task Title)
- **공통 보안 미들웨어(CSRF, Rate Limiter, SafeDialer SSRF 방어, 보안 헤더) 구축**

## 3. 목적 (Purpose)
- 전 API에 적용되는 CSRF 검증, Redis 슬라이딩 윈도우 Rate Limiter, SSRF 차단 HTTP 클라이언트, 해시 체인 감사 로거를 구축한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Security Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-ARCH-001, TB-AUTH-001`

## 6. 구현 범위 (Implementation Scope)
- Double-Submit CSRF, IP/User Rate Limiter, SafeDialer(사설/루프백/메타데이터 IP 차단), 불변 AuditLogger

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/pkg/security/, internal/pkg/safenet/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `audit_logs` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `N/A (All Endpoints Middleware)` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `N/A` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) DNS Rebinding 방어를 위해 net.Dialer의 Control 훅에서 리졸브된 최종 IP 검사. 2) audit_logs 기록 시 이전 해시(prev_hash)와 결합한 SHA-256 entry_hash 계산.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/pkg/safenet/dialer.go, internal/pkg/security/csrf.go, internal/pkg/security/ratelimit.go, internal/pkg/security/audit.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-SEC-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 127.0.0.1 및 169.254.169.254 연결 시도 시 소켓 생성 전 100% 차단되어야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 프록시 환경 클라이언트 IP 오인 -> 신뢰된 서브넷의 X-Forwarded-For만 파싱

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-SEC-001/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

# TB-BILLING-002 — 불변 복식부기 회계 원장(ledger_entries), 정산금 출금(Payout), 지급 보류 및 부정수익 환수 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 8`

---

## 1. Work Order ID
- `TB-BILLING-002`

## 2. 작업명 (Task Title)
- **불변 복식부기 회계 원장(ledger_entries), 정산금 출금(Payout), 지급 보류 및 부정수익 환수 구현**

## 3. 목적 (Purpose)
- 광고 수익, 멤버십, 후원금, 수수료, 원천징수 세액, 환불, 부정수익 환수(Clawback)를 기록하는 불변 복식부기 원장과 MFA 기반 출금 신청/승인 시스템을 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Backend Developer Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-ADS-001, TB-BILLING-001`

## 6. 구현 범위 (Implementation Scope)
- Double-Entry Ledger 트랜잭션 엔진(SUM(DEBIT)==SUM(CREDIT) 강제), 정산 신청(계좌 AES-256-GCM 암호화), 블로거 수익 센터 UI

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/monetization/ledger.go, payout.go, apps/blog-console/src/pages/studio/[blogId]/monetization/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/user-web/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `ledger_entries, payouts` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-PAY-005, EP-PAY-006, EP-ADM-007` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-MEM-006, SCR-ADM-007` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) ledger_entries는 절대 UPDATE/DELETE 하지 않으며 환불/환수는 역분개 전표(Reversal Entry) INSERT로만 처리.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/monetization/ledger_service.go, internal/modules/monetization/payout_service.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-MON-001, TEST-ADM-004` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 결제·정산·환불·환수 전 시나리오에서 원장 전체 차변 합계와 대변 합계가 1원 오차 없이 일치해야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 부동소수점 반올림 오차 -> 모든 금액을 정수(amount_minor, BIGINT)로만 연산

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-BILLING-002/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

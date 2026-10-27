# TB-BILLING-001 — 유료 콘텐츠(서버 사이드 Paywall), 유료 멤버십 구독 및 독자 후원(Tipping) 결제 엔진 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 8`

---

## 1. Work Order ID
- `TB-BILLING-001`

## 2. 작업명 (Task Title)
- **유료 콘텐츠(서버 사이드 Paywall), 유료 멤버십 구독 및 독자 후원(Tipping) 결제 엔진 구현**

## 3. 목적 (Purpose)
- 제12.3조 및 제12.4조의 유료 멤버십, 단건 유료글 구매, 독자 후원 주문/결제 어댑터와 미결제 시 서버에서 유료 본문을 절삭하는 Paywall을 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Backend Developer Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-POST-001, TB-DB-002`

## 6. 구현 범위 (Implementation Scope)
- 멤버십 플랜 관리, 주문 생성(Idempotency-Key 강제), PG 웹훅 HMAC 검증, Paywall AST/HTML 서버 절삭기, 청약철회 환불

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/monetization/membership.go, orders.go, paywall.go`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/admin-console/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `memberships, subscriptions, orders, payments, posts` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-PAY-001, EP-PAY-002, EP-PAY-003, EP-PAY-004` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-BLOG-005, SCR-MEM-006, SCR-MEM-011` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) PAID_MEMBERS_ONLY 게시글을 비구매자가 조회 시 paywallDivider 이후 본문 데이터를 서버에서 완전히 제거한 후 티저만 반환.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/monetization/paywall.go, internal/modules/monetization/order_service.go, payment_webhook.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-MON-001, TEST-MON-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 미결제 응답 바디에 유료 구간 텍스트가 0바이트이고 결제 완료 즉시 전문이 열람되어야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 웹훅 중복 수신으로 이중 적립 -> pg_transaction_id UNIQUE 및 idempotency_key로 차단

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-BILLING-001/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

# M_MONETIZATION — Monetization (광고·수익배분·유료콘텐츠·후원·정산 엔진)
> **부제**: 광고, 광고 수익 배분, 후원, 유료 콘텐츠, 멤버십, 구독, 결제, 수익 통계, 정산 시스템
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-M_MONETIZATION`
> **적용 조항**: 제3조 (M_MONETIZATION 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- 크리에이터와 전문 블로거가 TreeBlog에서 양질의 콘텐츠를 지속적으로 생산할 수 있는 강력한 경제적 동기를 제공하기 위해 **4대 수익화 파이프라인(12.1 자체 광고 네트워크, 12.2 블로거 광고 수익 배분, 12.3 유료 콘텐츠/멤버십, 12.4 독자 후원)**과 **불변 복식부기 정산 원장(`ledger_entries`)**을 구축한다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **광고주 (`ADVERTISER`)**: 광고 콘솔에서 캠페인 예산, 입찰 단가(CPM/CPC), 타기팅(카테고리/키워드/언어/디바이스)을 설정하고 배너/네이티브 소재를 등록하여 심사를 받는다.
- **크리에이터 (`BLOG_OWNER`)**: 블로그 내 광고 슬롯(본문 상단/중간/하단, 사이드바)을 활성화하여 일별 광고 수익 배분액을 확인하고, 심층 시리즈를 유료 멤버십 전용(`PAID_MEMBERS_ONLY`)으로 발행하거나 독자 후원을 받아 월별 정산금을 출금한다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-MON-001 (12.1 광고 시스템)**:
  - 광고주 가입, 캠페인/일예산/총예산 설정, 소재 등록(이미지/텍스트, 랜딩 URL), 관리자 사전 심사(`PENDING_REVIEW` → `APPROVED` / `REJECTED`), 카테고리·언어·문맥 타기팅 매칭, 노출(`IMPRESSION`) 및 클릭(`CLICK`) 이벤트 수집, 부정 클릭/봇 트래픽 필터링(Invalid Traffic Detection).
- **REQ-MON-002 (12.2 블로거 광고 수익 배분)**:
  - 유효 광고 노출/클릭 수익을 집계하여 플랫폼 운영 수수료(예: 기본 30%, 설정 가능)를 제외한 **크리에이터 기여 수익(예: 70%)**을 일 단위로 정산 원장에 기록한다.
  - 어뷰징(자가 클릭, 봇 유도) 적발 시 지급 보류(`ON_HOLD`) 및 환수/차감 정정(Clawback Reversal Entry) 전표를 생성한다.
- **REQ-MON-003 (12.3 유료 콘텐츠 및 멤버십)**:
  - 블로거가 월간/연간 유료 멤버십 티어(`memberships`)를 개설하거나 개별 게시글 단건 유료 구매를 설정할 수 있다.
  - 유료 게시글은 서버 측에서 무료 미리보기 구간(`paywall_divider` 블록 이전)까지만 내려주고 이후 본문은 구매 권한 확인 후에만 전송한다.
  - 구독 해지(만료일까지 권한 유지) 및 청약철회(열람 이력 없는 결제건 환불 정책) 지원.
- **REQ-MON-004 (12.4 독자 후원 — Tipping)**:
  - 독자가 응원 메시지와 함께 일회성 후원금을 결제하면 수수료 차감 후 블로거 정산 가능 잔액으로 적립되며, 이상 거래(도난 카드 등) 탐지 시 정산 대기 기간(예: D+7 Escrow)을 거쳐 확정된다.
- **REQ-MON-005 (투명한 광고 표시 및 단계적 출시 — 제12조 준수)**:
  - 모든 광고 슬롯과 유료/협찬 콘텐츠에는 이용자가 명확히 식별할 수 있는 `광고(Ad)` / `유료 멤버십 전용` 배지를 강제 표시한다.
  - 초기(Phase 1~7)에는 무료 블로그 발행을 우선 완성하고, 수익화 모듈은 Phase 8에서 독립 활성화할 수 있도록 설계한다.

## 4. 데이터 구조 (Data Structures & Domain Models)
```go
// 복식부기 원장 엔트리 (모든 금액은 통화 최소 단위 정수 minor units 또는 NUMERIC(18,4) 사용, 부동소수점 float 금지)
type LedgerEntry struct {
    ID             uuid.UUID       `json:"id"`
    TransactionID  uuid.UUID       `json:"transaction_id"`  // 차변/대변 묶음 키
    IdempotencyKey string          `json:"idempotency_key"`
    AccountType    string          `json:"account_type"`    // PLATFORM_CASH, ADVERTISER_PREPAID, CREATOR_PENDING, CREATOR_PAYABLE, PLATFORM_FEE_REVENUE
    OwnerUserID    *uuid.UUID      `json:"owner_user_id,omitempty"`
    BlogID         *uuid.UUID      `json:"blog_id,omitempty"`
    Direction      string          `json:"direction"`       // DEBIT, CREDIT
    AmountMinor    int64           `json:"amount_minor"`    // KRW 원 단위 또는 USD 센트 단위
    Currency       string          `json:"currency"`        // "KRW", "USD"
    ReferenceType  string          `json:"reference_type"`  // "AD_SETTLEMENT", "MEMBERSHIP_ORDER", "DONATION", "PAYOUT", "REFUND", "CLAWBACK"
    ReferenceID    uuid.UUID       `json:"reference_id"`
    CreatedAt      time.Time       `json:"created_at"`
}
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- 블로거 수익 센터(`apps/blog-console`): 광고 슬롯 배치 토글, 수익 통계 차트, 멤버십 플랜 설정, 정산 내역 및 출금 신청 UI.
- 광고주 콘솔 및 플랫폼 관리자 광고 심사/정산 관리 화면(`SCR-ADM-006`).

## 6. 백엔드 로직 (Backend Logic & Processing)
- `internal/modules/monetization`:
  - Ad Server Decision Engine (비공개/성인/제재 글에는 광고 송출 자동 제외 — Brand Safety).
  - Invalid Traffic(IVT) 필터: 동일 IP/디바이스 지문의 반복 클릭, 비정상적 짧은 체류 시간 클릭 무효화.
  - Double-Entry Ledger Service: 모든 거래에서 `SUM(DEBIT) == SUM(CREDIT)` 불변식을 트랜잭션 내에서 검증.

## 7. API 계약 (API Contracts)
- `POST /api/v1/ads/accounts` / `POST /api/v1/ads/campaigns` / `POST /api/v1/ads/creatives`
- `GET /api/v1/public/ads/serve?blogId={id}&slot={position}` (문맥 타기팅 광고 호출)
- `POST /api/v1/public/ads/events` (노출/클릭 비콘 수집)
- `GET /api/v1/blogs/{blogId}/monetization/summary` / `GET /api/v1/blogs/{blogId}/monetization/ledger`
- `POST /api/v1/blogs/{blogId}/memberships` / `POST /api/v1/blogs/{blogId}/posts/{postId}/purchase`
- `POST /api/v1/blogs/{blogId}/donate` (독자 후원 결제 생성)
- `POST /api/v1/me/payouts/request` (정산금 지급 신청)

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- 주요 테이블: `ad_accounts`, `ad_campaigns`, `ad_creatives`, `ad_placements`, `ad_events` (월별 파티셔닝), `memberships`, `user_subscriptions`, `orders`, `payments`, `payouts`, `ledger_entries`.

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **환불 및 지급 보류 시 원장 처리**: 이미 기록된 `ledger_entries` 행은 절대 `UPDATE`하거나 `DELETE`하지 않는다. 환불이나 부정 수익 환수 발생 시 반드시 반대 방향(`Direction` 역전)의 **조정 전표(Reversal Entry)**를 새로 INSERT하여 회계 감사 추적성을 100% 보존한다.

## 10. UX/UI 구조 (UX/UI Structure)
- 본문 가독성을 해치는 전면 팝업 광고나 본문을 가리는 플로팅 광고를 금지하고, CLS(레이아웃 밀림)가 발생하지 않도록 광고 슬롯 컨테이너의 최소 높이를 사전 예약한다.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- 카드 번호나 계좌 원본 등 민감 결제 정보(PCI-DSS 대상)는 TreeBlog DB에 직접 저장하지 않고 PG사 토큰(`pg_customer_token`, `pg_billing_key`)만 저장한다.
- 유료 콘텐츠 Paywall 우회 차단: 클라이언트 CSS(`display: none`이나 블러 효과)로만 본문을 가리는 방식을 절대 금지하고, 서버 API에서 미결제 사용자에게는 유료 구간 AST/HTML 자체를 내려주지 않는다.

## 12. 성능 및 확장성 (Performance & Scalability)
- 광고 노출/클릭 비콘(`POST /api/v1/public/ads/events`)은 Redis Streams 버퍼에 비동기 적재 후 워커가 5초 단위 마이크로 배치로 PostgreSQL 파티션 테이블에 Bulk Insert하여 초당 수만 건의 이벤트를 지연 없이 수용한다.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- `github.com/shopspring/decimal` (정밀 통화/비율 연산), PG 웹훅 서명 검증(HMAC-SHA256).

## 14. 개발 우선순위 (Development Priority)
- **Phase 8 (통계·광고·수익화 단계)** (`TB-ADS-001`, `TB-BILLING-001`, `TB-BILLING-002`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 선행: `I_IDENTITY_AUTHORIZATION`, `B_BLOGGING_EXPERIENCE`, `Y_ANALYTICS`.
- 후행: `W_WORKFLOW_ADMIN` (광고 심사 및 정산 승인).

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-MON-001`: 10,000원 멤버십 결제 → 플랫폼 수수료 1,000원 + 크리에이터 정산 대기금 9,000원 분배 → 환불 처리 시 `ledger_entries`의 전체 차변 합계와 대변 합계가 정확히 0원으로 일치하는지 검증.
- `TEST-MON-002`: 비구독자가 `PAID_MEMBERS_ONLY` 게시글 조회 시 페이월 이후 유료 본문이 HTTP 응답 바디에 1바이트도 포함되지 않는지 검증.

## 17. Acceptance Criteria (인수 기준)
1. 광고 캠페인 집행, 수익 배분, 유료 콘텐츠 페이월, 독자 후원, 복식부기 정산 원장이 데이터 불일치 없이 동작해야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- 복식부기 원장 대사(Reconciliation) 단위/통합 테스트 로그 및 페이월 서버 절삭 검증 응답 로그.

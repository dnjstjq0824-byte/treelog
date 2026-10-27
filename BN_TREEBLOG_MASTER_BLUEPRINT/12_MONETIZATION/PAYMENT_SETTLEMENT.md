# PAYMENT_SETTLEMENT.md — 복식부기 회계 원장(`ledger_entries`) 및 정산 지급 명세

## 1. 불변 복식부기 원장 (Immutable Double-Entry Ledger) 규칙
모든 금전 흐름은 반드시 동일한 `transaction_group_id` 아래 차변(`DEBIT`) 합계와 대변(`CREDIT`) 합계가 정확히 일치하는 2개 이상의 `ledger_entries` 행으로 기장된다.

### 거래 시나리오별 분개(Journal Entry) 규칙표
| 거래 이벤트 (`reference_type`) | 차변 (`DEBIT`, `+`자산/`-`부채) | 대변 (`CREDIT`, `-`자산/`+`부채·수익) | 설명 |
| :--- | :--- | :--- | :--- |
| **1. 유료글/멤버십/후원 결제 (`ORDER_PAYMENT`, 10,000원)** | `PLATFORM_CLEARING` (10,000) | `CREATOR_PENDING_ESCROW` (9,000)<br>`PLATFORM_FEE_REVENUE` (1,000) | 결제 대금 유입 후 에스크로 대기(7일) 및 수수료 10% 분리 |
| **2. 에스크로 기간 경과 확정** | `CREATOR_PENDING_ESCROW` (9,000) | `CREATOR_PAYABLE` (9,000) | 환불/분쟁 없는 건을 출금 가능 잔액으로 전환 |
| **3. 결제 환불 (`ORDER_REFUND`, 10,000원)** | `CREATOR_PENDING_ESCROW` (9,000)<br>`PLATFORM_FEE_REVENUE` (1,000) | `PLATFORM_CLEARING` (10,000) | 기존 행 수정 없이 반대 방향 역분개 전표 INSERT |
| **4. 블로거 정산금 지급 (`PAYOUT_SETTLEMENT`, 100,000원)** | `CREATOR_PAYABLE` (100,000) | `TAX_WITHHOLDING` (3,300)<br>`PAYOUT_DISBURSED` (96,700) | 원천징수 세액(예: 3.3%) 공제 후 실지급액 송금 완료 |

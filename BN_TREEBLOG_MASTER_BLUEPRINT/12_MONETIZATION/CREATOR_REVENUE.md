# CREATOR_REVENUE.md — 12.2 블로거 광고 수익 배분 및 기여도 산정 명세

## 1. 광고 수익 배분(Revenue Share) 공식
- 개별 블로그의 광고 슬롯(`ad_placements`)에서 발생한 유효 광고 과금액($C_{\text{valid}}$)은 블로그에 설정된 배분율(`revenue_share_bps`, 기본값 `7000` = **크리에이터 70% : 플랫폼 30%**)에 따라 정수 연산으로 분할된다.
  - $\text{CreatorShare} = \lfloor C_{\text{valid}} \times \text{revenue\_share\_bps} / 10000 \rfloor$
  - $\text{PlatformFee} = C_{\text{valid}} - \text{CreatorShare}$
- 일일 정산 배치 워커가 전일 유효 이벤트를 집계하여 `ledger_entries` 원장에 `ADVERTISER_PREPAID (DEBIT)` → `CREATOR_PENDING_ESCROW (CREDIT)` 및 `PLATFORM_FEE_REVENUE (CREDIT)` 전표를 생성한다.
- 사후 감사에서 조직적 부정 클릭이 적발된 경우, 해당 기간의 부정 수익을 환수하는 `FRAUD_CLAWBACK` 역분개 전표를 생성하고 `payouts.status = 'ON_HOLD'`로 지급을 보류한다.

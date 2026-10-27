# MEMBERSHIP.md — 12.3 유료 콘텐츠 · 멤버십 구독 및 12.4 독자 후원 명세

## 1. 유료 콘텐츠(`PAID_MEMBERS_ONLY`) 및 서버 사이드 Paywall
1. **유료 열람 모델**:
   - 블로거는 월간/연간 정기 구독형 **유료 멤버십(`memberships`)**을 운영하거나, 개별 심층 게시글에 **단건 구매 가격(`posts.paywall_price_minor`)**을 설정할 수 있다.
2. **서버 사이드 본문 절삭 (Server-side Truncation)**:
   - 에디터에서 작성자가 삽입한 `paywallDivider` 블록 이전까지만 무료 미리보기(Teaser)로 제공한다.
   - 권한이 없는 방문자에게는 백엔드 API가 `paywallDivider` 이후의 JSON AST 노드와 HTML을 **서버 메모리에서 완전히 제거한 상태**로 응답하므로, 브라우저 개발자 도구로도 유료 본문을 절대 열람할 수 없다.

## 2. 독자 후원 (Tipping / Donation)
- 독자는 유료 결제가 걸려 있지 않은 무료 공개 게시글을 읽고도 감사의 표시로 응원 메시지(`donation_message`)와 함께 자율 금액을 후원(`order_type = 'READER_DONATION'`)할 수 있다.

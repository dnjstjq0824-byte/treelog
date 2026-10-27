# ADVERTISING.md — 12.1 자체 광고 네트워크 및 문맥 타기팅 엔진 설계

## 1. 광고 캠페인 및 송출 파이프라인
1. **캠페인 및 소재 심사 (`ad_campaigns`)**:
   - 광고주(`ADVERTISER`)가 입찰 방식(`CPM` 1,000회 노출당 과금 / `CPC` 클릭당 과금), 일일 예산(`daily_budget_minor`), 타기팅 주제(`target_topics_json`), 타기팅 언어, 소재 이미지/문구/랜딩 URL을 등록하면 초기 상태는 `PENDING_REVIEW`가 된다.
   - 플랫폼 광고 심사원(`AD_REVIEWER`)이 랜딩 페이지 안전성과 소재 가이드라인 준수 여부를 검토하여 `APPROVED`로 승인한 캠페인만 송출 후보군에 진입한다.
2. **문맥 타기팅 및 Brand Safety**:
   - 독자가 게시글을 열람할 때 게시글의 카테고리(`topic_code`), 태그, 언어를 기반으로 유효 입찰가가 가장 높은(eCPM 기준) 광고를 매칭한다.
   - 비공개 글, 임시조치(`QUARANTINED`) 글, 성인 주의(`adult_content_warning = true`) 글, 스팸 의심 글에는 광고를 절대 송출하지 않는다.
3. **부정 클릭 및 무효 트래픽(IVT) 방어**:
   - 알려진 봇 User-Agent, 데이터센터 IP 대역, 동일 방문자 해시(`visitor_hash`)의 1시간 내 반복 클릭, 블로거 본인의 자가 클릭(Self-Click), 노출 후 300ms 이내에 발생한 매크로 클릭은 `is_invalid_traffic = true`로 마킹하여 광고주 과금 및 블로거 수익 정산에서 자동 제외한다.

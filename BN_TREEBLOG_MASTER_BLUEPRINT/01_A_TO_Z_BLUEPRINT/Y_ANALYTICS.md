# Y_ANALYTICS — Yield & Analytics (방문자·유입·전환·수익 분석 통계 시스템)
> **부제**: 방문자, 페이지뷰, 검색 유입, 구독 전환, 광고 실적, 작성자 통계, 검색엔진 통합 분석
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-Y_ANALYTICS`
> **적용 조항**: 제3조 (Y_ANALYTICS 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- 블로거와 기업 미디어팀이 자신의 블로그 성장세를 데이터로 파악할 수 있도록 **실시간 및 일/주/월별 방문자(UV), 페이지뷰(PV), 평균 체류시간/완독률, 유입 경로(검색엔진별/검색어별/소셜/직접), 구독 전환율, 수익 실적**을 제공하는 프라이버시 친화적 애널리틱스 엔진을 구축한다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **블로거 / 팀 애널리스트 (`BLOG_ANALYST`)**: 블로그 관리 센터의 통계 대시보드에서 어제 어떤 검색어(TreeScan, Google, Naver, Bing 등)로 독자들이 유입되었는지, 어떤 글에서 이웃/구독 전환이 가장 많이 일어났는지, 시간대별 조회수 추이는 어떠한지 확인한다.
- **플랫폼 운영자**: 전체 플랫폼의 일간 활성 사용자(DAU/MAU), 신규 발행 글 수, 검색 유입 비중, 광고 CTR/RPM 추이를 모니터링한다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-ANL-001 (핵심 수집 지표)**:
  - 트래픽 지표: 페이지뷰(`pv`), 순방문자(`uv` — 일일 회전 솔트 해시 기반 비식별 지문), 평균 열람 시간(`avg_dwell_sec`), 스크롤 완독률(`read_completion_rate`: 25%, 50%, 75%, 100% 도달 비율).
  - 유입 분석(Referrer & Search Attribution): `Direct`, `Internal (TreeBlog 피드/검색)`, `TreeScan Search`, `Google`, `Naver`, `Daum`, `Bing`, `Social (X/Facebook/Instagram/Kakao)`, `External Link` 분류 및 UTM 파라미터(`utm_source`, `utm_medium`, `utm_campaign`) 집계.
  - 전환 및 인게이지먼트: 게시글별 신규 팔로우/이웃/구독 전환 수, 공감 수, 댓글 수, 북마크 수, 외부 공유 수.
  - 수익 실적(Yield): 광고 노출 수, 유효 클릭 수, CTR, eCPM, 예상 크리에이터 배분 수익, 멤버십/후원 결제액.
- **REQ-ANL-002 (프라이버시 보호 및 어뷰징 필터링)**:
  - 서드파티 추적 쿠키 없이도 동작하는 1st-Party 경량 비콘(`POST /api/v1/public/analytics/beacon`)을 사용하며, 원본 IP 주소를 그대로 저장하지 않고 `SHA256(IP + UserAgent + DailyRotatingSalt)`로 익명화하여 GDPR/개인정보보호법을 준수한다.
  - 크롤러/봇(`Googlebot`, `TreeScanBot`, `Yeti`, Headless 봇 등) 트래픽은 일반 독자 UV/PV 통계와 분리하여 **'검색엔진 크롤링 현황'** 탭에 별도로 집계한다.

## 4. 데이터 구조 (Data Structures & Domain Models)
```go
type BlogDailyStats struct {
    BlogID             uuid.UUID      `json:"blog_id"`
    StatDate           string         `json:"stat_date"` // "2026-10-27" (블로그 설정 타임존 기준)
    PageViews          int64          `json:"page_views"`
    UniqueVisitors     int64          `json:"unique_visitors"`
    AvgDwellSeconds    float64        `json:"avg_dwell_seconds"`
    ReadCompletionRate float64        `json:"read_completion_rate"`
    NewSubscribers     int            `json:"new_subscribers"`
    ReferrerBreakdown  map[string]int64 `json:"referrer_breakdown"` // {"treescan": 420, "google": 310, ...}
    AdImpressions      int64          `json:"ad_impressions"`
    EstimatedRevenueMinor int64       `json:"estimated_revenue_minor"`
}
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- `apps/blog-console`: 기간 선택기(오늘 실시간 / 최근 7일 / 30일 / 90일 / 커스텀), 시계열 라인 차트, 유입 도메인/검색어 Top 50 테이블, 인기 게시글 순위, 독자 디바이스/국가/시간대 분포 차트.

## 6. 백엔드 로직 (Backend Logic & Processing)
- `internal/modules/analytics`:
  - 비콘 수집기는 Redis HyperLogLog(`uv:{blogId}:{date}`)와 Hash 카운터(`pv:{blogId}:{date}`)를 실시간 갱신하여 '오늘 실시간 방문자/조회수'를 1초 지연 이내로 제공한다.
  - 매시간/매일 정각 워커가 집계 데이터를 PostgreSQL 요약 테이블에 영구 반영(Rollup)한다.

## 7. API 계약 (API Contracts)
- `POST /api/v1/public/analytics/beacon` (페이지 열람, 체류시간, 완독 이벤트 수집)
- `GET /api/v1/blogs/{blogId}/analytics/overview?from={date}&to={date}` (블로그 종합 통계)
- `GET /api/v1/blogs/{blogId}/analytics/posts` (게시글별 상세 성과 순위)
- `GET /api/v1/blogs/{blogId}/analytics/referrers` (유입 경로 및 유입 검색어 분석)
- `GET /api/v1/blogs/{blogId}/analytics/audience` (기기고유유형/언어/국가/시간대 분포)
- `GET /api/v1/blogs/{blogId}/analytics/crawlers` (TreeScanBot/Googlebot 등 검색봇 수집 통계)

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- 주요 테이블 및 집계 구조: `posts` 카운터, `search_click_events`, `ad_events`, 일별/게시글별 통계 롤업 뷰 및 테이블.

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **작성자 본인 조회 제외**: 블로그 소유자 및 편집 멤버가 자신의 블로그를 조회하거나 미리보기한 트래픽은 독자 PV/UV 및 광고 노출 집계에서 자동 제외하여 통계 왜곡을 방지한다.

## 10. UX/UI 구조 (UX/UI Structure)
- 블로거가 가장 자주 확인하는 '오늘 방문자 / 어제 방문자 / 누적 방문자' 요약 위젯을 블로그 관리 홈 최상단에 배치한다.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- 블로그 통계 API는 해당 블로그의 `OWNER`, `ADMIN`, `EDITOR`, `ANALYST` 역할을 가진 멤버만 조회할 수 있다(단, 블로거가 블로그 홈에 '방문자 카운터 위젯 공개'를 켠 경우 요약 숫자만 공개).

## 12. 성능 및 확장성 (Performance & Scalability)
- 비콘 수집 엔드포인트는 DB 동기 쓰기를 수행하지 않고 메모리/Redis 파이프라인만 거치므로 P99 < 10ms 이내에 `204 No Content`를 반환한다.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- Redis HyperLogLog, `github.com/mileusna/useragent` (디바이스/OS/봇 파싱), `echarts`.

## 14. 개발 우선순위 (Development Priority)
- **Phase 8 (통계·광고·수익화 단계)** (`TB-OPS-003`, `TB-ADS-001`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 선행: `B_BLOGGING_EXPERIENCE`, `C_CONTENT_EDITOR`, `H_HOSTING_INFRASTRUCTURE`.

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-ANL-001`: `User-Agent: Googlebot/2.1` 및 `TreeScanBot/1.0`으로 게시글 요청 시 일반 독자 `UniqueVisitors`는 증가하지 않고 `crawlers` 통계에만 정확히 집계되는지 검증.
- `TEST-ANL-002`: 블로그 소유자 로그인 세션으로 본인 글 조회 시 조회수가 증가하지 않는지 검증.

## 17. Acceptance Criteria (인수 기준)
1. 실시간 및 일별 UV/PV, 체류시간, 유입 경로, 구독 전환, 광고 수익 통계가 봇 트래픽 및 본인 조회를 제외하고 정확히 집계·표시되어야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- 비콘 이벤트 주입 및 롤업 워커 실행 후 통계 API JSON 응답 검증 로그.

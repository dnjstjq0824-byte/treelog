# R_RECOMMENDATIONS — Recommendation & Ranking (추천 피드 및 공정 랭킹 엔진)
> **부제**: 사용자 피드, 관심사 추천, 인기 블로그, 관련 글, 최신 글, 추천 공정성 및 조작 방지
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-R_RECOMMENDATIONS`
> **적용 조항**: 제3조 (R_RECOMMENDATIONS 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- 독자에게는 관심사와 구독 관계에 기반한 맞춤형 피드와 양질의 토픽별 인기 글·관련 글을 제공하고, 신진 블로거에게는 공정한 발견 기회(Cold-start Discovery Boost)를 제공하며, 클릭베이트나 어뷰징 봇에 의한 랭킹 조작을 원천 차단한다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **로그인 독자**: 플랫폼 메인에서 '추천 피드(`SCR-EXP-009`)'와 '구독/이웃 새 글 피드(`SCR-EXP-010`)'를 전환하며 탐색하고, 게시글 하단에서 현재 읽은 글과 주제가 유사한 '관련 글(Related Posts)'을 이어서 읽는다.
- **비로그인 방문자**: 플랫폼 메인에서 실시간/일간/주간 인기 게시글(`SCR-EXP-003`), 인기 블로그(`SCR-EXP-002`), 카테고리별 추천 콘텐츠(`SCR-EXP-005`, `SCR-EXP-006`)를 탐색한다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-REC-001 (5대 디스커버리 지면 설계)**:
  1) **구독/이웃 피드 (Following/Neighbor Feed)**: 내가 팔로우/이웃/상호이웃/구독한 블로그의 최신 글을 시간 역순으로 정확히 모아보는 피드(관계 권한에 따라 `NEIGHBOR_ONLY`/`MUTUAL_ONLY` 글 포함).
  2) **개인화 추천 피드 (For You Feed)**: 사용자의 최근 읽은 글/공감/북마크 토픽 벡터와 유사한 고품질 공개 게시글 추천.
  3) **트렌딩 인기 게시글 (Trending Posts)**: 시간 감쇠(Time-decay Gravity) 공식을 적용한 실시간/일간/주간 인기 글.
  4) **주제별 인기 블로그 (Featured Blogs)**: 활동성, 완독률, 독자 증가율, 콘텐츠 품질이 우수한 블로그 추천.
  5) **게시글 하단 관련 글 (Related Posts)**: OpenSearch `more_like_this` 쿼리 및 동일 태그/카테고리 기반 연관 글 4~6건 추천.
- **REQ-REC-002 (투명한 랭킹 공식 및 어뷰징 방지)**:
  - 인기 점수 산식:
    $$\text{TrendingScore} = \frac{(V_{\text{valid}} \times 1.0 + R_{\text{unique}} \times 4.0 + C_{\text{unique}} \times 5.0 + B_{\text{unique}} \times 6.0) \times Q_{\text{factor}}}{(\text{AgeHours} + 2)^{1.5}}$$
  - 여기서 $V_{\text{valid}}$는 최소 15초 이상 체류한 고유 방문자 조회수, $R, C, B$는 각각 고유 사용자의 공감, 댓글, 북마크 수, $Q_{\text{factor}}$는 콘텐츠 품질 계수(`0.2` ~ `1.5`)이다.
  - **다양성 강제 (Author/Blog Diversity Cap)**: 인기 차트 상위 20위 내에 동일 블로그의 글은 최대 2개까지만 노출되도록 디듀플리케이션(MMR: Maximal Marginal Relevance)을 적용한다.

## 4. 데이터 구조 (Data Structures & Domain Models)
```go
type FeedItemDTO struct {
    PostID           uuid.UUID `json:"post_id"`
    BlogID           uuid.UUID `json:"blog_id"`
    BlogHandle       string    `json:"blog_handle"`
    BlogTitle        string    `json:"blog_title"`
    Title            string    `json:"title"`
    Excerpt          string    `json:"excerpt"`
    CoverImageURL    string    `json:"cover_image_url,omitempty"`
    ReasonBadge      string    `json:"reason_badge"` // "구독 중인 블로그", "IT·테크 인기 글", "관심 태그 #golang"
    TrendingScore    float64   `json:"trending_score"`
    PublishedAt      time.Time `json:"published_at"`
}
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- 플랫폼 탐색 메인(`SCR-EXP-001`~`010`)의 무한 스크롤(IntersectionObserver 커서 페이지네이션) 카드/리스트 뷰 및 추천 사유 배지(`왜 이 글이 추천되었나요?` 툴팁 + '이 블로그 관심 없음' 피드백 버튼) 제공.

## 6. 백엔드 로직 (Backend Logic & Processing)
- `internal/modules/recommendation`:
  - 매 5분마다 최근 72시간 내 활성 공개 게시글의 `TrendingScore`를 재계산하여 Redis Sorted Set(`trending:posts:{topic}:{period}`)과 OpenSearch `popularity_score`를 갱신하는 워커.
  - **조작 방지 (Anti-Gaming)**: 새로고침 연타로 인한 조회수 부풀리기를 막기 위해 Redis HyperLogLog 및 `(post_id, user_id_or_ip_ua_hash)` 6시간 쿨다운 키를 통과한 유효 열람만 랭킹에 반영한다. 또한 상호 담합(예: 소수 계정끼리 짧은 시간 내 수백 번 교차 공감) 패턴 감지 시 가중치를 0으로 감쇠시킨다.

## 7. API 계약 (API Contracts)
- `GET /api/v1/feed/following` (내가 구독/이웃한 블로그 피드)
- `GET /api/v1/feed/recommended` (개인화 추천 피드)
- `GET /api/v1/discover/trending-posts?period={24h|7d|30d}&topic={slug}` (인기 게시글)
- `GET /api/v1/discover/popular-blogs?topic={slug}` (인기 블로그)
- `GET /api/v1/public/posts/{postId}/related` (관련 글 목록)
- `POST /api/v1/feed/feedback` ('관심 없음' / '이 주제 숨기기' 피드백 반영)

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- 주요 연관 테이블: `posts`, `content_quality_signals`, `follows`, `subscriptions`, `reactions`, `bookmarks`, `search_click_events`.

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **콜드 스타트(신규 가입자/비로그인) 처리**: 열람 이력이 없는 사용자에게는 가입 시 선택한 관심 토픽(최소 3개) 또는 전체 에디토리얼/고품질 트렌딩 피드를 즉시 제공하여 빈 피드가 노출되지 않도록 한다.

## 10. UX/UI 구조 (UX/UI Structure)
- 모든 추천 카드에는 작성자 아바타, 블로그명, 읽기 소요 시간(`4분 분량`), 공감/댓글 수, 북마크 바로가기 버튼을 일관되게 배치한다.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- **비공개 글 추천 노출 절대 금지**:개인화 추천 피드, 인기 게시글, 관련 글은 100% `visibility = 'PUBLIC'` 게시글만을 후보군으로 삼는다. 이웃 공개(`NEIGHBOR_ONLY`) 및 상호이웃 공개(`MUTUAL_ONLY`) 글은 오직 로그인한 본인의 '구독/이웃 피드(`/api/v1/feed/following`)'에서만 권한 검증 후 노출된다.

## 12. 성능 및 확장성 (Performance & Scalability)
- 트렌딩 및 토픽별 추천 피드는 Redis ZSET에 사전 계산된 ID 목록을 캐싱하여 P95 < 40ms 이내에 응답한다.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- Redis ZSET, OpenSearch `more_like_this` & `function_score`.

## 14. 개발 우선순위 (Development Priority)
- **Phase 5 (기본 탐색/관련글)** 및 **Phase 6 (구독 피드/개인화 추천)** (`TB-SEARCH-002`, `TB-COMMUNITY-003`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 선행: `O_OPENSEARCH`, `K_CONTENT_QUALITY`, `E_ENGAGEMENT_COMMUNITY`.

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-REC-001`: 동일 IP에서 단일 게시글을 100회 연속 새로고침 조회하더라도 `view_count` 및 `TrendingScore`가 1회분만 증가하는지 어뷰징 방어 검증.
- `TEST-REC-002`: 인기 게시글 Top 20 조회 시 단일 블로그의 글이 2개를 초과하여 독점 노출되지 않는지 다양성 캡(Diversity Cap) 검증.

## 17. Acceptance Criteria (인수 기준)
1. 구독 피드, 추천 피드, 인기 게시글, 인기 블로그, 관련 글 API가 모두 정상 동작하며 어뷰징 방어 및 다양성 규칙이 적용되어야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- 랭킹 스코어 계산 단위 테스트 및 조회수 중복 방지 통합 테스트 로그.

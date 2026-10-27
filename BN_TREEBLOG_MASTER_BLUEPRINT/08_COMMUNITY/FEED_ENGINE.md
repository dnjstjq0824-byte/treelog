# FEED_ENGINE.md — 구독 피드 및 개인화 추천 피드 엔진 아키텍처

## 1. 구독/이웃 피드 (`GET /api/v1/feed/following`) 설계
- 로그인한 사용자의 `follows` (`status = 'ACCEPTED'`) 및 `subscriptions` (`status = 'ACTIVE'`) 대상 블로그 목록을 기반으로 최신 발행 글을 시간 역순 커서 페이지네이션으로 조회한다.
- 조회 시 요청자와 각 블로그 간의 관계(`FOLLOW` vs `NEIGHBOR` vs `MUTUAL_NEIGHBOR`)를 조인 조건에 반영하여, 상호이웃인 블로그의 `MUTUAL_ONLY` 글과 이웃인 블로그의 `NEIGHBOR_ONLY` 글이 내 피드에서 자연스럽게 함께 보이도록 구성한다.

## 2. 개인화 추천 피드 (`GET /api/v1/feed/recommended`) 설계
- 후보군 생성(Candidate Generation): 최근 14일 내 발행된 `visibility = 'PUBLIC'` 및 `content_quality_signals.index_eligibility = 'ELIGIBLE'` 게시글 중 사용자의 선호 토픽/태그와 일치하는 상위 500건을 OpenSearch에서 1차 추출.
- 리랭킹(Re-ranking): `TrendingScore` × `TopicAffinity` × `FreshnessDecay`로 점수를 매긴 뒤, 이미 읽은 글과 차단한 블로그를 제외하고 블로그 다양성 제한(동일 블로그 연속 노출 금지)을 적용하여 최종 20건을 반환한다.

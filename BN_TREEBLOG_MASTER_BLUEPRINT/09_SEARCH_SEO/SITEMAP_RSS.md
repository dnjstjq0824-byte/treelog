# SITEMAP_RSS.md — 동적 Sitemap 샤딩, RSS/Atom 및 WebSub 신디케이션 명세

## 1. 계층형 Sitemap 아키텍처
1. **플랫폼 전체 Sitemap Index (`GET /sitemap-index.xml`)**:
   - 활성 공개 블로그 및 공개 게시글을 25,000건 단위로 샤딩한 `/sitemaps/sitemap-posts-0001.xml`, `/sitemaps/sitemap-blogs-0001.xml` 목록을 제공한다.
2. **개별 블로그(서브도메인/개인도메인) 전용 Sitemap (`GET /sitemap.xml`)**:
   - 커스텀 도메인(`https://blog.user.com/sitemap.xml`) 요청 시 해당 블로그의 `status = 'PUBLISHED' AND visibility = 'PUBLIC'` 게시글과 공개 카테고리 URL만 포함된 독립 Sitemap을 실시간(Redis 10분 캐시, 새 글 발행 시 즉시 갱신) 반환한다.
3. **비공개 글 제외 보장**: `blog_settings.allow_search_indexing == false` 이거나 게시글이 `PUBLIC`이 아닌 경우 Sitemap 쿼리 조건에서 물리적으로 배제된다.

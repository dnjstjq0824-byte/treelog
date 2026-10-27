# SEO_TEST_PLAN.md — 글로벌 SEO 및 검색엔진 최적화 자동 검증 계획서 (제9조)

## 1. 자동화 SEO 검증 항목 (`tests/e2e/seo_audit.spec.ts`)
1. **No-JS SSR 무결성 테스트**: JS를 완전히 끈 HTTP 클라이언트(`curl` / Go `net/http`)로 공개 게시글을 요청했을 때 응답 HTML 내에 `<h1>`(글 제목), 본문 단락, `<title>`, `<meta name="description">`, `<link rel="canonical">`, `og:title`, `og:image`, `twitter:card`가 모두 존재해야 한다.
2. **구조화 데이터(JSON-LD) 스키마 검증**: `<script type="application/ld+json">`을 파싱하여 `@type: "BlogPosting"`의 필수 속성(`headline`, `datePublished`, `dateModified`, `author.name`, `mainEntityOfPage.@id`)과 `BreadcrumbList`가 유효한 JSON인지 검증한다.
3. **비공개/제한 공개 게시글 색인 차단 검증**: `UNLISTED`, `NEIGHBOR_ONLY`, `MUTUAL_ONLY`, `PRIVATE`, `QUARANTINED`, `TRASHED` 상태의 게시글 URL 요청 시 응답 HTTP 헤더에 `X-Robots-Tag: noindex, nofollow`가 100% 포함되는지 검증한다.
4. **Canonical 단일화 및 301 리다이렉트 검증**: 서브도메인, 경로형(`/@{handle}`), 개인 도메인 중 대표 주소가 아닌 경로로 접근하거나 변경 전 구 슬러그(`old_slug`)로 접근할 때 정확히 `301 Moved Permanently`와 대표 `Location` 헤더를 반환하는지 검증한다.
5. **Sitemap & RSS XML 유효성 검증**: `/sitemap-index.xml`, `/sitemap.xml`, `/rss.xml`, `/atom.xml`이 XML 파서 에러 없이 파싱되며, 비공개 글이 단 1건도 포함되지 않았는지 검증한다.

# SEO_ARCHITECTURE.md — 제9조 27대 필수 SEO 요구사항 구현 아키텍처

## 1. 제9조 27개 필수 SEO 항목 전수 구현 매핑

| 번호 | 필수 항목 | 기술적 구현 방법 (`apps/user-web` & `internal/modules/seo`) |
| :--- | :--- | :--- |
| 1 | **SSR 및 정적 렌더링** | Nuxt 3 Universal SSR + 공개 게시글 Edge SWR 캐싱으로 크롤러 요청 시 100% 완성된 HTML 즉시 반환 |
| 2 | **Canonical URL** | 블로그 대표 주소(`primary_url_type`) 및 `posts.canonical_url` 기준으로 `<link rel="canonical">` 단일화 |
| 3~4 | **XML Sitemap & Sitemap Index** | 플랫폼 루트 `/sitemap-index.xml` 및 개별 커스텀 도메인별 `/sitemap.xml` (50,000 URL 단위 자동 분할) |
| 5 | **RSS / Atom** | `/rss.xml`, `/atom.xml`, `/feed.json` 및 `<head>` 내 자동 발견(`rel="alternate"`) 태그 삽입 |
| 6 | **robots.txt** | 도메인별 동적 `/robots.txt` 생성 (비공개 경로 `/studio/`, `/admin/`, `/api/` 차단 및 Sitemap 경로 명시) |
| 7~8 | **Meta Title & Description** | `seo_title`(미지정 시 `{글제목} — {블로그명}`), `seo_description`(미지정 시 본문 첫 155자) 자동 생성 |
| 9~10| **Open Graph & Twitter/X Card** | `og:type="article"`, `og:title`, `og:description`, `og:image`(1200x630 최적화), `twitter:card="summary_large_image"` |
| 11~14| **구조화 데이터 (JSON-LD)** | `BlogPosting`, `BreadcrumbList`, `Person`/`Organization`(작성자 정보 `author`, 발행일 `datePublished`, 수정일 `dateModified`) 내장 |
| 15 | **공개·비공개 색인 제어** | `PUBLIC` 외 모든 글에 HTTP 헤더 `X-Robots-Tag: noindex, nofollow` 및 `<meta name="robots" content="noindex, nofollow">` 이중 적용 |
| 16~17| **삭제 페이지 및 리다이렉트** | 휴지통 `404`, 영구삭제 `410 Gone`, 법적차단 `451`, 슬러그/도메인 변경 시 `post_slug_history` 기반 `301 Moved Permanently` |
| 18~19| **페이지네이션 & 중복 색인 방지** | 목록 페이지 정식 URL 정규화 및 얇은 태그/검색 결과 페이지 `noindex, follow` 정책으로 크롤 버짓(Crawl Budget) 집중 |
| 20 | **개인 도메인 SEO** | 커스텀 도메인 접속 시 모든 내부 링크, Sitemap, Canonical, JSON-LD URL이 해당 개인 도메인 기준으로 동적 변환됨 |
| 21 | **검색 결과 미리보기** | 에디터 발행 드로어 내 Google / TreeScan / 소셜 공유 카드 실시간 SERP 미리보기 위젯 제공 |
| 22~24| **Core Web Vitals·이미지·모바일** | LCP 이미지 `fetchpriority="high"`, 나머지 `loading="lazy"`, AVIF/WebP `<picture>`, `width`/`height` 명시로 CLS < 0.05 |
| 25~27| **중복 관리·크롤링 허용·색인 모니터링**| SimHash 중복 감지, `TreeScanBot`/`Googlebot` 크롤링 통계 집계 및 IndexNow / Sitemap Ping 자동 전송 |

## 2. `BlogPosting` JSON-LD 출력 표준 템플릿
```html
<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "BlogPosting",
  "mainEntityOfPage": {
    "@type": "WebPage",
    "@id": "https://tech-pioneer.treeblog.example.com/posts/go-opensearch-architecture"
  },
  "headline": "Go와 OpenSearch로 구축하는 글로벌 블로그 검색 아키텍처",
  "description": "PostgreSQL을 원본으로 유지하며 OpenSearch를 비동기 동기화하는 방법",
  "image": ["https://media.treeblog.example.com/variants/018f.../w1200.webp"],
  "datePublished": "2026-10-27T12:00:00Z",
  "dateModified": "2026-10-27T14:30:00Z",
  "author": {
    "@type": "Person",
    "name": "김아키",
    "url": "https://tech-pioneer.treeblog.example.com/profile"
  },
  "publisher": {
    "@type": "Organization",
    "name": "Tech Pioneer 아키텍처 연구소"
  }
}
</script>
```

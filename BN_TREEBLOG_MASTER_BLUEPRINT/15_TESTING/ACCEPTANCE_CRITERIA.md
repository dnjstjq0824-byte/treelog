# ACCEPTANCE_CRITERIA.md — TreeBlog 상용화 인수 기준 총괄 명세 (제23조, 제25조, 제27조)

## 1. 플랫폼 전체 완료 승인(Sign-off)을 위한 10대 절대 인수 기준
후속 개발 에이전트가 프로젝트를 완료했다고 선언하기 위해서는 아래 10가지 인수 기준을 **실제 구동되는 컨테이너와 브라우저 환경에서 100% 충족**하고 증빙 로그를 제출해야 한다.

1. **독립 구동 완전성 (Zero TreeScan Dependency)**: TreeScan 관련 서버나 환경변수가 전혀 없는 상태에서 전체 컨테이너(`postgres`, `redis`, `opensearch`, `minio`, `api`, `worker`, `user-web`, `blog-console`, `admin-console`)가 정상 기동되고 모든 사용자/관리자 기능이 동작해야 한다.
2. **DB 무결성 및 영속성**: 인메모리 Mock이나 LocalStorage 위장이 아닌 실제 PostgreSQL 57개 테이블에 데이터가 트랜잭션으로 저장되고, 컨테이너 재시작 후에도 데이터가 100% 유지되어야 한다.
3. **핵심 사용자 여정 E2E 완주 (제27조)**: 회원가입 → 로그인 → 블로그 개설 → 에디터 글 작성(이미지 첨부·서식·태그) → 즉시/예약 발행 → OpenSearch 검색 노출 → 이웃/상호이웃 신청·수락 → 공감·댓글·대댓글 → 실시간 알림 → 관리자 신고 심사·임시조치·복원의 전 흐름이 브라우저 E2E 테스트를 통과해야 한다.
4. **비공개 데이터 및 프라이버시 절대 보호**: `PRIVATE`, `NEIGHBOR_ONLY`, `MUTUAL_ONLY` 게시글은 권한 없는 사용자의 직접 URL 접근 시 `403`/`404`로 차단되며, OpenSearch 공개 인덱스, Sitemap, RSS, 추천 피드, TreeScan 동기화 피드에 **단 1바이트도 노출되지 않아야 한다**.
5. **보안 무결성**: 모든 에디터 본문과 커스텀 CSS는 서버 측 Sanitizer를 통과하여 XSS 실행이 0건이어야 하며, 외부 URL 수집기는 SSRF `SafeDialer`를 통과하고, 업로드된 모든 공개 이미지는 GPS EXIF가 100% 제거되어야 한다.
6. **검색 및 글로벌 SEO 표준 준수**: 공개 게시글의 SSR 초기 HTML에 `<title>`, `<link rel="canonical">`, OpenGraph, `BlogPosting` JSON-LD가 정확히 포함되고, 도메인별 `/sitemap.xml` 및 `/rss.xml`이 유효하게 생성되어야 한다.
7. **반응형 및 웹 접근성**: 320px 모바일 뷰포트부터 데스크톱까지 45개 전체 화면에서 가로 스크롤 깨짐이 없고 Loading/Empty/Error 상태가 모두 구현되며 `axe-core` 심각한 접근성 위반이 0건이어야 한다.
8. **복식부기 정산 원장 무결성**: 광고 수익 배분, 유료 멤버십/단건 구매, 후원, 환불, 정산 출금 시 `ledger_entries` 테이블의 차변(`DEBIT`) 합계와 대변(`CREDIT`) 합계가 항상 정확히 일치해야 한다.
9. **데이터 이동의 자유**: 블로그 전체 글을 Markdown/HTML/WXR ZIP으로 내보내고 다시 가져올 때 본문, 이미지, 메타데이터, 구 URL 리다이렉트 매핑이 보존되어야 한다.
10. **독립 감사 통과**: 구현 에이전트가 아닌 `Independent Audit Agent`가 재실행한 테스트 로그와 정적/동적 감사에서 실패(`FAILED`) 항목이 0건이어야 한다.

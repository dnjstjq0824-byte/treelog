# MASTER_WORK_ORDER.md — TreeBlog 전체 개발 Work Order 마스터 지시서 (45+ 원자적 작업 분해)

> **총 Work Order 수**: **49개** (`13_WORK_ORDERS/individual/TB-*.md`에 개별 18개 항목 상세 명세 수록)
> **적용 조항**: 제19조 (개발 Work Order 전수 작성) 및 제22조 (Phase 0~10 로드맵)

---

## 1. Work Order 편성 원칙
1. **원자적 검증 가능성**: 각 Work Order는 단일 전문 에이전트가 독립된 디렉터리 소유권 내에서 구현하고 즉시 단위·통합·E2E 테스트로 검증할 수 있는 크기로 분해되었다.
2. **전 영역 추적성**: A~Z 26개 블루프린트 명세, 57개 DB 테이블, 24개 API 그룹(106개 엔드포인트), 45개 UX/UI 화면이 단 하나의 누락도 없이 아래 45개 Work Order에 매핑되어 있다.

## 2. 전체 Work Order 마스터 목록

| Work Order ID | Phase | 작업명 | 담당 에이전트 | 선행 작업 | 검증 테스트 ID | 개별 명세 파일 |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `TB-ARCH-001` | Phase 0 | **모노레포 디렉터리 스캐폴딩 및 Go 모듈형 모놀리스 코어 프레임워크 구축** | Lead Architect Agent | `NONE` | `TEST-ARCH-001` | [`individual/TB-ARCH-001.md`](./individual/TB-ARCH-001.md) |
| `TB-OPS-001` | Phase 0 | **Docker Compose 로컬/CI 표준 스택 및 GitHub Actions 파이프라인 구축** | DevOps Agent | `TB-ARCH-001` | `TEST-INF-001, TEST-INF-002` | [`individual/TB-OPS-001.md`](./individual/TB-OPS-001.md) |
| `TB-DB-001` | Phase 0 | **PostgreSQL 마이그레이션 엔진 및 Identity/Blog 기초 스키마(0001~0002) 구축** | PostgreSQL Agent | `TB-ARCH-001, TB-OPS-001` | `TEST-DB-001` | [`individual/TB-DB-001.md`](./individual/TB-DB-001.md) |
| `TB-ARCH-002` | Phase 0 | **Transactional Outbox 이벤트 릴레이 및 Asynq 멀티 큐 워커 엔진 구축** | Lead Architect Agent | `TB-DB-001` | `TEST-ARCH-002, TEST-JOB-001` | [`individual/TB-ARCH-002.md`](./individual/TB-ARCH-002.md) |
| `TB-DB-002` | Phase 1 | **Content, Media, Community, Search, Moderation, Monetization 전체 스키마(0003~0010) 구축** | PostgreSQL Agent | `TB-DB-001` | `TEST-DB-001, TEST-DB-002` | [`individual/TB-DB-002.md`](./individual/TB-DB-002.md) |
| `TB-AUTH-001` | Phase 1 | **독립 이메일 회원가입·로그인·Argon2id·JWT Refresh Token Rotation 엔진 구현** | Backend Developer Agent | `TB-DB-002` | `TEST-USR-001, TEST-USR-002, TEST-USR-003, TEST-IAM-001` | [`individual/TB-AUTH-001.md`](./individual/TB-AUTH-001.md) |
| `TB-AUTH-002` | Phase 1 | **소셜 OAuth/OIDC 어댑터, TOTP MFA 및 2계층 RBAC 인가 미들웨어 구현** | Backend Developer Agent | `TB-AUTH-001` | `TEST-USR-004, TEST-USR-005, TEST-USR-006, TEST-IAM-002` | [`individual/TB-AUTH-002.md`](./individual/TB-AUTH-002.md) |
| `TB-SEC-001` | Phase 1 | **공통 보안 미들웨어(CSRF, Rate Limiter, SafeDialer SSRF 방어, 보안 헤더) 구축** | Security Agent | `TB-ARCH-001, TB-AUTH-001` | `TEST-SEC-002` | [`individual/TB-SEC-001.md`](./individual/TB-SEC-001.md) |
| `TB-API-001` | Phase 1 | **OpenAPI 3.1.0 스키마 검증기 및 프론트엔드 타입 안전 API 클라이언트 생성 파이프라인** | Lead Architect Agent | `TB-ARCH-001, TB-AUTH-001` | `TEST-FE-001` | [`individual/TB-API-001.md`](./individual/TB-API-001.md) |
| `TB-WEB-001` | Phase 1 | **디자인 시스템(packages/ui-kit) 및 회원 인증·프로필·보안 설정 UI 구현** | Frontend Developer Agent | `TB-API-001, TB-AUTH-002` | `TEST-USR-001, TEST-UX-001` | [`individual/TB-WEB-001.md`](./individual/TB-WEB-001.md) |
| `TB-BLOG-001` | Phase 2 | **블로그 개설 엔진, 핸들 예약어 검증, 팀 멤버 RBAC 및 운영 설정 API 구현** | Backend Developer Agent | `TB-AUTH-002, TB-DB-002` | `TEST-BLOG-001, TEST-BLOG-002, TEST-BLOG-005` | [`individual/TB-BLOG-001.md`](./individual/TB-BLOG-001.md) |
| `TB-BLOG-002` | Phase 2 | **2뎁스 계층형 카테고리, 시리즈(연재물) 및 태그 사전 관리 API 구현** | Backend Developer Agent | `TB-BLOG-001` | `TEST-BLOG-003` | [`individual/TB-BLOG-002.md`](./individual/TB-BLOG-002.md) |
| `TB-BLOG-003` | Phase 2 | **멀티 도메인 라우팅(서브도메인/경로/개인 도메인), DNS 검증 및 자동 ACME TLS 엔진 구현** | Backend Developer Agent | `TB-BLOG-001, TB-SEC-001` | `TEST-DOM-001, TEST-DOM-002` | [`individual/TB-BLOG-003.md`](./individual/TB-BLOG-003.md) |
| `TB-WEB-002` | Phase 2 | **블로그 개설 위저드 및 내 블로그 관리 센터(설정·카테고리·멤버·도메인) UI 구현** | Frontend Developer Agent | `TB-WEB-001, TB-BLOG-001, TB-BLOG-002, TB-BLOG-003` | `TEST-BLOG-001, TEST-BLOG-003` | [`individual/TB-WEB-002.md`](./individual/TB-WEB-002.md) |
| `TB-DB-003` | Phase 2 | **데이터베이스 백업·시점복구(PITR) 스크립트 및 파티션 자동 관리 워커 구현** | PostgreSQL Agent | `TB-DB-002, TB-ARCH-002` | `TEST-INF-001` | [`individual/TB-DB-003.md`](./individual/TB-DB-003.md) |
| `TB-MEDIA-001` | Phase 3 | **S3 Presigned 업로드 세션, 매직바이트 검증, 이미지 리사이징(WebP/AVIF) 및 EXIF 제거 워커 구현** | Media Processing Agent | `TB-ARCH-002, TB-DB-002` | `TEST-MED-001, TEST-SEC-003` | [`individual/TB-MEDIA-001.md`](./individual/TB-MEDIA-001.md) |
| `TB-MEDIA-002` | Phase 3 | **허용 목록 기반 외부 동영상 안전 임베드(oEmbed) 및 자체 비디오 트랜스코딩 확장 인터페이스 구현** | Media Processing Agent | `TB-MEDIA-001, TB-SEC-001` | `TEST-MED-002` | [`individual/TB-MEDIA-002.md`](./individual/TB-MEDIA-002.md) |
| `TB-EDITOR-001` | Phase 3 | **Tiptap/ProseMirror 블록 스키마(packages/editor-core) 및 서버 AST 검증·HTML Sanitizer 구현** | Editor Engine Agent | `TB-ARCH-001, TB-MEDIA-001` | `TEST-EDIT-001, TEST-SEC-001` | [`individual/TB-EDITOR-001.md`](./individual/TB-EDITOR-001.md) |
| `TB-EDITOR-002` | Phase 3 | **듀얼 레이어 자동저장(IndexedDB + 서버), 버전 충돌 해결 및 리비전 Diff 복원 엔진 구현** | Editor Engine Agent | `TB-EDITOR-001, TB-DB-002` | `TEST-EDIT-002, TEST-EDIT-003` | [`individual/TB-EDITOR-002.md`](./individual/TB-EDITOR-002.md) |
| `TB-EDITOR-003` | Phase 3 | **블로그 글쓰기 에디터 스튜디오 UI(SCR-BLOG-013) 데스크톱 및 모바일 완벽 대응 구현** | Editor Engine Agent | `TB-EDITOR-001, TB-EDITOR-002, TB-MEDIA-002` | `TEST-EDIT-001, TEST-FE-001` | [`individual/TB-EDITOR-003.md`](./individual/TB-EDITOR-003.md) |
| `TB-POST-001` | Phase 3 | **게시글 즉시 발행, 수정, 공개범위(6단계) 제어, 휴지통(30일 보관)·복원·영구삭제 엔진 구현** | Backend Developer Agent | `TB-EDITOR-001, TB-BLOG-002` | `TEST-PUB-001, TEST-LFC-001, TEST-LFC-002` | [`individual/TB-POST-001.md`](./individual/TB-POST-001.md) |
| `TB-POST-002` | Phase 3 | **분 단위 정밀 예약 발행 스케줄러(scheduled_publications) 및 팀 블로그 원고 승인 워크플로 구현** | Backend Developer Agent | `TB-POST-001, TB-ARCH-002` | `TEST-JOB-001, TEST-JOB-002, TEST-GLB-002` | [`individual/TB-POST-002.md`](./individual/TB-POST-002.md) |
| `TB-THEME-001` | Phase 4 | **블로그 테마·스킨 엔진, 디자인 토큰 주입기 및 커스텀 CSS 보안 파서/샌드박스 구현** | Frontend Developer Agent | `TB-BLOG-001, TB-SEC-001` | `TEST-DSN-001` | [`individual/TB-THEME-001.md`](./individual/TB-THEME-001.md) |
| `TB-THEME-002` | Phase 4 | **블로그 관리 센터 내 실시간 비주얼 테마 스튜디오(모바일/데스크톱 프리뷰어) UI 구현** | Frontend Developer Agent | `TB-THEME-001, TB-WEB-002` | `TEST-DSN-001` | [`individual/TB-THEME-002.md`](./individual/TB-THEME-002.md) |
| `TB-WEB-003` | Phase 4 | **개인 블로그 전체 사용자 화면(SCR-BLOG-001~012) Nuxt 3 SSR 및 반응형 UI 구현** | Frontend Developer Agent | `TB-POST-001, TB-THEME-001` | `TEST-FE-001, TEST-FE-002, TEST-UX-001` | [`individual/TB-WEB-003.md`](./individual/TB-WEB-003.md) |
| `TB-WEB-004` | Phase 4 | **플랫폼 탐색 메인, 인기 블로그/게시글, 카테고리/주제/태그 탐색 화면(SCR-EXP-001~010) 구현** | Frontend Developer Agent | `TB-WEB-003` | `TEST-UX-001, TEST-UX-002` | [`individual/TB-WEB-004.md`](./individual/TB-WEB-004.md) |
| `TB-SEARCH-001` | Phase 5 | **OpenSearch 다국어 인덱스(Nori/ICU), Outbox 비동기 인덱서 워커 및 통합 검색 API 구현** | Search & SEO Agent | `TB-POST-001, TB-ARCH-002` | `TEST-SRC-001, TEST-SRC-002` | [`individual/TB-SEARCH-001.md`](./individual/TB-SEARCH-001.md) |
| `TB-SEARCH-002` | Phase 5 | **콘텐츠 품질 시그널(SimHash 중복 탐지) 산출 및 트렌딩/개인화 추천 랭킹 엔진 구현** | Search & SEO Agent | `TB-SEARCH-001` | `TEST-QLT-001, TEST-REC-001, TEST-REC-002` | [`individual/TB-SEARCH-002.md`](./individual/TB-SEARCH-002.md) |
| `TB-SEO-001` | Phase 5 | **글로벌 SEO 엔진(SSR 메타, Canonical, JSON-LD 구조화 데이터, hreflang, noindex 제어) 구현** | Search & SEO Agent | `TB-WEB-003, TB-BLOG-003` | `TEST-FE-002, TEST-SEO-001` | [`individual/TB-SEO-001.md`](./individual/TB-SEO-001.md) |
| `TB-SEO-002` | Phase 5 | **동적 XML Sitemap Index, 도메인별 Sitemap, robots.txt, RSS 2.0 / Atom 1.0 / WebSub 구현** | Search & SEO Agent | `TB-SEO-001` | `TEST-SEO-002` | [`individual/TB-SEO-002.md`](./individual/TB-SEO-002.md) |
| `TB-COMMUNITY-001` | Phase 6 | **팔로우·이웃·상호이웃(서로이웃) 상태 머신, 이웃 그룹, 구독 및 사용자 차단 API 구현** | Community Agent | `TB-BLOG-001, TB-POST-001` | `TEST-COM-001, TEST-COM-002` | [`individual/TB-COMMUNITY-001.md`](./individual/TB-COMMUNITY-001.md) |
| `TB-COMMUNITY-002` | Phase 6 | **2뎁스 댓글·대댓글, 비밀댓글 서버 마스킹, 승인제 댓글, 6종 공감 리액션 및 북마크 구현** | Community Agent | `TB-COMMUNITY-001` | `TEST-COM-003, TEST-COM-004, TEST-COM-005` | [`individual/TB-COMMUNITY-002.md`](./individual/TB-COMMUNITY-002.md) |
| `TB-COMMUNITY-003` | Phase 6 | **멀티채널 알림 엔진(인앱 SSE, VAPID 웹 푸시, 이메일), Rollup 병합 및 커뮤니티 UI 구현** | Community Agent | `TB-COMMUNITY-001, TB-COMMUNITY-002, TB-WEB-003` | `TEST-NOT-001, TEST-NOT-002` | [`individual/TB-COMMUNITY-003.md`](./individual/TB-COMMUNITY-003.md) |
| `TB-POST-003` | Phase 6 | **외부 블로그 가져오기(WXR/Markdown Import) 및 블로그 전체 백업 내보내기(Export ZIP) 구현** | Backend Developer Agent | `TB-POST-001, TB-MEDIA-001, TB-SEO-002` | `TEST-INT-001, TEST-INT-002` | [`individual/TB-POST-003.md`](./individual/TB-POST-003.md) |
| `TB-SEC-002` | Phase 7 | **자동 스팸 탐지 엔진, 신고 접수·병합, 임시조치(QUARANTINE) 및 이의신청(Appeals) 백엔드 구현** | Security Agent | `TB-POST-001, TB-COMMUNITY-002` | `TEST-QLT-002, TEST-ADM-002` | [`individual/TB-SEC-002.md`](./individual/TB-SEC-002.md) |
| `TB-ADMIN-001` | Phase 7 | **플랫폼 관리자 백엔드 API(회원·블로그·게시글·댓글·검색·시스템·감사로그) 구현** | Admin Console Agent | `TB-SEC-002, TB-SEARCH-001` | `TEST-ADM-001, TEST-ADM-003` | [`individual/TB-ADMIN-001.md`](./individual/TB-ADMIN-001.md) |
| `TB-ADMIN-002` | Phase 7 | **통합 관리자 콘솔 프론트엔드(apps/admin-console) 회원·블로그·게시글·커뮤니티·이의신청 UI 구현** | Admin Console Agent | `TB-ADMIN-001, TB-WEB-001` | `TEST-ADM-001, TEST-ADM-002` | [`individual/TB-ADMIN-002.md`](./individual/TB-ADMIN-002.md) |
| `TB-ADMIN-003` | Phase 7 | **통합 관리자 콘솔 검색 관리·광고/정산 심사·시스템 설정·감사 로그 UI(SCR-ADM-006~008) 구현** | Admin Console Agent | `TB-ADMIN-002` | `TEST-ADM-003, TEST-ADM-004` | [`individual/TB-ADMIN-003.md`](./individual/TB-ADMIN-003.md) |
| `TB-ANL-001` | Phase 8 | **프라이버시 친화적 방문자·체류시간·완독률 비콘 수집기 및 유입 경로/크롤러 분석 엔진 구현** | Backend Developer Agent | `TB-POST-001, TB-OPS-001` | `TEST-ANL-001, TEST-ANL-002` | [`individual/TB-ANL-001.md`](./individual/TB-ANL-001.md) |
| `TB-ADS-001` | Phase 8 | **자체 광고 네트워크(캠페인/소재/타기팅 송출/노출·클릭 비콘/IVT 부정클릭 방어) 및 수익 배분 엔진 구현** | Backend Developer Agent | `TB-ANL-001, TB-DB-002` | `TEST-MON-003` | [`individual/TB-ADS-001.md`](./individual/TB-ADS-001.md) |
| `TB-BILLING-001` | Phase 8 | **유료 콘텐츠(서버 사이드 Paywall), 유료 멤버십 구독 및 독자 후원(Tipping) 결제 엔진 구현** | Backend Developer Agent | `TB-POST-001, TB-DB-002` | `TEST-MON-001, TEST-MON-002` | [`individual/TB-BILLING-001.md`](./individual/TB-BILLING-001.md) |
| `TB-BILLING-002` | Phase 8 | **불변 복식부기 회계 원장(ledger_entries), 정산금 출금(Payout), 지급 보류 및 부정수익 환수 구현** | Backend Developer Agent | `TB-ADS-001, TB-BILLING-001` | `TEST-MON-001, TEST-ADM-004` | [`individual/TB-BILLING-002.md`](./individual/TB-BILLING-002.md) |
| `TB-OPS-002` | Phase 9 | **OpenTelemetry 분산 추적·Prometheus 메트릭·구조화 로깅 및 운영 런북/알람 체계 구축** | DevOps Agent | `TB-ARCH-002, TB-ADMIN-001` | `TEST-INF-002` | [`individual/TB-OPS-002.md`](./individual/TB-OPS-002.md) |
| `TB-OPS-003` | Phase 9 | **글로벌 다국어(ko/en/ja), 타임존 정규화, CDN Edge 캐시 퍼지 및 Core Web Vitals 성능 최적화** | Frontend Developer Agent | `TB-WEB-004, TB-SEO-001` | `TEST-GLB-001, TEST-GLB-002, TEST-FE-001` | [`individual/TB-OPS-003.md`](./individual/TB-OPS-003.md) |
| `TB-QA-001` | Phase 9 | **전체 도메인 단위·실제 DB/OpenSearch 통합 테스트 및 보안 침투(XSS/SSRF/CSRF/IDOR) 테스트 스위트 구축** | QA / E2E Agent | `TB-ADMIN-003, TB-BILLING-002` | `TEST-SEC-001, TEST-SEC-002, TEST-SEC-003` | [`individual/TB-QA-001.md`](./individual/TB-QA-001.md) |
| `TB-QA-002` | Phase 9 | **Playwright 멀티 브라우저 E2E(가입->블로그개설->글쓰기->발행->검색->이웃/댓글->어드민) 및 a11y/성능 검증** | QA / E2E Agent | `TB-QA-001, TB-OPS-003` | `TEST-FE-001, TEST-UX-001, TEST-UX-002` | [`individual/TB-QA-002.md`](./individual/TB-QA-002.md) |
| `TB-TREESCAN-001` | Phase 10 | **TreeScan 검색 색인 동기화 어댑터(공개 콘텐츠 전용 증분 피드 & 삭제 툼스톤) 및 디스커버리 API 구현** | Search & SEO Agent | `TB-SEARCH-001, TB-SEC-001` | `TEST-ZTR-001, TEST-ZTR-002, TEST-ZTR-003, TEST-ZTR-004` | [`individual/TB-TREESCAN-001.md`](./individual/TB-TREESCAN-001.md) |
| `TB-TREESCAN-002` | Phase 10 | **Booking Network / TreeScan OIDC 통합 SSO 계정 연결 및 데이터 소유권 무손실 병합 구현** | Backend Developer Agent | `TB-AUTH-002, TB-TREESCAN-001` | `TEST-IAM-002, TEST-ZTR-001` | [`individual/TB-TREESCAN-002.md`](./individual/TB-TREESCAN-002.md) |
| `TB-QA-003` | Phase 10 | **Independent Audit Agent 전수 독립 감사, TreeScan 무의존 단독 구동 검증 및 최종 납품 승인** | Independent Audit Agent | `TB-QA-002, TB-TREESCAN-001, TB-TREESCAN-002` | `TEST-ZTR-001, TEST-ZTR-002, TEST-ZTR-003` | [`individual/TB-QA-003.md`](./individual/TB-QA-003.md) |

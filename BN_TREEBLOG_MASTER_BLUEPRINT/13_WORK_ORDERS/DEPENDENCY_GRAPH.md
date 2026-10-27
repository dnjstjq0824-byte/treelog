# DEPENDENCY_GRAPH.md — Work Order 실행 의존성 DAG 및 병렬화 설계

## 1. Phase 0 ~ Phase 10 실행 의존성 그래프 (Mermaid DAG)

```mermaid
graph TD
    %% Phase 0
    TB_ARCH_001[TB-ARCH-001: 모노레포 & Go 코어] --> TB_OPS_001[TB-OPS-001: Docker & CI 스택]
    TB_OPS_001 --> TB_DB_001[TB-DB-001: DB 기초 & 0001~0002]
    TB_DB_001 --> TB_ARCH_002[TB-ARCH-002: Outbox & Worker 큐]

    %% Phase 1
    TB_DB_001 --> TB_DB_002[TB-DB-002: 전체 스키마 0003~0010]
    TB_DB_002 --> TB_AUTH_001[TB-AUTH-001: 회원가입/로그인/JWT]
    TB_AUTH_001 --> TB_AUTH_002[TB-AUTH-002: OIDC/MFA/2계층 RBAC]
    TB_AUTH_001 --> TB_SEC_001[TB-SEC-001: CSRF/RateLimit/SafeDialer]
    TB_AUTH_001 --> TB_API_001[TB-API-001: OpenAPI 코드젠 클라이언트]
    TB_API_001 --> TB_WEB_001[TB-WEB-001: UI-Kit & 회원/인증 화면]

    %% Phase 2
    TB_AUTH_002 --> TB_BLOG_001[TB-BLOG-001: 블로그 개설/멤버/설정]
    TB_BLOG_001 --> TB_BLOG_002[TB-BLOG-002: 2뎁스 카테고리/시리즈/태그]
    TB_BLOG_001 --> TB_BLOG_003[TB-BLOG-003: 커스텀 도메인/DNS/ACME TLS]
    TB_BLOG_003 --> TB_WEB_002[TB-WEB-002: 블로그 개설 & 관리 센터 UI]
    TB_ARCH_002 --> TB_DB_003[TB-DB-003: 백업/PITR & 파티션 매니저]

    %% Phase 3
    TB_ARCH_002 --> TB_MEDIA_001[TB-MEDIA-001: S3 업로드/WebP/EXIF 제거]
    TB_MEDIA_001 --> TB_MEDIA_002[TB-MEDIA-002: 외부 동영상 안전 임베드]
    TB_MEDIA_001 --> TB_EDITOR_001[TB-EDITOR-001: Tiptap 스키마 & 서버 Sanitizer]
    TB_EDITOR_001 --> TB_EDITOR_002[TB-EDITOR-002: 듀얼 자동저장 & 리비전 Diff]
    TB_EDITOR_002 --> TB_EDITOR_003[TB-EDITOR-003: 에디터 스튜디오 UI]
    TB_EDITOR_001 --> TB_POST_001[TB-POST-001: 게시글 발행/가시성/휴지통]
    TB_POST_001 --> TB_POST_002[TB-POST-002: 정밀 예약 발행 스케줄러]

    %% Phase 4
    TB_BLOG_001 --> TB_THEME_001[TB-THEME-001: 테마 엔진 & CSS 샌드박스]
    TB_THEME_001 --> TB_THEME_002[TB-THEME-002: 비주얼 테마 스튜디오 UI]
    TB_POST_001 --> TB_WEB_003[TB-WEB-003: 개인 블로그 12개 화면 SSR]
    TB_WEB_003 --> TB_WEB_004[TB-WEB-004: 플랫폼 탐색 10개 화면 SSR]

    %% Phase 5
    TB_POST_001 --> TB_SEARCH_001[TB-SEARCH-001: OpenSearch 다국어 색인/검색]
    TB_SEARCH_001 --> TB_SEARCH_002[TB-SEARCH-002: SimHash 품질 & 추천 랭킹]
    TB_WEB_003 --> TB_SEO_001[TB-SEO-001: SSR 메타/Canonical/JSON-LD]
    TB_SEO_001 --> TB_SEO_002[TB-SEO-002: 동적 Sitemap/RSS/robots.txt]

    %% Phase 6
    TB_POST_001 --> TB_COMMUNITY_001[TB-COMMUNITY-001: 팔로우/이웃/상호이웃/차단]
    TB_COMMUNITY_001 --> TB_COMMUNITY_002[TB-COMMUNITY-002: 댓글/비밀댓글/6종공감/북마크]
    TB_COMMUNITY_002 --> TB_COMMUNITY_003[TB-COMMUNITY-003: 멀티채널 알림/Rollup & UI]
    TB_SEO_002 --> TB_POST_003[TB-POST-003: 블로그 Import/Export ZIP]

    %% Phase 7
    TB_COMMUNITY_002 --> TB_SEC_002[TB-SEC-002: 스팸탐지/신고/QUARANTINE/이의신청]
    TB_SEC_002 --> TB_ADMIN_001[TB-ADMIN-001: 통합 관리자 백엔드 API]
    TB_ADMIN_001 --> TB_ADMIN_002[TB-ADMIN-002: 어드민 콘솔 UI Part 1]
    TB_ADMIN_002 --> TB_ADMIN_003[TB-ADMIN-003: 어드민 콘솔 UI Part 2]

    %% Phase 8
    TB_POST_001 --> TB_ANL_001[TB-ANL-001: 방문자/유입/크롤러 통계 엔진]
    TB_ANL_001 --> TB_ADS_001[TB-ADS-001: 광고 서버/IVT방어/수익배분]
    TB_POST_001 --> TB_BILLING_001[TB-BILLING-001: 유료 멤버십/페이월/후원]
    TB_BILLING_001 --> TB_BILLING_002[TB-BILLING-002: 복식부기 원장 & 정산 출금]

    %% Phase 9 & 10
    TB_ADMIN_003 --> TB_OPS_002[TB-OPS-002: OpenTelemetry 관측성/런북]
    TB_WEB_004 --> TB_OPS_003[TB-OPS-003: 다국어/타임존/CWV 최적화]
    TB_BILLING_002 --> TB_QA_001[TB-QA-001: 전 도메인 통합 & 보안 테스트]
    TB_QA_001 --> TB_QA_002[TB-QA-002: Playwright E2E & 접근성/부하 검증]
    TB_SEARCH_001 --> TB_TREESCAN_001[TB-TREESCAN-001: TreeScan 색인/디스커버리 어댑터]
    TB_TREESCAN_001 --> TB_TREESCAN_002[TB-TREESCAN-002: TreeScan OIDC 계정 연동]
    TB_QA_002 --> TB_QA_003[TB-QA-003: 최종 독립 감사 & 납품 승인]
    TB_TREESCAN_002 --> TB_QA_003
```

## 2. DAG 무결성 및 병렬 실행 슬롯 검증
- **순환 의존성(Cycle) 검사 결과**: `0 Cycles Detected (Topological Sort Validated)`.
- **최대 병렬 실행 구간**: `TB-BLOG-001` 및 `TB-POST-001` 완료 직후, `Media Agent`, `Editor Agent`, `Theme/Frontend Agent`, `Search/SEO Agent`, `Community Agent`, `Analytics/Billing Agent`가 각자 독점 소유한 디렉터리에서 충돌 없이 6-Way 병렬 개발을 수행할 수 있다.

# COMPONENT_DIAGRAMS.md — 모듈 경계 및 컴포넌트 상세 구조

## 1. Go 백엔드 15대 도메인 모듈 경계 및 의존성 규칙

무조건적인 마이크로서비스 분할을 지양하고 **모듈형 모놀리스(Modular Monolith)**의 장점을 극대화하기 위해 다음 모듈 계층 규칙을 강제한다.

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                         API Gateway / HTTP Router (Chi)                      │
│     Middleware: RequestID -> OpenTelemetry -> RealIP -> CORS -> RateLimit    │
│                 -> Auth(JWT/Session) -> RBAC -> CSRF -> AuditLogger          │
└───────────────────────────────────────┬──────────────────────────────────────┘
                                        │
        ┌───────────────────────────────┼───────────────────────────────┐
        ▼                               ▼                               ▼
┌───────────────────┐         ┌───────────────────┐         ┌───────────────────┐
│ Presentation Core │         │  Engagement Core  │         │  Platform Admin   │
│ - blog            │         │ - community       │         │ - moderation      │
│ - post            │         │ - notification    │         │ - monetization    │
│ - editor          │         │ - search          │         │ - analytics       │
│ - theme           │         │ - recommendation  │         │ - admin           │
│ - domain          │         │ - seo             │         │ - treescan        │
│ - media           │         └─────────┬─────────┘         └─────────┬─────────┘
└─────────┬─────────┘                   │                             │
          └─────────────────────────────┼─────────────────────────────┘
                                        ▼
                        ┌───────────────────────────────┐
                        │       Foundation Modules      │
                        │ - identity (Auth/OIDC/RBAC)   │
                        │ - user (Profiles/Settings)    │
                        │ - shared/outbox & events      │
                        └───────────────────────────────┘
```

### 1.1 모듈별 책임 및 소유 테이블 매핑

| 모듈명 (`internal/modules/*`) | 핵심 책임 | 독점 소유 DB 테이블 (Write Ownership) |
| :--- | :--- | :--- |
| **`identity`** | 회원가입, 로그인, OAuth/OIDC, 세션, MFA, RBAC 권한 평가 | `users`, `identities`, `sessions`, `roles`, `permissions`, `user_roles`, `user_consents` |
| **`user`** | 사용자 프로필, 기본 환경설정, 계정 탈퇴 및 휴면 처리 | `user_profiles` |
| **`blog`** | 블로그 개설, 핸들 검증, 팀 멤버 관리, 카테고리, 네비게이션, 블로그 생명주기 | `blogs`, `blog_members`, `blog_settings`, `blog_navigation`, `blog_categories` |
| **`domain`** | 서브도메인/경로/개인 도메인 매핑, DNS TXT/CNAME 검증, ACME TLS 인증서 상태 | `blog_domains` |
| **`theme`** | 기본 스킨/테마 카탈로그, 디자인 토큰 커스터마이징, 안전한 HTML/CSS 샌드박스 | `themes`, `blog_themes` |
| **`editor`** | Tiptap JSON AST 스키마 검증, 임시저장/자동저장, 리비전 생성 및 Diff 복원 | `post_revisions`, `autosave_drafts` |
| **`post`** | 게시글 발행, 공개 범위 제어(`PUBLIC`, `MUTUAL_ONLY` 등), 예약 발행, 태그, 시리즈, 데이터 이동(Import/Export) | `posts`, `post_categories`, `post_tags`, `tags`, `post_media`, `post_visibility_rules`, `scheduled_publications`, `series`, `migration_jobs` |
| **`media`** | Presigned 업로드 세션, 바이러스/매직바이트 검사, 이미지 리사이징(WebP/AVIF), EXIF 제거, 동영상 임베드 메타데이터 | `upload_sessions`, `media_assets`, `media_variants`, `media_processing_jobs` |
| **`community`** | 팔로우, 이웃, 상호이웃 신청/승인, 이메일/웹 구독, 댓글/대댓글, 공감(6종 반응), 북마크, 방명록, 사용자 차단 | `follows`, `subscriptions`, `reactions`, `comments`, `bookmarks`, `guestbook_entries`, `user_blocks` |
| **`notification`** | 인앱 알림, 웹 푸시, 이메일 발송 큐, 알림 수신 설정, 중복 알림 병합(Aggregation) | `notifications`, `notification_preferences` |
| **`search`** | OpenSearch 색인 작업 관리, 다국어 쿼리 빌더, 자동완성, 검색어 통계, 품질 시그널 반영 | `search_index_jobs`, `search_queries`, `search_click_events`, `content_quality_signals` |
| **`seo`** | 도메인별 Sitemap/Sitemap Index 생성, RSS/Atom 피드, Canonical/Robots 정책, IndexNow 푸시 | (Read-only projection from `blogs`, `posts`, `blog_domains`) |
| **`moderation`** | 사용자 신고 접수, 스팸 자동 탐지, 모더레이션 케이스, 제재 조치, 이의신청(Appeals) | `user_reports`, `moderation_cases`, `moderation_actions`, `appeals`, `spam_detection_results` |
| **`monetization`** | 광고주/캠페인/소재/노출·클릭, 블로거 수익 배분, 유료 멤버십/단건 구매, 후원, 결제, 복식부기 정산 원장 | `ad_accounts`, `ad_campaigns`, `ad_placements`, `ad_events`, `memberships`, `user_subscriptions`, `orders`, `payments`, `payouts`, `ledger_entries` |
| **`admin`** | 시스템 설정, Feature Flag, 운영 공지, 감사 로그, 백그라운드 잡 모니터링 | `audit_logs`, `feature_flags`, `system_settings`, `outbox_events`, `background_jobs` |
| **`treescan`** | TreeScan Search API 어댑터, OIDC 계정 연동 토큰 교환, 공개 콘텐츠 색인 피드 제공 | `treescan_sync_state` (및 `outbox_events` 구독) |

## 2. 프론트엔드 모노레포 컴포넌트 구조 (Turborepo / pnpm workspace)

```
treeblog-frontend/
├── apps/
│   ├── user-web/              # Nuxt 3 SSR: 블로그 탐색 + 개별 블로그 렌더링 + SEO 엔드포인트
│   ├── blog-console/          # Vue 3 + Vite: 블로그 관리 센터 + Tiptap 글쓰기 스튜디오
│   └── admin-console/         # Vue 3 + Vite: 플랫폼 통합 관리자 콘솔
└── packages/
    ├── ui-kit/                # 공통 디자인 시스템 (Button, Modal, DataTable, Badge, Form, Toast)
    ├── editor-core/           # Tiptap/ProseMirror 커스텀 확장 노드, 툴바, 슬래시 메뉴, Markdown 변환기
    ├── api-client/            # OpenAPI 3.1.0 기반 자동 생성 TypeScript Fetch/Axios 클라이언트 & Zod 스키마
    ├── theme-renderer/        # 블로그 테마 토큰 주입기 및 안전한 블록 콘텐츠 렌더러
    └── shared-utils/          # 다국어(i18n), 날짜/시간대(Intl), 권한 체크 헬퍼, 상수 정의
```

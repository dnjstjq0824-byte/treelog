# B_BLOGGING_EXPERIENCE — Blogging Experience (블로그 개설 및 운영 경험)
> **부제**: 블로그 개설, 블로그 주소, 개인·기업 블로그, 프로필, 카테고리, 글 관리, 블로그 운영 설정 및 전체 사용자 경험
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-B_BLOGGING_EXPERIENCE`
> **적용 조항**: 제3조 (B_BLOGGING_EXPERIENCE 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- 회원가입 직후 누구나 30초 이내에 자신의 목적(일상 기록, 전문 기술 블로그, 기업 공식 블로그)에 맞는 블로그를 개설하고 운영할 수 있는 직관적이고 강력한 블로그 운영 경험을 제공한다.
- 단일 계정이 여러 개의 독립된 블로그(개인용, 부캐용, 브랜드용)를 개설하거나 팀 블로그에 참여할 수 있는 멀티 블로그(Multi-Blog per User) 구조를 지원한다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **개인 블로거 (`BLOG_OWNER`)**: 개인 블로그를 개설하고 블로그 제목, 한줄 소개, 파비콘, 대표 이미지, 2단계 카테고리 트리, 기본 공개 범위를 설정한다.
- **기업 블로그 관리자 (`BLOG_ADMIN`)**: 기업용 블로그를 생성하고 마케터/작성자를 `BLOG_EDITOR` 또는 `BLOG_AUTHOR`로 초대하여 역할을 부여한다.
- **독자 (`ANON` / `USER`)**: 블로그 홈에 방문하여 공지사항, 대표글(Featured Posts), 카테고리 목록, 시리즈, 프로필, 방명록을 탐색한다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-BLOG-001 (블로그 개설 및 핸들 부여)**: 사용자는 계정당 기본 최대 5개의 블로그를 개설할 수 있다. 각 블로그는 고유 `handle`(3~30자, `^[a-z0-9][a-z0-9-]{1,28}[a-z0-9]$`)을 가지며 시스템 예약어(`admin`, `api`, `www`, `mail`, `help`, `treescan`, `login`, `static` 등 150개 목록)와 중복될 수 없다.
- **REQ-BLOG-002 (개인·기업 블로그 유형)**: `blog_type`은 `PERSONAL`, `PROFESSIONAL`, `CORPORATE`, `COMMUNITY`를 지원하며, 기업/커뮤니티 블로그는 다중 멤버(`blog_members`) 및 원고 승인 옵션(`require_approval_for_authors`)을 활성화할 수 있다.
- **REQ-BLOG-003 (계층형 카테고리 및 시리즈)**: 블로그별로 최대 2뎁스(상위-하위) 카테고리를 무제한(권장 100개 이하) 생성하고 드래그앤드롭으로 순서를 변경할 수 있다. 카테고리별 공개/비공개 및 목록 스타일(카드형/목록형/갤러리형)을 지정할 수 있다.
- **REQ-BLOG-004 (공지사항 및 대표글 고정)**: 블로그 상단에 최대 5개의 공지 게시글(`is_notice = true`)과 최대 3개의 대표 고정글(`is_pinned = true`)을 배치할 수 있다.
- **REQ-BLOG-005 (방명록 및 블로그 소개)**: 블로그 소유자 설정에 따라 방명록(`guestbook_entries`, 비밀글 지원)과 상세 소개 페이지(`/about`)를 활성화/비활성화할 수 있다.

## 4. 데이터 구조 (Data Structures & Domain Models)
```go
type Blog struct {
    ID               uuid.UUID `json:"id"`
    OwnerUserID      uuid.UUID `json:"owner_user_id"`
    Handle           string    `json:"handle"`
    BlogType         string    `json:"blog_type"` // PERSONAL, PROFESSIONAL, CORPORATE, COMMUNITY
    Title            string    `json:"title"`
    Subtitle         string    `json:"subtitle"`
    Description      string    `json:"description"`
    LogoMediaID      *uuid.UUID `json:"logo_media_id,omitempty"`
    CoverMediaID     *uuid.UUID `json:"cover_media_id,omitempty"`
    DefaultLanguage  string    `json:"default_language"` // e.g., "ko-KR", "en-US"
    Timezone         string    `json:"timezone"`         // e.g., "Asia/Seoul"
    Status           string    `json:"status"`           // ACTIVE, SUSPENDED, CLOSING, CLOSED
    SubscriberCount  int64     `json:"subscriber_count"`
    PostCount        int64     `json:"post_count"`
    CreatedAt        time.Time `json:"created_at"`
}
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- `apps/user-web`: 블로그 홈(`SCR-BLOG-001`), 카테고리별 글 목록(`SCR-BLOG-004`), 블로그 소개(`SCR-BLOG-009`), 방명록(`SCR-BLOG-008`)을 SSR로 렌더링.
- `apps/blog-console`: 블로그 개설 위저드(`SCR-MEM-005`), 블로그 기본 설정(`SCR-MEM-006`), 카테고리 트리 편집기(SortableJS 기반 드래그앤드롭), 팀 멤버 초대/권한 관리 UI.

## 6. 백엔드 로직 (Backend Logic & Processing)
- `internal/modules/blog`:
  - 핸들 유효성 및 예약어 검증기(`ValidateBlogHandle`).
  - 블로그 개설 시 단일 트랜잭션으로 `blogs`, `blog_members`(Owner 등록), `blog_settings`(기본 설정), `blog_themes`(기본 테마 연결), `blog_categories`('미분류' 기본 카테고리)를 원자적으로 생성.
  - 카테고리 삭제 시 소속 게시글을 일괄 '미분류(Uncategorized)'로 이동하거나 다른 카테고리로 이전하는 안전 로직 구현.

## 7. API 계약 (API Contracts)
- `POST /api/v1/blogs` (신규 블로그 개설)
- `GET /api/v1/blogs/check-handle?handle={handle}` (블로그 주소 중복 및 예약어 실시간 검사)
- `GET /api/v1/blogs/{blogId}` / `PATCH /api/v1/blogs/{blogId}` (블로그 정보 조회 및 수정)
- `GET /api/v1/blogs/{blogId}/settings` / `PUT /api/v1/blogs/{blogId}/settings` (운영 설정 변경)
- `GET /api/v1/blogs/{blogId}/categories` / `POST /api/v1/blogs/{blogId}/categories` / `PUT /api/v1/blogs/{blogId}/categories/reorder` (카테고리 관리)
- `GET /api/v1/blogs/{blogId}/members` / `POST /api/v1/blogs/{blogId}/members/invite` (팀 블로그 멤버 관리)

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- 주요 테이블: `blogs`, `blog_members`, `blog_settings`, `blog_categories`, `blog_navigation`, `series`, `guestbook_entries`.
- 제약조건: `blogs.handle`에 대소문자 구분 없는 고유 인덱스(`UNIQUE (LOWER(handle))`), `blog_categories`에 `(blog_id, slug)` UNIQUE 제약 적용.

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **블로그 상태**: `ACTIVE` → `SUSPENDED` (운영 정책 위반 시) / `CLOSING` (소유자 폐쇄 신청 후 30일 유예기간) → `CLOSED` (영구 폐쇄).
- **예외 처리**: 이미 게시글이 존재하는 카테고리를 삭제 시도할 때 `target_category_id`가 지정되지 않으면 기본 미분류 카테고리로 자동 재매핑하여 게시글 유실을 방지한다.

## 10. UX/UI 구조 (UX/UI Structure)
- 네이버 블로그의 친숙한 좌측/상단 프로필+카테고리 구조와 미디엄/티스토리의 깔끔한 매거진형 헤더 중 선택 가능.
- 모바일(320px~767px)에서는 카테고리 트리와 프로필이 슬라이드 드로어(Bottom Sheet / Side Drawer)로 자동 전환된다.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- 팀 블로그 멤버 초대는 암호학적 난수 토큰(유효기간 48시간) 이메일 발송 및 수락 절차를 거친다.
- 방명록 비밀글(`is_secret = true`)은 작성자 본인과 해당 블로그의 `OWNER`/`ADMIN`에게만 조회 권한이 부여된다.

## 12. 성능 및 확장성 (Performance & Scalability)
- 블로그 메타데이터, 설정, 카테고리 트리는 Redis에 `blog:meta:{blogId}` 키로 캐싱(TTL 10분, 설정 변경 시 즉시 무효화)하여 매 페이지 렌더링 시 DB 부하를 최소화한다.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- Backend: `github.com/gosimple/slug` (다국어 슬러그 정규화).
- Frontend: `@vueuse/core`, `@formkit/drag-and-drop` (카테고리 순서 편집).

## 14. 개발 우선순위 (Development Priority)
- **Phase 2 (핵심 선행)**: 회원/인증 직후 구현해야 하는 최우선 코어 모듈 (`TB-BLOG-001`, `TB-BLOG-002`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 선행: `I_IDENTITY_AUTHORIZATION` (`users` 테이블 및 인증 세션).
- 후행: `C_CONTENT_EDITOR`, `D_DOMAINS_DESIGN`, `E_ENGAGEMENT_COMMUNITY`.

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-BLOG-001`: 예약어(`admin`, `api` 등) 또는 잘못된 형식(`-abc`, `ab`)으로 블로그 개설 시 `400 INVALID_BLOG_HANDLE` 반환 검증.
- `TEST-BLOG-002`: 기업 블로그에서 `BLOG_AUTHOR` 권한 멤버가 블로그 기본 설정 변경 시도 시 `403 FORBIDDEN` 차단 검증.
- `TEST-BLOG-003`: 2단계 카테고리 순서 변경 및 상위 카테고리 삭제 시 하위 게시글 안전 이동 검증.

## 17. Acceptance Criteria (인수 기준)
1. 신규 회원이 블로그 개설 화면에서 핸들 중복 확인 후 1초 이내에 블로그를 생성하고 즉시 `/{handle}` 주소로 접속할 수 있어야 한다.
2. 카테고리 생성/수정/순서변경/삭제가 블로그 사이드바 및 네비게이션에 즉시 반영되어야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- 블로그 개설 E2E 테스트 영상/스크린샷, `blogs`/`blog_settings`/`blog_categories` DB 레코드 조회 로그.

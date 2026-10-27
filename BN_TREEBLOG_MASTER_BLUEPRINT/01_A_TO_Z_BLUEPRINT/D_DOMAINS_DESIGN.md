# D_DOMAINS_DESIGN — Domains & Design Customization (도메인 및 디자인 커스터마이징)
> **부제**: 블로그 스킨, 테마, 템플릿, 사용자 서브도메인, 개인 도메인, HTTPS, DNS 연결, 레이아웃 커스터마이징
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-D_DOMAINS_DESIGN`
> **적용 조항**: 제3조 (D_DOMAINS_DESIGN 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- 블로거가 플랫폼 기본 주소뿐만 아니라 자신만의 독립 브랜드 도메인(서브도메인, 경로형, 개인 소유 커스텀 도메인)을 무료 자동 HTTPS(TLS)와 함께 운영할 수 있도록 지원한다.
- 안전한 디자인 토큰 기반 테마 편집기와 단계적 고급 CSS 커스터마이징을 통해 보안 위험(악성 스크립트 주입) 없이 높은 디자인 자유도를 제공한다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **개인 블로거**: 기본 제공되는 8종의 공식 반응형 테마(Magazine, Minimal, TechLog, PhotoGrid, Editorial, ClassicNaverStyle, DarkStudio, Portfolio) 중 하나를 선택하고 포인트 컬러, 폰트, 헤더 레이아웃, 사이드바 위젯 구성을 실시간 미리보기로 변경한다.
- **전문가/기업 블로거**: 가비아/Cloudflare/Route53 등에서 구매한 `blog.mycompany.com` 도메인을 등록하고 DNS 안내에 따라 TXT/CNAME 레코드를 설정하여 5분 내에 HTTPS가 적용된 독립 도메인 블로그로 전환한다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-DOM-001 (3종 주소 체계 지원 및 Canonical 단일화)**:
  1) 플랫폼 경로형: `{base_domain}/@{handle}`
  2) 플랫폼 서브도메인형: `{handle}.{base_domain}`
  3) 사용자 개인 도메인: `blog.user-owned-domain.com` (Apex 도메인 `A`/`ALIAS` 및 서브도메인 `CNAME` 모두 지원)
  - 블로그마다 `primary_domain_type`을 1개 지정하며, 나머지 주소로 유입된 요청은 `301 Moved Permanently`로 대표 주소로 리다이렉트하거나 `<link rel="canonical">`을 대표 주소로 고정하여 검색엔진 중복 문서 패널티를 원천 방지한다.
- **REQ-DOM-002 (DNS 소유권 검증 및 자동 ACME TLS)**:
  - 도메인 등록 시 `_treeblog-verify.{domain}` TXT 레코드 값을 발급한다.
  - 소유권이 `VERIFIED`된 도메인만 역방향 프록시(Caddy/Cert-Manager)의 On-Demand TLS 허용 목록에 등록되어 Let's Encrypt 인증서를 자동 발급하고 만료 30일 전 자동 갱신한다.
- **REQ-DSN-001 (제7조 19대 테마·디자인 필수 기능)**:
  - 기본 스킨, 테마 변경, 시스템 연동/수동 다크모드, 색상 팔레트 설정, 타이포그래피(본문/제목 폰트, 줄간격, 글자 크기), 블로그 대표 커버 이미지, 프로필 레이아웃(상단형/좌측 사이드바형/미니멀형), 게시글 카드 스타일, 목록형/갤러리형/피드형 레이아웃, 사이드바 위젯(프로필, 카테고리, 최근글, 인기글, 태그 클라우드, 이웃 목록, 검색창), 카테고리 표시 옵션, 상단 커스텀 메뉴 편집, 푸터 편집, 모바일/데스크톱 실시간 미리보기, 테마 설정 저장, 기본값 원클릭 복원.
- **REQ-DSN-002 (보안 격리된 커스텀 CSS)**:
  - 사용자가 임의의 서버 사이드 코드나 `<script>`를 실행하는 것을 엄격히 금지한다.
  - 커스텀 CSS 입력 시 서버 측 CSS AST 파서가 `@import` 외부 악성 URL, `expression()`, `javascript:` 바인딩을 제거하고 블로그 루트 컨테이너(`.tb-blog-theme-root`) 하위로 스코핑하여 플랫폼 공통 헤더나 보안 UI(신고 버튼, 유료 광고 표시 등)를 숨기거나 위조(UI Redressing)하지 못하도록 차단한다.

## 4. 데이터 구조 (Data Structures & Domain Models)
```go
type BlogDomain struct {
    ID                 uuid.UUID  `json:"id"`
    BlogID             uuid.UUID  `json:"blog_id"`
    DomainName         string     `json:"domain_name"`         // "blog.acme.com"
    DomainType         string     `json:"domain_type"`         // "SUBDOMAIN", "CUSTOM_CNAME", "CUSTOM_APEX"
    IsPrimary          bool       `json:"is_primary"`
    VerificationToken  string     `json:"verification_token"`  // "tb-verify=8f9a2b..."
    VerificationStatus string     `json:"verification_status"` // "PENDING_DNS", "VERIFIED", "FAILED", "SUSPENDED"
    TLSStatus          string     `json:"tls_status"`          // "NONE", "PROVISIONING", "ACTIVE", "RENEWAL_FAILED"
    LastVerifiedAt     *time.Time `json:"last_verified_at"`
}
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- `packages/theme-renderer`: JSON 기반 테마 설정(`theme_config`)을 CSS Custom Properties(`--tb-color-primary`, `--tb-font-heading`, `--tb-layout-width` 등)로 변환하여 SSR 시점부터 깜빡임(FOUC) 없이 주입.
- `apps/blog-console`: 좌측 설정 컨트롤러 + 우측 반응형 iframe 프리뷰어(375px Mobile / 768px Tablet / 1440px Desktop 전환)로 구성된 비주얼 테마 스튜디오(`SCR-MEM-006`).

## 6. 백엔드 로직 (Backend Logic & Processing)
- `internal/modules/domain`:
  - `net.Resolver`를 사용한 비동기 DNS TXT/CNAME/A 레코드 검증 워커.
  - 도메인 변경 시 구 도메인/구 슬러그 요청을 신 도메인으로 `301` 리다이렉트하는 라우팅 테이블 관리.
  - 도메인 탈취 방지: 이미 다른 활성 블로그에서 `VERIFIED` 상태로 사용 중인 도메인은 새 검증 토큰이 확인되기 전까지 절대 소유권을 이전하지 않는다.

## 7. API 계약 (API Contracts)
- `GET /api/v1/blogs/{blogId}/domains` / `POST /api/v1/blogs/{blogId}/domains` (도메인 목록 및 신규 등록)
- `POST /api/v1/blogs/{blogId}/domains/{domainId}/verify` (DNS 소유권 즉시 검증 트리거)
- `PUT /api/v1/blogs/{blogId}/domains/{domainId}/primary` (대표 Canonical 도메인 지정)
- `DELETE /api/v1/blogs/{blogId}/domains/{domainId}` (도메인 연결 해제)
- `GET /api/v1/themes` (공식 테마 카탈로그 조회)
- `GET /api/v1/blogs/{blogId}/theme` / `PUT /api/v1/blogs/{blogId}/theme` (블로그 테마 및 커스텀 설정 저장)
- `POST /api/v1/blogs/{blogId}/theme/reset` (테마 설정 기본값 복원)

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- 주요 테이블: `blog_domains`, `themes`, `blog_themes`, `blog_navigation`.
- 제약조건: `blog_domains.domain_name`에 `UNIQUE (LOWER(domain_name))` 적용, 블로그당 `is_primary = true`는 Partial Unique Index(`UNIQUE (blog_id) WHERE is_primary = true`)로 단 1개만 허용.

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **도메인 상태 전이**: `PENDING_DNS` → `VERIFIED` → (TLS 발급 성공) → `TLS_ACTIVE`. 만약 7일 이내 미검증 시 `EXPIRED` 처리, 운영 중 DNS 레코드가 제거되어 24시간 이상 연속 실패 시 `SUSPENDED`로 전환하고 블로거에게 긴급 이메일/알림 발송 후 기본 서브도메인으로 자동 폴백.

## 10. UX/UI 구조 (UX/UI Structure)
- DNS 설정에 익숙하지 않은 사용자를 위해 주요 DNS 제공업체(Cloudflare, Route53, 가비아, 카페24, Namecheap)별 복사-붙여넣기 가이드 모달과 실시간 DNS 전파 상태 표시기를 제공한다.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- **Subdomain Takeover 원천 차단**: TXT 소유권 토큰 검증 없는 CNAME 단독 연결을 금지한다.
- **Custom Domain Cookie 격리**: 개인 도메인(`blog.user.com`)에서는 플랫폼 루트 세션 쿠키가 직접 전송되지 않으므로, 읽기 전용 공개 뷰는 무세션으로 즉시 서빙하고 로그인/댓글/공감 등 인증 액션은 OAuth PKCE 기반 Silent Token Exchange 또는 보안 위젯 프레임으로 안전하게 처리한다.

## 12. 성능 및 확장성 (Performance & Scalability)
- 호스트명(`Host` 헤더) → `blog_id` 및 테마 설정 매핑은 Redis + 로컬 인메모리 LRU 캐시(TTL 60초, Pub/Sub 즉시 무효화)를 사용하여 도메인 라우팅 오버헤드를 1ms 미만으로 유지한다.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- Backend: `github.com/caddyserver/certmagic` (또는 Caddy Admin API 연동), `github.com/miekg/dns` (정밀 DNS 레코드 조회), `github.com/tdewolff/parse/v2/css` (커스텀 CSS 파싱 및 안전성 검증).

## 14. 개발 우선순위 (Development Priority)
- **Phase 2 (도메인 기초)** 및 **Phase 4 (테마 스튜디오 및 커스텀 도메인 완성)** (`TB-BLOG-003`, `TB-THEME-001`, `TB-THEME-002`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 선행: `B_BLOGGING_EXPERIENCE`, `H_HOSTING_INFRASTRUCTURE`.
- 후행: `I_IDENTITY_AUTHORIZATION` (멀티 도메인 인증), `O_OPENSEARCH` / `SEO` (도메인별 Sitemap/Canonical).

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-DOM-001`: TXT 검증 레코드가 없는 상태에서 커스텀 도메인 검증 호출 시 `422 DNS_TXT_NOT_FOUND` 반환 및 TLS 발급 차단 확인.
- `TEST-DOM-002`: 대표 도메인을 `blog.custom.com`으로 설정한 뒤 `handle.treeblog.example.com/posts/1` 접속 시 `301` 리다이렉트 및 Canonical URL이 `https://blog.custom.com/posts/1`로 출력되는지 검증.
- `TEST-DSN-001`: 커스텀 CSS에 `background: url("javascript:alert(1)")` 입력 시 Sanitizer가 해당 속성을 제거하는지 검증.

## 17. Acceptance Criteria (인수 기준)
1. 서브도메인, 경로형, 개인 도메인 3가지 방식 모두에서 블로그가 정상 렌더링되고 대표 주소로 Canonical이 통일되어야 한다.
2. 테마 스튜디오에서 색상/폰트/레이아웃 변경 후 저장 시 1초 이내에 실제 블로그에 반영되고 '기본값 복원'이 정상 동작해야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- DNS Mock 서버 연동 도메인 검증 통합 테스트 로그, 테마 변경 전/후 Playwright 반응형 스크린샷, 리다이렉트 헤더(`curl -I`) 출력 로그.

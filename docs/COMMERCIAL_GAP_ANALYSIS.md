# TreeBlog 상용급(Tistory·WordPress급) 갭 분석과 로드맵

작성 2026-11-24. 블루프린트 158개 문서 전수 점검과 저장소 실측을 바탕으로,
현재 구현이 실제 상용 블로그 플랫폼 수준에 도달하기 위해 필요한 잔여 과제를
우선순위 순으로 정리한다. 모든 현황 서술은 파일 근거를 달았으며, 미확인 항목은
`미확인`으로 표기했다.

관련 문서: [IMPLEMENTATION_STATUS.md](../IMPLEMENTATION_STATUS.md),
[BLUEPRINT_ERRATA.md](./BLUEPRINT_ERRATA.md).

## 기준선 요약

- 백엔드: 16개 모듈·177개 핸들러·마이그레이션 0001~0011·테스트 44개로 골격 완성.
  `go vet` 클린, `go test` 26개 패키지 전체 통과.
- 프론트엔드: user-web 19페이지, blog-console 15개·admin-console 7개 기능 라우트 배선.
- 상용 출시를 막는 결함(P0)은 보안 정화·피드/SEO·라우트 규격·예약 정밀도에 집중돼 있다.

## P0 — 출시 차단 항목

### G-01 발행 본문 서버측 HTML 정화 없음 — ✅ 해소 (2026-11-24)

- 조치: `internal/pkg/security/sanitize.go`의 `SanitizePostHTML`(bluemonday UGC
  기반 + 코드 언어 클래스·이미지 속성·YouTube/Vimeo 임베드 허용)을
  `editor.CreateDraft`와 `post.PublishPost`(본문 + 발행 스냅샷)에 적용.
  `sanitize_test.go` 13케이스(스크립트·이벤트 핸들러·javascript: 링크·
  비허용 iframe·폼·SVG + 허용 콘텐츠 보존) 통과.
- 잔여: 댓글·방명록 정화는 G-08에서 같은 헬퍼 재사용 예정.

### G-02 Content-Security-Policy 미설정 — ✅ 해소 (2026-11-24)

- 조치: `internal/pkg/security/headers.go`에 JSON API용 잠금 정책
  `default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'`
  추가, `TestSecurityHeaders`에 존재 검증 추가.
- 잔여: 프론트엔드 3종 문서 CSP는 각 앱 배포 시 설정 필요.

### G-03 테마 CSS 정화가 정규식 기반 — ✅ 해소 (2026-11-24)

- 조치: `SanitizeCSS`를 douceur CSS 파서 기반으로 교체(신규 의존성 없음,
  bluemonday 경유로 이미 존재). 허용 at-rule 4종(@media·@supports·@keyframes·
  @font-face), 플랫폼 셀렉터 3종 차단, 위험 선언·URL 스킴 제거, 전 셀렉터
  `.tb-blog-theme-root` 스코프 강제, 파싱 불가 입력은 빈 문자열로 실패 닫힘.
  `sanitize_test.go` 16케이스 + 기존 `TestSanitizeCSS` 통과.

### G-04 RSS/Atom/WebSub 피드 미구현 — ✅ 해소 (2026-11-24, WebSub 제외)

- 조치: `internal/modules/seo/feed.go`에 RSS 2.0·Atom 1.0 빌더와 조회 추가.
  라우트 6종(`/b/{handle}/feed|/atom`, `/tags/{tag}/feed|/atom`,
  `/comments/feed|/atom`), 공개·발행·비휴지통·비밀 제외 필터, 본문은 정화 후
  수록, `Cache-Control: max-age=900`. 블로그 메인 페이지에 자동발견
  `<link rel=alternate>` 2종 추가. `feed_test.go` 5개 테스트 통과.
- 잔여: WebSub 허브 발행(폴백인 폴링 구독은 피드 자체로 충족), JSON Feed는 미정.

### G-05 공개 라우트 규격이 명세와 상이, 핵심 공개 화면 누락 — ✅ 부분 해소 (2026-11-24)

- 확정: `/b/{handle}` 경로형을 정본 규격으로 확정(서브도메인·커스텀 도메인
  라우팅은 후속 단계로 이월).
- 조치: slug 이력 301은 백엔드에 이미 구현돼 있었고(`GetPublicPost`→`handleError`
  Location), 이번에 `redirect_test.go`로 계약 고정. 상세 페이지에 canonical
  링크 + 구 slug 접속 시 현 slug로 client replace 추가. 레거시 `/@{handle}`
  → `/b/{handle}` 301 미들웨어(`server/middleware/legacy-blog-redirect.ts`)
  신설, 개발 서버 curl로 3건 실측 확인(`/@forest`→`/b/forest`,
  `/@forest/my-post`→`/b/forest/my-post`, `/b/forest`→200).
- 잔여: SSR 단계 서버 301(페이지는 onMounted 클라이언트 fetch), 누락 공개 화면
  충원(탐색·태그·피드·방명록 등), 서브도메인·커스텀 도메인 라우팅.

### G-06 예약 발행 정밀도 미달 — ✅ 해소 (2026-11-24)

- 조치: `cmd/worker/main.go` 스케줄 루프를 30초→`schedulerInterval = 5초`로 변경,
  명세 계약 충족. 루프는 순차 실행이라 느린 스윕이 다음 틱과 겹치지 않으며,
  SKIP LOCKED로 다중 워커 중복 처리도 방지됨. `main_test.go`에 5초 계약
  고정 테스트 + 취소 시 종료 테스트 추가.
- 잔여: 지연 발행 알림(스윕 지연 관측)은 관측성(G-18)과 함께 후속.

## P1 — 상용 품질 항목

### G-07 비밀번호 보호글 없음 — ✅ 해소 (2026-11-24)

- 현황: `internal/modules/post/service.go`의 공개범위는
  PUBLIC·UNLISTED·NEIGHBOR_ONLY·MUTUAL_ONLY·PAID_MEMBERS_ONLY·PRIVATE 6종이며
  비밀번호 입력 흐름이 없다.
- 기준: Tistory 보호글·WordPress 비밀번호 글과 동등 기능 필요.
- 수용 기준: 비밀번호 해시 저장·세션 유지·피드/검색 노출 제외.
- 조치: 마이그레이션 0012(`protection_password_hash` + visibility `PROTECTED`).
  발행 시 Argon2id 해시 저장(미제공 시 기존 유지·해제 시 NULL), 조회 시 작성자
  외 비밀번호 요구(`X-Post-Password` 헤더·`password` 쿼리), 피드·검색·사이트맵·
  목록은 PUBLIC 전용이라 자동 제외. user-web 상세에 비밀번호 게이트 +
  sessionStorage 세션 유지, blog-console 발행 모달에 PROTECTED + 비밀번호 입력.
  단위 3종 + DB 통합 흐름 테스트 통과.
- 부수 수정(최초 DB 실행에서 발견): outbox `payload_json`에 `[]byte` 전달 시
  22P02로 실패하는 4곳(post 발행/발행취소/스케줄러·미디어)을 `string()`으로 수정.

### G-08 댓글·방명록 운영 기능 부족 — ✅ 부분 해소 (2026-11-24)

- 조치: 승인·반려·고정/해제·대기목록 API 신설(`POST /comments/{id}/approve|reject|pin`,
  `DELETE .../pin`, `GET /blogs/{id}/comments/pending`), 소유자·관리자·편집자
  권한, 승인 시 댓글 수 증가, 고정 우선 정렬. 신고는 기존 moderation 신고가
  COMMENT 타깃을 지원함을 통합 테스트로 고정, 상세 페이지 댓글별 신고 버튼 추가.
- 잔여: blog-console 승인 큐 화면, 댓글 좋아요·수정 이력·비회원·금칙어·스팸 점수.

### G-09 알림 발송 경로 없음 — ✅ 부분 해소 (2026-11-24)

- 현황: `internal/modules/notification/service.go`는 목록·미읽음·읽음 처리만
  제공하며 SSE·웹푸시·이메일 발송·묶음·수신거부가 없다.
- 기준: 구독 새 글·댓글·공감 실시간 알림은 리텐션 핵심이다.
- 수용 기준: SSE 스트림 + 웹푸시(VAPID) + 이메일 다이제스트, 조용한 시간·채널 설정.
- 조치: 알림 생성 API + 프로세스 내 허브 + SSE 스트림(`GET /me/notifications/stream`,
  ready·heartbeat·연결 종료 테스트), 카테고리별 인앱 음소거 적용, 댓글(답글 포함)·
  공감·이웃·구독 생성 시 발송 연동(cmd/api 배선), 알림 페이지 EventSource 실시간
  prepend. 스트림 전용 `access_token` 쿼리 인증 미들웨어 추가.
- 잔여: 웹푸시(VAPID)·이메일 발송, 조용한 시간, 다중 인스턴스 Redis fanout.

### G-10 에디터 상용 블록 부족, 공용 코어 스텁 — ✅ 부분 해소 (2026-11-24)

- 현황: `packages/editor-core/src/index.ts` 65바이트 스텁.
  blog-console은 Tiptap v3 starter-kit + image/link/placeholder 수준이며
  표·슬래시 메뉴·접기·목차·수식·투표·파일 첨부가 없다.
- 기준: Notion급 블록 UX와 Tistory·WordPress(Gutenberg) 패리티 필요.
- 수용 기준: `editor-core` 실구현, 표·코드·인용·구분선·토글·목차 블록,
  단축키·접근성·마크다운 입출력 회귀 테스트.
- 조치: `editor-core` 실구현(AST 타입·허용목록 검증·본문 추출·단어 수, vitest 6건).
  콘솔에 표 확장(@tiptap/extension-table 3.32.3 고정)·표 툴바·저장/발행 전 검증·
  단어 수 표시 추가. 서버 정화기는 표를 허용하므로 종단 정합.
- 잔여: 슬래시 메뉴·토글·목차·수식·투표·파일 첨부, 단축키·접근성 체계화.

### G-11 자동저장 오프라인 복구 없음 — ✅ 해소 (2026-11-24)

- 현황: `apps/blog-console/src/views/EditorView.vue`는 2초 디바운스 서버 직접
  저장 + 충돌 감지만 하며 IndexedDB·로컬 캐시·오프라인 배지가 없다.
- 기준: Tistory·WordPress·Notion은 오프라인 초안을 로컬에 보존한다.
- 수용 기준: IndexedDB 2계층 저장, 재접속 병합, 오프라인 배지.
- 조치: `lib/drafts.ts` 2계층 캐시(IndexedDB + 메모리 폴백, 스냅샷 격리, 복원
  판정, vitest 2건). 에디터에 1초 쓰로틀 로컬 저장·온라인 복귀 시 서버 저장·
  오프라인/로컬초안 배지·신규 로드 시 복원·발행 시 정리 추가.

### G-12 미디어 파생본 JPEG 전용, 서명 URL 없음 — ✅ 부분 해소 (2026-11-24)

- 현황: `internal/modules/media/processor.go`는 gif/jpeg/png/webp를 디코드해
  JPEG 파생본을 만들며(EXIF 제거 포함), WebP/AVIF 출력·반응형 srcset·서명 URL·
  고아 수집기가 없다.
- 기준: Core Web Vitals(LCP < 2.0s, 명세 요구) 달성에 차세대 포맷이 사실상 필수다.
- 수용 기준: WebP(필수)·AVIF(권장) 파생, srcset, 비공개 원본 서명 URL, 핫링크 정책.
- 조치: pure-Go 무손실 WebP 인코더 도입, JPEG보다 작을 때만 병행 방출(실측 규칙,
  단위 테스트 고정). 포맷 협상(`?format=` + Accept 폴백), 파생본 목록 API(srcset
  구성용), 작성자 전용 원본 서명 다운로드, Referer 핫링크 정책(당사 host 허용).
  콘솔 썸네일 파생본 사용 + srcset 복사 버튼.
- 잔여: 손실 WebP·AVIF(cgo 없이 인코더 부재, 프로덕션 빌드가 CGO_ENABLED=0),
  고아 수집기, 원본 보관 정책.

### G-13 유료 차단 미리보기 불일치 — ✅ 해소 (2026-11-24)

- 현황: 명세는 AST 절삭 기반 차단을 정의하나 실제 멤버십 가림은 문자 미리보기
  수준이다(1차 점검 확인, `internal/modules/monetization/memberships.go` 대조 필요).
- 기준: 유료 분량은 피드·검색·HTML 소스에 노출되면 안 된다.
- 수용 기준: 서버 렌더 시점에 AST 절삭 후 정화된 미리보기만 반환, 소스 노출 회귀 테스트.
- 조치: `CutPaywallPreview`로 선행 2블록·500자 절삭 + 이스케이프 렌더러 + 정화기
  2차 통과, 본문 평문 제거(발췌는 티저로 유지). 기존 미이스케이프 삽입 XSS도 함께
  제거. 4개 단위 테스트(경계·문자절삭·이스케이프·축퇴 입력).

### G-14 검색 제안 인덱스·스팸 운영 없음 — ✅ 부분 해소 (2026-11-24)

- 현황: `Suggest`는 `search_queries` 테이블 SQL 조회이며 별도 suggest 인덱스·
  동의어 사전·스팸 신고·인기 급상승 운영 도구가 없다.
  랭크 검색·SQL 폴백·클릭로깅·SimHash는 구현돼 있다.
- 기준: 상용 검색은 오타 교정·동의어·어뷰징 대응을 운영 프로세스로 가진다.
- 수용 기준: OpenSearch suggester 전환(또는 SQL 고도화 결정 기록), 금칙어·신고 루프.
- 조치: 결정 = SQL 고도화 유지(제안·인기는 SQL 집계로 충분, OpenSearch는 랭크
  검색 전담). 스팸 용어 차단/해제·목록 API(운영자 전용) + 인기 검색어 API +
  제안 정렬 안정화. DB 통합 테스트로 차단 루프 고정.
- 잔여: 동의어 사전 운영, 오타 교정 고도화, OpenSearch suggester 전환(규모 시).

### G-15 SEO 구조화 데이터·헤더 미흡 — ✅ 해소 (2026-11-24)

- 현황: BlogMeta·Sitemap·Robots는 있으나 포스트 단위 JSON-LD(BlogPosting·
  BreadcrumbList)·Canonical·`noindex` 분기·410 Gone·영구 이동 처리가 미흡하다.
- 기준: WordPress SEO 플러그인 수준의 메타 완결성 필요.
- 수용 기준: 화면별 메타 매트릭스 테스트, 비공개/삭제 글 헤더 회귀 테스트.
- 조치: 게시글 크롤러 메타 API(canonical·robots·BlogPosting+BreadcrumbList JSON-LD),
  가시성 매트릭스(공개 index·제한 noindex·비공개 404), 삭제 글 410 Gone.
  상세 페이지에 robots·JSON-LD 주입 + 410 전용 상태. 단위 2종 + 통합 2종.

### G-16 OpenAPI 코드젠 파이프라인 없음 (TB-API-001) — ✅ 부분 해소 (2026-11-24)

- 현황: 113 paths / 125 operations 계약이 있으나 생성 클라이언트 없이
  수동 `lib/api.ts`를 사용한다. 계약 변경 시 프론트 드리프트가 발생한다.
- 수용 기준: CI에서 openapi.yaml→TS 클라이언트 생성·타입 검사 실패 시 빌드 차단.
- 조치: 명세에 신규 17경로 추가(128 paths / 142 ops) + `packages/api-client`
  신설(openapi-typescript 타입 생성, 인증·토큰갱신·요청ID·CSRF·멱등키 래퍼,
  6건 테스트) + CI 코드젠 드리프트 게이트(재생성 후 diff 검사).
- 잔여: 기존 수동 클라이언트의 생성 클라이언트 전환, 백엔드 요청/응답 검증기
  (`internal/platform/validator/`, TB-API-001 §7).

### G-17 이웃·구독·차단 관계 미완성 — ✅ 부분 해소 (2026-11-24)

- 현황: 팔로우 골격은 있으나 이웃 신청/수락·상호이웃·그룹 폴더·즐겨찾기·
  차단 사용자 관리가 미흡하다.
- 기준: Naver/Tistory 이웃·서로이웃 패리티가 커뮤니티 핵심이다.
- 수용 기준: 관계 상태머신 + 상한(인당 이웃 수 등) + 모아보기 피드.
- 조치: 이웃 신청/수락/거절 + 상호이웃 자동 승급, 폴더 CRUD·이동·즐겨찾기,
  인당 1000·폴더 20 상한, 구독 일시정지/재개·채널 토글, 팔로우+구독 모아보기
  피드(커서 페이지). 구독 페이지에 일시정지 UI. DB 통합 테스트.
- 잔여: 이웃 관리 페이지·모아보기 UI, 멘션·유료등급·수신거부 세분화.

## P2 — 확장·운영 항목

### G-18 관측성 없음

- 현황: `go.mod`에 OpenTelemetry가 없고 분산 추적·메트릭·구조화 중앙 로그가 없다.
- 수용 기준: OTel 추적 + RED 메트릭 + JSON 로그 파이프라인, SLO 대시보드.

### G-19 단일 노드 전제 인프라

- 현황: Compose 8서비스 단일 노드 구성. 다중 AZ·읽기 복제본·오토스케일·CDN·
  WAF·백업/PITR·복구 목표가 문서 전제로만 존재한다.
- 수용 기준: 스테이지/운영 분리, 관리형 Postgres·Redis, R2/S3, 백업 복원 훈련 기록.

### G-20 수익화 미완성 (광고·정산·정기결제)

- 현황: 멤버십·지급·Toss 웹훅 골격만 있으며 광고 송출·무효 트래픽 판정·
  일일 정산·환불·정기결제(빌링키)·독촉·세금계산서가 없다.
- 수용 기준: Phase 8 범위 확정 후 광고/멤버십 중 하나를 종단 완성.

### G-21 테마 미리보기·버전·위젯 API 없음

- 현황: 6종 시드와 조회·변경·초기화만 있으며 실시간 미리보기·버전 복원·
  자식 테마·위젯 API가 없다.
- 수용 기준: 미리보기 렌더 + 버전 스냅샷 + 위젯 슬롯 규격.

### G-22 데이터 이동성–가져오기/내보내기 없음

- 현황: Tistory·WordPress·Blogger 가져오기, WXR/HTML 내보내기, 백업 ZIP이 없다.
- 기준: WordPress WXR는 사실상 표준이다.
- 수용 기준: WXR 내보내기 → Tistory/WordPress 가져오기 순으로 구현.

### G-23 접근성·다국어·준법 운영 항목

- 현황: 키보드·스크린리더·고대비·인쇄·단축키 안내, 다국어 확장, 72시간 침해
  대응·국외 이전·연령 게이트·투명성 보고가 미비하다.
- 수용 기준: WCAG 2.2 AA 자가 점검표, PII 보관표, 약관·동의 버전 관리.

## 미확인 항목 (후속 점검 필요)

1. 발행 상태 전이 가드·410 Gone·slug 이력 구현 여부.
2. user-web SSR 권한 검사·`X-Robots-Tag`·사이트맵 캐싱 구현 여부.
3. Caddy/reverse-proxy + ACME 자동 TLS 실배선 여부(`FF_CUSTOM_DOMAINS_ACME`는 선언만 확인).
4. 영구 삭제 배치·휴지통 복원 범위·미완료 멀티파트 정리(24시간) 실행 여부.
5. 격리 글의 검색 색인 제거·복원 연동 여부.
6. 비밀번호 재설정·이메일 인증 메일 실제 발송 경로(SMTP 설정은 존재).

## 권장 실행 순서

1. P0 G-01~G-06 (보안·피드·주소·예약) — 출시 전제 조건.
2. G-16 코드젠 파이프라인 — 이후 모든 API 작업의 드리프트 방지.
3. P1 G-07~G-11 (보호글·댓글·알림·에디터·자동저장) — 블로깅 핵심 루프.
4. P1 G-12~G-17 (미디어·유료차단·검색·SEO·관계) — 콘텐츠 경쟁력.
5. Compose 전 구간 검증 + E2E + 독립 감사 후 모듈별 `VERIFIED` 승격.
6. P2 G-18~G-23 (관측·인프라·수익화·테마·이식·준법) — 스케일 단계.

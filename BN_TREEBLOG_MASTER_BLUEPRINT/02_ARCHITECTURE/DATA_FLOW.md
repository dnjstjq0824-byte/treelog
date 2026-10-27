# DATA_FLOW.md — 핵심 시나리오별 엔드투엔드 데이터 흐름

## 1. 게시글 작성 · 자동저장 · 공개 발행 데이터 흐름

```
[작성자 브라우저 (Tiptap Editor)]
  │
  ├─ 1. 타이핑 중 (3초 디바운스) ──► IndexedDB 로컬 스냅샷 저장 (네트워크 단절 대비)
  │
  ├─ 2. 자동저장 주기 (15초) ──► PUT /api/v1/blogs/{blogId}/posts/{postId}/autosave
  │                                  │
  │                                  ▼
  │                         [Go API: editor module]
  │                          - Tiptap JSON 스키마 유효성 검증
  │                          - `autosave_drafts` 테이블 UPSERT (Optimistic Lock `version` 체크)
  │
  └─ 3. [발행(Publish) 클릭] ──► POST /api/v1/blogs/{blogId}/posts/{postId}/publish
                                     │ (Idempotency-Key 포함)
                                     ▼
                            [Go API: post module]
                             ┌── BEGIN POSTGRES TRANSACTION ──────────────────────────────┐
                             │ 1) JSON AST -> Safe HTML 렌더링 (`bluemonday` XSS 정화)    │
                             │ 2) JSON AST -> Plain Text 추출, 단어 수/읽기 시간 계산     │
                             │ 3) `posts` 테이블 상태 `PUBLISHED`, `published_at` 갱신    │
                             │ 4) `post_revisions` 불변 스냅샷 INSERT                     │
                             │ 5) `search_index_jobs` (`UPSERT_POST`) INSERT              │
                             │ 6) `outbox_events` (`post.published.v1`) INSERT            │
                             └── COMMIT TRANSACTION ──────────────────────────────────────┘
                                     │
                                     ▼ (비동기, < 500ms 이내)
                            [Go Worker: Outbox Relay]
                             ├─► [Search Worker] ──► OpenSearch `posts_v1` 색인 갱신
                             ├─► [Cache Worker]  ──► Redis SSR 캐시 및 Sitemap 캐시 무효화
                             ├─► [Notif Worker]  ──► 구독자/이웃에게 웹/푸시/이메일 알림 발송
                             └─► [TreeScan Worker]─► (`visibility == PUBLIC` 일 때만) TreeScan Sync 이벤트 발행
```

## 2. 게시글 열람 및 접근 제어(공개/이웃공개/상호이웃공개/비공개) 데이터 흐름

```
[독자/검색봇 요청: GET https://{custom-or-sub-domain}/posts/{slug}]
  │
  ▼
[Ingress -> Nuxt 3 SSR (`apps/user-web`) -> Go API: GET /api/v1/public/blogs/{blogKey}/posts/{slug}]
  │
  ▼
[Go API: Access Control Engine]
  ├─ 1. 블로그 상태 확인: `blog.status != 'ACTIVE'` 이면 즉시 404/451 반환 (`X-Robots-Tag: noindex`)
  ├─ 2. 게시글 상태 확인: `post.status != 'PUBLISHED'` 이면 작성자/블로그 편집자 외 즉시 404 반환
  ├─ 3. 공개 범위(`post.visibility`) 평가:
  │     ├─ `PUBLIC`: 즉시 본문 + SEO 메타데이터 + JSON-LD 반환 (CDN SWR 캐시 허용)
  │     ├─ `UNLISTED` (링크 공개): 본문 반환하되 `X-Robots-Tag: noindex, nofollow` 강제 (검색/피드 제외)
  │     ├─ `NEIGHBOR_ONLY` (이웃 공개): 요청자 세션 확인 -> `follows` 테이블에서 `NEIGHBOR` 또는 `MUTUAL_NEIGHBOR` 확인
  │     ├─ `MUTUAL_ONLY` (상호이웃 공개): 요청자 세션 확인 -> `follows` 테이블에서 양방향 `MUTUAL_NEIGHBOR (ACCEPTED)` 확인
  │     ├─ `PAID_MEMBERS_ONLY` (유료 멤버십 전용): `user_subscriptions` 활성 여부 또는 단건 구매(`orders`) 확인 -> 미구매 시 티저(Preview)만 반환하고 본문은 서버에서 절삭(Truncation)
  │     └─ `PRIVATE` (비공개): 오직 작성자 및 해당 블로그의 `OWNER`/`ADMIN`/`EDITOR`에게만 반환
  └─ 4. 차단 관계 확인: 블로그 소유자가 요청자를 차단(`user_blocks`)한 경우 접근 거부 또는 상호작용 차단
```

## 3. 개인 도메인(Custom Domain) 연결 및 자동 TLS 데이터 흐름

1. 블로거가 관리 센터에서 `blog.mybrand.com` 등록 (`POST /api/v1/blogs/{blogId}/domains`).
2. 서버가 예약어/중복 여부를 검사하고 고유 검증 토큰(`_treeblog-verify.blog.mybrand.com TXT "tb-verify=..."`)을 발급하여 `blog_domains` 테이블에 `PENDING_DNS` 상태로 저장.
3. 사용자가 자신의 DNS에 `TXT` 레코드와 `CNAME`(`cname.treeblog.example.com`)을 설정 후 검증 요청.
4. 워커가 DNS Lookup을 수행하여 소유권 확인 시 `verification_status = 'VERIFIED'`, `tls_status = 'PROVISIONING'`으로 전이.
5. 첫 HTTPS 접속 또는 백그라운드 프로비저닝 시 Caddy/ACME 모듈이 `/internal/v1/domains/ask?domain=blog.mybrand.com`을 호출하여 `200 OK`를 확인한 뒤 Let's Encrypt 인증서를 자동 발급하고 `tls_status = 'ACTIVE'`로 갱신.

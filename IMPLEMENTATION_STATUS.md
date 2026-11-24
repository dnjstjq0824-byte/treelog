# TreeBlog implementation status

Updated 2026-11-24. Source: `BN_TREEBLOG_MASTER_BLUEPRINT_V1_0.zip`, SHA-256 `A08A723A5125DE4F9BD600D4B8F62F395539FED031BAF33188DF35AF95445108`.

모든 수치는 같은 날짜에 저장소에서 직접 실측한 값이다. 블루프린트 원문과의
불일치는 [BLUEPRINT_ERRATA.md](./docs/BLUEPRINT_ERRATA.md), 상용 서비스 수준까지의
잔여 과제는 [COMMERCIAL_GAP_ANALYSIS.md](./docs/COMMERCIAL_GAP_ANALYSIS.md)를 본다.

## Intake findings (유지)

- The supplied workspace was empty before extraction. The ZIP contained specifications and SQL, with no preexisting application source.
- `START_HERE.md` calls for `MANIFEST.json` and `FILE_CHECKSUMS.sha256`, but neither exists in the archive. File-level checksum verification therefore cannot be performed from the supplied package.
- The blueprint says there are 48 individual work orders; 49 `TB-*.md` work-order files are present.
- `MASTER_WORK_ORDER.md` contradicts itself: the header declares 49 work orders while section 1 still says 45.
- The architecture describes 15 domain modules in prose but names 16 in `02_ARCHITECTURE/COMPONENT_DIAGRAMS.md`. The implementation creates all 16 named boundaries.

## Measured inventory (2026-11-24 실측)

| 영역 | 실측값 | 비고 |
| --- | --- | --- |
| 블루프린트 문서 | md 158개 | `BN_TREEBLOG_MASTER_BLUEPRINT/` 재귀 집계 |
| 워크오더 파일 | `TB-*.md` 49개 | 명세서 본문의 48/45 표기와 불일치 |
| DB 마이그레이션 | `0001`~`0013` (up/down 완비) | 0012 보호글, 0013 정산 컬럼 추가 |
| `CREATE TABLE` 문 | 78개 | 명세서의 57/58 표기와 불일치 |
| Go 도메인 모듈 | 16개 | admin, blog, community, domain, editor, identity, media, moderation, monetization, notification, post, search, seo, theme, treescan, user |
| HTTP 핸들러 메서드 | 207개 | `internal/modules/*/handler.go` exported 메서드 집계 |
| Go 테스트 파일 | 63개 | 단위 + 통합(`*_integration_test.go`, `*_sql_test.go`) 포함 |
| OpenAPI 계약 | 128 paths / 142 operations | `04_API/openapi.yaml` 실측, G-16에서 신규 17경로 추가 |
| 화면 명세 | `SCR-*` 46개 | EXP 10 + BLOG 13 + MEM 15 + ADM 8, 인벤토리 제목의 "45개"와 불일치 |
| user-web 페이지 | `.vue` 19개 | 라우트는 `/b/[handle]`형, 명세의 `/@{handle}`형과 상이 |
| blog-console 라우트 | 기능 15개 + login/404 | posts, editor, media, categories, series, guestbook, members, settings, theme, domains, memberships, payout 등 |
| admin-console 라우트 | 기능 7개 + login/404 | dashboard, users, blogs, moderation(+상세), appeals, payouts |
| 에디터 의존성 | Tiptap v3 (`^3.32.3`) | 명세서의 "Tiptap v2" 표기와 불일치 |
| `packages/editor-core` | `index.ts` 65바이트 스텁 | 공용 에디터 코어 미구현 |
| Go 버전 | 1.23.0 | 명세서 "1.22+" 하한 충족 |

## Module status

판정 기준은 `START_HERE.md` Step 4 상태값을 따른다.
`VERIFIED`는 실제 PostgreSQL/OpenSearch/Redis 연동 검증과 독립 감사 통과 전까지 사용하지 않는다.

| 모듈 | 상태 | 근거 |
| --- | --- | --- |
| `identity` (인증·인가·MFA·OAuth) | `IMPLEMENTED_UNVERIFIED` | handler/service/jwt/mfa/oauth/rbac + 단위·통합 테스트 |
| `blog` (블로그·카테고리·시리즈) | `IMPLEMENTED_UNVERIFIED` | category/series 서비스 + 테스트 |
| `domain` (도메인 연결) | `IMPLEMENTED_UNVERIFIED` | handler/service + 테스트 |
| `editor` (편집·마크다운) | `PARTIAL` | 마크다운 변환 + 테스트는 있으나 공용 에디터 코어는 스텁, Tiptap 연동은 starter-kit 수준 |
| `post` (포스트 생명주기) | `IMPLEMENTED_UNVERIFIED` | 공개범위 7종(PROTECTED·Argon2id 추가) + 휴지통 + 스케줄러 + DB 통합 테스트 |
| `theme` (테마) | `PARTIAL` | 6종 시드와 조회·변경·초기화 구현, CSS 정화는 파서 기반으로 해소(G-03), 미리보기·버전 없음 |
| `media` (미디어 파이프라인) | `PARTIAL` | JPEG+WebP(작을 때만) 파생·포맷 협상·파생본 목록·원본 서명·핫링크 해소(G-12), 손실 WebP·AVIF·고아 수집 잔여 |
| `community` (댓글·반응·방명록) | `PARTIAL` | 댓글 심사(G-08) + 이웃 상태머신·폴더·상한·일시정지·모아보기 해소(G-17), 좋아요·멘션 잔여 |
| `search` (검색·추천 대용) | `PARTIAL` | 스팸 용어 차단·인기 검색어 해소(G-14), 동의어·오타교정·추천 모듈 잔여 |
| `seo` (사이트맵·robots·메타) | `PARTIAL` | 포스트 메타·JSON-LD·robots·410 해소(G-15), WebSub 잔여 |
| `notification` (알림) | `PARTIAL` | 생성 API·SSE 스트림·음소거·댓글/공감/이웃 발송 해소(G-09), 웹푸시·이메일 잔여 |
| `monetization` (멤버십·정산) | `PARTIAL` | 주문·결제·원장·지급 SQL 드리프트 수정 + DB 통합 통과, 광고·일일정산·환불·정기결제 없음 |
| `moderation` (신고·심사·이의) | `IMPLEMENTED_UNVERIFIED` | DecideAppeal 미스캔 버그 수정 + DB 통합 통과 |
| `admin` (운영 콘솔 API) | `IMPLEMENTED_UNVERIFIED` | handler/service + 단위·통합 테스트 |
| `treescan` (외부 연동 어댑터) | `IMPLEMENTED_UNVERIFIED` | handler/service + 테스트, 기본 비활성 플래그 유지 |
| `user` (프로필) | `IMPLEMENTED_UNVERIFIED` | handler/service + 테스트 |

## Frontend status

| 앱 | 상태 | 근거 |
| --- | --- | --- |
| user-web (Nuxt) | `PARTIAL` | 19페이지(인증·검색·과금·내정보·블로그 2종) 배선, 탐색·태그·피드·방명록·이웃·미디어·공지·소개·신고 화면 없음 |
| blog-console (Vue) | `PARTIAL` | 15개 기능 라우트 배선, 에디터는 starter-kit 수준, 통계·수익 상세·백업 UI 없음 |
| admin-console (Vue) | `PARTIAL` | 7개 기능 라우트 배선, 감사 로그·공지·광고 심사 UI 없음 |
| OpenAPI 코드젠 파이프라인 | `NOT_STARTED` | `TB-API-001` 해당, 수동 `lib/api.ts` 사용 중 |

## Verification performed (2026-11-24)

- `go vet ./...` — passed, 출력 없음.
- `go test ./...` — passed, 28개 패키지 전체 `ok`. `TREEBLOG_TEST_DATABASE_URL`을
  일회용 postgres:16.6 컨테이너(`treeblog_test` DB)에 연결해 통합 테스트까지
  포함한 사상 첫 전체 실행이다. 기존에 한 번도 DB와 만나지 못했던 경로에서
  실결함 3건(outbox `[]byte`→jsonb, 정산 SQL 드리프트, 이의신청 미스캔)을 찾아 수정했다.
- `go build ./...` — passed.
- 마이그레이션 전 구간 Up(0001~0013)과 Down → Up 왕복을 일회용 컨테이너에서 검증.
  (`treeblog` DB는 마이그레이터 테스트의 안전 가드가 거부하므로 `treeblog_test` 사용)
- `pnpm typecheck` — 전체 워크스페이스 통과.
- 프론트 단위 테스트 — api-client 6건, editor-core 6건, ui-kit 16건, blog-console 5건, user-web 4건 통과.
- `pnpm typecheck` — 전체 워크스페이스(6개 패키지) 통과.
- Playwright E2E와 독립 감사는 미수행이므로 `VERIFIED` 판정은 부여하지 않는다.

## Next order

1. ~~Compose 기동 후 마이그레이션 전 구간 검증~~ — 완료(일회용 컨테이너, Compose 본체는 미기동).
2. `docs/COMMERCIAL_GAP_ANALYSIS.md` P0 6건 + P1 11건 해소 완료(2026-11-24):
  G-12(WebP·서명·핫링크, 부분), G-13(AST 유료차단), G-14(검색 스팸·인기어, 부분),
  G-15(SEO 메타·410), G-16(코드젠 파이프라인, 부분), G-17(관계 상태머신, 부분).
  다음은 P2(G-18 관측성·G-19 인프라·G-20 수익화 완성·G-21 테마·G-22 이식성·G-23 접근성/준법).
3. `TB-API-001` OpenAPI 코드젠 파이프라인 도입으로 수동 API 클라이언트 제거.
4. 에디터 상용 기능(`editor-core` 실구현, 표·슬래시 메뉴·버전 비교)과 알림 발송 경로(SSE/웹푸시) 구현.
5. E2E 시나리오와 독립 감사를 통과한 모듈부터 `VERIFIED`로 승격.

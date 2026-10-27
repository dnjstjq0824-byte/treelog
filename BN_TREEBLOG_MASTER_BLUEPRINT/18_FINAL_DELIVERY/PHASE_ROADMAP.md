# PHASE_ROADMAP.md — 제22조 Phase 0 ~ Phase 10 전체 개발 로드맵 상세 명세

> **적용 조항**: 제22조 (Phase 0~10 전체 개발 로드맵)
> **필수 정의 항목**: 각 Phase별 1) 개발 목표, 2) Work Order 목록, 3) 선행조건, 4) 병렬 가능 작업, 5) 산출물, 6) 테스트, 7) 실패 대응, 8) 검증 증빙, 9) 완료 승인 기준을 완비한다.

---

## Phase 0 — 요구사항·아키텍처·저장소 기초
- **1. 개발 목표**: TreeScan 의존성 없는 독립 모노레포 구조, Docker Compose 로컬/CI 스택, PostgreSQL 기초 마이그레이션 및 Transactional Outbox 워커 프레임워크 확립.
- **2. Work Order 목록**: `TB-ARCH-001`, `TB-OPS-001`, `TB-DB-001`, `TB-ARCH-002` (4개)
- **3. 선행조건**: 블루프린트 패키지(`BN_TREEBLOG_MASTER_BLUEPRINT_V1_0.zip`) 인수 및 `START_HERE.md` 숙지.
- **4. 병렬 가능 작업**: `TB-ARCH-001` 완료 후 `TB-OPS-001`과 초기 프론트엔드 패키지 스캐폴딩 병렬 진행.
- **5. 산출물**: `cmd/api`, `cmd/worker`, `docker-compose.yml`, `.github/workflows/ci.yml`, `0001`~`0002` 마이그레이션, Outbox 릴레이 엔진.
- **6. 테스트**: `TEST-ARCH-001`, `TEST-ARCH-002`, `TEST-INF-001`, `TEST-DB-001`.
- **7. 실패 대응**: 컨테이너 헬스체크 실패 시 볼륨 초기화 후 포트/메모리 할당 점검; 마이그레이션 실패 시 즉시 `.down.sql` 롤백 검증.
- **8. 검증 증빙**: `docker compose ps` 정상 기동 로그, `/api/v1/health/ready` `200 OK` 응답, Outbox 이벤트 소비 로그.
- **9. 완료 승인 기준**: 단일 명령으로 전체 인프라가 기동되고 Outbox 트랜잭션 원자성 테스트가 `PASS`할 것.

## Phase 1 — 회원·인증·DB 전체 스키마·API 기초
- **1. 개발 목표**: 전체 57개 DB 테이블 스키마 완성, 자체 회원가입/로그인/JWT Refresh Rotation/OIDC/MFA/2계층 RBAC, 공통 보안 미들웨어 및 인증 UI 구축.
- **2. Work Order 목록**: `TB-DB-002`, `TB-AUTH-001`, `TB-AUTH-002`, `TB-SEC-001`, `TB-API-001`, `TB-WEB-001` (6개)
- **3. 선행조건**: Phase 0 완료 (`VERIFIED`).
- **4. 병렬 가능 작업**: `TB-DB-002` 완료 후 `TB-AUTH-001`과 `TB-SEC-001` 병렬 개발, 이후 `TB-AUTH-002`와 `TB-API-001` 병렬 개발.
- **5. 산출물**: `0003`~`0010` SQL 마이그레이션, `internal/modules/identity`, `internal/pkg/security`, `packages/api-client`, `packages/ui-kit`, 인증/계정 설정 화면 5종.
- **6. 테스트**: `TEST-USR-001`~`006`, `TEST-IAM-001`, `TEST-SEC-002`, `TEST-DB-002`.
- **7. 실패 대응**: 토큰 재사용 감지 오작동 시 Grace Window(5초) 튜닝; SSRF 우회 발견 시 `SafeDialer` IP 대역 필터 즉시 패치.
- **8. 검증 증빙**: 57개 테이블 생성 확인 로그, 인증/MFA/RBAC 통합 테스트 로그, Playwright 회원가입/로그인 E2E 트레이스.
- **9. 완료 승인 기준**: 모든 비인가/권한부족 요청이 `401`/`403`으로 정확히 차단되고 회원가입부터 MFA 로그인까지 E2E 통과할 것.

## Phase 2 — 블로그 개설·카테고리·도메인·관리 기초
- **1. 개발 목표**: 멀티 블로그 개설, 150개 예약어 차단, 팀 블로그 멤버 관리, 2뎁스 카테고리/시리즈, 서브도메인/경로/개인 커스텀 도메인(DNS 검증 및 자동 ACME TLS) 및 관리 콘솔 구축.
- **2. Work Order 목록**: `TB-BLOG-001`, `TB-BLOG-002`, `TB-BLOG-003`, `TB-WEB-002`, `TB-DB-003` (5개)
- **3. 선행조건**: Phase 1 완료 (`VERIFIED`).
- **4. 병렬 가능 작업**: `TB-BLOG-001` 완료 후 `TB-BLOG-002`(카테고리), `TB-BLOG-003`(도메인), `TB-DB-003`(백업/파티션) 3-Way 병렬 실행.
- **5. 산출물**: `internal/modules/blog`, `internal/modules/domain`, `infra/caddy/Caddyfile`, 블로그 개설 위저드(`SCR-MEM-005`) 및 관리 콘솔(`SCR-MEM-006`).
- **6. 테스트**: `TEST-BLOG-001`~`005`, `TEST-DOM-001`~`002`, `TEST-INF-001`.
- **7. 실패 대응**: 커스텀 도메인 DNS 미전파 시 `PENDING_DNS` 상태 유지 및 서브도메인 기본 접근 보장(서비스 중단 없음).
- **8. 검증 증빙**: 블로그 개설/카테고리 이동/도메인 검증 통합 테스트 로그 및 백업 복원 리허설 로그.
- **9. 완료 승인 기준**: 30초 내 블로그 개설 완료, 카테고리 삭제 시 글 안전 이동, 대표 도메인 `301` 리다이렉트 정상 작동.

## Phase 3 — 글쓰기 에디터·미디어 처리·게시글 발행
- **1. 개발 목표**: 제5조 35대 기능의 Tiptap 블록 에디터, 듀얼 자동저장(IndexedDB+서버), 리비전 Diff 복원, S3 이미지 업로드/WebP·AVIF 변환/EXIF 제거, 동영상 안전 임베드, 6단계 공개범위 발행 및 예약 발행 완성.
- **2. Work Order 목록**: `TB-MEDIA-001`, `TB-MEDIA-002`, `TB-EDITOR-001`, `TB-EDITOR-002`, `TB-EDITOR-003`, `TB-POST-001`, `TB-POST-002` (7개)
- **3. 선행조건**: Phase 2 완료 (`VERIFIED`).
- **4. 병렬 가능 작업**: `TB-MEDIA-001`과 `TB-EDITOR-001` 병렬 착수 후 `TB-EDITOR-002`/`003`(프론트 에디터)과 `TB-POST-001`/`002`(백엔드 발행/스케줄러) 병렬 진행.
- **5. 산출물**: `packages/editor-core`, `internal/modules/{media,editor,post}`, 에디터 스튜디오 화면(`SCR-BLOG-013`).
- **6. 테스트**: `TEST-MED-001`~`002`, `TEST-EDIT-001`~`004`, `TEST-PUB-001`~`003`, `TEST-JOB-001`~`002`, `TEST-LFC-001`~`002`, `TEST-SEC-001`, `TEST-SEC-003`.
- **7. 실패 대응**: 이미지 변환 워커 지연 시 원본 임시 서빙 후 비동기 교체; 네트워크 단절 시 IndexedDB 로컬 스냅샷으로 100% 복원.
- **8. 검증 증빙**: 에디터 35개 기능 E2E 영상/로그, XSS 50종 무력화 로그, 이미지 EXIF GPS 0바이트 검증 로그, 다중 워커 예약 발행 로그.
- **9. 완료 승인 기준**: 글 작성·이미지 첨부·자동저장·발행·예약·수정·버전복원·휴지통 전 과정이 모바일/데스크톱에서 무결하게 동작할 것.

## Phase 4 — 사용자 블로그 UX/UI·테마 스튜디오
- **1. 개발 목표**: 6종 공식 테마, 제7조 19대 디자인 커스터마이징 및 안전 CSS 샌드박스, 비주얼 테마 스튜디오, 개인 블로그 12개 화면(`SCR-BLOG-001`~`012`) 및 탐색 10개 화면(`SCR-EXP-001`~`010`) SSR 완성.
- **2. Work Order 목록**: `TB-THEME-001`, `TB-THEME-002`, `TB-WEB-003`, `TB-WEB-004` (4개)
- **3. 선행조건**: Phase 3 완료 (`VERIFIED`).
- **4. 병렬 가능 작업**: `TB-THEME-002`(테마 스튜디오)와 `TB-WEB-003`/`004`(사용자 웹 SSR) 병렬 개발.
- **5. 산출물**: `internal/modules/theme`, `packages/theme-renderer`, `apps/user-web` 전체 페이지.
- **6. 테스트**: `TEST-DSN-001`, `TEST-FE-001`~`002`, `TEST-UX-001`~`002`.
- **7. 실패 대응**: 커스텀 CSS 오류 시 `POST .../theme/reset`으로 즉시 기본 테마 복구.
- **8. 검증 증빙**: 320px/768px/1440px 뷰포트별 SSR 렌더링 스냅샷 및 No-JS HTML 출력 로그.
- **9. 완료 승인 기준**: 모든 탐색 및 개인 블로그 화면이 SSR로 렌더링되고 테마 변경 및 기본값 복원이 즉시 반영될 것.

## Phase 5 — SEO·OpenSearch·RSS·추천 랭킹
- **1. 개발 목표**: OpenSearch 한국어(Nori)/다국어(ICU) 실시간 색인·검색·자동완성, SimHash 중복 탐지 및 품질/인기 추천 랭킹, 제9조 27대 글로벌 SEO(JSON-LD, Canonical, 동적 Sitemap, RSS/Atom, 비공개 `noindex`) 완성.
- **2. Work Order 목록**: `TB-SEARCH-001`, `TB-SEARCH-002`, `TB-SEO-001`, `TB-SEO-002` (4개)
- **3. 선행조건**: Phase 3 & Phase 4 완료 (`VERIFIED`).
- **4. 병렬 가능 작업**: `TB-SEARCH-001`/`002`(OpenSearch)와 `TB-SEO-001`/`002`(SSR SEO/Sitemap/RSS) 병렬 개발.
- **5. 산출물**: `internal/modules/{search,recommendation,seo}`.
- **6. 테스트**: `TEST-SRC-001`~`002`, `TEST-QLT-001`, `TEST-REC-001`~`002`, `TEST-SEO-001`~`002`.
- **7. 실패 대응**: 색인 불일치 발견 시 무중단 Alias Reindex 실행.
- **8. 검증 증빙**: 공개 글 발행 후 검색 노출 및 비공개 전환 후 1초 내 색인 삭제 로그, JSON-LD/Sitemap/RSS 유효성 검증 로그.
- **9. 완료 승인 기준**: 비공개 글 색인 노출 0건 보장 및 제9조/제10조 모든 검색·SEO 요구사항 통과.

## Phase 6 — 이웃·댓글·구독·알림·콘텐츠 이동성
- **1. 개발 목표**: 팔로우·이웃·상호이웃(서로이웃) 상태 머신, 2뎁스 댓글/비밀댓글/승인제, 6종 공감, 북마크, 멀티채널 알림(SSE/웹푸시/이메일) 및 Rollup 병합, 외부 블로그 Import/Export 완성.
- **2. Work Order 목록**: `TB-COMMUNITY-001`, `TB-COMMUNITY-002`, `TB-COMMUNITY-003`, `TB-POST-003` (4개)
- **3. 선행조건**: Phase 3 & Phase 5 완료 (`VERIFIED`).
- **4. 병렬 가능 작업**: `TB-COMMUNITY-001`~`003`(커뮤니티/알림)과 `TB-POST-003`(블로그 Import/Export) 병렬 개발.
- **5. 산출물**: `internal/modules/{community,notification}`, `internal/modules/post/portability.go`, 회원 활동 화면(`SCR-MEM-007`~`014`).
- **6. 테스트**: `TEST-COM-001`~`005`, `TEST-NOT-001`~`002`, `TEST-INT-001`~`002`.
- **7. 실패 대응**: 외부 블로그 가져오기 중 깨진 외부 이미지 발견 시 본문은 정상 이관하고 `PARTIAL_SUCCESS` 오류 리포트 제공.
- **8. 검증 증빙**: 상호이웃 수락 전후 `MUTUAL_ONLY` 글 접근 테스트, 비밀댓글 제3자 마스킹 로그, 알림 병합 로그, Export/Import Round-trip 로그.
- **9. 완료 승인 기준**: 관계별 게시글/댓글 접근 권한이 100% 정확히 제어되고 백업 내보내기/가져오기가 무손실 동작할 것.

## Phase 7 — 관리자 통합 콘솔·보안·콘텐츠 모더레이션
- **1. 개발 목표**: 자동 스팸 탐지, 신고 접수·병합 → 임시조치(`QUARANTINE`) → 작성자 이의신청(`Appeals`) → 복원 워크플로, 그리고 제11조 7대 모듈이 실제 DB/백엔드와 직결된 관리자 콘솔(`SCR-ADM-001`~`008`) 완성.
- **2. Work Order 목록**: `TB-SEC-002`, `TB-ADMIN-001`, `TB-ADMIN-002`, `TB-ADMIN-003` (4개)
- **3. 선행조건**: Phase 5 & Phase 6 완료 (`VERIFIED`).
- **4. 병렬 가능 작업**: `TB-SEC-002`와 `TB-ADMIN-001` 완료 후 `TB-ADMIN-002`와 `TB-ADMIN-003` 병렬 구현.
- **5. 산출물**: `internal/modules/{moderation,admin}`, `apps/admin-console` 전체 화면.
- **6. 테스트**: `TEST-QLT-002`, `TEST-ADM-001`~`004`.
- **7. 실패 대응**: 자동 스팸 오탐 시 관리자 콘솔 또는 작성자 이의신청을 통해 원클릭 복원 및 재색인.
- **8. 검증 증빙**: 신고→임시조치→이의신청→복원 E2E 로그 및 `audit_logs` 해시 체인 검증 로그.
- **9. 완료 승인 기준**: 모든 관리자 메뉴가 실제 DB 데이터를 조작·조회하며 감사 로그가 100% 기록될 것.

## Phase 8 — 통계·광고·크리에이터 수익화·정산 원장
- **1. 개발 목표**: 방문자/유입/완독률/크롤러 애널리틱스, 자체 광고 캠페인/타기팅/IVT 부정클릭 방어, 블로거 광고 수익 배분, 유료 멤버십/단건 유료글(서버 Paywall), 독자 후원, 불변 복식부기 원장(`ledger_entries`) 및 정산 출금 완성.
- **2. Work Order 목록**: `TB-ANL-001`, `TB-ADS-001`, `TB-BILLING-001`, `TB-BILLING-002` (4개)
- **3. 선행조건**: Phase 7 완료 (`VERIFIED`).
- **4. 병렬 가능 작업**: `TB-ANL-001`(통계)과 `TB-BILLING-001`(유료콘텐츠/주문) 병렬 착수 후 `TB-ADS-001` 및 `TB-BILLING-002` 진행.
- **5. 산출물**: `internal/modules/{analytics,monetization}`, 수익/통계 스튜디오 화면.
- **6. 테스트**: `TEST-ANL-001`~`002`, `TEST-MON-001`~`003`.
- **7. 실패 대응**: PG 웹훅 지연/중복 시 `idempotency_key`로 중복 기장 차단; 원장 불균형 시 트랜잭션 즉시 패닉 롤백.
- **8. 검증 증빙**: 복식부기 원장 차변==대변 대사 테스트 로그, 비구매자 유료 본문 0바이트 절삭 응답 로그.
- **9. 완료 승인 기준**: 모든 금전 거래가 복식부기 불변식을 통과하고 광고/유료콘텐츠 표시 및 페이월이 완벽히 작동할 것.

## Phase 9 — 글로벌 운영·관측성·성능·상용화 종합 QA
- **1. 개발 목표**: OpenTelemetry 전 구간 관측성, 3개 국어(한/영/일) 및 글로벌 타임존 완성, Core Web Vitals 최적화, 전 도메인 실제 컨테이너 통합/보안/E2E/부하 테스트 완주.
- **2. Work Order 목록**: `TB-OPS-002`, `TB-OPS-003`, `TB-QA-001`, `TB-QA-002` (4개)
- **3. 선행조건**: Phase 0 ~ Phase 8 전체 완료.
- **4. 병렬 가능 작업**: `TB-OPS-002`(관측성)와 `TB-OPS-003`(다국어/CWV) 병렬 완료 후 `TB-QA-001` → `TB-QA-002` 실행.
- **5. 산출물**: `internal/platform/telemetry`, 다국어 번들, `tests/{integration,security,e2e,load}` 전체 테스트 스위트 및 실행 리포트.
- **6. 테스트**: `TEST-INF-002`, `TEST-GLB-001`~`002`, `TEST-FE-001`, `TEST-UX-001`~`002`, `TEST-SEC-001`~`003`.
- **7. 실패 대응**: 성능/접근성 기준 미달 화면 발견 시 해당 Work Order를 `FAILED`로 되돌려 즉시 수정 후 재검증.
- **8. 검증 증빙**: Lighthouse / axe-core / k6 / Playwright / Go Integration 전체 테스트 리포트.
- **9. 완료 승인 기준**: 모든 기능·보안·성능·접근성 테스트가 실제 환경에서 100% 통과할 것.

## Phase 10 — TreeScan 통합 준비 및 Zero-Trust 최종 독립 감사
- **1. 개발 목표**: TreeScan 기존 코드 수정 없이 향후 즉시 연결 가능한 검색 색인 피드 어댑터(`EP-TSC-001`), 자매 플랫폼 디스커버리 API(`EP-TSC-002`), OIDC 통합 계정 연결(`EP-TSC-003`)을 구현하고, **Independent Audit Agent**가 단독 운영 완전성과 전체 품질을 최종 감사한다. *(주의: 실제 TreeScan 측 배포 변경은 본 Phase에 포함하지 않으며 별도 PO 승인 후 진행한다 — 제22조 준수)*
- **2. Work Order 목록**: `TB-TREESCAN-001`, `TB-TREESCAN-002`, `TB-QA-003` (3개)
- **3. 선행조건**: Phase 9 완료 (`VERIFIED`).
- **4. 병렬 가능 작업**: `TB-TREESCAN-001`과 `TB-TREESCAN-002` 병렬 구현 후 최종 `TB-QA-003` 독립 감사 수행.
- **5. 산출물**: `internal/modules/treescan/*`, 최종 독립 감사 보고서.
- **6. 테스트**: `TEST-ZTR-001`~`004`, `TEST-EXE-001`~`002`.
- **7. 실패 대응**: 비공개 게시글이 Sync Feed에 단 1건이라도 노출되거나 단독 구동 시 에러 발생하면 즉시 반려(`FAILED`).
- **8. 검증 증빙**: `TEST-ZTR-001`(TreeScan 무의존 단독 E2E 로그), `TEST-ZTR-002`~`003`(프라이버시 필터링 및 툼스톤 로그), 독립 감사 최종 서명 리포트.
- **9. 완료 승인 기준**: 제27조 14대 금지사항 위반 0건 및 49개 전 Work Order `VERIFIED` 달성.

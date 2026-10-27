# E2E_TEST_PLAN.md — Playwright 멀티 브라우저 End-to-End 테스트 계획서

## 1. 테스트 환경 및 대상 브라우저 매트릭스
- **구동 환경**: `docker compose -f docker-compose.yml -f docker-compose.test.yml up -d` (실제 PostgreSQL, Redis, OpenSearch, MinIO, Mailpit 구동).
- **대상 브라우저 및 뷰포트**:
  1. `Desktop Chromium` (`1440x900`)
  2. `Desktop Firefox` (`1440x900`)
  3. `Mobile WebKit (iPhone SE)` (`320x568` — 최소 폭 320px 레이아웃 검증 필수)
  4. `Mobile Chromium (Pixel 7)` (`412x915`)

## 2. 8대 필수 E2E 스위트 (`tests/e2e/*.spec.ts`)
1. `01_auth_and_account.spec.ts`: 회원가입 → Mailpit에서 인증코드 추출 → 이메일 인증 → 로그인 → TOTP MFA 등록 → 프로필 수정.
2. `02_blog_creation_and_domains.spec.ts`: 예약어 차단 확인 → 블로그 개설 → 2뎁스 카테고리 생성 및 드래그 순서 변경 → 커스텀 도메인 등록 및 DNS 검증 시뮬레이션.
3. `03_editor_and_publishing.spec.ts`: Tiptap 에디터 슬래시 메뉴로 제목/코드블록/표/이미지 삽입 → 오프라인 전환 후 로컬 자동저장 복구 확인 → 공개 발행 → 리비전 수정 및 이전 버전 복원.
4. `04_visibility_and_community.spec.ts`: 사용자 A가 `MUTUAL_ONLY` 글 발행 → 사용자 B 접근 시 `403` 확인 → B가 상호이웃 신청 → A가 수락 → B가 재접속 시 본문 정상 열람 및 공감/비밀댓글 작성 → A에게 실시간 SSE 알림 도착 확인.
5. `05_search_and_discovery.spec.ts`: 발행된 글 키워드 검색 → 기간/언어 필터링 → 비공개 전환 즉시 검색 결과에서 사라짐 확인.
6. `06_theme_and_responsive.spec.ts`: 테마 스튜디오에서 스킨 및 다크모드 변경 → 320px/768px/1440px 뷰포트 오버플로 0건 및 `axe-core` 접근성 검사.
7. `07_monetization_and_portability.spec.ts`: 유료 전용 글(`PAID_MEMBERS_ONLY`) 비구매자 티저 절삭 확인 → 단건 구매 결제 완료 후 전문 열람 → Markdown ZIP 내보내기 및 새 블로그로 가져오기 확인.
8. `08_admin_moderation_and_appeals.spec.ts`: 독자가 게시글 신고 → 관리자 콘솔(`SCR-ADM-004`)에서 `QUARANTINE` 조치 → 공개 페이지 `451` 확인 → 작성자 이의신청(`Appeal`) 제출 → 관리자 복원 승인 → `200 OK` 원상복구 확인.

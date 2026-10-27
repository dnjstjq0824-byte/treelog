# API_TEST_CASES.md — 24개 API 그룹 핵심 계약 및 보안 검증 테스트 케이스

본 문서는 `04_API/openapi.yaml`의 모든 API 그룹이 정상 응답뿐만 아니라 인증 우회, 권한 상승(IDOR), 입력값 오염, 동시성 경쟁, 멱등성 위반 상황에서 계약대로 동작하는지 검증하는 API 통합 테스트 명세이다.

| Test Case ID | 대상 엔드포인트 | 테스트 목적 및 입력 조건 | 기대 HTTP 응답 및 DB/부수효과 검증 기준 |
| :--- | :--- | :--- | :--- |
| `API-TC-001` | `POST /api/v1/auth/register` (`EP-AUTH-001`) | 유효한 이메일과 비밀번호로 가입 후 동일 이메일로 재가입 시도 | 첫 요청 `201 Created`, 두 번째 요청 `409 DUPLICATE_EMAIL_OR_HANDLE` |
| `API-TC-002` | `POST /api/v1/auth/refresh` (`EP-AUTH-004`) | 이미 Rotation되어 폐기된 구 Refresh Token을 재전송 (Token Replay 공격 시뮬레이션) | `401 AUTH_REFRESH_REVOKED` 반환 및 동일 `token_family_id`의 모든 세션 즉시 `is_revoked = true` 처리 |
| `API-TC-003` | `POST /api/v1/blogs` (`EP-BCRT-001`) | 시스템 예약어(`admin`, `treescan`, `api`)로 블로그 개설 요청 | `400 INVALID_BLOG_HANDLE` 반환 및 DB 미생성 |
| `API-TC-004` | `PATCH /api/v1/blogs/{blogId}` (`EP-BMGT-002`) | 타인의 `blogId`에 대해 일반 사용자 토큰으로 제목 변경 시도 (IDOR 테스트) | `403 INSUFFICIENT_BLOG_ROLE` 반환 및 `audit_logs`에 권한 위반 시도 기록 |
| `API-TC-005` | `PUT /api/v1/blogs/{blogId}/posts/{postId}/autosave` (`EP-EDIT-002`) | 서버의 현재 `revision_no = 5`인데 클라이언트가 구 버전 `base_revision_no = 3`으로 덮어쓰기 시도 | `409 EDIT_VERSION_CONFLICT` 및 최신 서버 스냅샷 함께 반환 |
| `API-TC-006` | `POST /api/v1/blogs/{blogId}/editor/link-preview` (`EP-EDIT-005`) | `url: "http://169.254.169.254/latest/meta-data/"` 또는 `http://localhost:5432` 전송 | `400 SSRF_FORBIDDEN_HOST` 즉시 반환 (아웃바운드 소켓 연결 0건) |
| `API-TC-007` | `POST /api/v1/blogs/{blogId}/posts/{postId}/publish` (`EP-PUB-001`) | 본문 JSON AST 내 `<script>alert(1)</script>` 또는 `javascript:` 링크 포함 발행 | `200 OK` 발행되되 `posts.content_html`에서 악성 태그/속성이 100% 제거됨 |
| `API-TC-008` | `GET /api/v1/public/blogs/{blogKey}/posts/{slug}` (`EP-PUB-007`) | `visibility = 'MUTUAL_ONLY'` 게시글을 비로그인, 단순 팔로워, 단방향 이웃이 조회 | 모두 `403 FORBIDDEN_MUTUAL_ONLY` 반환, 응답 JSON에 `content_html`/`content_json` 미포함 및 `X-Robots-Tag: noindex` 헤더 포함 |
| `API-TC-009` | `POST /api/v1/blogs/{blogId}/media/upload-sessions/{id}/complete` (`EP-MED-002`) | GPS 좌표가 포함된 JPEG 업로드 완료 처리 | 워커 처리 후 `media_assets.exif_stripped = true` 및 공개 WebP 파일 내 GPS 메타데이터 0바이트 확인 |
| `API-TC-010` | `GET /api/v1/posts/{postId}/comments` (`EP-CMT-001`) | `is_secret = true`인 비밀 댓글이 포함된 게시글의 댓글 목록을 제3자가 조회 | 해당 댓글의 `content_text`가 서버에서 `"비밀 댓글입니다."`로 치환되어 반환됨 (원문 노출 0%) |
| `API-TC-011` | `GET /api/v1/search/posts` (`EP-SRC-001`) | 공개 게시글을 `UNPUBLISHED` 또는 `PRIVATE`으로 전환한 직후 키워드 검색 | 전환 후 1초 이내 검색 결과 배열에서 해당 게시글이 완전히 제외됨 |
| `API-TC-012` | `POST /api/v1/blogs/{blogId}/domains/{domainId}/verify` (`EP-DOM-003`) | DNS TXT 레코드가 불일치하는 도메인에 대해 검증 요청 | `422 DNS_TXT_VERIFICATION_FAILED` 반환 및 `tls_status = 'NONE'` 유지 |
| `API-TC-013` | `POST /api/v1/orders` (`EP-PAY-002`) | 동일한 `Idempotency-Key`로 5번 동시에 단건 유료글 구매 요청 전송 | DB `orders` 테이블에 정확히 1건만 생성되며 5개 요청 모두 동일한 주문 결과를 반환 |
| `API-TC-014` | `POST /api/v1/admin/posts/{postId}/action` (`EP-ADM-003`) | 모더레이터가 신고된 게시글을 `QUARANTINE` 처리 | `posts.status = 'QUARANTINED'`, `moderation_actions` 및 `audit_logs` 기록, OpenSearch 색인 즉시 제거 |
| `API-TC-015` | `GET /api/v1/treescan/sync/feed` (`EP-TSC-001`) | `PUBLIC` 글 2건, `PRIVATE` 글 2건, `MUTUAL_ONLY` 글 2건 생성 후 TreeScan 색인 피드 조회 | 오직 `PUBLIC` 글 2건만 피드에 포함되며 비공개/상호이웃 글은 절대 노출되지 않음 |

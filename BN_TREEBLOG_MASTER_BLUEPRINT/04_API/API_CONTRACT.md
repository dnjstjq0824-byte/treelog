# API_CONTRACT.md — 전체 24개 API 그룹 상세 개발 명세서

> **적용 조항**: 제15조 (전체 Backend API 개발 명세) 및 `04_API/openapi.yaml`
> **API 그룹 수**: 24개 그룹 전수 포함 (`G01` ~ `G24`)
> **총 엔드포인트 수**: 125개 REST API 엔드포인트

---

## 1. API 설계 표준 원칙

1. **URL 및 버전 규칙**: 모든 공개/인증 클라이언트 API는 `/api/v1/...` 프리픽스를 사용하며, 리소스 명칭은 복수형 명사(`blogs`, `posts`, `comments`)를 기본으로 한다.
2. **공통 요청/응답 봉투(Envelope)**:
   - 성공 응답(`200 OK`, `201 Created`): `{"data": {...}, "meta": {"next_cursor": "...", "has_next": true}, "request_id": "req_..."}`
   - 오류 응답(`4xx`, `5xx`): `{"error": {"code": "DOMAIN_ERROR_CODE", "message": "...", "request_id": "req_...", "timestamp": "...", "details": [...]}}`
3. **제15조 13대 필수 속성 전수 명세**:
   - 모든 엔드포인트는 **1) HTTP Method, 2) Path, 3) Request Schema, 4) Response Schema, 5) Authentication, 6) Authorization, 7) Validation, 8) Pagination, 9) Error Codes, 10) Rate Limit, 11) Idempotency, 12) Audit Policy, 13) Test Cases**를 엄격히 정의한다.

---

## 2. 24개 API 그룹별 엔드포인트 마스터 계약 테이블

### [G01] Authentication — 회원가입, 로그인, 토큰 갱신, 비밀번호 복구, OAuth/OIDC, MFA

| ID | Method | Path | 기능 요약 | Auth / RBAC | Validation & Pagination | Error Codes | Rate Limit | Idempotency / Audit | Test ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `EP-AUTH-001` | `POST` | `/api/v1/auth/register` | 이메일 신규 회원가입 | `PUBLIC` / `ANON` | SchemaStrict / `NONE` | `400,409,422,429` | `10/hour/IP` | Idempotent / AuditLog | `TEST-USR-001` |
| `EP-AUTH-002` | `POST` | `/api/v1/auth/verify-email` | 이메일 인증 코드 확인 | `PUBLIC` / `ANON` | SchemaStrict / `NONE` | `400,410,429` | `20/hour/IP` | Idempotent / AuditLog | `TEST-USR-001` |
| `EP-AUTH-003` | `POST` | `/api/v1/auth/login` | 이메일/비밀번호 로그인 및 토큰 발급 | `PUBLIC` / `ANON` | SchemaStrict / `NONE` | `400,401,403,423,429` | `10/5m/IP` | Non-Idem / AuditLog | `TEST-USR-002` |
| `EP-AUTH-004` | `POST` | `/api/v1/auth/refresh` | Refresh Token Rotation 기반 토큰 갱신 | `COOKIE_REFRESH` / `USER` | SchemaStrict / `NONE` | `401,403` | `30/5m/User` | Non-Idem / AuditLog | `TEST-IAM-001` |
| `EP-AUTH-005` | `POST` | `/api/v1/auth/logout` | 현재 세션 로그아웃 및 토큰 폐기 | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `401` | `60/1m/User` | Idempotent / AuditLog | `TEST-USR-002` |
| `EP-AUTH-006` | `POST` | `/api/v1/auth/password/forgot` | 비밀번호 재설정 링크 발송 | `PUBLIC` / `ANON` | SchemaStrict / `NONE` | `400,429` | `5/hour/IP` | Idempotent / AuditLog | `TEST-USR-003` |
| `EP-AUTH-007` | `POST` | `/api/v1/auth/password/reset` | 재설정 토큰으로 새 비밀번호 설정 | `PUBLIC` / `ANON` | SchemaStrict / `NONE` | `400,410,422` | `10/hour/IP` | Idempotent / AuditLog | `TEST-USR-003` |
| `EP-AUTH-008` | `GET` | `/api/v1/auth/oauth/{provider}/start` | 소셜/OIDC 로그인 PKCE 시작 | `PUBLIC` / `ANON` | SchemaStrict / `NONE` | `400,404` | `30/5m/IP` | Idempotent / AccessLog | `TEST-IAM-002` |
| `EP-AUTH-009` | `POST` | `/api/v1/auth/mfa/verify` | TOTP 2단계 인증 검증 및 설정 활성화 | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `400,401,429` | `10/5m/User` | Non-Idem / AuditLog | `TEST-USR-004` |

### [G02] User — 내 프로필 조회/수정, 환경설정, 세션 기기 관리, 동의 내역, 계정 탈퇴

| ID | Method | Path | 기능 요약 | Auth / RBAC | Validation & Pagination | Error Codes | Rate Limit | Idempotency / Audit | Test ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `EP-USER-001` | `GET` | `/api/v1/me` | 내 계정 및 프로필 상세 조회 | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `401` | `120/1m/User` | Idempotent / AccessLog | `TEST-USR-005` |
| `EP-USER-002` | `PATCH` | `/api/v1/me/profile` | 내 프로필(닉네임, 소개, 아바타) 수정 | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `400,409,422` | `30/1m/User` | Idempotent / AuditLog | `TEST-USR-005` |
| `EP-USER-003` | `PATCH` | `/api/v1/me/preferences` | 인터페이스 언어(Locale) 및 타임존 설정 변경 | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `400,422` | `30/1m/User` | Idempotent / AccessLog | `TEST-GLB-002` |
| `EP-USER-004` | `GET` | `/api/v1/me/sessions` | 현재 활성 로그인 기기 및 세션 목록 조회 | `BEARER_JWT` / `USER` | SchemaStrict / `CURSOR` | `401` | `60/1m/User` | Idempotent / AccessLog | `TEST-USR-004` |
| `EP-USER-005` | `DELETE` | `/api/v1/me/sessions/{sessionId}` | 특정 기기 세션 원격 강제 로그아웃 | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `401,404` | `30/1m/User` | Idempotent / AuditLog | `TEST-USR-004` |
| `EP-USER-006` | `POST` | `/api/v1/me/deactivate` | 계정 탈퇴 신청 (30일 유예 및 블로그 처리) | `BEARER_JWT_MFA` / `USER` | SchemaStrict / `NONE` | `400,403,409` | `5/hour/User` | Idempotent / AuditLog | `TEST-USR-006` |

### [G03] Blog Creation — 블로그 개설, 핸들 중복/예약어 실시간 검증, 내 소속 블로그 목록

| ID | Method | Path | 기능 요약 | Auth / RBAC | Validation & Pagination | Error Codes | Rate Limit | Idempotency / Audit | Test ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `EP-BCRT-001` | `POST` | `/api/v1/blogs` | 신규 개인/전문/기업 블로그 개설 | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `400,403,409,422` | `10/hour/User` | Idempotent / AuditLog | `TEST-BLOG-001` |
| `EP-BCRT-002` | `GET` | `/api/v1/blogs/check-handle` | 블로그 핸들 중복 및 예약어 실시간 검사 | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `400,422` | `60/1m/User` | Idempotent / AccessLog | `TEST-BLOG-001` |
| `EP-BCRT-003` | `GET` | `/api/v1/me/blogs` | 내가 소유하거나 팀 멤버로 참여 중인 블로그 목록 | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `401` | `120/1m/User` | Idempotent / AccessLog | `TEST-BLOG-002` |

### [G04] Blog Management — 블로그 정보/설정 변경, 팀 멤버 초대/권한 변경, 네비게이션, 소유권 이전, 블로그 폐쇄

| ID | Method | Path | 기능 요약 | Auth / RBAC | Validation & Pagination | Error Codes | Rate Limit | Idempotency / Audit | Test ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `EP-BMGT-001` | `GET` | `/api/v1/blogs/{blogId}` | 블로그 상세 설정 및 메타데이터 조회 | `BEARER_JWT` / `BLOG_ANALYST+` | SchemaStrict / `NONE` | `401,403,404` | `120/1m/User` | Idempotent / AccessLog | `TEST-BLOG-002` |
| `EP-BMGT-002` | `PATCH` | `/api/v1/blogs/{blogId}` | 블로그 제목, 소개, 로고, 커버, 언어 수정 | `BEARER_JWT` / `BLOG_ADMIN+` | SchemaStrict / `NONE` | `400,403,404,422` | `30/1m/User` | Idempotent / AuditLog | `TEST-BLOG-002` |
| `EP-BMGT-003` | `PUT` | `/api/v1/blogs/{blogId}/settings` | 블로그 운영 정책(기본 공개범위, 댓글, RSS, 우클릭방지) 변경 | `BEARER_JWT` / `BLOG_ADMIN+` | SchemaStrict / `NONE` | `400,403,422` | `30/1m/User` | Idempotent / AuditLog | `TEST-BLOG-002` |
| `EP-BMGT-004` | `GET` | `/api/v1/blogs/{blogId}/members` | 팀 블로그 멤버 및 역할 목록 조회 | `BEARER_JWT` / `BLOG_ANALYST+` | SchemaStrict / `NONE` | `403,404` | `60/1m/User` | Idempotent / AccessLog | `TEST-BLOG-002` |
| `EP-BMGT-005` | `POST` | `/api/v1/blogs/{blogId}/members/invite` | 팀 블로그 멤버 초대 및 역할(ADMIN/EDITOR/AUTHOR/ANALYST) 부여 | `BEARER_JWT` / `BLOG_ADMIN+` | SchemaStrict / `NONE` | `400,403,409` | `20/hour/User` | Idempotent / AuditLog | `TEST-BLOG-002` |
| `EP-BMGT-006` | `PUT` | `/api/v1/blogs/{blogId}/navigation` | 블로그 상단 커스텀 메뉴 구성 저장 | `BEARER_JWT` / `BLOG_ADMIN+` | SchemaStrict / `NONE` | `400,403,422` | `30/1m/User` | Idempotent / AuditLog | `TEST-BLOG-004` |
| `EP-BMGT-007` | `POST` | `/api/v1/blogs/{blogId}/close` | 블로그 폐쇄 신청 (30일 유예 기간 시작) | `BEARER_JWT_MFA` / `BLOG_OWNER` | SchemaStrict / `NONE` | `400,403,409` | `5/hour/User` | Idempotent / AuditLog | `TEST-LFC-003` |

### [G05] Post Editor — 초안 생성, 듀얼 자동저장(Optimistic Lock), 리비전 목록/Diff 복원, 링크 OG 프리뷰, 마크다운 변환

| ID | Method | Path | 기능 요약 | Auth / RBAC | Validation & Pagination | Error Codes | Rate Limit | Idempotency / Audit | Test ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `EP-EDIT-001` | `POST` | `/api/v1/blogs/{blogId}/posts` | 신규 게시글 초안(DRAFT) 생성 | `BEARER_JWT` / `BLOG_AUTHOR+` | SchemaStrict / `NONE` | `400,403,422` | `60/1m/User` | Idempotent / AuditLog | `TEST-EDIT-001` |
| `EP-EDIT-002` | `PUT` | `/api/v1/blogs/{blogId}/posts/{postId}/autosave` | 작성 중 본문 JSON AST 자동저장 (Optimistic Lock 검증) | `BEARER_JWT` / `BLOG_AUTHOR+` | SchemaStrict / `NONE` | `400,403,404,409` | `120/1m/User` | Idempotent / AccessLog | `TEST-EDIT-002` |
| `EP-EDIT-003` | `GET` | `/api/v1/blogs/{blogId}/posts/{postId}/revisions` | 게시글 수정 이력(Revisions) 목록 조회 | `BEARER_JWT` / `BLOG_AUTHOR+` | SchemaStrict / `CURSOR` | `403,404` | `60/1m/User` | Idempotent / AccessLog | `TEST-EDIT-003` |
| `EP-EDIT-004` | `POST` | `/api/v1/blogs/{blogId}/posts/{postId}/revisions/{revisionId}/restore` | 특정 버전 리비전으로 본문 복원 | `BEARER_JWT` / `BLOG_AUTHOR+` | SchemaStrict / `NONE` | `403,404` | `20/1m/User` | Idempotent / AuditLog | `TEST-EDIT-003` |
| `EP-EDIT-005` | `POST` | `/api/v1/blogs/{blogId}/editor/link-preview` | 외부 링크 OpenGraph 메타데이터 안전 추출 (SSRF 방어) | `BEARER_JWT` / `BLOG_AUTHOR+` | SchemaStrict / `NONE` | `400,422,502` | `30/1m/User` | Idempotent / AccessLog | `TEST-SEC-002` |
| `EP-EDIT-006` | `POST` | `/api/v1/blogs/{blogId}/editor/markdown-convert` | Markdown 원문 <-> Tiptap JSON AST 상호 변환 | `BEARER_JWT` / `BLOG_AUTHOR+` | SchemaStrict / `NONE` | `400,422` | `60/1m/User` | Idempotent / AccessLog | `TEST-EDIT-004` |

### [G06] Post Publishing — 즉시 발행, 예약 발행 등록/취소, 공개 범위 변경, 발행 취소, 휴지통 이동/복원/영구삭제, 공개 열람

| ID | Method | Path | 기능 요약 | Auth / RBAC | Validation & Pagination | Error Codes | Rate Limit | Idempotency / Audit | Test ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `EP-PUB-001` | `POST` | `/api/v1/blogs/{blogId}/posts/{postId}/publish` | 게시글 즉시 발행 또는 예약 발행(SCHEDULED) 설정 | `BEARER_JWT` / `BLOG_AUTHOR+` | SchemaStrict / `NONE` | `400,403,404,409,422` | `30/1m/User` | Idempotent / AuditLog | `TEST-PUB-001` |
| `EP-PUB-002` | `PATCH` | `/api/v1/blogs/{blogId}/posts/{postId}` | 발행된 게시글 수정 및 슬러그/SEO/공개범위 변경 | `BEARER_JWT` / `BLOG_AUTHOR+` | SchemaStrict / `NONE` | `400,403,404,409,422` | `30/1m/User` | Idempotent / AuditLog | `TEST-PUB-002` |
| `EP-PUB-003` | `POST` | `/api/v1/blogs/{blogId}/posts/{postId}/unpublish` | 발행 취소 및 비공개 초안 전환 (검색 색인 즉시 삭제) | `BEARER_JWT` / `BLOG_AUTHOR+` | SchemaStrict / `NONE` | `403,404` | `30/1m/User` | Idempotent / AuditLog | `TEST-LFC-001` |
| `EP-PUB-004` | `DELETE` | `/api/v1/blogs/{blogId}/posts/{postId}` | 게시글 휴지통 이동 (Soft Delete, 30일 보관) | `BEARER_JWT` / `BLOG_AUTHOR+` | SchemaStrict / `NONE` | `403,404` | `30/1m/User` | Idempotent / AuditLog | `TEST-LFC-001` |
| `EP-PUB-005` | `POST` | `/api/v1/blogs/{blogId}/posts/{postId}/restore` | 휴지통 게시글 복원 | `BEARER_JWT` / `BLOG_AUTHOR+` | SchemaStrict / `NONE` | `403,404` | `30/1m/User` | Idempotent / AuditLog | `TEST-LFC-002` |
| `EP-PUB-006` | `GET` | `/api/v1/public/blogs/{blogKey}/posts` | 특정 블로그의 공개(또는 권한 허용) 게시글 목록 조회 | `OPTIONAL_JWT` / `PUBLIC_OR_RELATION` | SchemaStrict / `CURSOR` | `404` | `300/1m/IP` | Idempotent / AccessLog | `TEST-PUB-003` |
| `EP-PUB-007` | `GET` | `/api/v1/public/blogs/{blogKey}/posts/{slug}` | 게시글 상세 조회 (SSR/SEO 메타 및 공개범위 권한 검증) | `OPTIONAL_JWT` / `PUBLIC_OR_RELATION` | SchemaStrict / `NONE` | `403,404,410,451` | `300/1m/IP` | Idempotent / AccessLog | `TEST-COM-001` |

### [G07] Media — S3 Presigned 업로드 세션 발급/완료, 이미지 파생본/대체텍스트 관리, 외부 동영상 안전 임베드

| ID | Method | Path | 기능 요약 | Auth / RBAC | Validation & Pagination | Error Codes | Rate Limit | Idempotency / Audit | Test ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `EP-MED-001` | `POST` | `/api/v1/blogs/{blogId}/media/upload-sessions` | S3 Presigned 업로드 세션 발급 및 용량/형식 사전 검증 | `BEARER_JWT` / `BLOG_AUTHOR+` | SchemaStrict / `NONE` | `400,403,413,415,429` | `60/1m/User` | Idempotent / AccessLog | `TEST-MED-001` |
| `EP-MED-002` | `POST` | `/api/v1/blogs/{blogId}/media/upload-sessions/{sessionId}/complete` | 업로드 완료 확인 및 EXIF 제거·WebP/AVIF 변환 잡 트리거 | `BEARER_JWT` / `BLOG_AUTHOR+` | SchemaStrict / `NONE` | `400,404,415,422` | `60/1m/User` | Idempotent / AuditLog | `TEST-MED-001` |
| `EP-MED-003` | `GET` | `/api/v1/blogs/{blogId}/media` | 블로그 미디어 라이브러리 목록 조회 | `BEARER_JWT` / `BLOG_AUTHOR+` | SchemaStrict / `CURSOR` | `403` | `60/1m/User` | Idempotent / AccessLog | `TEST-MED-001` |
| `EP-MED-004` | `PATCH` | `/api/v1/blogs/{blogId}/media/{assetId}` | 이미지 대체 텍스트(alt_text) 및 캡션 수정 | `BEARER_JWT` / `BLOG_AUTHOR+` | SchemaStrict / `NONE` | `400,403,404` | `60/1m/User` | Idempotent / AccessLog | `TEST-MED-001` |
| `EP-MED-005` | `POST` | `/api/v1/blogs/{blogId}/media/video-embed` | 외부 동영상(YouTube/Vimeo 등) oEmbed 검증 및 안전 임베드 생성 | `BEARER_JWT` / `BLOG_AUTHOR+` | SchemaStrict / `NONE` | `400,422` | `30/1m/User` | Idempotent / AccessLog | `TEST-MED-002` |

### [G08] Category & Tags — 2뎁스 카테고리 CRUD/순서변경, 시리즈(연재물) 관리, 태그 자동완성 및 인기 태그

| ID | Method | Path | 기능 요약 | Auth / RBAC | Validation & Pagination | Error Codes | Rate Limit | Idempotency / Audit | Test ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `EP-CAT-001` | `GET` | `/api/v1/blogs/{blogId}/categories` | 블로그 2뎁스 카테고리 트리 조회 | `OPTIONAL_JWT` / `PUBLIC` | SchemaStrict / `NONE` | `404` | `300/1m/IP` | Idempotent / AccessLog | `TEST-BLOG-003` |
| `EP-CAT-002` | `POST` | `/api/v1/blogs/{blogId}/categories` | 신규 카테고리 생성 | `BEARER_JWT` / `BLOG_EDITOR+` | SchemaStrict / `NONE` | `400,403,409` | `30/1m/User` | Idempotent / AuditLog | `TEST-BLOG-003` |
| `EP-CAT-003` | `PUT` | `/api/v1/blogs/{blogId}/categories/reorder` | 카테고리 계층 및 정렬 순서 일괄 변경 | `BEARER_JWT` / `BLOG_EDITOR+` | SchemaStrict / `NONE` | `400,403` | `30/1m/User` | Idempotent / AuditLog | `TEST-BLOG-003` |
| `EP-CAT-004` | `DELETE` | `/api/v1/blogs/{blogId}/categories/{categoryId}` | 카테고리 삭제 및 소속 게시글 안전 이동 | `BEARER_JWT` / `BLOG_EDITOR+` | SchemaStrict / `NONE` | `400,403,404` | `20/1m/User` | Idempotent / AuditLog | `TEST-BLOG-003` |
| `EP-CAT-005` | `GET` | `/api/v1/tags/popular` | 플랫폼 및 주제별 인기 태그 목록 조회 | `PUBLIC` / `ANON` | SchemaStrict / `CURSOR` | `400` | `300/1m/IP` | Idempotent / AccessLog | `TEST-SRC-002` |

### [G09] Follow & Subscription — 팔로우, 이웃 추가, 상호이웃(서로이웃) 신청/수락/거절, 이웃 그룹 관리, 알림/뉴스레터 구독

| ID | Method | Path | 기능 요약 | Auth / RBAC | Validation & Pagination | Error Codes | Rate Limit | Idempotency / Audit | Test ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `EP-FOL-001` | `POST` | `/api/v1/blogs/{blogId}/follow` | 팔로우 / 이웃 추가 / 상호이웃(서로이웃) 신청 | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `400,403,404,409` | `60/1m/User` | Idempotent / AccessLog | `TEST-COM-002` |
| `EP-FOL-002` | `DELETE` | `/api/v1/blogs/{blogId}/follow` | 팔로우 / 이웃 / 상호이웃 관계 해제 | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `404` | `60/1m/User` | Idempotent / AccessLog | `TEST-COM-002` |
| `EP-FOL-003` | `GET` | `/api/v1/me/mutual-requests` | 수신/발신 상호이웃 신청 목록 조회 | `BEARER_JWT` / `USER` | SchemaStrict / `CURSOR` | `401` | `60/1m/User` | Idempotent / AccessLog | `TEST-COM-002` |
| `EP-FOL-004` | `POST` | `/api/v1/me/mutual-requests/{requestId}/respond` | 상호이웃 신청 수락(ACCEPTED) 또는 거절(REJECTED) | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `400,403,404` | `60/1m/User` | Idempotent / AuditLog | `TEST-COM-002` |
| `EP-FOL-005` | `PUT` | `/api/v1/blogs/{blogId}/subscription` | 블로그 새 글 알림(인앱/푸시/이메일 뉴스레터) 구독 설정 | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `400,404` | `60/1m/User` | Idempotent / AccessLog | `TEST-NOT-001` |

### [G10] Comments — 2뎁스 댓글/대댓글 작성/수정/삭제, 비밀댓글, 댓글 승인/고정, 방명록 CRUD

| ID | Method | Path | 기능 요약 | Auth / RBAC | Validation & Pagination | Error Codes | Rate Limit | Idempotency / Audit | Test ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `EP-CMT-001` | `GET` | `/api/v1/posts/{postId}/comments` | 게시글 댓글·대댓글 목록 조회 (비밀댓글 마스킹 및 권한 검증) | `OPTIONAL_JWT` / `PUBLIC_OR_RELATION` | SchemaStrict / `CURSOR` | `403,404` | `180/1m/IP` | Idempotent / AccessLog | `TEST-COM-004` |
| `EP-CMT-002` | `POST` | `/api/v1/posts/{postId}/comments` | 댓글 또는 대댓글 작성 (비밀댓글/멘션 지원) | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `400,403,404,422,429` | `20/1m/User` | Idempotent / AccessLog | `TEST-COM-004` |
| `EP-CMT-003` | `PATCH` | `/api/v1/comments/{commentId}` | 내 댓글 수정 또는 블로그 주인의 댓글 승인/고정 처리 | `BEARER_JWT` / `COMMENT_OWNER_OR_BLOG_MOD` | SchemaStrict / `NONE` | `400,403,404` | `30/1m/User` | Idempotent / AuditLog | `TEST-COM-004` |
| `EP-CMT-004` | `DELETE` | `/api/v1/comments/{commentId}` | 댓글 삭제 (대댓글 존재 시 마스킹 소프트 삭제) | `BEARER_JWT` / `COMMENT_OWNER_OR_BLOG_MOD` | SchemaStrict / `NONE` | `403,404` | `30/1m/User` | Idempotent / AuditLog | `TEST-COM-004` |
| `EP-CMT-005` | `POST` | `/api/v1/blogs/{blogId}/guestbook` | 블로그 방명록 글/답글 작성 (비밀글 지원) | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `400,403,404` | `20/1m/User` | Idempotent / AccessLog | `TEST-BLOG-005` |

### [G11] Reactions — 게시글 6종 공감 리액션 토글 및 공감자 목록, 댓글 좋아요 토글

| ID | Method | Path | 기능 요약 | Auth / RBAC | Validation & Pagination | Error Codes | Rate Limit | Idempotency / Audit | Test ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `EP-RCT-001` | `POST` | `/api/v1/posts/{postId}/reactions` | 게시글 6종 공감 리액션(LIKE/LOVE/HELPFUL 등) 등록/변경 | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `400,403,404` | `60/1m/User` | Idempotent / AccessLog | `TEST-COM-003` |
| `EP-RCT-002` | `DELETE` | `/api/v1/posts/{postId}/reactions` | 게시글 공감 리액션 취소 | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `404` | `60/1m/User` | Idempotent / AccessLog | `TEST-COM-003` |
| `EP-RCT-003` | `POST` | `/api/v1/comments/{commentId}/like` | 댓글 좋아요 토글 | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `403,404` | `60/1m/User` | Idempotent / AccessLog | `TEST-COM-004` |

### [G12] Bookmarks — 게시글 북마크 저장/해제, 폴더별 북마크 목록, 읽기 진행률 동기화

| ID | Method | Path | 기능 요약 | Auth / RBAC | Validation & Pagination | Error Codes | Rate Limit | Idempotency / Audit | Test ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `EP-BMK-001` | `POST` | `/api/v1/posts/{postId}/bookmarks` | 게시글 북마크 추가 및 폴더/읽기진행률 저장 | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `400,403,404` | `60/1m/User` | Idempotent / AccessLog | `TEST-COM-005` |
| `EP-BMK-002` | `DELETE` | `/api/v1/posts/{postId}/bookmarks` | 게시글 북마크 해제 | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `404` | `60/1m/User` | Idempotent / AccessLog | `TEST-COM-005` |
| `EP-BMK-003` | `GET` | `/api/v1/me/bookmarks` | 내 북마크 보관함 목록 조회 | `BEARER_JWT` / `USER` | SchemaStrict / `CURSOR` | `401` | `60/1m/User` | Idempotent / AccessLog | `TEST-COM-005` |

### [G13] Notifications — 인앱 알림 목록, 미확인 카운트, SSE 실시간 스트림, 일괄 읽음, 채널별 수신 설정, Web Push 등록

| ID | Method | Path | 기능 요약 | Auth / RBAC | Validation & Pagination | Error Codes | Rate Limit | Idempotency / Audit | Test ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `EP-NOT-001` | `GET` | `/api/v1/me/notifications` | 내 알림 목록 조회 (Rollup 병합 적용) | `BEARER_JWT` / `USER` | SchemaStrict / `CURSOR` | `401` | `120/1m/User` | Idempotent / AccessLog | `TEST-NOT-001` |
| `EP-NOT-002` | `POST` | `/api/v1/me/notifications/read-all` | 미확인 알림 전체 읽음 처리 | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `401` | `30/1m/User` | Idempotent / AccessLog | `TEST-NOT-001` |
| `EP-NOT-003` | `GET` | `/api/v1/me/notifications/stream` | 실시간 인앱 알림 Server-Sent Events(SSE) 스트림 | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `401` | `10/1m/User` | Idempotent / AccessLog | `TEST-NOT-001` |
| `EP-NOT-004` | `PUT` | `/api/v1/me/notification-preferences` | 유형별/채널별 알림 수신 설정 및 방해금지 시간 저장 | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `400,422` | `30/1m/User` | Idempotent / AccessLog | `TEST-NOT-002` |
| `EP-NOT-005` | `POST` | `/api/v1/me/push-subscriptions` | W3C VAPID 웹 푸시 구독 엔드포인트 등록 | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `400,422` | `20/1m/User` | Idempotent / AccessLog | `TEST-NOT-002` |

### [G14] Search — 통합/블로그별 게시글 검색, 블로그 검색, 태그 검색, 검색어 자동완성, 검색 클릭 비콘

| ID | Method | Path | 기능 요약 | Auth / RBAC | Validation & Pagination | Error Codes | Rate Limit | Idempotency / Audit | Test ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `EP-SRC-001` | `GET` | `/api/v1/search/posts` | OpenSearch 기반 게시글 통합 검색 및 블로그 내 검색 | `OPTIONAL_JWT` / `PUBLIC` | SchemaStrict / `CURSOR` | `400,429,503` | `120/1m/IP` | Idempotent / AccessLog | `TEST-SRC-001` |
| `EP-SRC-002` | `GET` | `/api/v1/search/blogs` | 블로그 이름·핸들·주제 검색 | `OPTIONAL_JWT` / `PUBLIC` | SchemaStrict / `CURSOR` | `400,429` | `120/1m/IP` | Idempotent / AccessLog | `TEST-SRC-002` |
| `EP-SRC-003` | `GET` | `/api/v1/search/tags` | 태그별 공개 게시글 검색 | `OPTIONAL_JWT` / `PUBLIC` | SchemaStrict / `CURSOR` | `400,404` | `120/1m/IP` | Idempotent / AccessLog | `TEST-SRC-002` |
| `EP-SRC-004` | `GET` | `/api/v1/search/suggest` | 실시간 검색어 자동완성 및 오타 교정 제안 | `PUBLIC` / `ANON` | SchemaStrict / `NONE` | `400,429` | `300/1m/IP` | Idempotent / AccessLog | `TEST-SRC-002` |
| `EP-SRC-005` | `POST` | `/api/v1/search/click` | 검색 품질 평가용 결과 클릭/체류 비콘 기록 | `OPTIONAL_JWT` / `PUBLIC` | SchemaStrict / `NONE` | `400` | `300/1m/IP` | Idempotent / AccessLog | `TEST-SRC-002` |

### [G15] Recommendations — 구독/이웃 새 글 피드, 개인화 추천 피드, 실시간/일간/주간 인기 게시글, 인기 블로그, 관련 글, 피드백

| ID | Method | Path | 기능 요약 | Auth / RBAC | Validation & Pagination | Error Codes | Rate Limit | Idempotency / Audit | Test ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `EP-REC-001` | `GET` | `/api/v1/feed/following` | 구독 및 이웃/상호이웃 블로그의 최신 글 모아보기 피드 | `BEARER_JWT` / `USER` | SchemaStrict / `CURSOR` | `401` | `120/1m/User` | Idempotent / AccessLog | `TEST-REC-001` |
| `EP-REC-002` | `GET` | `/api/v1/feed/recommended` | 관심사 기반 개인화 추천 게시글 피드 | `OPTIONAL_JWT` / `PUBLIC` | SchemaStrict / `CURSOR` | `400` | `120/1m/IP` | Idempotent / AccessLog | `TEST-REC-002` |
| `EP-REC-003` | `GET` | `/api/v1/discover/trending-posts` | 시간 감쇠+품질 기반 실시간/일간/주간 인기 게시글 | `PUBLIC` / `ANON` | SchemaStrict / `CURSOR` | `400` | `300/1m/IP` | Idempotent / AccessLog | `TEST-REC-002` |
| `EP-REC-004` | `GET` | `/api/v1/discover/popular-blogs` | 주제 카테고리별 추천 인기 블로그 | `PUBLIC` / `ANON` | SchemaStrict / `CURSOR` | `400` | `300/1m/IP` | Idempotent / AccessLog | `TEST-REC-002` |
| `EP-REC-005` | `GET` | `/api/v1/public/posts/{postId}/related` | 현재 게시글과 연관된 추천 글 목록 | `PUBLIC` / `ANON` | SchemaStrict / `NONE` | `404` | `300/1m/IP` | Idempotent / AccessLog | `TEST-REC-002` |

### [G16] Analytics — 열람/완독 비콘 수집, 방문자/페이지뷰 개요, 게시글별 성과, 유입 경로/검색어, 독자 분포, 크롤러 통계

| ID | Method | Path | 기능 요약 | Auth / RBAC | Validation & Pagination | Error Codes | Rate Limit | Idempotency / Audit | Test ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `EP-ANL-001` | `POST` | `/api/v1/public/analytics/beacon` | 1st-Party 비식별 열람·체류시간·완독률 비콘 수집 | `PUBLIC` / `ANON` | SchemaStrict / `NONE` | `204,400` | `600/1m/IP` | Idempotent / AccessLog | `TEST-ANL-001` |
| `EP-ANL-002` | `GET` | `/api/v1/blogs/{blogId}/analytics/overview` | 블로그 기간별 방문자(UV)·페이지뷰(PV)·구독전환 요약 | `BEARER_JWT` / `BLOG_ANALYST+` | SchemaStrict / `NONE` | `403,404` | `60/1m/User` | Idempotent / AccessLog | `TEST-ANL-002` |
| `EP-ANL-003` | `GET` | `/api/v1/blogs/{blogId}/analytics/posts` | 블로그 내 게시글별 상세 조회·완독·공감 성과 순위 | `BEARER_JWT` / `BLOG_ANALYST+` | SchemaStrict / `CURSOR` | `403` | `60/1m/User` | Idempotent / AccessLog | `TEST-ANL-002` |
| `EP-ANL-004` | `GET` | `/api/v1/blogs/{blogId}/analytics/referrers` | 유입 경로(TreeScan/Google/Naver/소셜) 및 유입 검색어 분석 | `BEARER_JWT` / `BLOG_ANALYST+` | SchemaStrict / `NONE` | `403` | `60/1m/User` | Idempotent / AccessLog | `TEST-ANL-002` |
| `EP-ANL-005` | `GET` | `/api/v1/blogs/{blogId}/analytics/crawlers` | 검색엔진 봇(TreeScanBot, Googlebot 등) 크롤링 통계 | `BEARER_JWT` / `BLOG_ANALYST+` | SchemaStrict / `NONE` | `403` | `60/1m/User` | Idempotent / AccessLog | `TEST-ANL-001` |

### [G17] Custom Domains — 서브도메인/개인 도메인 등록, DNS TXT/CNAME 소유권 검증, 대표 Canonical 도메인 지정, 내부 호스트 리졸브

| ID | Method | Path | 기능 요약 | Auth / RBAC | Validation & Pagination | Error Codes | Rate Limit | Idempotency / Audit | Test ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `EP-DOM-001` | `GET` | `/api/v1/blogs/{blogId}/domains` | 블로그 연결 도메인 및 TLS 상태 목록 조회 | `BEARER_JWT` / `BLOG_ADMIN+` | SchemaStrict / `NONE` | `403,404` | `60/1m/User` | Idempotent / AccessLog | `TEST-DOM-001` |
| `EP-DOM-002` | `POST` | `/api/v1/blogs/{blogId}/domains` | 개인 소유 커스텀 도메인 등록 및 TXT 검증 토큰 발급 | `BEARER_JWT` / `BLOG_ADMIN+` | SchemaStrict / `NONE` | `400,403,409,422` | `10/hour/User` | Idempotent / AuditLog | `TEST-DOM-001` |
| `EP-DOM-003` | `POST` | `/api/v1/blogs/{blogId}/domains/{domainId}/verify` | DNS TXT 및 CNAME/A 설정 즉시 검증 및 ACME TLS 프로비저닝 요청 | `BEARER_JWT` / `BLOG_ADMIN+` | SchemaStrict / `NONE` | `400,403,404,422` | `20/hour/User` | Idempotent / AuditLog | `TEST-DOM-001` |
| `EP-DOM-004` | `PUT` | `/api/v1/blogs/{blogId}/domains/{domainId}/primary` | 대표 Canonical 도메인 지정 (타 주소 301 리다이렉트 활성화) | `BEARER_JWT` / `BLOG_ADMIN+` | SchemaStrict / `NONE` | `400,403,404` | `20/hour/User` | Idempotent / AuditLog | `TEST-DOM-002` |
| `EP-DOM-005` | `DELETE` | `/api/v1/blogs/{blogId}/domains/{domainId}` | 커스텀 도메인 연결 해제 | `BEARER_JWT` / `BLOG_ADMIN+` | SchemaStrict / `NONE` | `403,404` | `20/hour/User` | Idempotent / AuditLog | `TEST-DOM-001` |

### [G18] Themes — 공식 테마 카탈로그 조회, 블로그 테마 적용/토큰 커스터마이징/안전 CSS 저장, 기본값 원클릭 복원

| ID | Method | Path | 기능 요약 | Auth / RBAC | Validation & Pagination | Error Codes | Rate Limit | Idempotency / Audit | Test ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `EP-THM-001` | `GET` | `/api/v1/themes` | 플랫폼 공식 반응형 테마 카탈로그 조회 | `PUBLIC` / `ANON` | SchemaStrict / `NONE` | `200` | `300/1m/IP` | Idempotent / AccessLog | `TEST-DSN-001` |
| `EP-THM-002` | `GET` | `/api/v1/blogs/{blogId}/theme` | 블로그 현재 테마 및 커스터마이징 토큰 조회 | `OPTIONAL_JWT` / `PUBLIC` | SchemaStrict / `NONE` | `404` | `300/1m/IP` | Idempotent / AccessLog | `TEST-DSN-001` |
| `EP-THM-003` | `PUT` | `/api/v1/blogs/{blogId}/theme` | 블로그 테마 변경, 레이아웃/위젯/다크모드 및 안전 CSS 저장 | `BEARER_JWT` / `BLOG_ADMIN+` | SchemaStrict / `NONE` | `400,403,422` | `30/1m/User` | Idempotent / AuditLog | `TEST-DSN-001` |
| `EP-THM-004` | `POST` | `/api/v1/blogs/{blogId}/theme/reset` | 블로그 테마 디자인 설정 기본값 원클릭 복원 | `BEARER_JWT` / `BLOG_ADMIN+` | SchemaStrict / `NONE` | `403,404` | `15/1m/User` | Idempotent / AuditLog | `TEST-DSN-001` |

### [G19] Advertising — 광고주 계정/캠페인/소재 등록, 문맥 타기팅 광고 송출, 노출/클릭 비콘, 블로그 광고 슬롯 설정

| ID | Method | Path | 기능 요약 | Auth / RBAC | Validation & Pagination | Error Codes | Rate Limit | Idempotency / Audit | Test ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `EP-ADS-001` | `POST` | `/api/v1/ads/campaigns` | 광고주 신규 광고 캠페인 및 소재 등록 (심사 대기) | `BEARER_JWT` / `ADVERTISER` | SchemaStrict / `NONE` | `400,402,403,422` | `20/hour/User` | Idempotent / AuditLog | `TEST-MON-003` |
| `EP-ADS-002` | `GET` | `/api/v1/ads/campaigns` | 광고주 캠페인 목록 및 노출/클릭/소진액 성과 조회 | `BEARER_JWT` / `ADVERTISER` | SchemaStrict / `CURSOR` | `403` | `60/1m/User` | Idempotent / AccessLog | `TEST-MON-003` |
| `EP-ADS-003` | `PUT` | `/api/v1/blogs/{blogId}/ad-placements` | 블로거 광고 슬롯(상단/중간/하단/사이드바) 활성화 설정 | `BEARER_JWT` / `BLOG_OWNER` | SchemaStrict / `NONE` | `400,403` | `30/1m/User` | Idempotent / AuditLog | `TEST-MON-003` |
| `EP-ADS-004` | `GET` | `/api/v1/public/ads/serve` | 문맥/카테고리 타기팅 광고 송출 결정 (Brand Safety 검증) | `PUBLIC` / `ANON` | SchemaStrict / `NONE` | `200,204` | `600/1m/IP` | Idempotent / AccessLog | `TEST-MON-003` |
| `EP-ADS-005` | `POST` | `/api/v1/public/ads/events` | 광고 노출(IMPRESSION) 및 클릭(CLICK) 비콘 수집 (IVT 필터) | `PUBLIC` / `ANON` | SchemaStrict / `NONE` | `204,400` | `600/1m/IP` | Idempotent / AccessLog | `TEST-MON-003` |

### [G20] Membership & Payments — 유료 멤버십 플랜 관리, 단건 유료글 구매, 독자 후원, PG 결제 웹훅, 복식부기 원장 및 정산 출금 신청

| ID | Method | Path | 기능 요약 | Auth / RBAC | Validation & Pagination | Error Codes | Rate Limit | Idempotency / Audit | Test ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `EP-PAY-001` | `POST` | `/api/v1/blogs/{blogId}/memberships` | 블로그 유료 멤버십 티어 생성/수정 | `BEARER_JWT` / `BLOG_OWNER` | SchemaStrict / `NONE` | `400,403,422` | `20/hour/User` | Idempotent / AuditLog | `TEST-MON-002` |
| `EP-PAY-002` | `POST` | `/api/v1/orders` | 유료 멤버십 구독 / 단건 유료글 구매 / 독자 후원 주문 생성 | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `400,403,409,422` | `30/1m/User` | Idempotent / AuditLog | `TEST-MON-001` |
| `EP-PAY-003` | `POST` | `/api/v1/payments/webhook/{provider}` | PG사 결제 승인/취소/환불 웹훅 수신 및 복식부기 원장 기장 | `HMAC_SIGNATURE` / `PG_WEBHOOK` | SchemaStrict / `NONE` | `200,400,401` | `1000/1m/IP` | Idempotent / AuditLog | `TEST-MON-001` |
| `EP-PAY-004` | `POST` | `/api/v1/orders/{orderId}/refund` | 주문 취소 및 청약철회 환불 요청 (역분개 원장 전표 생성) | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `400,403,409` | `10/hour/User` | Idempotent / AuditLog | `TEST-MON-001` |
| `EP-PAY-005` | `GET` | `/api/v1/blogs/{blogId}/monetization/ledger` | 블로거 수익 배분 내역 및 복식부기 정산 원장 조회 | `BEARER_JWT` / `BLOG_OWNER` | SchemaStrict / `CURSOR` | `403` | `60/1m/User` | Idempotent / AccessLog | `TEST-MON-001` |
| `EP-PAY-006` | `POST` | `/api/v1/blogs/{blogId}/monetization/payouts` | 크리에이터 정산금 지급(출금) 신청 (MFA 필수) | `BEARER_JWT_MFA` / `BLOG_OWNER` | SchemaStrict / `NONE` | `400,403,409,422` | `5/hour/User` | Idempotent / AuditLog | `TEST-MON-001` |

### [G21] Moderation — 콘텐츠/블로그/댓글/저작권 신고 접수, 내 제재 내역 조회 및 이의신청(Appeal) 제출, 사용자 차단

| ID | Method | Path | 기능 요약 | Auth / RBAC | Validation & Pagination | Error Codes | Rate Limit | Idempotency / Audit | Test ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `EP-MOD-001` | `POST` | `/api/v1/reports` | 게시글·댓글·블로그·저작권(DMCA)·명예훼손 신고 접수 | `OPTIONAL_JWT` / `PUBLIC` | SchemaStrict / `NONE` | `400,422,429` | `20/hour/IP` | Idempotent / AuditLog | `TEST-ADM-002` |
| `EP-MOD-002` | `POST` | `/api/v1/me/appeals` | 게시 중단 또는 블로그 제재에 대한 작성자 이의신청(소명) 제출 | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `400,403,404,409` | `10/hour/User` | Idempotent / AuditLog | `TEST-ADM-002` |
| `EP-MOD-003` | `POST` | `/api/v1/users/{userId}/block` | 특정 사용자 차단 (댓글·공감·상호이웃 원천 차단) | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `400,404` | `60/1m/User` | Idempotent / AuditLog | `TEST-COM-002` |
| `EP-MOD-004` | `DELETE` | `/api/v1/users/{userId}/block` | 사용자 차단 해제 | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `404` | `60/1m/User` | Idempotent / AuditLog | `TEST-COM-002` |

### [G22] Admin — 회원/블로그/게시글/댓글/모더레이션/검색/광고심사/정산승인/시스템설정/기능플래그/감사로그 통합 관리

| ID | Method | Path | 기능 요약 | Auth / RBAC | Validation & Pagination | Error Codes | Rate Limit | Idempotency / Audit | Test ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `EP-ADM-001` | `GET` | `/api/v1/admin/users` | 전체 회원 검색·상태·가입통계 조회 및 정지/권한 제어 | `BEARER_JWT_MFA` / `SUPPORT_AGENT+` | SchemaStrict / `CURSOR` | `401,403` | `120/1m/Admin` | Idempotent / AuditLog | `TEST-ADM-001` |
| `EP-ADM-002` | `POST` | `/api/v1/admin/blogs/{blogId}/action` | 블로그 강제 정지/해제 및 도메인 연결 강제 차단 | `BEARER_JWT_MFA` / `TRUST_SAFETY_MOD+` | SchemaStrict / `NONE` | `400,403,404` | `60/1m/Admin` | Idempotent / AuditLog | `TEST-ADM-001` |
| `EP-ADM-003` | `POST` | `/api/v1/admin/posts/{postId}/action` | 게시글 임시 게시중단(QUARANTINE), 비공개 전환, 복원 처리 | `BEARER_JWT_MFA` / `TRUST_SAFETY_MOD+` | SchemaStrict / `NONE` | `400,403,404` | `60/1m/Admin` | Idempotent / AuditLog | `TEST-ADM-002` |
| `EP-ADM-004` | `POST` | `/api/v1/admin/moderation/cases/{caseId}/decide` | 신고 모더레이션 케이스 심사 결정 및 제재/기각 처분 | `BEARER_JWT_MFA` / `TRUST_SAFETY_MOD+` | SchemaStrict / `NONE` | `400,403,404` | `60/1m/Admin` | Idempotent / AuditLog | `TEST-ADM-002` |
| `EP-ADM-005` | `POST` | `/api/v1/admin/moderation/appeals/{appealId}/resolve` | 이의신청 재심사 승인(원상복구) 또는 기각 결정 | `BEARER_JWT_MFA` / `TRUST_SAFETY_MOD+` | SchemaStrict / `NONE` | `400,403,404` | `60/1m/Admin` | Idempotent / AuditLog | `TEST-ADM-002` |
| `EP-ADM-006` | `POST` | `/api/v1/admin/search/control` | 검색 색인 재구축, 스팸 검색어 차단, 문서 품질 오버라이드 | `BEARER_JWT_MFA` / `TRUST_SAFETY_MOD+` | SchemaStrict / `NONE` | `400,403` | `30/1m/Admin` | Idempotent / AuditLog | `TEST-ADM-003` |
| `EP-ADM-007` | `POST` | `/api/v1/admin/monetization/review` | 광고 캠페인 소재 승인/거절 및 크리에이터 정산 지급 승인/보류 | `BEARER_JWT_MFA` / `AD_REVIEWER_OR_FINANCE` | SchemaStrict / `NONE` | `400,403,404` | `60/1m/Admin` | Idempotent / AuditLog | `TEST-ADM-004` |
| `EP-ADM-008` | `GET` | `/api/v1/admin/system/audit-logs` | 불변 감사 로그(Audit Logs), 백업 현황, Feature Flag 통합 관리 | `BEARER_JWT_MFA` / `SUPER_ADMIN` | SchemaStrict / `CURSOR` | `401,403` | `60/1m/Admin` | Idempotent / AuditLog | `TEST-ADM-001` |

### [G23] Import / Export — 블로그 전체 백업 내보내기(Markdown/HTML/WXR ZIP), 외부 블로그 가져오기 및 오류 리포트, RSS/Atom 피드

| ID | Method | Path | 기능 요약 | Auth / RBAC | Validation & Pagination | Error Codes | Rate Limit | Idempotency / Audit | Test ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `EP-PRT-001` | `POST` | `/api/v1/blogs/{blogId}/exports` | 블로그 콘텐츠 전체 백업 내보내기(Markdown/HTML/WXR ZIP) 잡 생성 | `BEARER_JWT` / `BLOG_OWNER` | SchemaStrict / `NONE` | `400,403,429` | `5/hour/User` | Idempotent / AuditLog | `TEST-INT-001` |
| `EP-PRT-002` | `GET` | `/api/v1/blogs/{blogId}/exports/{jobId}` | 내보내기 진행 상태 조회 및 완료된 ZIP Signed URL 다운로드 | `BEARER_JWT` / `BLOG_OWNER` | SchemaStrict / `NONE` | `403,404` | `60/1m/User` | Idempotent / AccessLog | `TEST-INT-001` |
| `EP-PRT-003` | `POST` | `/api/v1/blogs/{blogId}/imports` | 외부 블로그 백업 파일(WXR/Markdown ZIP) 가져오기 실행 (소유권 동의 필수) | `BEARER_JWT` / `BLOG_OWNER` | SchemaStrict / `NONE` | `400,403,413,422` | `5/hour/User` | Idempotent / AuditLog | `TEST-INT-001` |
| `EP-PRT-004` | `GET` | `/api/v1/blogs/{blogId}/imports/{jobId}` | 가져오기 진행률, 구 URL->신 URL 매핑표 및 실패 항목 오류 리포트 조회 | `BEARER_JWT` / `BLOG_OWNER` | SchemaStrict / `NONE` | `403,404` | `60/1m/User` | Idempotent / AccessLog | `TEST-INT-001` |
| `EP-PRT-005` | `GET` | `/api/v1/public/blogs/{blogKey}/rss.xml` | 블로그 표준 RSS 2.0 / Atom 1.0 신디케이션 피드 조회 | `PUBLIC` / `ANON` | SchemaStrict / `NONE` | `200,404` | `300/1m/IP` | Idempotent / AccessLog | `TEST-SEO-002` |

### [G24] TreeScan Integration — TreeScan 증분 공개 색인 동기화 피드, 자매 플랫폼 디스커버리 API, OIDC 통합 계정 매핑, 연동 헬스체크

| ID | Method | Path | 기능 요약 | Auth / RBAC | Validation & Pagination | Error Codes | Rate Limit | Idempotency / Audit | Test ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `EP-TSC-001` | `GET` | `/api/v1/treescan/sync/feed` | TreeScan 검색엔진용 공개 게시글 증분 색인 피드 및 삭제 툼스톤 스트림 | `SERVICE_MTLS_HMAC` / `TREESCAN_SYNC_BOT` | SchemaStrict / `CURSOR` | `401,403` | `600/1m/Service` | Idempotent / AuditLog | `TEST-ZTR-002` |
| `EP-TSC-002` | `GET` | `/api/v1/treescan/discovery/posts` | Booking Network 자매 플랫폼(getforTrip/Treewiki 등) 연동용 공개 콘텐츠 조회 | `SERVICE_MTLS_HMAC` / `BN_PARTNER_SERVICE` | SchemaStrict / `CURSOR` | `400,401,403` | `600/1m/Service` | Idempotent / AccessLog | `TEST-ZTR-004` |
| `EP-TSC-003` | `POST` | `/api/v1/treescan/sso/link` | Booking Network / TreeScan OIDC 통합 계정 안전 연결 (데이터 소유권 보존) | `BEARER_JWT` / `USER` | SchemaStrict / `NONE` | `400,401,409` | `10/hour/User` | Idempotent / AuditLog | `TEST-IAM-002` |
| `EP-TSC-004` | `GET` | `/api/v1/treescan/health` | TreeScan 통합 어댑터 상태 및 동기화 시퀀스 지연(Lag) 확인 | `SERVICE_MTLS_HMAC` / `TREESCAN_SYNC_BOT` | SchemaStrict / `NONE` | `200,503` | `120/1m/Service` | Idempotent / AccessLog | `TEST-ZTR-001` |

---

## 3. 핵심 도메인별 Request / Response JSON Schema 및 Validation 상세

### 3.1 블로그 개설 (`POST /api/v1/blogs` — `EP-BCRT-001`)
- **Validation Rules**:
  - `handle`: 필수, 3~30자, 정규식 `^[a-z0-9][a-z0-9-]{1,28}[a-z0-9]$`, 시스템 예약어 150개와 충돌 금지.
  - `title`: 필수, 1~120자, 제어문자 제외.
  - `blog_type`: `PERSONAL` | `PROFESSIONAL` | `CORPORATE` | `COMMUNITY`.
- **Request Schema**:
```json
{
  "handle": "tech-pioneer",
  "title": "Tech Pioneer 아키텍처 연구소",
  "subtitle": "클라우드와 분산 시스템을 기록합니다",
  "blog_type": "PROFESSIONAL",
  "default_language": "ko-KR",
  "timezone": "Asia/Seoul",
  "theme_slug": "tech-monolith"
}
```
- **Response Schema (`201 Created`)**:
```json
{
  "data": {
    "id": "018f2a10-7b11-7000-8000-000000000010",
    "handle": "tech-pioneer",
    "primary_url": "https://tech-pioneer.treeblog.example.com",
    "status": "ACTIVE",
    "created_at": "2026-10-27T14:00:00Z"
  },
  "request_id": "req_018f2a107b11"
}
```

### 3.2 게시글 발행 및 예약 발행 (`POST /api/v1/blogs/{blogId}/posts/{postId}/publish` — `EP-PUB-001`)
- **Validation Rules**:
  - `title`: 필수, 1~300자.
  - `content_json`: 필수, 유효한 Tiptap/ProseMirror Root `doc` 노드, 최소 1개 이상의 블록 포함.
  - `visibility`: `PUBLIC` | `UNLISTED` | `NEIGHBOR_ONLY` | `MUTUAL_ONLY` | `PAID_MEMBERS_ONLY` | `PRIVATE`.
  - `scheduled_for`: 예약 발행일 경우 현재 시각보다 최소 1분 이후의 미래 UTC 시각이어야 함.
- **Request Schema**:
```json
{
  "title": "Go와 OpenSearch로 구축하는 글로벌 블로그 검색 아키텍처",
  "subtitle": "Transactional Outbox와 Nori 형태소 분석기 실전 가이드",
  "slug": "go-opensearch-architecture",
  "category_id": "018f2a10-7b11-7000-8000-000000000050",
  "tags": ["golang", "opensearch", "architecture"],
  "visibility": "PUBLIC",
  "allow_comments": "ALL_ALLOWED",
  "allow_reactions": true,
  "seo_title": "Go + OpenSearch 블로그 검색엔진 구축 가이드",
  "seo_description": "PostgreSQL을 원본으로 유지하며 OpenSearch를 비동기 동기화하는 방법",
  "license_type": "CC_BY_NC",
  "ai_disclosure_level": "NONE",
  "publish_mode": "IMMEDIATE"
}
```

### 3.3 TreeScan 증분 색인 피드 (`GET /api/v1/treescan/sync/feed` — `EP-TSC-001`)
- **Security & Validation**:
  - `X-TreeBlog-Service-Signature` HMAC 검증 및 mTLS 인증서 검증 필수.
  - **절대 원칙**: 오직 `visibility == 'PUBLIC'` 게시글의 `UPSERT` 또는 비공개 전환/삭제된 글의 `DELETE_TOMBSTONE`(본문 미포함, 문서 ID만 포함)만 반환한다.

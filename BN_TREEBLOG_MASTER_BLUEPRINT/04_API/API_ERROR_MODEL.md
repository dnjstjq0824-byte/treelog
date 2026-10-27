# API_ERROR_MODEL.md — 표준 에러 응답 구조 및 도메인 에러 코드 사전

## 1. 표준 HTTP 상태 코드 및 에러 봉투 규약

모든 API 오류는 HTTP 상태 코드와 함께 기계 판독 가능한 `error.code` 문자열, 사용자 친화적 다국어 `error.message`, 분산 추적용 `error.request_id`를 반환한다. 스택 트레이스나 DB 내부 테이블/제약조건 원본 에러 메시지는 외부로 절대 노출하지 않는다.

```json
{
  "error": {
    "code": "FORBIDDEN_MUTUAL_ONLY",
    "message": "이 게시글은 작성자와 상호이웃(서로이웃) 관계인 사용자만 열람할 수 있습니다.",
    "request_id": "req_018f2a1099ac7000",
    "timestamp": "2026-10-27T14:35:12Z",
    "details": [
      {
        "field": "visibility",
        "reason": "REQUIRED_RELATION_MUTUAL_NEIGHBOR"
      }
    ]
  }
}
```

## 2. 도메인별 표준 에러 코드 사전 (Error Code Dictionary)

| HTTP 상태 | 에러 코드 (`error.code`) | 발생 조건 및 클라이언트 처리 지침 |
| :--- | :--- | :--- |
| `400 Bad Request` | `INVALID_REQUEST_PAYLOAD` | JSON 문법 오류 또는 필수 필드 누락 |
| `400 Bad Request` | `INVALID_BLOG_HANDLE` | 블로그 핸들 정규식 위반 또는 시스템 예약어(`admin`, `api` 등) 사용 시도 |
| `400 Bad Request` | `SSRF_FORBIDDEN_HOST` | 링크 프리뷰/가져오기에서 사설 IP, 로컬호스트, 비표준 포트 접근 시도 |
| `400 Bad Request` | `UNSUPPORTED_VIDEO_PROVIDER` | 허용되지 않은 외부 동영상 임베드 도메인 요청 |
| `401 Unauthorized` | `AUTH_TOKEN_MISSING` | Authorization 헤더 또는 세션 쿠키 누락 |
| `401 Unauthorized` | `AUTH_TOKEN_EXPIRED` | Access Token 만료 (`POST /api/v1/auth/refresh` 호출 유도) |
| `401 Unauthorized` | `AUTH_REFRESH_REVOKED` | 만료되었거나 재사용(탈취 의심)이 감지된 Refresh Token |
| `401 Unauthorized` | `MFA_REQUIRED` | 2단계 인증(TOTP) 검증이 필요한 계정 또는 민감 작업 |
| `403 Forbidden` | `INSUFFICIENT_BLOG_ROLE` | 해당 블로그에 대한 요구 권한(`OWNER`/`ADMIN`/`EDITOR`/`AUTHOR`) 부족 |
| `403 Forbidden` | `INSUFFICIENT_ADMIN_ROLE` | 플랫폼 관리자 콘솔 RBAC 권한 부족 |
| `403 Forbidden` | `FORBIDDEN_PRIVATE_POST` | 비공개(`PRIVATE`) 게시글에 제3자가 접근 시도 |
| `403 Forbidden` | `FORBIDDEN_NEIGHBOR_ONLY` | 이웃 공개(`NEIGHBOR_ONLY`) 게시글에 비이웃이 접근 시도 |
| `403 Forbidden` | `FORBIDDEN_MUTUAL_ONLY` | 상호이웃 공개(`MUTUAL_ONLY`) 게시글에 비상호이웃이 접근 시도 |
| `403 Forbidden` | `PAYWALL_PAYMENT_REQUIRED` | 유료 멤버십/단건 구매 전용 구간에 미결제 접근 (티저만 함께 반환) |
| `403 Forbidden` | `USER_BLOCKED_BY_OWNER` | 블로그 소유자에 의해 차단(`user_blocks`)된 사용자 |
| `403 Forbidden` | `CSRF_TOKEN_MISMATCH` | 상태 변경 요청 시 `X-CSRF-Token` 불일치 |
| `404 Not Found` | `RESOURCE_NOT_FOUND` | 존재하지 않거나 휴지통(`TRASHED`)으로 이동된 블로그/게시글/댓글 |
| `409 Conflict` | `DUPLICATE_EMAIL_OR_HANDLE` | 이미 사용 중인 이메일, 사용자명, 블로그 핸들, 커스텀 도메인 |
| `409 Conflict` | `EDIT_VERSION_CONFLICT` | 자동저장/수정 시 `expected_revision_no` 불일치 (충돌 해결 모달 호출) |
| `409 Conflict` | `IDEMPOTENCY_KEY_IN_FLIGHT` | 동일한 `Idempotency-Key`로 요청이 이미 처리 중이거나 다른 페이로드로 재전송됨 |
| `410 Gone` | `POST_PERMANENTLY_PURGED` | 영구 삭제(`PURGED`)된 게시글 (검색엔진 색인 즉시 제거 신호) |
| `413 Payload Too Large` | `MEDIA_SIZE_LIMIT_EXCEEDED` | 허용된 파일 크기(이미지 25MB 등) 초과 |
| `415 Unsupported Media` | `MEDIA_MAGIC_BYTE_MISMATCH` | 파일 확장자와 실제 바이너리 헤더(Magic Bytes) 불일치 또는 금지된 포맷 |
| `422 Unprocessable` | `DNS_TXT_VERIFICATION_FAILED` | 커스텀 도메인 `_treeblog-verify` TXT 레코드 미발견 또는 값 불일치 |
| `422 Unprocessable` | `EDITOR_SCHEMA_INVALID` | 허용되지 않은 Tiptap/ProseMirror 노드 또는 속성 포함 |
| `423 Locked` | `ACCOUNT_TEMPORARILY_LOCKED` | 로그인 연속 실패로 인한 일시적 계정 잠금 (15분) |
| `429 Too Many Requests` | `RATE_LIMIT_EXCEEDED` | 슬라이딩 윈도우 호출 제한 초과 (`Retry-After` 헤더 참조) |
| `451 Legal Reasons` | `CONTENT_QUARANTINED` | 저작권(DMCA)/명예훼손/불법정보 신고로 임시 게시 중단된 콘텐츠 |
| `503 Unavailable` | `SEARCH_TEMPORARILY_UNAVAILABLE`| OpenSearch 점검 또는 서킷 브레이커 작동 중 |

# S_SECURITY — Security & Abuse Prevention (보안 아키텍처 및 어뷰징 방어)
> **부제**: XSS, CSRF, SSRF, 악성 업로드, 스팸, 계정 탈취, 봇, 콘텐츠 사칭, 관리자 권한, 개인정보 보호
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-S_SECURITY`
> **적용 조항**: 제3조 (S_SECURITY 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- 사용자가 직접 리치 텍스트, 이미지, 커스텀 도메인, 테마 CSS를 등록하는 UGC(User-Generated Content) 플랫폼의 특성상 발생할 수 있는 **제16조의 20대 핵심 보안 위협**을 심층 방어(Defense-in-Depth) 아키텍처로 원천 차단한다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **Security Agent**: 모든 입력 검증, HTML/CSS 정화, SSRF 방지 Dialer, 파일 매직바이트 검증, Rate Limiting, 관리자 MFA 및 감사 로그를 구현하고 침투 테스트를 수행한다.
- **Trust & Safety Moderator**: 피싱 블로그, 사칭 계정, 불법 스팸 캠페인을 탐지하여 즉시 격리하고 피해 확산을 차단한다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-SEC-001 (제16조 20대 위협 전수 방어 매트릭스)**:
  1) **XSS (Stored/Reflected/DOM)**: Tiptap JSON AST 원본 검증 + 서버 측 `bluemonday` 허용목록 Sanitizer + 응답 `Content-Security-Policy` (`script-src 'self' 'nonce-...'`, `object-src 'none'`).
  2) **CSRF**: `SameSite=Lax` HttpOnly 쿠키 + 상태 변경 API 전체에 Double-Submit Cookie (`X-CSRF-Token`) 및 `Origin`/`Referer` 검증 강제.
  3) **SSRF**: 외부 URL 프리뷰/가져오기/웹훅 발송 시 `SafeDialer`를 통해 루프백/사설망/클라우드 메타데이터 IP(`169.254.169.254`) 연결을 소켓 레벨에서 차단.
  4) **SQL Injection**: 문자열 결합 쿼리 전면 금지. 100% `pgx`/`sqlc` 파라미터 바인딩(`$1, $2`)만 사용.
  5) **업로드 파일 악용**: 확장자 화이트리스트 + 첫 512바이트 Magic Bytes(`http.DetectContentType`) 일치 검증 + `libvips` 이미지 재인코딩(폴리글랏 악성코드 파괴) + 미디어 전용 분리 도메인(`media.treeblog.example.com`)에서 `Content-Disposition: inline` 및 `X-Content-Type-Options: nosniff` 서빙(SVG 업로드는 스크립트 내장 위험으로 인해 기본 금지하거나 래스터 변환만 허용).
  6) **악성 HTML/CSS**: 커스텀 테마 내 `<script>`, `<iframe>`, `javascript:`, 외부 `@import` 차단.
  7) **개인정보 노출**: 업로드 사진 GPS/EXIF 메타데이터 자동 삭제, 비공개/이웃공개 글 서버 측 접근 제어, 에러 응답에 스택 트레이스 노출 금지.
  8) **계정 탈취 (ATO)**: Argon2id 해싱, 유출 비밀번호 차단, Refresh Token Rotation, 새 기기 로그인 알림, 민감 작업 시 MFA 재인증.
  9~16) **스팸 게시글 / 악성 댓글 / 피싱 블로그 / 검색엔진 스팸 / 저작권 침해 / 명예훼손·사칭 / 대량 자동 가입 / 봇 트래픽**: Redis Sliding Window Rate Limiter + 적응형 Turnstile/Captcha挑战 + 신규 가입 계정 신뢰도 등급(Trust Tier)에 따른 일일 발행/외부링크 제한 + SimHash 중복 필터 + Safe Browsing URL 검사.
  17~20) **불법 광고 / 관리자 권한 탈취 / 도메인 연결 악용 / 데이터 삭제 및 복원 사고**: 광고 소재 사전 심사 필수, 관리자 콘솔 MFA 및 IP/세션 분리 + 불변 `audit_logs`, 도메인 TXT 소유권 검증 강제, 휴지통 30일 보관 및 PITR 백업.

## 4. 데이터 구조 (Data Structures & Domain Models)
```go
// Rate Limit 정책 정의 구조체
type RateLimitRule struct {
    EndpointGroup string        // "auth_login", "post_create", "comment_create", "media_upload"
    Window        time.Duration // 1m, 1h, 24h
    MaxRequests   int           // IP 및 User ID 기준 허용치
    Burst         int
}
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- CSP Nonce 지원 SSR 렌더링, 민감 액션(블로그 폐쇄, 도메인 삭제, 정산 계좌 변경, 관리자 제재) 실행 시 Step-up Authentication(비밀번호/OTP 재입력) 모달 호출.

## 6. 백엔드 로직 (Backend Logic & Processing)
- `internal/pkg/security`:
  - `SanitizePostHTML(raw string) string`
  - `SanitizeCustomCSS(raw string) (string, error)`
  - `NewSafeHTTPClient() *http.Client` (SSRF 차단)
  - `VerifyFileMagicBytes(r io.Reader, allowedMimeTypes []string) error`

## 7. API 계약 (API Contracts)
- 모든 API 응답에 표준 보안 헤더 자동 부착:
  - `Strict-Transport-Security: max-age=63072000; includeSubDomains; preload`
  - `X-Content-Type-Options: nosniff`
  - `X-Frame-Options: SAMEORIGIN` (커스텀 임베드 경로 제외)
  - `Referrer-Policy: strict-origin-when-cross-origin`
  - `Permissions-Policy: camera=(), microphone=(), geolocation=()`

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- 주요 보안/감사 테이블: `audit_logs`, `sessions`, `user_reports`, `moderation_cases`, `moderation_actions`, `appeals`, `spam_detection_results`.
- 개인정보 암호화: 정산용 은행 계좌번호나 세무 식별 정보는 애플리케이션 레벨 AES-256-GCM(Envelope Encryption)으로 암호화하여 저장한다.

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **자동 보안 차단 시 오탐 복구**: Rate Limit(`429 Too Many Requests`) 응답 시 `Retry-After` 초를 명확히 반환하고, 자동 스팸 격리 시 작성자가 즉시 이의신청(`POST /api/v1/me/appeals`)을 제출할 수 있는 경로는 차단하지 않는다.

## 10. UX/UI 구조 (UX/UI Structure)
- 외부 도메인 링크 클릭 시 피싱 의심 도메인인 경우 `외부 사이트로 이동합니다. 신뢰할 수 있는 주소인지 확인하세요.` 경고 인터스티셜 화면을 표시한다.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- 상세 운영 정책 및 글로벌 국가별 법률 준수 체크리스트(한국 정보통신망법/개인정보보호법, EU GDPR/DSA, 미국 DMCA/CCPA, 일본 APPI)는 `11_SECURITY_MODERATION/COMPLIANCE_CHECKLIST.md`에 별도 문서로 상세 정의됨.

## 12. 성능 및 확장성 (Performance & Scalability)
- 보안 필터(JWT 검증, Redis Rate Limit, Bluemonday HTML 정화)의 총 합산 오버헤드는 요청당 3ms 미만으로 유지한다.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- `github.com/microcosm-cc/bluemonday`, `golang.org/x/time/rate`, `crypto/aes`, `crypto/cipher`.

## 14. 개발 우선순위 (Development Priority)
- **Phase 1 (인증/입력보안)**, **Phase 3 (에디터 XSS/미디어 보안)**, **Phase 7 (모더레이션/감사 완성)** (`TB-SEC-001`, `TB-SEC-002`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 전 모듈에 횡단 관심사(Cross-cutting Concern)로 적용됨.

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-SEC-001`: OWASP Top 10 주요 XSS 벡터 50종을 게시글/댓글/프로필/테마에 주입 시 스크립트 실행이 0건인지 검증.
- `TEST-SEC-002`: 링크 프리뷰 API에 `http://127.0.0.1:8080`, `http://169.254.169.254/latest/meta-data/` 입력 시 즉시 `400 SSRF_FORBIDDEN_HOST`로 차단되는지 검증.
- `TEST-SEC-003`: 확장자를 `.jpg`로 위장한 실행 파일(ELF/PE 헤더) 업로드 시 매직바이트 검사에서 `415 UNSUPPORTED_MEDIA_TYPE`으로 거부되는지 검증.

## 17. Acceptance Criteria (인수 기준)
1. 제16조의 20개 보안 위협 시나리오에 대한 방어 코드와 자동화 테스트가 모두 통과해야 하며, 저장소 내에 하드코딩된 비밀키가 0건이어야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- `15_TESTING/SECURITY_TEST_PLAN.md` 실행 로그, 정적 보안 스캔(`gosec`) 및 시크릿 스캔 통과 리포트.

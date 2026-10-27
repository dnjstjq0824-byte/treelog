# I_IDENTITY_AUTHORIZATION — Identity & Authorization (인증·인가 및 계정 보안)
> **부제**: 회원가입, 로그인, 소셜 로그인, OAuth/OIDC, MFA, 계정 보안, RBAC, 기업 블로그 팀 관리
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-I_IDENTITY_AUTHORIZATION`
> **적용 조항**: 제3조 (I_IDENTITY_AUTHORIZATION 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- 초기에는 **TreeBlog 자체 독립 회원·인증 시스템**으로 완벽히 동작하면서, 향후 **TreeScan 및 Booking Network 생태계와 OIDC(OpenID Connect) 기반 통합 SSO**로 계정 손실 없이 연동할 수 있는 확장형 Identity 아키텍처를 구현한다.
- 플랫폼 전체 역할(Platform RBAC)과 개별 블로그 단위 팀 역할(Blog-level RBAC)의 2계층 인가 모델을 수립한다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **신규 사용자**: 이메일+비밀번호 또는 소셜 계정(Google, Apple, GitHub, Naver, Kakao OIDC 어댑터)으로 가입하고, 보안 강화를 위해 TOTP 2단계 인증(MFA)을 활성화한다.
- **기존 사용자 (향후 TreeScan 연동 시)**: TreeBlog 단독 계정을 보유한 상태에서 향후 'Booking Network / TreeScan 통합 계정 연결'을 수행하면 `identities` 테이블에 OIDC Subject가 매핑되어 기존 블로그, 게시글, 수익 원장 소유권이 100% 그대로 유지된다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-IAM-001 (독립 회원가입 및 로그인)**:
  - 이메일 인증(6자리 OTP 또는 매직 링크), 비밀번호 정책(최소 10자, 영문/숫자/특수문자 조합 또는 15자 이상 패스프레이즈, HaveIBeenPwned 유출 비밀번호 사전 차단), 비밀번호 해싱(`Argon2id`: time=3, memory=64MB, threads=4).
- **REQ-IAM-002 (세션 및 토큰 관리)**:
  - 단기 Access Token(JWT, 유효기간 15분, `RS256` 또는 `EdDSA` 비대칭키 서명) + 장기 Refresh Token(유효기간 14일, `sessions` 테이블에 SHA-256 해시 저장 및 **Refresh Token Rotation** 적용).
  - 탈취된 구 Refresh Token 재사용 감지 시 해당 세션 패밀리 전체를 즉시 강제 만료(Revoke)시킨다.
- **REQ-IAM-003 (다중 인증 수단 `identities` 분리 설계 — 제1.3조 공통 계정 확장성)**:
  - `users` 테이블(고유 사용자 엔터티)과 `identities` 테이블(`provider`: `LOCAL`, `GOOGLE`, `GITHUB`, `NAVER`, `KAKAO`, `TREESCAN_OIDC`)을 1:N으로 분리한다.
  - 이를 통해 향후 TreeScan OIDC 연동 시 `users.id` 변경 없이 `identities` 레코드만 추가하여 블로그/데이터 소유권 유실을 원천 방지한다.
- **REQ-IAM-004 (MFA 및 기기 세션 관리)**:
  - RFC 6238 표준 TOTP(Google Authenticator 등) 및 1회용 복구 코드(Backup Codes 10개, 해시 저장) 지원.
  - 관리자와 정산 출금 요청자는 MFA 필수 강제. 현재 로그인된 기기 목록 조회 및 원격 강제 로그아웃 지원.
- **REQ-IAM-005 (2계층 RBAC 인가 모델)**:
  - **Layer 1 (플랫폼 전역 역할)**: `USER`, `ADVERTISER`, `SUPPORT_AGENT`, `TRUST_SAFETY_MOD`, `AD_REVIEWER`, `FINANCE_ADMIN`, `SUPER_ADMIN`.
  - **Layer 2 (개별 블로그 멤버 역할)**: `OWNER`(소유자 1인), `ADMIN`(관리자), `EDITOR`(편집자), `AUTHOR`(필진), `ANALYST`(통계 열람자).

## 4. 데이터 구조 (Data Structures & Domain Models)
```go
type IdentityClaims struct {
    UserID        uuid.UUID `json:"sub"`
    SessionID     uuid.UUID `json:"sid"`
    PlatformRoles []string  `json:"roles"`
    MFAVerified   bool      `json:"amr_mfa"`
    IssuedAt      int64     `json:"iat"`
    ExpiresAt     int64     `json:"exp"`
}
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- 회원가입(`SCR-MEM-001`), 로그인(`SCR-MEM-002`), 비밀번호 복구(`SCR-MEM-003`), 프로필 및 보안/세션/MFA 설정(`SCR-MEM-004`), 계정 탈퇴(`SCR-MEM-015`) 화면 제공.
- 라우트 가드(`middleware/auth.ts`, `middleware/blog-role.ts`, `middleware/admin-role.ts`)를 통한 권한별 페이지 접근 제어.

## 6. 백엔드 로직 (Backend Logic & Processing)
- `internal/modules/identity`:
  - `Argon2id` 해셔, JWT 발급/검증기, OIDC Relying Party 클라이언트, TOTP 검증기, RBAC Policy Evaluator.
  - 로그인 브루트포스 방어: 동일 이메일 또는 IP에서 5회 연속 실패 시 지수적 대기 시간 부여 및 10회 실패 시 15분간 잠금 + 보안 알림 이메일 발송.

## 7. API 계약 (API Contracts)
- `POST /api/v1/auth/register` / `POST /api/v1/auth/verify-email`
- `POST /api/v1/auth/login` / `POST /api/v1/auth/logout` / `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/password/forgot` / `POST /api/v1/auth/password/reset`
- `GET /api/v1/auth/oauth/{provider}/start` / `GET /api/v1/auth/oauth/{provider}/callback`
- `POST /api/v1/auth/mfa/setup` / `POST /api/v1/auth/mfa/verify` / `DELETE /api/v1/auth/mfa`
- `GET /api/v1/me/sessions` / `DELETE /api/v1/me/sessions/{sessionId}`
- `POST /api/v1/me/deactivate` (계정 탈퇴 신청 — 유예기간 및 소유 블로그 처리 옵션 포함)

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- 주요 테이블: `users`, `user_profiles`, `identities`, `sessions`, `roles`, `permissions`, `role_permissions`, `user_roles`, `user_consents`, `blog_members`.
- 상세 권한 매트릭스는 `04_API/AUTHORIZATION_MATRIX.md`에 정의됨.

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **계정 탈퇴 시 소유 블로그 처리 예외**: 단독 소유 블로그가 있거나 미정산 수익금이 남아 있는 경우, 탈퇴 위저드에서 1) 블로그 공동 관리자에게 소유권 양도 또는 2) 블로그 함께 폐쇄 신청을 명시적으로 선택해야 탈퇴가 진행된다(고아 블로그 발생 방지).

## 10. UX/UI 구조 (UX/UI Structure)
- 비밀번호 입력 시 실시간 강도 게이지 제공, 활성 세션 목록에서 현재 접속 기기 표시 및 '다른 모든 기기에서 로그아웃' 원클릭 버튼 제공.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- 비밀번호 평문 로깅 절대 금지.
- OAuth/OIDC 로그인 시 `state` 파라미터와 `PKCE (code_challenge / code_verifier)`를 필수로 사용하여 CSRF 및 인가 코드 가로채기 공격을 방어한다.

## 12. 성능 및 확장성 (Performance & Scalability)
- Access Token 검증은 DB 조회 없이 메모리에 로드된 공개키(JWKS)로 0.2ms 이내에 수행하며, 긴급 정지된 계정/세션 블랙리스트만 Redis(`revoked:session:{sid}`)에서 O(1)로 확인한다.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- `golang.org/x/crypto/argon2`, `github.com/golang-jwt/jwt/v5`, `github.com/coreos/go-oidc/v3/oidc`, `github.com/pquerna/otp`.

## 14. 개발 우선순위 (Development Priority)
- **Phase 1 (최우선 코어 모듈)**: 모든 인증/권한 기반 (`TB-AUTH-001`, `TB-AUTH-002`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 선행: `A_ARCHITECTURE`, `P_POSTGRESQL`.
- 후행: 전체 모듈.

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-IAM-001`: 이미 사용된 Refresh Token으로 재갱신 시도 시 해당 세션 전체가 즉시 무효화(Revoked)되는지 검증.
- `TEST-IAM-002`: 로컬 계정 사용자가 `TREESCAN_OIDC` 프로바이더를 추가 연동한 뒤에도 기존 `blogs.owner_user_id`와 `posts.author_user_id`가 100% 보존되는지 검증.

## 17. Acceptance Criteria (인수 기준)
1. 회원가입, 로그인, 토큰 갱신, MFA 설정, 소셜/OIDC 연동, 2계층 RBAC 권한 검사가 모든 API에서 누락 없이 동작해야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- 인증/인가 통합 테스트 통과 로그, RBAC 권한 위반 시 `403` 응답 테스트 결과.

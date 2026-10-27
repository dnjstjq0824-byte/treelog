# SSO_INTEGRATION.md — Booking Network / TreeScan OIDC 통합 로그인 및 계정 무손실 병합 명세 (제1.3조)

## 1. 기존 데이터 및 블로그 소유권 100% 보존 아키텍처
1. **분리된 Identity 구조**:
   - TreeBlog의 모든 소유권(`blogs.owner_user_id`, `posts.author_user_id`, `follows.follower_user_id`, `ledger_entries.owner_user_id`)은 외부 인증 제공자와 무관한 TreeBlog 내부 불변 PK인 **`users.id` (UUID)**에 귀속되어 있다.
2. **안전한 계정 연동(Account Linking) 절차 (`POST /api/v1/treescan/sso/link`)**:
   - Step 1: 기존 TreeBlog 사용자가 자신의 TreeBlog 계정으로 로그인(필요 시 MFA 검증)한다.
   - Step 2: 사용자가 'Booking Network / TreeScan 통합 계정 연결'을 시작하여 OIDC Authorization Code + PKCE 플로우를 완료하고 ID Token을 획득한다.
   - Step 3: TreeBlog 백엔드는 ID Token의 서명(`JWKS`), `iss`, `aud`, `exp`를 검증한 뒤, `identities` 테이블에 `(user_id = 기존_users_id, provider = 'TREESCAN_OIDC', provider_subject = sub)` 행을 추가한다.
   - **결과**: `users.id`가 전혀 바뀌지 않으므로 기존에 개설한 블로그, 작성한 게시글, 이웃 관계, 정산 잔액이 **단 1건의 손실이나 마이그레이션 다운타임 없이 100% 그대로 보존**된다.

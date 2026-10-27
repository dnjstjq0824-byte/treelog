# SECURITY_ARCHITECTURE.md — 심층 방어(Defense-in-Depth) 보안 아키텍처 명세 (제16조)

## 1. 계층별 보안 통제 요약
1. **Edge / Network Layer**: TLS 1.3 강제, HSTS Preload, DDoS L3/L4/L7 방어, IP 평판 필터링.
2. **API Middleware Layer**:
   - JWT 서명(`RS256`/`EdDSA`) 검증 및 세션 폐기 여부 확인.
   - 상태 변경 요청(`POST`, `PUT`, `PATCH`, `DELETE`) 전체에 CSRF 토큰(`X-CSRF-Token`) + `Origin` 헤더 교차 검증.
   - Redis Sliding Window Rate Limiter (계정별/IP별/엔드포인트별 세분화).
3. **Application & Domain Layer**:
   - 2계층 RBAC (`RequireBlogRole`, `RequireAdminRole`) 및 게시글 가시성(`CanReadPost`) 강제.
   - Tiptap AST 스키마 검증 + `bluemonday` HTML 정화 + 커스텀 CSS AST 정화.
   - 외부 아웃바운드 HTTP 호출 시 `SafeDialer`를 통한 SSRF 원천 차단.
4. **Data Layer**:
   - 100% 파라미터화 SQL(`pgx`/`sqlc`)로 SQL Injection 원천 배제.
   - 비밀번호 `Argon2id` 해싱, MFA 시크릿 및 정산 계좌번호 `AES-256-GCM` 암호화 저장.
   - 관리자 및 보안 이벤트의 해시 체인 기반 불변 `audit_logs` 기록.

# BLOG_DOMAINS.md — 서브도메인 · 경로형 · 개인 커스텀 도메인 및 자동 TLS 아키텍처 (제6조)

## 1. 3종 주소 라우팅 및 Canonical 정규화 규칙
TreeBlog는 모든 블로그에 대해 다음 3가지 접근 경로를 지원하되, 검색엔진 중복 색인(Duplicate Content Penalty)을 막기 위해 블로그당 단 하나의 **Primary URL**만 Canonical로 지정한다.

| 주소 유형 | URL 구조 예시 (운영 도메인 확정 전 예시) | DNS 요구사항 | 기본 동작 |
| :--- | :--- | :--- | :--- |
| **1) 플랫폼 서브도메인** | `https://{handle}.blog.example.com` | 플랫폼 와일드카드 DNS (`*.blog.example.com`) | 기본 Primary URL |
| **2) 플랫폼 경로형** | `https://blog.example.com/@{handle}` | 플랫폼 루트 도메인 | 서브도메인이 Primary일 경우 `301` 리다이렉트 (또는 선택 시 Primary로 지정 가능) |
| **3) 개인 커스텀 도메인** | `https://blog.user-domain.com` 또는 `https://user-domain.com` | 사용자 DNS에 `_treeblog-verify` TXT + CNAME/A 설정 | 검증 및 TLS 활성화 후 Primary 지정 시 서브도메인/경로형 유입을 개인 도메인으로 `301` 리다이렉트 |

## 2. 개인 도메인 소유권 검증 및 도메인 탈취(Hijacking) 방어
1. **TXT 소유권 검증 선행 필수**:
   - 사용자가 `blog.alice.com`을 등록하면 서버는 고유 난수 토큰 `tb-verify=9f8e7d6c5b4a...`를 생성한다.
   - 사용자가 자신의 DNS에 `_treeblog-verify.blog.alice.com TXT "tb-verify=9f8e7d6c5b4a..."` 및 `blog.alice.com CNAME cname.treeblog.example.com`을 등록한 뒤 검증을 통과해야만 라우팅이 활성화된다.
2. **On-Demand ACME TLS 자동 발급·갱신**:
   - 역방향 프록시(Caddy / Cert-Manager)는 알 수 없는 호스트의 TLS ClientHello 수신 시 내부 API `GET /internal/v1/domains/ask?domain={host}`를 호출한다.
   - DB에서 `verification_status == 'VERIFIED'`가 확인된 도메인에 대해서만 `200 OK`를 반환하여 Let's Encrypt 인증서를 즉시 발급하고 만료 30일 전 자동 갱신한다.

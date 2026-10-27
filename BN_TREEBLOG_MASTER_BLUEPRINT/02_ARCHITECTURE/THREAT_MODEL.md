# THREAT_MODEL.md — TreeBlog STRIDE 위협 모델링 및 방어 아키텍처

## 1. STRIDE 기반 핵심 위협 분석 및 통제 대책 (제16조 연계)

| 위협 분류 | 구체적 공격 시나리오 | 영향도 | 아키텍처 및 코드 수준 방어 대책 |
| :--- | :--- | :--- | :--- |
| **Spoofing (위장)** | 타인의 커스텀 도메인을 자신의 블로그에 먼저 등록하여 트래픽 가로채기 (Domain Hijacking / Subdomain Takeover) | Critical | 1) 모든 커스텀 도메인은 고유 난수 TXT 레코드(`_treeblog-verify`) 검증 통과 전까지 라우팅 및 TLS 발급 절대 비활성화.<br>2) 주기적 DNS 재검증 워커(매 6시간)가 CNAME 이탈 감지 시 즉시 도메인 연결을 `SUSPENDED`로 전환. |
| **Tampering (변조)** | 에디터 API 요청 시 정제되지 않은 악성 `<script>`나 `onerror` 속성을 포함한 HTML/JSON을 직접 주입 (Stored XSS) | Critical | 1) 클라이언트가 보낸 HTML은 신뢰하지 않으며, 오직 Tiptap JSON AST만 수신하여 서버에서 HTML로 직렬화.<br>2) 직렬화된 HTML도 엄격한 허용목록 기반 `bluemonday` 정책으로 2차 정화.<br>3) 응답 헤더에 강력한 `Content-Security-Policy` 적용. |
| **Repudiation (부인)** | 관리자가 특정 블로그를 정지시키거나 정산 금액을 조작한 뒤 행위를 부인 | High | 1) 모든 관리자 액션, 권한 변경, 도메인 변경, 금전 정산 처리는 불변 테이블 `audit_logs`에 행위자 ID, IP, User-Agent, 변경 전/후 JSON Diff, 해시 체인을 기록.<br>2) `audit_logs` 테이블은 애플리케이션 DB 계정에서 `UPDATE`/`DELETE` 권한을 회수(INSERT-only). |
| **Information Disclosure (정보 노출)** | 비공개(`PRIVATE`) 또는 상호이웃 공개(`MUTUAL_ONLY`) 게시글이 OpenSearch 검색 결과, Sitemap, RSS, 또는 캐시를 통해 제3자에게 노출 | Critical | 1) 비공개/제한 공개 게시글은 OpenSearch 공개 인덱스 색인 대상에서 원천 제외.<br>2) 공개에서 비공개로 전환 시 최우선 큐로 색인 즉시 삭제 + API 반환 전 PostgreSQL 실시간 권한 재검증.<br>3) 비공개/이웃공개 응답에는 `Cache-Control: private, no-store` 및 `X-Robots-Tag: noindex, nofollow` 강제. |
| **Denial of Service (서비스 거부)** | 대용량 이미지 폭탄(Decompression Bomb / Pixel Flood) 업로드 또는 검색 API 대량 호출로 서버 자원 고갈 | High | 1) 업로드는 S3 Presigned URL(최대 크기 Content-Length 강제)을 사용.<br>2) 미디어 워커는 파일 전체를 메모리에 로드하기 전 헤더만 읽어 최대 해상도(예: 16,384 x 16,384px) 초과 시 즉시 파기.<br>3) IP/계정별 Redis Sliding Window Rate Limiter 적용. |
| **Elevation of Privilege (권한 상승)** | 일반 블로거가 요청 파라미터의 `blog_id`를 조작(IDOR)하여 타인의 블로그 설정 변경 또는 유료 정산금 출금 시도 | Critical | 1) 모든 블로그 관리 API는 미들웨어 수준에서 `RequireBlogRole(blogID, userID, minRole)`을 통과해야 실행됨.<br>2) 정산 출금은 2FA(MFA) 재인증 및 본인 명의 계좌 검증 완료 상태에서만 허용. |

## 2. SSRF (Server-Side Request Forgery) 특별 방어 설계
블로그 에디터의 **외부 링크 OG 메타태그 프리뷰 수집** 및 **외부 블로그 가져오기(Import)** 기능은 서버가 사용자가 입력한 URL로 HTTP 요청을 보내므로 SSRF의 주요 표적이 된다.
- **방어 설계 (`internal/pkg/safenet/safe_dialer.go`)**:
  1. `http.Transport`의 커스텀 `DialContext` 단계에서 DNS 리졸브 직후 연결할 대상 IP를 검사한다 (`DNS Rebinding` 공격 완벽 차단).
  2. 루프백(`127.0.0.0/8`, `::1`), 사설 대역(`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`), 링크 로컬 및 클라우드 메타데이터 IP(`169.254.0.0/16`, `fd00::/8`)로의 연결을 즉시 차단(`ErrForbiddenIPRange`)한다.
  3. 프로토콜은 오직 `http://`와 `https://`만 허용하며, 포트 `80`, `443` 외 비표준 포트 접근을 차단한다.
  4. 최대 리다이렉트 횟수 3회 제한, 응답 본문 최대 크기 2MB(`io.LimitReader`), 타임아웃 5초를 강제한다.

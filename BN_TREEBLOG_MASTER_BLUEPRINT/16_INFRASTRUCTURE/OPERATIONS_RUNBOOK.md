# OPERATIONS_RUNBOOK.md — 주요 장애 시나리오별 표준 운영 런북 (SOP)

## 1. 시나리오별 즉각 대응 절차
### SOP-01: OpenSearch 클러스터 과부하 또는 응답 불가 시
1. **증상**: `/api/v1/search/*` 지연 증가 및 `SEARCH_TEMPORARILY_UNAVAILABLE` 알람 발생.
2. **자동 격리 동작**: Go API의 Circuit Breaker가 자동으로 열려 게시글 작성·발행·열람·댓글 등 코어 PostgreSQL 트랜잭션은 100% 정상 유지되며, 색인 작업은 `search_index_jobs` 테이블에 안전하게 큐잉된다.
3. **복구 절차**: OpenSearch 노드 복구 후 관리자 콘솔(`SCR-ADM-006`) 또는 CLI에서 `treeblog-worker --replay-search-jobs`를 실행하여 적체된 색인 잡을 순차 반영한다.

### SOP-02: 특정 스팸 봇넷의 대량 회원가입 및 불법 게시글 살포 공격 시
1. 관리자 콘솔(`SCR-ADM-008`)에서 `FF_STRICT_CAPTCHA_CHALLENGE` 플래그를 즉시 `true`로 전환하여 모든 가입/글쓰기에 Turnstile Captcha를 강제한다.
2. 공격 IP 대역 또는 스팸 패턴으로 생성된 블로그들을 `SCR-ADM-003`에서 일괄 `SUSPENDED` 처리하면 Outbox `critical` 워커가 1초 이내에 모든 관련 문서를 검색 색인과 캐시에서 삭제한다.

### SOP-03: 사용자 개인 커스텀 도메인 TLS 갱신 실패 시
1. `blog_domains.tls_status = 'RENEWAL_FAILED'` 알람 확인 → 대상 도메인의 CNAME/A 레코드가 여전히 TreeBlog Ingress를 가리키는지 `POST /api/v1/blogs/{id}/domains/{domId}/verify`로 점검.
2. 사용자가 DNS를 변경한 경우 블로거에게 안내 알림을 자동 발송하고, 블로그 접근은 기본 서브도메인(`{handle}.treeblog.example.com`)으로 무중단 폴백시킨다.

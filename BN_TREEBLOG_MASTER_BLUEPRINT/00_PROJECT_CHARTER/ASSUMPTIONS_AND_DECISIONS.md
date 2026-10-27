# ASSUMPTIONS_AND_DECISIONS.md — 아키텍처 의사결정 기록 (ADR) 및 설계 가정

프로젝트 요구사항 중 운영 환경 확정 전이거나 합리적인 기술적 선택이 필요한 사항에 대해 다음과 같이 설계 가정(Assumptions)과 아키텍처 결정(ADR)을 수립하고 그 근거를 명시한다.

---

## 1. 합리적 설계 가정 (Design Assumptions)

| ID | 영역 | 설계 가정 내용 | 선택 근거 및 변경 가능성 |
| :--- | :--- | :--- | :--- |
| **ASM-001** | **기본 도메인** | 명세서 내 예시 기본 도메인은 `treeblog.example.com` (또는 환경변수 `PLATFORM_BASE_DOMAIN`)으로 표기한다. | 제6조 지침에 따라 특정 운영 도메인을 하드코딩하지 않고 환경변수 주입 방식으로 설계하여 도메인 확정 즉시 무수정 배포가 가능하도록 함. |
| **ASM-002** | **블로그 주소 체계** | 블로그 생성 시 고유 `handle`(영문 소문자, 숫자, 하이픈, 3~30자)을 발급하며, 이를 통해 경로형(`/{handle}` 또는 `/@{handle}`)과 서브도메인형(`{handle}.{base_domain}`)을 동시에 해석할 수 있도록 설계한다. 단, 대표 Canonical URL은 블로그 설정(`primary_url_type`)의 값 1개로만 고정한다. | 사용자 편의성과 검색엔진 중복 색인 방지(제6조, 제9조)를 동시에 만족하기 위함. |
| **ASM-003** | **결제 대행사(PG)** | 글로벌 결제는 **Stripe**(해외 카드/구독), 한국 로컬 결제는 **Toss Payments / PortOne** 어댑터 인터페이스를 표준으로 설계한다. | 실제 PG 가맹점 심사 및 금융 계약은 법인/운영 국가 확정 후 PO 승인이 필요하므로, 코어 결제 모듈은 `PaymentGatewayProvider` 인터페이스로 추상화함. |
| **ASM-004** | **이메일/푸시 발송** | 트랜잭션 이메일 및 뉴스레터는 **AWS SES / SMTP** 추상화, 웹 푸시는 **W3C Web Push (VAPID)** 표준을 기본 사용한다. | 벤더 종속성을 제거하고 로컬 개발 환경에서는 Mailpit/Mock으로 100% 검증 가능하도록 구성함. |
| **ASM-005** | **커스텀 도메인 TLS** | 사용자 개인 도메인 연결 시 **Caddy On-Demand TLS** 또는 **Cert-Manager / Go ACME Client (Let's Encrypt)**를 통해 도메인 소유권(`blog_domains.verification_status = 'VERIFIED'`)이 확인된 도메인에 한해서만 인증서를 발급한다. | 미검증 도메인에 대한 무차별 ACME 요청 공격(Rate Limit 고갈)을 방어하기 위함. |

---

## 2. 핵심 아키텍처 의사결정 기록 (Architecture Decision Records)

### ADR-001: 마이크로서비스 대신 Go 모듈형 모놀리스 + 비동기 워커 채택
- **상태**: 채택됨 (Accepted — 제2조 준수)
- **맥락**: 블로그 플랫폼은 게시글 조회 시 작성자 프로필, 블로그 설정, 테마, 카테고리, 이웃 권한, 공감/댓글 수 등 다수의 도메인 데이터가 결합되어야 한다. 초기부터 MSA로 분리하면 네트워크 홉(gRPC/REST) 증가로 SSR 응답 시간이 느려지고 트랜잭션 일관성 유지가 어렵다.
- **결정**:
  1. 동기 요청을 처리하는 `treeblog-api`는 단일 Go 바이너리로 빌드하되, 내부 패키지를 15개 도메인 모듈로 엄격히 분리한다.
  2. 무거운 작업(미디어 리사이징, 검색 색인, 알림 발송, 예약 발행, 정산 배치)은 `treeblog-worker` 바이너리로 분리하여 Redis Queue 및 PostgreSQL `outbox_events`를 통해 비동기 실행한다.
- **결과**: 단일 DB 트랜잭션으로 데이터 정합성을 완벽히 보장하면서도, 향후 트래픽 급증 시 특정 모듈(예: Media, Search, Ad Server)만 별도 서비스로 손쉽게 추출 가능.

### ADR-002: 에디터 원본 포맷으로 Tiptap/ProseMirror JSON AST 채택 및 3중 저장 모델
- **상태**: 채택됨 (Accepted — 제5조 준수)
- **맥락**: 순수 HTML만 저장하면 에디터 버그나 보안 취약점(XSS)에 취약하며, 순수 Markdown만 저장하면 복잡한 레이아웃(갤러리 그리드, 커스텀 콜아웃, 임베드 카드, 이미지 정렬/크기) 표현에 한계가 있다.
- **결정**: `posts` 및 `post_revisions` 테이블에 다음 3가지 표현을 분리 저장한다.
  1. `content_json` (`JSONB`): Tiptap/ProseMirror 표준 AST 원본 데이터 (에디터 로드 및 리비전 관리용)
  2. `content_html` (`TEXT`): 서버 사이드 렌더러 및 `bluemonday` Sanitizer를 거친 안전한 렌더링용 HTML (SSR 고속 응답용)
  3. `content_plain` (`TEXT`): 모든 마크업이 제거된 순수 텍스트 (OpenSearch 전문 검색 색인, 요약문 추출, 읽기 소요 시간 계산, 스팸/품질 분석용)
- **결과**: 에디터 확장성, XSS 완벽 방어, SSR 렌더링 속도, 검색 색인 품질을 동시에 달성. Markdown은 가져오기/내보내기(Import/Export) 및 에디터 내 단축키 변환으로 완벽 지원.

### ADR-003: 팔로우(Follow) · 이웃(Neighbor) · 상호이웃(Mutual Neighbor) · 구독(Subscription) 통합 관계 모델
- **상태**: 채택됨 (Accepted — 제8조 준수)
- **맥락**: 글로벌 플랫폼(Medium/X)의 가벼운 팔로우/뉴스레터 구독과 네이버 블로그 특유의 끈끈한 이웃/서로이웃 모델을 충돌 없이 하나의 소셜 그래프로 통합해야 한다.
- **결정**:
  1. **팔로우 (`follows.relation_type = 'FOLLOW'`)**: 상대방 승인 없이 즉시 성립하는 단방향 관계. 추천 피드 및 내 피드 모아보기에 반영.
  2. **이웃 (`follows.relation_type = 'NEIGHBOR'`)**: 상대방 블로그를 내 '이웃 목록(그룹별 정리 가능)'에 추가한 상태. 단방향 성립.
  3. **상호이웃 (`follows.relation_type = 'MUTUAL_NEIGHBOR'`)**: 내가 신청(`PENDING`)하고 상대방이 수락(`ACCEPTED`)할 때 양방향으로 성립하는 친밀 관계. **`visibility = 'MUTUAL_ONLY'`(상호이웃 공개) 게시글을 열람할 수 있는 유일한 권한**을 부여.
  4. **구독 (`subscriptions`)**: 새 글 발행 시 알림 수신 채널(웹 알림, 앱 푸시, 이메일 뉴스레터) 및 유료 멤버십 티어(`FREE`, `PAID_TIER`)를 제어하는 전달 계약.
- **결과**: 사용자에게 친숙한 한국형 '서로이웃' 커뮤니티 경험과 글로벌 '뉴스레터/유료 구독' 경험을 하나의 일관된 권한 매트릭스로 제공.

### ADR-004: Transactional Outbox 기반 OpenSearch 및 TreeScan 비동기 색인 동기화
- **상태**: 채택됨 (Accepted — 제10조, 제18조 준수)
- **맥락**: 게시글 발행/비공개 전환/삭제 시 PostgreSQL과 OpenSearch에 동시 쓰기(Dual Write)를 수행하면 네트워크 일시 장애 시 DB에는 삭제되었으나 검색엔진에는 남아있는 치명적 보안/프라이버시 사고가 발생할 수 있다.
- **결정**: 모든 콘텐츠 상태 변경 트랜잭션은 PostgreSQL `outbox_events` 및 `search_index_jobs` 테이블에 작업을 기록하고 커밋한다. 워커가 이를 읽어 OpenSearch에 반영하며, 특히 **비공개 전환/삭제/제재(`DELETE_FROM_INDEX`) 작업은 최우선 순위 큐(`priority = 100`)로 즉시 처리**하고 API 조회 시점에도 DB 상태를 교차 확인하여 비공개 글 노출을 0%로 차단한다.

---

## 3. 외부 승인 및 법무/비즈니스 확정 필요 사항 (Unverified / Pending PO Approval)

제25조 및 제28조에 따라, 기술 설계는 완료되었으나 실제 프로덕션 런칭 전 발주자(PO) 및 법무/재무팀의 최종 확정이 필요한 항목을 명확히 분리한다.

1. **PG사 실계약 및 가맹점 키 발급 (`PENDING_EXTERNAL_APPROVAL`)**: Stripe / Toss Payments 실운영 API 키, 정산 주기, 국가별 원천징수 세율(한국 3.3% 사업소득세 또는 미국 W-8BEN/W-9 세무 폼 처리 정책).
2. **국가별 개인정보처리방침 및 이용약관 법률 감수 (`PENDING_LEGAL_REVIEW`)**: `11_SECURITY_MODERATION/COMPLIANCE_CHECKLIST.md`에 기술적 요구사항(GDPR, CCPA, 한국 개인정보보호법 및 정보통신망법 임시조치 30일 규정, DMCA Safe Harbor)을 모두 설계하였으나, 실제 약관 문안은 법무 검토가 필요함.
3. **TreeScan 프로덕션 연동 엔드포인트 및 OIDC Client Secret (`PENDING_TREESCAN_READINESS`)**: Phase 10 이후 TreeScan 운영팀과의 합의를 통해 실제 엔드포인트 URL 및 상호 mTLS/OIDC 크레덴셜 발급 필요.

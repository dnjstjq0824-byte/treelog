# DEVELOPMENT_PRINCIPLES.md — 핵심 아키텍처 및 개발 원칙

## 제1원칙: 100% 독립 운영 우선 및 느슨한 TreeScan 통합 (Independence First)
1. **Zero Dependency on TreeScan at Runtime**: TreeBlog는 TreeScan의 서버, DB, 인증, 검색 클러스터가 전혀 존재하지 않는 상태에서도 회원가입, 로그인, 블로그 개설, 글 발행, 자체 검색(OpenSearch), 미디어 서빙, 수익화, 관리자 콘솔이 100% 완벽히 동작해야 한다.
2. **Adapter Pattern for Future Integration**: 향후 TreeScan 및 Booking Network 타 플랫폼과의 연동은 코어 도메인 로직을 수정하지 않고 `internal/integration/treescan/` 패키지의 Adapter와 `outbox_events` 테이블 기반 비동기 이벤트 스트림으로만 처리한다.
3. **검색 공정성 (Search Neutrality)**: 향후 TreeScan에 TreeBlog 콘텐츠를 색인 연동하더라도 외부 웹 문서보다 인위적으로 가중치를 부여하는 조작 로직을 절대 포함하지 않는다. 오직 표준 검색 품질 시그널(본문 충실도, 구조화 데이터, 인용/참조, 스팸 점수)만을 투명하게 제공한다.

## 제2원칙: 모듈형 모놀리스 우선, 단계적 분리 (Modular Monolith + Async Workers)
1. **무조건적인 마이크로서비스(MSA) 분할 금지**: 초기부터 수십 개의 마이크로서비스로 쪼개어 발생하는 분산 트랜잭션 장애, 네트워크 지연, 운영 복잡도를 방지한다.
2. **엄격한 모듈 경계(Strict Module Boundaries)**: Go 백엔드는 단일 배포 바이너리(`treeblog-api`)와 워커 바이너리(`treeblog-worker`)로 구성하되, `internal/modules/{auth,user,blog,post,editor,media,community,search,seo,theme,domain,monetization,moderation,admin,treescan}` 디렉터리 간에는 오직 정의된 인터페이스(Service Port)와 도메인 이벤트로만 통신한다. 타 모듈의 DB 테이블을 직접 JOIN하거나 수정하는 쿼리를 금지한다(조회 성능을 위한 명시적 Read Model 뷰 제외).
3. **워커 프로세스 분리**: 이미지 변환(libvips), OpenSearch 색인 동기화, 대량 이메일/푸시 알림, 예약 발행 스케줄러, 정산 집계는 API 서버의 응답 지연(Latency)에 영향을 주지 않도록 독립된 `treeblog-worker` 컨테이너 그룹에서 처리한다.

## 제3원칙: 원본 데이터 단일 진실 공급원 (PostgreSQL as Single Source of Truth)
1. **PostgreSQL 16+ 중심 설계**: 모든 사용자, 블로그, 게시글, 리비전, 커뮤니티 관계, 금전 원장(Ledger)의 원본은 PostgreSQL에 저장한다.
2. **Transactional Outbox 패턴**: 게시글 발행/수정/삭제 시 PostgreSQL 트랜잭션 내부에서 `posts` 테이블 변경과 함께 `outbox_events` 테이블에 이벤트를 기록한다. 워커가 Outbox를 폴링(또는 CDC)하여 OpenSearch 색인, 캐시 무효화, 알림 발송, TreeScan 웹훅을 수행함으로써 Dual-Write 불일치 문제를 원천 해결한다.
3. **OpenSearch 재구축 가능성**: OpenSearch 클러스터가 전소되더라도 PostgreSQL의 원본 데이터를 기반으로 전체 검색 색인을 100% 무손실 재구축(Reindex)할 수 있어야 한다.

## 제4원칙: 보안 기본 내장 (Security & Privacy by Design)
1. **비공개 데이터 완벽 격리**:
   - 게시글 상태(`DRAFT`, `SCHEDULED`, `PUBLISHED`, `UNLISTED`, `PRIVATE`, `MUTUAL_ONLY`, `NEIGHBOR_ONLY`, `QUARANTINED`, `DELETED`)를 API 레이어와 DB 쿼리 양쪽에서 검증한다.
   - `PUBLISHED` + `visibility = 'PUBLIC'` + `blog.status = 'ACTIVE'` 조건을 모두 만족하는 게시글만 OpenSearch 공개 인덱스, Sitemap, RSS, 추천 피드에 포함된다.
2. **사용자 입력 무결성 검증 및 XSS 원천 차단**:
   - 에디터 본문은 Tiptap/ProseMirror JSON AST(`content_json`)를 원본으로 저장하고, 서버에서 엄격한 스키마 검증 후 허용된 HTML 태그/속성만 통과시키는 Sanitizer(`bluemonday`)를 거쳐 `content_html`을 생성한다.
   - 사용자 커스텀 스킨/테마는 CSS 변수(Design Tokens)와 안전한 레이아웃 템플릿 조합을 기본으로 하며, 고급 HTML/CSS 모드는 스크립트 제거 및 격리된 도메인/CSP 정책 하에서만 동작한다.
3. **EXIF 개인정보 자동 파기**: 사용자가 업로드하는 모든 사진(JPEG, PNG, WebP, HEIC 등)은 공개 스토리지로 복사되기 전 GPS 위치 좌표, 카메라 시리얼 번호 등 민감 EXIF 메타데이터를 강제로 제거한다(색공간 ICC 프로필과 회전 방향만 반영 후 정규화).

## 제5원칙: 금전 거래 및 정산의 무결성 (Double-Entry Ledger)
1. **복식부기 원장(`ledger_entries`)**: 광고 충전, 광고비 차감, 블로거 광고 수익 배분, 멤버십 결제, 후원금, 플랫폼 수수료, 정산 지급, 환불/환수는 모두 불변(Append-only) 복식부기 원장 테이블(`ledger_entries`)에 차변(Debit)/대변(Credit) 쌍으로 기록한다.
2. **Idempotency 보장**: 모든 결제, 정산, 포인트/수익 변동 API 및 워커 잡은 고유 `idempotency_key`를 강제하여 네트워크 재시도로 인한 이중 결제나 이중 지급을 방지한다.

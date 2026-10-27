# SCOPE_AND_NON_GOALS.md — 프로젝트 범위(In-Scope) 및 비목표(Non-Goals)

## 1. 프로젝트 포함 범위 (In-Scope)

### 1.1 이번 마스터 블루프린트 패키지(`BN_TREEBLOG_MASTER_BLUEPRINT_V1_0.zip`)의 범위
- **A~Z 26개 전 영역 상세 기술 명세서** (18개 필수 섹션 완비)
- **전체 시스템 아키텍처, 데이터 흐름, 위협 모델**
- **PostgreSQL 58개 테이블 전체 DDL, 인덱스, 파티셔닝, 마이그레이션 SQL**
- **OpenAPI 3.1.0 기반 24개 그룹 122개 REST API 엔드포인트 명세**
- **45개 전체 화면(사용자 웹, 개인 블로그, 에디터, 블로그 관리, 플랫폼 어드민) 반응형 UX/UI 및 접근성/SEO 설계**
- **Tiptap/ProseMirror 블록 에디터 엔진, 자동저장, 리비전, 발행 워크플로 명세**
- **블로그 개설, 멀티 도메인(서브도메인/경로/개인 도메인) 및 자동 TLS, 테마 엔진 명세**
- **팔로우·이웃·상호이웃·구독 커뮤니티 및 멀티채널 알림 엔진 명세**
- **OpenSearch 다국어 검색엔진 및 글로벌 SEO/Sitemap/RSS 명세**
- **미디어 업로드·이미지 변환·동영상 임베드 파이프라인 명세**
- **광고 네트워크, 블로거 수익 배분, 멤버십, 후원, 정산 원장 명세**
- **보안, 스팸 탐지, 신고·제재·이의신청 모더레이션 및 글로벌 법률 검토 체크리스트**
- **48개 실행형 Work Order(`TB-*-*`) 및 13개 에이전트 협업 프로토콜, Phase 0~10 로드맵, 90개 마스터 테스트 매트릭스**
- **TreeScan 및 Booking Network 연동을 위한 독립 Adapter/SSO/색인 계약 명세**

### 1.2 후속 개발 에이전트가 본 패키지를 통해 구현할 시스템 범위
- **Phase 0 ~ Phase 7 (코어 플랫폼 상용화 필수 범위)**: 인프라/저장소 초기화, 회원/인증/RBAC, 블로그 개설/도메인/테마, 블록 에디터/발행, 사용자 웹/반응형 UI, OpenSearch/SEO/RSS, 이웃/댓글/공감/알림, 관리자 콘솔/보안/모더레이션, 데이터 이동(Import/Export).
- **Phase 8 ~ Phase 9 (수익화 및 글로벌 상용화 범위)**: 통계/애널리틱스, 광고 시스템/수익 배분, 유료 멤버십/후원/정산, 다국어/성능/장애 복구 감사.
- **Phase 10 (TreeScan 통합 준비 범위)**: TreeScan Search API Adapter, OIDC SSO 연동 모듈, 색인 동기화 Outbox Publisher 구현 및 Contract Test 검증.

---

## 2. 명시적 비목표 (Non-Goals & Out-of-Scope)

과도한 엔지니어링(Over-engineering)과 일정 지연을 방지하기 위해 다음 항목은 초기 구현 범위에서 명시적으로 제외하거나 별도 확장 단계로 분리한다.

1. **TreeScan 기존 서버·DB·소스코드의 직접 수정 금지 (제1조, 제18조)**
   - 본 프로젝트에서는 TreeScan 측 저장소나 데이터베이스를 일절 건드리지 않는다. TreeBlog 측에 Adapter와 Webhook/API 엔드포인트를 준비하는 것까지만 수행한다.
2. **실시간 다중 사용자 동시 편집 (CRDT/Yjs Collaborative Editing) 초기 제외 (제5조)**
   - 팀 블로그에서도 초기에는 **단일 작성자 편집 잠금(Optimistic Lock + Lease Lock) 및 자동저장/버전 충돌 방지**를 우선 구현한다. Google Docs 식의 실시간 다중 커서 동시 편집(Yjs WebSocket 서버)은 확장 아키텍처 인터페이스만 정의하고 초기 출시 범위에서는 제외한다.
3. **사용자의 서버 사이드 임의 코드(PHP, Node.js, 서버 템플릿 스크립트) 실행 금지 (제7조)**
   - WordPress처럼 사용자가 서버에서 실행되는 플러그인이나 스크립트를 업로드하는 기능은 보안상 영구히 허용하지 않는다. 스킨 커스터마이징은 안전한 선언형 템플릿 변수와 Sanitized HTML/CSS 범위 내에서만 동작한다.
4. **초기 자체 동영상 트랜스코딩 클러스터 구축 보류 (제13조)**
   - Phase 1~7 초기 출시에서는 **고성능 이미지 처리(WebP/AVIF 변환, 썸네일)와 검증된 외부 동영상 임베드(YouTube, Vimeo, Global Video 등 oEmbed/iframe 샌드박스)**를 우선 제공한다. 대용량 원본 동영상 직접 업로드 및 HLS/DASH 멀티 비트레이트 트랜스코딩 워커는 스토리지/대역폭 비용과 운영 복잡도를 고려하여 Phase 8 이후 Feature Flag(`FF_NATIVE_VIDEO_TRANSCODE`)로 단계적 활성화할 수 있도록 큐와 테이블 구조(`media_processing_jobs`)만 사전 완비한다.
5. **타사 블로그 비공개 콘텐츠 크롤링/우회 수집 금지 (제17조)**
   - 외부 블로그 이전(Import) 기능은 사용자가 적법하게 내보낸 파일(WordPress WXR XML, Markdown ZIP, 표준 RSS/Atom) 또는 공식 공개 API만을 사용한다. 타 플랫폼의 로그인 세션을 가로채거나 약관을 위반하는 스크래핑 도구는 개발하지 않는다.

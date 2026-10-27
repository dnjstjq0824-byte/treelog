# TreeBlog (가칭) — 글로벌 독립 블로그 플랫폼 마스터 개발 블루프린트

- **문서 번호**: `BN-TREEBLOG-GEMINI4-ARGON-MASTER-BLUEPRINT-001`
- **버전**: `v1.0.0`
- **작성 주체**: Gemini 4 Argon (Lead System Architect & Master Development Planning Agent)
- **발주처**: Booking Network 플랫폼 총괄 PO
- **패키지명**: `BN_TREEBLOG_MASTER_BLUEPRINT_V1_0.zip`

---

## 1. 프로젝트 개요

**TreeBlog**는 Google Blogger의 간편한 개설·도메인 연결, Tistory의 높은 스킨·도메인 자유도 및 수익화, Naver Blog의 탄탄한 이웃·공감 커뮤니티와 친숙한 에디터, WordPress의 강력한 SEO·데이터 이동성을 하나로 결합한 **상용 등급(Commercial-Grade) 글로벌 독립 블로그 플랫폼**입니다.

### 10대 핵심 플랫폼 지향점 (제0조)
1. **개인 블로그 플랫폼**: 일상 기록, 리뷰, 취미, 포트폴리오를 손쉽게 발행하는 개인 공간
2. **전문 콘텐츠 블로그 플랫폼**: 기술 블로그, 칼럼, 심층 리포트 등 코드 하이라이팅·수식·목차·시리즈를 지원하는 전문가 환경
3. **기업·브랜드 블로그 플랫폼**: 다중 멤버 역할(Owner, Admin, Editor, Author, Analyst)과 승인 워크플로를 갖춘 브랜드 미디어룸
4. **커뮤니티형 블로그 플랫폼**: 팔로우(단방향), 이웃(열람 그룹), 상호이웃(양방향 친밀 관계), 방명록, 공감·댓글 중심의 관계형 생태계
5. **사용자 구독·이웃 시스템**: 이메일 뉴스레터, 웹/앱 푸시, 피드 모아보기를 결합한 독자 리텐션 엔진
6. **독립 도메인형 블로그**: 서브도메인(`{handle}.treeblog.com`), 경로형(`treeblog.com/@{handle}`), 개인 소유 커스텀 도메인(`blog.userdomain.com`) 및 자동 TLS(Let's Encrypt ACME) 지원
7. **광고 및 크리에이터 수익화 플랫폼**: 플랫폼 네이티브 광고 네트워크, 크리에이터 광고 수익 배분(Revenue Share), 유료 멤버십/유료 포스트, 독자 후원 및 복식부기 정산 원장(Ledger)
8. **콘텐츠 검색·추천 플랫폼**: OpenSearch 기반 다국어 형태소 분석(Nori/Kuromoji/ICU), 오타 교정, 토픽 클러스터링 및 개인화 하이브리드 피드
9. **SEO 중심 공개 콘텐츠 플랫폼**: Nuxt 3 SSR/ISR, 완벽한 JSON-LD(`BlogPosting`, `BreadcrumbList`, `Person`), 동적 Sitemap 샤딩, RSS/Atom/WebSub, Core Web Vitals 최적화
10. **향후 TreeScan 검색엔진 통합 플랫폼**: 초기에는 100% 독립 운영되며, 향후 Booking Network의 **TreeScan** 검색엔진 및 자매 서비스(Treewiki, getforTrip, TripScanner 등)와 Adapter/Event 기반으로 즉시 연동 가능한 아키텍처

---

## 2. 지정 기술 스택 요약 (제2조)

| 영역 | 지정 기술 스택 | 아키텍처 채택 근거 및 운영 전략 |
| :--- | :--- | :--- |
| **사용자 웹 (User Web)** | Vue 3 + Nuxt 3 + TypeScript + Tailwind CSS | SSR/ISR 하이브리드 렌더링을 통한 최상위 SEO 확보 및 Core Web Vitals(LCP < 2.0s) 달성 |
| **블로그 관리 센터 (Blog Console)** | Vue 3 + TypeScript + Pinia + Vite | 블로그 개설, 스킨 편집, 통계, 수익 정산, 멤버 관리용 고성능 SPA/SSR 하이브리드 |
| **관리자 콘솔 (Admin Console)** | Vue 3 + TypeScript + Pinia | RBAC 기반 플랫폼 통합 운영, 신고/모더레이션, 광고 심사, 시스템 모니터링 콘솔 |
| **글쓰기 에디터 (Rich Text Editor)** | Tiptap v2 / ProseMirror + TypeScript | 엄격한 JSON Document AST 기반 블록 편집, 슬래시 커맨드, 드래그앤드롭, 오프라인 복구 |
| **백엔드 API (Backend)** | Go (Golang 1.22+) + Chi/Echo + sqlc/pgx | 낮은 메모리 점유율, 높은 동시성 처리, 명확한 도메인 경계를 가진 **모듈형 모놀리스(Modular Monolith)** |
| **백그라운드 워커 (Workers)** | Go Workers + Redis Streams / Asynq + Transactional Outbox | 예약 발행, OpenSearch 색인 동기화, 이미지 변환, 대량 알림, 정산 배치의 비동기 격리 처리 |
| **메인 데이터베이스 (Database)** | PostgreSQL 16+ | 시스템 유일의 원본 데이터 저장소(Single Source of Truth), ACID 트랜잭션, 파티셔닝 적용 |
| **검색 엔진 (Search Engine)** | OpenSearch 2.15+ | 다국어 전문(Full-text) 검색, 자동완성, 패싯 필터, 품질 시그널 결합 부스팅 (원본 DB로 간주하지 않음) |
| **캐시 및 세션 (Cache)** | Redis 7.2+ (Cluster/Sentinel) | 세션 스토어, Rate Limiting, 조회수 어뷰징 방지, 피드 캐시, 분산 락 |
| **미디어 스토리지 & CDN** | S3-compatible Object Storage + Global CDN | 원본(Private Bucket)과 공개 파생본(Public/CDN Bucket) 물리 분리, Signed URL 및 자동 WebP/AVIF 최적화 |
| **API 문서화** | OpenAPI 3.1.0 (`04_API/openapi.yaml`) | 프론트엔드/백엔드/외부 연동 간 엄격한 스키마 계약 및 자동 클라이언트/문서 생성 |
| **컨테이너 & 관측성** | Docker + OpenTelemetry (Traces/Metrics/Logs) | 벤더 중립적 분산 추적(OTLP), Prometheus 메트릭, 구조화 JSON 로깅 |
| **선택적 AI 서비스** | Python 3.11+ (FastAPI Worker) | 스팸/유해 콘텐츠 분류, 언어 감지, 요약/태그 추천 보조 워커 (코어 경로와 비동기 분리) |
| **미래 통합 인터페이스** | REST API + OIDC + Outbox CloudEvents | TreeScan 및 Booking Network 생태계 연동을 위한 느슨한 결합(Loosely Coupled) 어댑터 |

---

## 3. 패키지 전체 구조 및 통계 요약

본 패키지는 제24조의 필수 디렉터리 구조를 100% 준수하며, 총 **163개 파일**로 구성되어 있습니다.

- **루트 제어 문서**: `START_HERE.md`, `README.md`, `MANIFEST.json`, `FILE_CHECKSUMS.sha256` (4개)
- **`00_PROJECT_CHARTER/`**: 프로젝트 헌장 및 핵심 아키텍처 결정 기록 (4개)
- **`01_A_TO_Z_BLUEPRINT/`**: A~Z 26개 분야별 18개 표준 항목 완비 상세 개발 명세서 (26개)
- **`02_ARCHITECTURE/`**: 시스템 아키텍처, 컴포넌트 다이어그램, 배포, 데이터 흐름, 위협 모델 (5개)
- **`03_DATABASE/`**: 58개 테이블 ERD, 데이터 사전, 인덱스 전략, 실행 가능한 SQL 마이그레이션 12개 파일 (15개)
- **`04_API/`**: OpenAPI 3.1.0(`openapi.yaml`) 및 API 계약, 에러 모델, 인가 매트릭스, 테스트 케이스 (5개)
- **`05_UXUI/`**: 45개 화면 인벤토리 및 사용자 웹/블로그 관리/에디터/어드민 상세 화면 설계서 (7개)
- **`06_EDITOR_ENGINE/`**: Tiptap 블록 스키마, 자동저장, 발행 파이프라인, 리비전 시스템 (5개)
- **`07_BLOG_ENGINE/`**: 블로그 개설, 멀티 도메인/TLS, 테마 엔진, 게시글 생명주기, 데이터 이동성 (5개)
- **`08_COMMUNITY/`**: 이웃/팔로우/구독, 댓글/공감, 알림 엔진, 추천 피드 (4개)
- **`09_SEARCH_SEO/`**: OpenSearch 스키마, 랭킹 모델, SEO 아키텍처, Sitemap/RSS, 색인 파이프라인 (5개)
- **`10_MEDIA/`**: 업로드 세션, 이미지 리사이징/EXIF 제거, 비디오 임베드, 스토리지/CDN (4개)
- **`11_SECURITY_MODERATION/`**: 보안 아키텍처, 콘텐츠 모더레이션, 안티스팸, 어뷰징 방지, 글로벌 컴플라이언스 (5개)
- **`12_MONETIZATION/`**: 자체 광고 서버, 크리에이터 수익 배분, 멤버십, 결제 및 복식부기 정산 (4개)
- **`13_WORK_ORDERS/`**: 마스터 작업지시서, CSV 인덱스, 의존성 그래프, 48개 개별 Work Order(`individual/TB-*.md`) (51개)
- **`14_AGENT_EXECUTION/`**: 13개 에이전트 역할 매트릭스, 핸드오프 프로토콜, 실행 순서, 충돌 방지, 보고 템플릿 (5개)
- **`15_TESTING/`**: 90개 마스터 테스트 매트릭스(`MASTER_TEST_MATRIX.csv`), 인수 기준, E2E/SEO/보안/독립 감사 계획 (6개)
- **`16_INFRASTRUCTURE/`**: 배포 계획, CI/CD, 모니터링, 백업/복구, 운영 런북 (5개)
- **`17_TREESCAN_INTEGRATION/`**: 통합 아키텍처, 검색 어댑터, SSO/OIDC 연동, 색인 계약, 독립 운영 요구사항 (5개)
- **`18_FINAL_DELIVERY/`**: Phase 0~10 로드맵, 요구사항 추적 매트릭스, 납품 체크리스트, 한계점, 최종 보고서 (5개)

---

## 4. 후속 개발 에이전트 시작 방법

1. **즉시 [`START_HERE.md`](./START_HERE.md)를 열람하십시오.**
2. `FILE_CHECKSUMS.sha256`으로 파일 무결성을 검증하십시오.
3. `18_FINAL_DELIVERY/PHASE_ROADMAP.md`의 **Phase 0 (`TB-ARCH-001`, `TB-DB-001`, `TB-OPS-001`)**부터 순서대로 착수하십시오.

# START_HERE.md — TreeBlog 마스터 개발 블루프린트 인수 및 실행 프로토콜

> **문서 식별자**: `BN-TREEBLOG-GEMINI4-ARGON-MASTER-BLUEPRINT-001`
> **대상**: 후속 구현 에이전트 (Lead Architect, Frontend/Backend/Editor/DB/Search/Security/QA/DevOps Agents) 및 개발팀
> **명령 등급**: MASTER EXECUTION PROTOCOL (최상위 실행 지침)

본 문서는 `BN_TREEBLOG_MASTER_BLUEPRINT_V1_0.zip` 패키지를 인수한 후속 개발 에이전트가 **가장 먼저 읽고 준수해야 하는 최상위 실행 지침**이다. 후속 개발 에이전트는 이 ZIP 패키지를 단순 참고자료로 취급해서는 안 되며, 아래 **Step 1 ~ Step 10 프로토콜**과 **절대 금지 사항(제27조)**을 엄격히 따라야 한다.

---

## 1. 10단계 표준 실행 프로토콜 (Step 1 ~ Step 10)

### Step 1 — 패키지 인수 및 무결성 검증
1. 전달받은 `BN_TREEBLOG_MASTER_BLUEPRINT_V1_0.zip` 파일의 SHA-256 해시를 검증한다.
2. 압축 해제 후 루트 디렉터리의 `MANIFEST.json`과 `FILE_CHECKSUMS.sha256`을 대조하여 모든 파일이 누락·손상 없이 존재하는지 확인한다.
   ```bash
   sha256sum -c FILE_CHECKSUMS.sha256
   ```
3. 필수 디렉터리(`00_PROJECT_CHARTER` ~ `18_FINAL_DELIVERY`, 총 19개 하위 디렉터리 및 루트 문서)가 모두 존재하는지 확인한다.

### Step 2 — 명세 검토 (순차 독해 의무)
코드를 한 줄이라도 작성하기 전에 다음 순서대로 문서를 정독하여 전체 아키텍처와 도메인 계약을 숙지한다.
1. `README.md` 및 `00_PROJECT_CHARTER/PROJECT_OVERVIEW.md`, `DEVELOPMENT_PRINCIPLES.md`, `SCOPE_AND_NON_GOALS.md`, `ASSUMPTIONS_AND_DECISIONS.md`
2. `01_A_TO_Z_BLUEPRINT/A_ARCHITECTURE.md` ~ `Z_ZERO_TRUST_TREESCAN.md` (26개 전 분야 마스터 명세)
3. `02_ARCHITECTURE/*` (시스템 구조, 컴포넌트 다이어그램, 배포 구조, 데이터 흐름, 위협 모델)
4. `03_DATABASE/*` (ERD, 테이블 사전 58개 테이블, 인덱스 전략, `migrations/` SQL DDL)
5. `04_API/*` (`openapi.yaml` 24개 태그 그룹 122개 엔드포인트, 에러 모델, 인가 매트릭스)
6. `05_UXUI/*` ~ `12_MONETIZATION/*` (도메인별 심화 설계서)
7. `13_WORK_ORDERS/MASTER_WORK_ORDER.md`, `WORK_ORDER_INDEX.csv`, `DEPENDENCY_GRAPH.md` 및 `individual/TB-*.md` (총 48개 Work Order)
8. `14_AGENT_EXECUTION/*` (에이전트 핸드오프, 역할 매트릭스, 충돌 방지 규칙)
9. `15_TESTING/MASTER_TEST_MATRIX.csv` 및 테스트 계획서

### Step 3 — 코드베이스 실사 (Codebase Audit)
1. **기존 저장소가 존재하는 경우**: 현재 디렉터리 트리, Go 모듈(`go.mod`), Nuxt/Vue 패키지(`package.json`), DB 마이그레이션 상태, Docker Compose 설정을 전수 조사한다.
2. **신규 빈 저장소인 경우**: `TB-ARCH-001` (모노레포 디렉터리 구조 및 Docker 개발 환경 구축) 명세에 맞춰 초기 디렉터리(`backend/`, `apps/user-web/`, `apps/blog-console/`, `apps/admin-console/`, `packages/editor-core/`, `packages/ui-kit/`, `infra/`, `migrations/`)를 생성한다.

### Step 4 — Gap Analysis (현 상태 분류)
`18_FINAL_DELIVERY/REQUIREMENTS_TRACEABILITY.csv` 및 `13_WORK_ORDERS/WORK_ORDER_INDEX.csv`의 모든 항목에 대해 현재 코드베이스 상태를 다음 6가지 상태값 중 하나로 정확히 분류하여 초기 감사 리포트를 생성한다.
- `NOT_STARTED`: 코드 및 스키마가 전혀 작성되지 않은 상태
- `PARTIAL`: 스키마, API, UI 중 일부만 작성되었거나 핵심 비즈니스 로직/예외 처리가 누락된 상태
- `IMPLEMENTED_UNVERIFIED`: 코드는 작성되었으나 실제 PostgreSQL/OpenSearch/Redis 연동 테스트 및 E2E 검증이 수행되지 않은 상태
- `VERIFIED`: 단위·통합·E2E 테스트를 통과하고 실제 증빙(로그, 스크린샷, DB 쿼리 결과)이 확보된 상태
- `FAILED`: 테스트 실행 결과 오류가 발생하거나 명세(API 계약, DB 제약조건, 보안 정책)를 위반한 상태
- `BLOCKED`: 외부 승인(PG사 실계약, TreeScan 운영 키 발급 등) 또는 선행 Work Order 미완료로 인해 대기 중인 상태

### Step 5 — 실행 계획 수립 (Execution Plan)
1. `13_WORK_ORDERS/DEPENDENCY_GRAPH.md`와 `14_AGENT_EXECUTION/EXECUTION_SEQUENCE.md`를 분석한다.
2. 선행 작업(Prerequisites)이 `VERIFIED` 상태인 Work Order만 착수한다.
3. 멀티 에이전트 병렬 작업 시 `14_AGENT_EXECUTION/CONFLICT_PREVENTION.md`에 명시된 **파일 소유권(Directory Ownership)**을 엄격히 분리하여 Git Merge 충돌을 원천 차단한다.

### Step 6 — 코드 구현 (Work Order 기반 개발)
1. 각 에이전트는 할당된 `13_WORK_ORDERS/individual/TB-XXX-NNN.md`에 정의된 **수정 가능 디렉터리(Allowed Directories)** 내에서만 코드를 생성·수정한다.
2. DB 스키마(`03_DATABASE/migrations/`) 또는 API 계약(`04_API/openapi.yaml`) 변경이 불가피할 경우, 단독으로 수정하지 않고 **Lead Architect Agent**에게 변경 요청(RFC)을 제출하여 승인받은 후 영향받는 Work Order를 동시에 갱신한다.
3. 지정 기술 스택(Go 1.22+, PostgreSQL 16+, OpenSearch 2.x, Redis 7.x, Vue 3 + Nuxt 3 + TypeScript, Tiptap/ProseMirror, S3 호환 스토리지, OpenTelemetry)을 엄격히 준수한다.

### Step 7 — 단위·통합 테스트 (Unit & Integration Testing)
1. **Mock 테스트와 실제 연동 테스트의 엄격한 분리**:
   - 단위 테스트(Unit Test)에서는 순수 도메인 로직만 검증한다.
   - 통합 테스트(Integration Test)는 반드시 Docker Compose(`postgres`, `opensearch`, `redis`, `minio`)를 띄운 상태에서 **실제 DB 트랜잭션, 제약조건, OpenSearch 색인, S3 Presigned URL 업로드**를 검증해야 한다.
2. Mock 객체를 사용한 테스트만 통과한 상태에서 통합 테스트 완료로 보고하는 것을 엄격히 금지한다.

### Step 8 — 브라우저 E2E 검증 (Playwright E2E)
`15_TESTING/E2E_TEST_PLAN.md`에 정의된 시나리오에 따라 실제 헤드리스 브라우저(Playwright)를 구동하여 다음 핵심 사용자 여정을 검증한다.
1. 회원가입(`SCR-MEM-001`) → 이메일 인증 → 로그인(`SCR-MEM-002`)
2. 블로그 개설(`SCR-MEM-005`) → 주소 중복 검사 → 기본 스킨 적용
3. 글쓰기 에디터(`SCR-BLOG-013`) → Tiptap 블록 작성 + 이미지 업로드 + 태그/카테고리 설정 + 임시저장/자동저장 → 공개 발행
4. 블로그 상세(`SCR-BLOG-005`) SSR 렌더링 확인 → Canonical/OG/JSON-LD 메타태그 검증 → 비공개 전환 시 403/404 및 `noindex` 확인
5. OpenSearch 색인 전파 확인 → 플랫폼 통합 검색(`SCR-EXP-007`) 결과 노출 확인
6. 타 계정으로 이웃/구독(`SCR-BLOG-010`), 공감(`SCR-MEM-007`), 댓글/대댓글(`SCR-MEM-008`) 작성 및 실시간 알림(`SCR-MEM-013`) 수신 확인
7. 관리자 콘솔(`SCR-ADM-001`~`008`)에서 신고 접수 → 게시 중단 조치 → 검색 색인 즉시 제거 확인

### Step 9 — 독립 감사 (Independent Audit)
1. 구현을 담당한 에이전트가 스스로 `VERIFIED`를 최종 승인할 수 없다.
2. **Independent Audit Agent**가 `15_TESTING/INDEPENDENT_AUDIT.md` 절차에 따라 코드 정적 분석, 보안 취약점 점검(XSS/SSRF/CSRF/SQLi), DB 무결성, API 응답 스키마 일치성, 테스트 실행 로그의 진위를 독립적으로 재확인한다.

### Step 10 — Git·배포·최종 보고 (Git & Delivery Report)
1. 발주자(PO)가 허가한 Git 저장소와 브랜치 환경에서만 커밋·푸시·배포를 수행한다.
2. `14_AGENT_EXECUTION/REPORTING_TEMPLATE.md` 양식에 맞춰 다음 증빙을 포함한 최종 완료 보고서를 제출한다.
   - Git Commit SHA 및 브랜치명
   - Work Order별 완료 상태 (`VERIFIED` / `BLOCKED` 등)
   - 실제 테스트 실행 표준 출력 로그 (`go test`, `vitest`, `playwright test`)
   - 빌드 및 컨테이너 이미지 산출물 해시
   - 외부 승인 대기 항목 및 미완료 사유

---

## 2. 후속 개발 에이전트 절대 금지 행동 수칙 (제27조)

후속 에이전트는 어떠한 경우에도 다음 행위를 해서는 안 되며, 적발 시 전체 마일스톤이 즉시 반려(`FAILED`)된다.

1. **UI만 만든 뒤 전체 기능 완료 선언 금지**: 정적 퍼블리싱이나 더미 데이터를 바인딩한 화면만 만들고 기능 완료로 보고하지 않는다.
2. **Mock API를 운영 백엔드로 보고 금지**: 프론트엔드 내부 MSW(Mock Service Worker)나 하드코딩된 JSON 응답을 실제 Go 백엔드 연동으로 위장하지 않는다.
3. **실제 DB 미사용 저장 완료 선언 금지**: 인메모리 맵/배열이나 LocalStorage 저장만으로 영속성 구현 완료를 선언하지 않는다. 반드시 PostgreSQL 트랜잭션과 마이그레이션을 거친다.
4. **테스트 없이 완료 보고 금지**: 실행 가능한 자동화 테스트 코드와 실제 실행 결과 로그 없이 `VERIFIED`로 분류하지 않는다.
5. **가짜 브라우저 캡처 및 가짜 성능 수치 조작 금지**: 임의로 합성한 스크린샷이나 측정하지 않은 Lighthouse/k6 수치를 기재하지 않는다.
6. **존재하지 않는 Git 커밋 SHA 날조 금지**: 실제 `git log`로 검증 가능한 커밋 해시만 보고한다.
7. **승인 없는 프로덕션 배포 금지**: 스테이징/프로덕션 배포는 반드시 PO 및 DevOps 승인 게이트를 거친다.
8. **TreeScan 기존 코드 및 DB 무단 수정 금지**: TreeBlog는 100% 독립 운영 가능해야 하며, TreeScan 연동은 오직 `17_TREESCAN_INTEGRATION/`에 정의된 Adapter/Outbox 인터페이스로만 구현한다.
9. **타사 디자인 및 소스코드 무단 복제 금지**: Naver Blog, Tistory, Blogger, WordPress의 UX 장점은 참고하되, 고유 디자인 토큰(`05_UXUI/DESIGN_SYSTEM.md`)과 클린룸 구현을 준수한다.
10. **사용자 데이터 무단 공유 금지**: 플랫폼 간 계정/프로필/콘텐츠 연동 시 반드시 사용자의 명시적 OAuth/OIDC 동의(`user_consents`)를 검증한다.
11. **비공개 게시글의 검색 색인 노출 금지**: `visibility != 'PUBLIC'` 이거나 블로그가 정지/비공개 상태인 글은 OpenSearch 공개 색인, Sitemap, RSS, TreeScan 피드에서 물리적으로 제외하며, 서버 측에서 즉시 `403 Forbidden`/`404 Not Found` 및 `X-Robots-Tag: noindex, nofollow`를 반환한다.
12. **사용자 입력 HTML 무검증 실행 금지**: 에디터 JSON에서 렌더링된 HTML 및 사용자 커스텀 스킨 CSS/HTML은 서버 측 허용목록 Sanitizer(bluemonday 등)와 CSP(Content-Security-Policy)를 통과해야 하며, 스크립트 실행을 원천 차단한다.
13. **검증되지 않은 외부 라이브러리 의존성 무단 추가 금지**: `01_A_TO_Z_BLUEPRINT`에 명시된 허용 오픈소스 외에 라이선스 불명 또는 보안 취약점이 있는 패키지를 임의로 추가하지 않는다.

---

## 3. 빠른 탐색 가이드 (Directory Quick Map)

| 디렉터리 | 핵심 역할 및 포함 문서 |
| :--- | :--- |
| `00_PROJECT_CHARTER/` | 프로젝트 헌장, 개발 원칙, 범위 및 비목표(Non-goals), 아키텍처 의사결정 기록(ADR) |
| `01_A_TO_Z_BLUEPRINT/` | **A~Z 26개 분야 초정밀 개발 명세서** (각 문서당 18개 필수 섹션 완비) |
| `02_ARCHITECTURE/` | 모듈형 모놀리스 + 워커 아키텍처, 컴포넌트 다이어그램, 배포 구조, 데이터 흐름, 위협 모델 |
| `03_DATABASE/` | 58개 핵심 테이블 ERD, 상세 컬럼 사전, 인덱스/파티셔닝 전략, 실행 가능한 SQL 마이그레이션(`0001`~`0010`) |
| `04_API/` | OpenAPI 3.1.0 명세(`openapi.yaml`), 24개 API 그룹 122개 엔드포인트 계약, 에러 코드, RBAC 매트릭스 |
| `05_UXUI/` | 45개 전체 화면 인벤토리, 사용자 웹/블로그 관리/에디터/어드민 UX 명세, 디자인 시스템, 반응형·접근성 |
| `06_EDITOR_ENGINE/` | Tiptap/ProseMirror 블록 스키마, 듀얼 레이어 자동저장, 발행 워크플로, Diff 리비전 엔진 |
| `07_BLOG_ENGINE/` | 블로그 개설, 서브도메인/경로/개인 도메인 라우팅 및 자동 TLS, 테마/스킨 샌드박스, 글 생명주기, 이관(WXR/MD) |
| `08_COMMUNITY/` | 팔로우·이웃·상호이웃·구독 상태 머신, 댓글/공감, 멀티채널 알림, 하이브리드 추천 피드 |
| `09_SEARCH_SEO/` | OpenSearch 한국어/다국어 인덱스 매핑, BM25+품질 랭킹, SSR SEO, Sitemap/RSS/IndexNow |
| `10_MEDIA/` | S3 Presigned 멀티파트 업로드, libvips WebP/AVIF 변환, EXIF 제거, 동영상 임베드, 스토리지 수명주기 |
| `11_SECURITY_MODERATION/` | XSS/CSRF/SSRF 방어, 신고→심사→제재→이의신청 워크플로, 스팸 탐지, 글로벌 법규 준수(GDPR/PIPA/DMCA) |
| `12_MONETIZATION/` | 자체 광고 서버, 블로거 수익 배분(복식부기 원장), 유료 멤버십/단건 구매, 후원 및 정산 |
| `13_WORK_ORDERS/` | **48개 전체 Work Order** 마스터 문서, CSV 인덱스, DAG 의존성 그래프, `individual/TB-*.md` 개별 작업지시서 |
| `14_AGENT_EXECUTION/` | 13개 전문 에이전트 역할 매트릭스, 핸드오프 프로토콜, 병렬 실행 순서, 파일 충돌 방지 규칙 |
| `15_TESTING/` | 90개 마스터 테스트 매트릭스(`MASTER_TEST_MATRIX.csv`), 인수 기준, E2E/SEO/보안 테스트 및 독립 감사 계획 |
| `16_INFRASTRUCTURE/` | Docker/Kubernetes 배포 계획, GitHub Actions CI/CD, OpenTelemetry 모니터링, 백업/복구(RPO/RTO), 운영 런북 |
| `17_TREESCAN_INTEGRATION/` | 독립 운영 보장 원칙, TreeScan 검색 API 어댑터, OIDC 통합 SSO, 색인 동기화 이벤트 계약 |
| `18_FINAL_DELIVERY/` | Phase 0~10 로드맵, 요구사항 추적성 매트릭스(`REQUIREMENTS_TRACEABILITY.csv`), 납품 체크리스트, 한계점 및 최종 보고서 |

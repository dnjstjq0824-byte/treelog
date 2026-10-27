# X_AGENT_EXECUTION — eXecution & Multi-Agent Development (멀티 에이전트 개발 및 작업지시 실행 체계)
> **부제**: 개발 Work Order, 에이전트 역할, 선행 작업, 병렬 개발, 충돌 방지, 보고 체계, 완료 검증
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-X_AGENT_EXECUTION`
> **적용 조항**: 제3조 (X_AGENT_EXECUTION 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- 후속 개발을 수행할 **13개 전문 에이전트(제21조)**가 총 **48개의 원자적 Work Order(제19조)**를 기반으로 파일 충돌이나 계약 불일치 없이 일사불란하게 병렬 개발·통합 테스트·독립 감사를 완수하도록 실행 프로토콜을 정의한다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **Lead Architect Agent**: 전체 Work Order 의존성 DAG를 통제하고 DB 스키마(`03_DATABASE`) 및 OpenAPI 계약(`04_API`)의 변경 요청(RFC)을 승인/기각한다.
- **11개 도메인/계층 구현 에이전트**: 자신에게 할당된 Work Order의 `수정 가능 디렉터리(Allowed Directories)` 내에서만 코드를 구현하고 단위/통합 테스트를 통과시킨다.
- **Independent Audit Agent**: 구현 에이전트와 완전히 독립된 입장에서 산출물과 테스트 로그를 재검증하여 최종 `VERIFIED` 여부를 판정한다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-EXE-001 (제21조 13개 에이전트 역할 및 디렉터리 독점 소유권 분리)**:
  1) `Lead Architect Agent`: `00_PROJECT_CHARTER/`, `02_ARCHITECTURE/`, `04_API/openapi.yaml`, 아키텍처 린트 및 RFC 승인.
  2) `PostgreSQL Agent`: `03_DATABASE/migrations/`, `internal/db/`, DB 쿼리 성능 및 파티셔닝 관리.
  3) `Backend Developer Agent`: `cmd/api/`, `cmd/worker/`, `internal/modules/{identity,user,blog,domain,post,monetization}/`.
  4) `Editor Engine Agent`: `packages/editor-core/`, `internal/modules/editor/`, `apps/blog-console/src/pages/editor/`.
  5) `Frontend Developer Agent`: `apps/user-web/`, `apps/blog-console/`(에디터 제외), `packages/ui-kit/`, `packages/theme-renderer/`, `internal/modules/theme/`.
  6) `Search & SEO Agent`: `internal/modules/{search,seo}/`, OpenSearch 인덱스 매핑 및 Sitemap/RSS.
  7) `Media Processing Agent`: `internal/modules/media/`, 스토리지/이미지 변환 워커.
  8) `Community Agent`: `internal/modules/{community,notification,recommendation}/`.
  9) `Security Agent`: `internal/pkg/security/`, `internal/pkg/safenet/`, `internal/modules/moderation/`, 보안 미들웨어.
  10) `Admin Console Agent`: `apps/admin-console/`, `internal/modules/admin/`.
  11) `DevOps Agent`: `Dockerfile*`, `docker-compose*.yml`, `.github/workflows/`, `infra/`, OpenTelemetry/백업 스크립트.
  12) `QA / E2E Agent`: `tests/e2e/`, `tests/integration/`, `tests/load/`, `15_TESTING/`.
  13) `Independent Audit Agent`: 코드 수정 금지(Read-only). 전수 감사 스크립트 실행 및 최종 승인/반려 보고서 발행.
- **REQ-EXE-002 (제19조 18개 필수 항목을 갖춘 48개 Work Order 전수 정의)**:
  - 모든 Work Order(`13_WORK_ORDERS/individual/TB-*.md`)는 1) ID, 2) 작업명, 3) 목적, 4) 담당 에이전트, 5) 선행 작업, 6) 구현 범위, 7) 수정 가능 디렉터리, 8) 수정 제한 디렉터리, 9) 관련 DB 테이블, 10) 관련 API, 11) 관련 UI, 12) 구현 상세 요구사항, 13) 예상 파일 구조, 14) 테스트 시나리오, 15) Acceptance Criteria, 16) 예상 위험, 17) 최종 증빙, 18) 완료 선언 조건을 빠짐없이 포함한다.

## 4. 데이터 구조 (Data Structures & Domain Models)
```json
{
  "workOrderId": "TB-EDITOR-001",
  "title": "Tiptap/ProseMirror 블록 스키마 및 서버 HTML Sanitizer 코어 구현",
  "assignedAgent": "Editor Engine Agent",
  "phase": 3,
  "prerequisites": ["TB-ARCH-001", "TB-DB-001", "TB-BLOG-001"],
  "allowedDirectories": ["packages/editor-core/", "internal/modules/editor/"],
  "restrictedDirectories": ["03_DATABASE/migrations/", "apps/admin-console/"],
  "status": "READY_FOR_IMPLEMENTATION"
}
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- 프론트엔드 작업 시 `packages/api-client`는 `04_API/openapi.yaml`로부터 자동 생성하므로 수동 편집을 금지하여 에이전트 간 충돌을 방지한다.

## 6. 백엔드 로직 (Backend Logic & Processing)
- 각 백엔드 도메인 모듈은 독립된 Go 패키지(`internal/modules/{name}`)로 격리되어 있어 서로 다른 에이전트가 동시에 서로 다른 모듈을 개발하더라도 Git Merge Conflict가 발생하지 않는다.

## 7. API 계약 (API Contracts)
- API 계약 변경 시 반드시 `Lead Architect Agent`를 통해 `04_API/openapi.yaml`을 먼저 갱신하고, 영향받는 Work Order들을 동기화한다.

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- 마이그레이션 번호(`0001`~`0010`)는 `PostgreSQL Agent`가 중앙 통제하여 번호 충돌을 원천 차단한다.

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **블로커(Blocker) 발생 시 에스컬레이션 프로토콜**: 특정 Work Order 수행 중 선행 API 결함이나 외부 의존성 블로커가 발견되면 임의로 우회/하드코딩하지 않고 상태를 `BLOCKED`로 전환한 뒤 `14_AGENT_EXECUTION/REPORTING_TEMPLATE.md` 양식으로 즉시 보고한다.

## 10. UX/UI 구조 (UX/UI Structure)
- N/A (개발 실행 및 멀티 에이전트 오케스트레이션 명세).

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- 에이전트 작업 환경에는 프로덕션 DB 접근 권한이나 배포 시크릿을 부여하지 않으며, 격리된 로컬/CI 샌드박스에서만 코드를 실행·검증한다.

## 12. 성능 및 확장성 (Performance & Scalability)
- Phase별 병렬 실행 그래프(`13_WORK_ORDERS/DEPENDENCY_GRAPH.md`)를 통해 최대 6개 에이전트가 파일 간섭 없이 동시 개발을 수행할 수 있다.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- Git, Docker Compose, GNU Make / Taskfile, OpenAPI Generator.

## 14. 개발 우선순위 (Development Priority)
- **Phase 0 착수 전 전제 규약** (`13_WORK_ORDERS/*`, `14_AGENT_EXECUTION/*`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- `START_HERE.md` 및 `18_FINAL_DELIVERY/PHASE_ROADMAP.md`와 직결.

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-EXE-001`: 48개 전체 Work Order의 선행 작업(Prerequisites) 그래프에 순환 의존성(Cycle)이나 존재하지 않는 Work Order 참조가 0건인지 DAG 위상 정렬(Topological Sort) 검증.
- `TEST-EXE-002`: 모든 Work Order 문서에 18개 필수 섹션이 누락 없이 존재하는지 자동 파서 검증.

## 17. Acceptance Criteria (인수 기준)
1. 48개 개별 Work Order와 13개 에이전트 역할 매트릭스, 충돌 방지 규칙, 보고 양식이 100% 일관되게 작성되어야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- `13_WORK_ORDERS/WORK_ORDER_INDEX.csv`, `individual/TB-*.md` 48개 파일, DAG 무결성 검증 스크립트 출력.

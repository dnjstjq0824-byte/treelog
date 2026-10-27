# Q_QUALITY_ASSURANCE — Quality Assurance (품질 보증 및 종합 테스트 전략)
> **부제**: 단위 테스트, API 테스트, E2E, UI/UX, SEO, 접근성, 성능, 보안, 회귀 및 운영 테스트
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-Q_QUALITY_ASSURANCE`
> **적용 조항**: 제3조 (Q_QUALITY_ASSURANCE 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- 단순 문서 확인이나 Mock 테스트 통과만으로 구현 완료를 선언하는 것을 원천 차단하고, 단위·통합·E2E·SEO·접근성·보안·성능 전 영역에 걸쳐 **재현 가능하고 독립 감사 가능한 검증 체계(제23조, 제25조)**를 확립한다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **QA / E2E Agent**: 각 Work Order 구현 완료 시 `15_TESTING/MASTER_TEST_MATRIX.csv`의 90개 테스트 케이스를 실제 컨테이너 환경에서 실행하고 증빙 아티팩트(로그, 트레이스, 스크린샷)를 수집한다.
- **Independent Audit Agent**: 개발 에이전트와 분리된 독립 검증 환경에서 테스트 재현성과 보안/데이터 무결성을 최종 감사한다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-QA-001 (테스트 피라미드 및 Mock 분리 원칙)**:
  1) **Unit Tests (`go test ./...`, `vitest`)**: 도메인 상태 머신, 권한 계산(`CanReadPost`), 마크다운/AST 파서, 복식부기 검증 등 순수 로직 검증 (커버리지 목표 ≥ 85%).
  2) **Real Integration Tests**: Testcontainers 또는 Docker Compose(`postgres`, `redis`, `opensearch`, `minio`)를 실제로 띄운 상태에서 DB 제약조건, 트랜잭션 롤백, Outbox 색인 동기화, S3 업로드 검증. **Mock DB로 대체하는 것을 엄격히 금지한다**.
  3) **Browser E2E Tests (`Playwright`)**: Desktop Chrome, Mobile Safari(WebKit), Firefox 뷰포트에서 회원가입 → 블로그 개설 → 에디터 글쓰기/이미지 업로드 → 발행 → 검색 → 이웃/댓글/공감 → 관리자 제재의 전체 사용자 흐름 검증.
- **REQ-QA-002 (전문 영역별 자동화 검증)**:
  - **SEO 검증**: SSR 응답 HTML의 `<title>`, `<link rel="canonical">`, `og:*`, JSON-LD 스키마 유효성, `sitemap.xml`, `rss.xml`, 비공개 글 `X-Robots-Tag: noindex` 자동 검사.
  - **접근성(a11y) 검증**: `@axe-core/playwright`를 통해 WCAG 2.1 AA 중대 위반(Critical/Serious) 0건 검증.
  - **보안(DAST/SAST) 검증**: XSS 페이로드 퍼징, SSRF 로컬호스트 접근 차단, IDOR 타인 블로그 수정 차단, CSRF 토큰 누락 거부 자동 테스트.
  - **성능/부하 검증 (`k6`)**: 공개 게시글 조회, 검색, 동시 공감 클릭 시나리오 부하 테스트.

## 4. 데이터 구조 (Data Structures & Domain Models)
```csv
test_id,category,work_order_id,target_requirement,title,preconditions,steps,expected_result,evidence_path,execution_status
TEST-USR-001,User,TB-AUTH-001,REQ-IAM-001,이메일 회원가입 및 Argon2id 해시 저장 검증,Clean DB,POST /api/v1/auth/register 호출 후 users 테이블 조회,201 Created 반환 및 password_hash가 $argon2id$로 시작,artifacts/tests/TEST-USR-001.log,NOT_EXECUTED_IN_BLUEPRINT_PHASE
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- 모든 인터랙티브 UI 요소에 안정적인 E2E 셀렉터(`data-testid="..."`)와 접근성 역할(`role`, `aria-label`)을 부여하여 CSS 클래스 변경으로 인해 테스트가 깨지지 않도록 설계한다.

## 6. 백엔드 로직 (Backend Logic & Processing)
- `internal/testutil`: 테스트마다 격리된 PostgreSQL 스키마/트랜잭션 및 임시 OpenSearch 인덱스를 프로비저닝하는 픽스처 헬퍼 제공.

## 7. API 계약 (API Contracts)
- `04_API/API_TEST_CASES.md`에 24개 API 그룹 전체의 정상(2xx), 유효성 오류(400/422), 미인증(401), 권한 부족(403), 리소스 없음(404), 충돌(409), 속도 제한(429) 테스트 케이스를 정의한다.

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- 테스트 실행 후 DB 고아 레코드나 복식부기 불균형(`SUM(debit) != SUM(credit)`)이 없는지 검증하는 사후 불변식 체커(`AssertDatabaseInvariants`)를 모든 통합 테스트 Suite 종료 시 자동 실행한다.

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **실행하지 않은 테스트의 정직한 상태 표기 (제23조, 제25조 준수)**:
  - 현재 블루프린트 작성 단계에서는 명세 정합성 검증(JSON/YAML/CSV/SQL/Traceability 검증)은 즉시 실행하여 `VERIFIED`로 기록하고, 후속 개발 에이전트가 실제 애플리케이션 코드를 빌드한 뒤 실행해야 하는 런타임 E2E/부하 테스트는 `NOT_EXECUTED_YET (AWAITING_CODE_IMPLEMENTATION)`으로 정직하게 구분하여 표기한다. 실행하지 않은 런타임 테스트를 허위로 `PASS` 처리하는 것을 절대 금지한다.

## 10. UX/UI 구조 (UX/UI Structure)
- 시각적 회귀 테스트(Playwright Visual Comparisons)를 통해 테마 변경이나 반응형 브레이크포인트(320px, 768px, 1440px)에서 UI 깨짐을 자동 탐지한다.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- 테스트 데이터에 실제 운영 개인정보나 실제 외부 결제 카드를 절대 사용하지 않으며, 합성 시드 데이터(Synthetic Seed Data)와 PG Sandbox 키만 사용한다.

## 12. 성능 및 확장성 (Performance & Scalability)
- CI 파이프라인에서 전체 단위+통합+E2E 테스트가 병렬 샤딩을 통해 12분 이내에 완료되도록 구성한다.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- `@playwright/test`, `@axe-core/playwright`, `vitest`, `github.com/stretchr/testify`, `k6`.

## 14. 개발 우선순위 (Development Priority)
- **전 Phase 상시 병행** 및 **Phase 9 최종 상용화 종합 감사** (`TB-QA-001` ~ `TB-QA-003`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 모든 구현 Work Order(`TB-*`)의 완료 게이트 역할 수행.

## 16. 테스트 시나리오 (Test Scenarios)
- `15_TESTING/MASTER_TEST_MATRIX.csv`에 정의된 90개 전체 테스트 시나리오(`TEST-USR-001` ~ `TEST-ZTR-004`).

## 17. Acceptance Criteria (인수 기준)
1. 모든 핵심 요구사항이 최소 1개 이상의 고유 테스트 ID와 연결되어야 하며, 독립 감사 에이전트가 동일 명령으로 테스트를 재현할 수 있어야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- `15_TESTING/*` 테스트 계획서 6종 및 블루프린트 무결성 검증 스크립트 실행 결과.

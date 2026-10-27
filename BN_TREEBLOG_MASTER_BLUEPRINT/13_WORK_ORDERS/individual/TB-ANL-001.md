# TB-ANL-001 — 프라이버시 친화적 방문자·체류시간·완독률 비콘 수집기 및 유입 경로/크롤러 분석 엔진 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 8`

---

## 1. Work Order ID
- `TB-ANL-001`

## 2. 작업명 (Task Title)
- **프라이버시 친화적 방문자·체류시간·완독률 비콘 수집기 및 유입 경로/크롤러 분석 엔진 구현**

## 3. 목적 (Purpose)
- 1st-Party 비식별 열람 비콘을 수집하여 실시간/일별 UV·PV, 체류시간, 완독률, 유입 검색어(TreeScan/Google/Naver 등), 검색봇 크롤링 통계를 집계하고 차트 UI를 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Backend Developer Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-POST-001, TB-OPS-001`

## 6. 구현 범위 (Implementation Scope)
- 비콘 수집기(Redis HyperLogLog + 배치 롤업), 봇/작성자 본인 트래픽 분리 필터, 블로그 통계 대시보드 UI

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/analytics/, apps/blog-console/src/pages/studio/[blogId]/analytics/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/admin-console/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `posts, search_click_events` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-ANL-001, EP-ANL-002, EP-ANL-003, EP-ANL-004, EP-ANL-005` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-MEM-006` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) IP 원본 미저장(일일 회전 솔트 해시 사용). 2) 작성자 본인 조회 및 봇 트래픽은 일반 독자 UV에서 분리.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/analytics/collector.go, rollup_worker.go, apps/blog-console/src/pages/studio/[blogId]/analytics/index.vue`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-ANL-001, TEST-ANL-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 봇 User-Agent와 본인 조회가 독자 UV에서 제외되고 유입 도메인/완독률 통계가 정확히 조회되어야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 대량 비콘 트래픽 DB 부하 -> Redis 메모리 집계 후 주기적 배치 플러시

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-ANL-001/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

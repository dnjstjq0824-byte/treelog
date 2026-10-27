# TB-ADS-001 — 자체 광고 네트워크(캠페인/소재/타기팅 송출/노출·클릭 비콘/IVT 부정클릭 방어) 및 수익 배분 엔진 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 8`

---

## 1. Work Order ID
- `TB-ADS-001`

## 2. 작업명 (Task Title)
- **자체 광고 네트워크(캠페인/소재/타기팅 송출/노출·클릭 비콘/IVT 부정클릭 방어) 및 수익 배분 엔진 구현**

## 3. 목적 (Purpose)
- 제12.1조 및 제12.2조에 따른 광고주 캠페인 관리, 문맥 타기팅 광고 송출, Brand Safety 필터, 부정 클릭(IVT) 차단 및 블로거 광고 수익 일일 배분 정산을 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Backend Developer Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-ANL-001, TB-DB-002`

## 6. 구현 범위 (Implementation Scope)
- 광고 캠페인 CRUD, 문맥 매칭 엔진, 노출/클릭 비콘 및 IVT 필터, 블로거 수익 배분(기본 70:30) 일일 원장 기장 워커

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/monetization/ads.go, ivt_filter.go`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/user-web/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `ad_accounts, ad_campaigns, ad_placements, ad_events, ledger_entries` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-ADS-001, EP-ADS-002, EP-ADS-003, EP-ADS-004, EP-ADS-005` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-BLOG-005, SCR-MEM-006, SCR-ADM-007` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 비공개/임시조치/성인 글에는 광고 미송출. 2) 모든 광고 슬롯에 명확한 '광고(Ad)' 라벨 표시.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/monetization/ad_server.go, internal/modules/monetization/ad_settlement.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-MON-003` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 승인된 광고만 문맥에 맞게 송출되고 자가 클릭/반복 클릭은 과금 및 수익 배분에서 제외되어야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 광고 로딩으로 인한 본문 CLS 발생 -> 고정 높이 컨테이너 예약으로 해결

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-ADS-001/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

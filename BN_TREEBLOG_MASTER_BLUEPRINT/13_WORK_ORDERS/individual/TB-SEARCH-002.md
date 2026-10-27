# TB-SEARCH-002 — 콘텐츠 품질 시그널(SimHash 중복 탐지) 산출 및 트렌딩/개인화 추천 랭킹 엔진 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 5`

---

## 1. Work Order ID
- `TB-SEARCH-002`

## 2. 작업명 (Task Title)
- **콘텐츠 품질 시그널(SimHash 중복 탐지) 산출 및 트렌딩/개인화 추천 랭킹 엔진 구현**

## 3. 목적 (Purpose)
- 64비트 SimHash 기반 중복/도용 탐지, 품질 점수 계산, 시간 감쇠(Time-decay) 인기 차트 및 다양성 제한(Diversity Cap) 추천 엔진을 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Search & SEO Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-SEARCH-001`

## 6. 구현 범위 (Implementation Scope)
- SimHash 계산기, 품질 점수 산출 워커, Redis ZSET 트렌딩 랭커, OpenSearch more_like_this 관련글 추천

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/search/quality.go, internal/modules/recommendation/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `content_quality_signals, posts, reactions, bookmarks` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-REC-001, EP-REC-002, EP-REC-003, EP-REC-004, EP-REC-005` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-EXP-002, SCR-EXP-003, SCR-EXP-009, SCR-BLOG-005` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 인기 Top 20 내 동일 블로그 글 최대 2개 제한(MMR 다양성). 2) 새로고침 어뷰징 조회수 제외.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/search/quality.go, internal/modules/recommendation/ranker.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-QLT-001, TEST-REC-001, TEST-REC-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 중복 복제 글 감점 처리 및 인기글 다양성 캡, 관련글 추천이 정상 동작해야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 신규 블로그 노출 소외 -> 신규 고품질 글 Freshness Boost 적용

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-SEARCH-002/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

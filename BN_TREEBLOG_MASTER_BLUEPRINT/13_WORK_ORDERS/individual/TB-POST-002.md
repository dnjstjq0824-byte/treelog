# TB-POST-002 — 분 단위 정밀 예약 발행 스케줄러(scheduled_publications) 및 팀 블로그 원고 승인 워크플로 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 3`

---

## 1. Work Order ID
- `TB-POST-002`

## 2. 작업명 (Task Title)
- **분 단위 정밀 예약 발행 스케줄러(scheduled_publications) 및 팀 블로그 원고 승인 워크플로 구현**

## 3. 목적 (Purpose)
- 사용자 타임존을 고려한 예약 발행 등록·취소·변경 및 5초 주기 무중복(SKIP LOCKED) 자동 발행 워커와 기업 블로그 검토 승인(IN_REVIEW) 흐름을 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Backend Developer Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-POST-001, TB-ARCH-002`

## 6. 구현 범위 (Implementation Scope)
- 예약 등록/취소 API, 5초 주기 ScheduledPostPoller, 예약 시점 블로그 정지 상태 재확인 가드

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/post/scheduler.go`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `scheduled_publications, posts, outbox_events` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-PUB-001` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-BLOG-013, SCR-MEM-006` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 예약 대기 중인 글은 예약 시각 도래 전까지 외부에서 절대 조회/색인 불가. 2) 다중 워커 환경에서 정확히 1회만 발행 보장.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/post/scheduler.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-JOB-001, TEST-JOB-002, TEST-GLB-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 3대 워커 동시 폴링 시 중복 발행 0건 및 지정 시각 5초 이내 정확히 PUBLISHED로 전환되어야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- DST(일광절약시간) 오차 -> IANA 타임존 변환 후 UTC 저장으로 해결

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-POST-002/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

# TB-SEC-002 — 자동 스팸 탐지 엔진, 신고 접수·병합, 임시조치(QUARANTINE) 및 이의신청(Appeals) 백엔드 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 7`

---

## 1. Work Order ID
- `TB-SEC-002`

## 2. 작업명 (Task Title)
- **자동 스팸 탐지 엔진, 신고 접수·병합, 임시조치(QUARANTINE) 및 이의신청(Appeals) 백엔드 구현**

## 3. 목적 (Purpose)
- 제16조의 콘텐츠 신고 → 케이스 병합 → 모더레이터 심사 → 임시 게시중단/복원 → 작성자 이의신청 및 재심사 워크플로와 자동 스팸 탐지기를 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Security Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-POST-001, TB-COMMUNITY-002`

## 6. 구현 범위 (Implementation Scope)
- 신고 접수 및 중복 병합, 스팸 휴리스틱/빈도 스코어러, QUARANTINE 전환(색인 즉시 삭제 및 451 응답), 이의신청(Appeal) 처리

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/moderation/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `moderation_cases, user_reports, moderation_actions, appeals, spam_detection_results, audit_logs` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-MOD-001, EP-MOD-002, EP-ADM-003, EP-ADM-004, EP-ADM-005` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-BLOG-012, SCR-ADM-004, SCR-ADM-005` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 자동 스팸 격리 시에도 작성자에게 사유 통지 및 원클릭 이의신청 경로 100% 보장. 2) 제재 중 작성자 임의 증거 인멸 삭제 잠금.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/moderation/case_service.go, internal/modules/moderation/spam_detector.go, internal/modules/moderation/appeal_service.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-QLT-002, TEST-ADM-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 신고 -> QUARANTINE(공개 451 및 검색 제외) -> 이의신청 -> 승인 복원(공개 200 및 재색인) 전 과정이 동작해야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 악의적 대량 허위 신고(Mass Reporting)로 정상 글 자동 차단 -> 자동 격리는 스팸 시그널 결합 시에만 수행하고 일반 권리침해 신고는 인적 검토 큐로 라우팅

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-SEC-002/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

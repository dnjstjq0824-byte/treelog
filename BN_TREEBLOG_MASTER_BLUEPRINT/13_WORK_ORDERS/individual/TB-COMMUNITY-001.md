# TB-COMMUNITY-001 — 팔로우·이웃·상호이웃(서로이웃) 상태 머신, 이웃 그룹, 구독 및 사용자 차단 API 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 6`

---

## 1. Work Order ID
- `TB-COMMUNITY-001`

## 2. 작업명 (Task Title)
- **팔로우·이웃·상호이웃(서로이웃) 상태 머신, 이웃 그룹, 구독 및 사용자 차단 API 구현**

## 3. 목적 (Purpose)
- 팔로우, 단방향 이웃, 양방향 상호이웃 신청·수락·거절 상태 머신과 이웃 폴더 관리, 알림/뉴스레터 구독, 악성 사용자 차단을 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Community Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-BLOG-001, TB-POST-001`

## 6. 구현 범위 (Implementation Scope)
- 관계 상태 전이 트랜잭션, 게시글 열람 권한(NEIGHBOR_ONLY / MUTUAL_ONLY) 실시간 연동, 차단 시 기존 관계 즉시 해지

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/community/follow.go, block.go`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `neighbor_groups, follows, subscriptions, user_blocks` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-FOL-001, EP-FOL-002, EP-FOL-003, EP-FOL-004, EP-FOL-005, EP-MOD-003, EP-MOD-004` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-BLOG-010, SCR-MEM-011, SCR-MEM-012, SCR-MEM-014` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 상호이웃 수락 시 양방향 권한 부여 및 신청자 알림 발행. 2) 차단(user_blocks) 등록 즉시 댓글/공감/상호이웃 신청 및 비공개 글 접근 차단.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/community/follow_service.go, internal/modules/community/block_service.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-COM-001, TEST-COM-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 상호이웃 수락 전에는 MUTUAL_ONLY 글이 403으로 차단되고 수락 즉시 200 OK로 열람되며 차단 시 다시 403으로 회수되어야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 자기 자신 블로그 상호이웃 신청 -> 서비스 레이어 검증으로 차단

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-COMMUNITY-001/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

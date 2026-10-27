# TB-COMMUNITY-003 — 멀티채널 알림 엔진(인앱 SSE, VAPID 웹 푸시, 이메일), Rollup 병합 및 커뮤니티 UI 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 6`

---

## 1. Work Order ID
- `TB-COMMUNITY-003`

## 2. 작업명 (Task Title)
- **멀티채널 알림 엔진(인앱 SSE, VAPID 웹 푸시, 이메일), Rollup 병합 및 커뮤니티 UI 구현**

## 3. 목적 (Purpose)
- 댓글·공감·이웃·구독 새 글 이벤트를 인앱 SSE, 웹 푸시, 이메일로 발송하고 중복 방지(dedup_key) 및 스마트 병합(group_key)과 회원 커뮤니티 관리 화면을 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Community Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-COMMUNITY-001, TB-COMMUNITY-002, TB-WEB-003`

## 6. 구현 범위 (Implementation Scope)
- Outbox 알림 디스패처, 공감 알림 Rollup 병합기, 방해금지 시간(Quiet Hours) 스케줄링, 커뮤니티 UI 화면 8종

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/notification/, apps/user-web/src/components/community/, apps/blog-console/src/pages/activity/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/admin-console/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `notifications, notification_preferences` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-NOT-001, EP-NOT-002, EP-NOT-003, EP-NOT-004, EP-NOT-005` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-MEM-007, SCR-MEM-008, SCR-MEM-009, SCR-MEM-010, SCR-MEM-011, SCR-MEM-012, SCR-MEM-013, SCR-MEM-014` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 동일 글 연속 공감은 단일 알림 레코드로 병합(actor_count 증가). 2) RFC 8058 원클릭 이메일 수신거부 서명 링크 제공.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/notification/dispatcher.go, internal/modules/notification/sse.go, apps/user-web/src/pages/me/notifications.vue`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-NOT-001, TEST-NOT-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 10명 연속 공감 시 단일 롤업 알림으로 병합되고 실시간 SSE 스트림 및 일괄 읽음 처리가 동작해야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 대형 블로거 새 글 발행 시 알림 지연 -> 1,000명 단위 청크 워커 팬아웃 처리

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-COMMUNITY-003/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

# TB-POST-001 — 게시글 즉시 발행, 수정, 공개범위(6단계) 제어, 휴지통(30일 보관)·복원·영구삭제 엔진 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 3`

---

## 1. Work Order ID
- `TB-POST-001`

## 2. 작업명 (Task Title)
- **게시글 즉시 발행, 수정, 공개범위(6단계) 제어, 휴지통(30일 보관)·복원·영구삭제 엔진 구현**

## 3. 목적 (Purpose)
- 게시글 발행 트랜잭션, 6단계 가시성(PUBLIC/UNLISTED/NEIGHBOR_ONLY/MUTUAL_ONLY/PAID_MEMBERS_ONLY/PRIVATE) 서버 접근 제어, 슬러그 변경 301 이력 및 휴지통 생명주기를 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Backend Developer Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-EDITOR-001, TB-BLOG-002`

## 6. 구현 범위 (Implementation Scope)
- 발행 파이프라인, CanReadPost 권한 평가기, 휴지통 소프트 삭제/복원 및 30일 경과 자동 영구 파기 워커

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/post/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `posts, post_revisions, post_categories, post_tags, tags, post_media, post_slug_history, outbox_events` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-PUB-001, EP-PUB-002, EP-PUB-003, EP-PUB-004, EP-PUB-005, EP-PUB-006, EP-PUB-007` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-BLOG-003, SCR-BLOG-005` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 비공개 전환 또는 휴지통 이동 즉시 priority=100으로 검색 색인 삭제 Outbox 이벤트 발행 및 캐시 퍼지. 2) 비인가 접근 시 본문 필드를 응답에서 완전 제거.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/post/publish_service.go, internal/modules/post/access_control.go, internal/modules/post/lifecycle.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-PUB-001, TEST-LFC-001, TEST-LFC-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 발행/수정/비공개전환/휴지통이동/복원이 DB, 캐시, 검색 색인에 원자적으로 반영되어야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 슬러그 변경 시 기존 외부 링크 깨짐 -> post_slug_history 조회 후 301 리다이렉트로 완벽 방지

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-POST-001/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

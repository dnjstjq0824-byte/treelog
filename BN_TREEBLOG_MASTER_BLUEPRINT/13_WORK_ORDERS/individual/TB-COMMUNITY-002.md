# TB-COMMUNITY-002 — 2뎁스 댓글·대댓글, 비밀댓글 서버 마스킹, 승인제 댓글, 6종 공감 리액션 및 북마크 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 6`

---

## 1. Work Order ID
- `TB-COMMUNITY-002`

## 2. 작업명 (Task Title)
- **2뎁스 댓글·대댓글, 비밀댓글 서버 마스킹, 승인제 댓글, 6종 공감 리액션 및 북마크 구현**

## 3. 목적 (Purpose)
- 계층형 댓글/대댓글, 비밀댓글 보호, 댓글 승인제/고정, 게시글 6종 공감 리액션(LIKE/LOVE/HELPFUL/FUNNY/TOUCHING/INSIGHTFUL), 북마크 보관함을 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Community Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-COMMUNITY-001`

## 6. 구현 범위 (Implementation Scope)
- 댓글 CRUD 및 권한별 마스킹, 동시성 안전 공감 토글(UNIQUE 제약), 북마크 폴더 및 읽기 진행률 저장

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/community/comment.go, reaction.go, bookmark.go`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `comments, reactions, bookmarks, posts` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-CMT-001, EP-CMT-002, EP-CMT-003, EP-CMT-004, EP-RCT-001, EP-RCT-002, EP-RCT-003, EP-BMK-001, EP-BMK-002, EP-BMK-003` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-BLOG-005, SCR-MEM-007, SCR-MEM-008, SCR-MEM-009, SCR-MEM-010` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 비밀 댓글은 작성자·원댓글작성자·블로그주인 외 제3자에게 서버에서 원문을 절대 전송하지 않음. 2) 대댓글이 달린 원댓글 삭제 시 소프트 삭제 마스킹 처리.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/community/comment_service.go, internal/modules/community/reaction_service.go, internal/modules/community/bookmark_service.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-COM-003, TEST-COM-004, TEST-COM-005` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 비밀 댓글 제3자 원문 미노출 및 동시 공감 클릭 시 중복 카운트 0건이 검증되어야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 댓글 폭주로 인한 DB 부하 -> 분당 Rate Limit 및 카운터 원자적 업데이트 적용

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-COMMUNITY-002/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

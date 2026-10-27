# TB-EDITOR-001 — Tiptap/ProseMirror 블록 스키마(packages/editor-core) 및 서버 AST 검증·HTML Sanitizer 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 3`

---

## 1. Work Order ID
- `TB-EDITOR-001`

## 2. 작업명 (Task Title)
- **Tiptap/ProseMirror 블록 스키마(packages/editor-core) 및 서버 AST 검증·HTML Sanitizer 구현**

## 3. 목적 (Purpose)
- 제5조의 35대 에디터 기능을 지원하는 Tiptap 커스텀 블록 스키마와 백엔드 JSON AST 검증기, Bluemonday HTML 정화기, 평문 추출기를 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Editor Engine Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-ARCH-001, TB-MEDIA-001`

## 6. 구현 범위 (Implementation Scope)
- 문단, H2~H6, 목록, 인용, Shiki 코드블록, KaTeX 수식, 표, 이미지(정렬/크기/캡션/alt), 갤러리, 비디오임베드, 링크카드, 구분선, 페이월구분선 노드 및 Markdown 양방향 변환

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `packages/editor-core/, internal/modules/editor/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/admin-console/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `posts, post_revisions` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-EDIT-001, EP-EDIT-005, EP-EDIT-006` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-BLOG-013` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 클라이언트가 보낸 HTML은 신뢰하지 않고 서버에서 JSON AST를 HTML로 직렬화 후 bluemonday로 2차 정화. 2) SafeDialer 기반 외부 링크 OG 카드 추출기 포함.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `packages/editor-core/src/extensions/*, internal/modules/editor/ast_validator.go, internal/modules/editor/serializer.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-EDIT-001, TEST-SEC-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 악성 XSS 벡터가 포함된 AST/Markdown 입력 시 출력 HTML에서 스크립트 실행 가능 요소가 0건이어야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 프론트엔드와 백엔드 간 노드 스키마 불일치 -> JSON Schema 정의 공유로 해결

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-EDITOR-001/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

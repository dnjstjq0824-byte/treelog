# TB-POST-003 — 외부 블로그 가져오기(WXR/Markdown Import) 및 블로그 전체 백업 내보내기(Export ZIP) 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 6`

---

## 1. Work Order ID
- `TB-POST-003`

## 2. 작업명 (Task Title)
- **외부 블로그 가져오기(WXR/Markdown Import) 및 블로그 전체 백업 내보내기(Export ZIP) 구현**

## 3. 목적 (Purpose)
- 제17조에 따른 Markdown/HTML/WXR ZIP 전체 내보내기와 소유권 동의 기반 외부 블로그 가져오기, 외부 이미지 안전 이관, 구 URL 301 매핑 및 오류 리포트를 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Backend Developer Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-POST-001, TB-MEDIA-001, TB-SEO-002`

## 6. 구현 범위 (Implementation Scope)
- 스트리밍 ZIP/WXR 생성기, XXE/ZIP-Bomb 방어 파서, SafeDialer 외부 이미지 수집, URL 매핑 테이블, 프론트엔드 위저드 UI

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/post/portability.go, apps/blog-console/src/pages/studio/[blogId]/portability/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/admin-console/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `migration_jobs, posts, post_slug_history, user_consents` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-PRT-001, EP-PRT-002, EP-PRT-003, EP-PRT-004` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-MEM-006` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 타사 비공개 우회 수집 금지, 반드시 CONTENT_IMPORT_OWNERSHIP 동의 확인. 2) 일부 외부 이미지 실패 시 PARTIAL_SUCCESS 및 상세 에러 JSON 리포트 생성.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/post/export_worker.go, internal/modules/post/import_worker.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-INT-001, TEST-INT-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- Export한 Markdown ZIP을 새 블로그에 Import했을 때 글·태그·발행일·슬러그 매핑이 100% 복원되어야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 대용량 XML 파싱 OOM -> encoding/xml 스트리밍 토큰 디코더 사용

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-POST-003/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

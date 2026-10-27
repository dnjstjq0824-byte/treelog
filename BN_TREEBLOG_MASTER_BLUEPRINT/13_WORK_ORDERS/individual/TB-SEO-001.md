# TB-SEO-001 — 글로벌 SEO 엔진(SSR 메타, Canonical, JSON-LD 구조화 데이터, hreflang, noindex 제어) 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 5`

---

## 1. Work Order ID
- `TB-SEO-001`

## 2. 작업명 (Task Title)
- **글로벌 SEO 엔진(SSR 메타, Canonical, JSON-LD 구조화 데이터, hreflang, noindex 제어) 구현**

## 3. 목적 (Purpose)
- 제9조의 27대 SEO 요구사항 중 메타 태그, 대표 도메인 Canonical, BlogPosting/BreadcrumbList JSON-LD, 다국어 hreflang, 비공개 글 X-Robots-Tag 제어를 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Search & SEO Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-WEB-003, TB-BLOG-003`

## 6. 구현 범위 (Implementation Scope)
- SSR 헤드 메타 생성기, JSON-LD 직렬화기, 구 슬러그/구 도메인 301 리다이렉트 미들웨어, 비공개 페이지 noindex 헤더 강제

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/seo/, apps/user-web/src/composables/useBlogSeo.ts`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/admin-console/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `posts, blogs, blog_domains, post_slug_history` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-PUB-007` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-BLOG-001, SCR-BLOG-005` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 커스텀 도메인 접속 시 모든 Canonical 및 JSON-LD URL을 해당 커스텀 도메인으로 일치시킴. 2) 비공개 글은 HTTP 응답 헤더와 메타 태그 양쪽에 noindex, nofollow 부착.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/seo/metadata.go, apps/user-web/src/composables/useBlogSeo.ts`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-FE-002, TEST-SEO-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 공개 글 HTML에 유효한 BlogPosting JSON-LD와 Canonical이 포함되고 비공개 글에 X-Robots-Tag: noindex가 반환되어야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- JSON-LD 내부 문자열로 인한 스크립트 탈출 XSS -> 유니코드 이스케이프 직렬화 강제

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-SEO-001/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

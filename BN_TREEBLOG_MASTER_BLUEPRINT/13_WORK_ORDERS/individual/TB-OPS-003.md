# TB-OPS-003 — 글로벌 다국어(ko/en/ja), 타임존 정규화, CDN Edge 캐시 퍼지 및 Core Web Vitals 성능 최적화

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 9`

---

## 1. Work Order ID
- `TB-OPS-003`

## 2. 작업명 (Task Title)
- **글로벌 다국어(ko/en/ja), 타임존 정규화, CDN Edge 캐시 퍼지 및 Core Web Vitals 성능 최적화**

## 3. 목적 (Purpose)
- 한국어·영어·일본어 완벽 번역 번들, IANA 타임존 포맷팅, Surrogate-Key 기반 CDN 즉시 캐시 무효화 및 LCP/CLS/INP 최적화를 완료한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Frontend Developer Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-WEB-004, TB-SEO-001`

## 6. 구현 범위 (Implementation Scope)
- 3개 국어 JSON 메시지 번들, Intl 포맷터, 이미지 프리로드/코드 스플리팅, CDN Purge 연동

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `apps/user-web/, apps/blog-console/, packages/shared-utils/i18n/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `03_DATABASE/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `users, blogs, posts` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-USER-003` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `All 45 Screens` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) LCP < 2.0s, INP < 150ms, CLS < 0.05 달성.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `packages/shared-utils/src/i18n/*, apps/user-web/nuxt.config.ts`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-GLB-001, TEST-GLB-002, TEST-FE-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- 3개 국어 전환 시 미번역 문자열 0건 및 Lighthouse 성능/접근성/SEO 점수 기준 충족
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 다국어 폰트 용량 증가 -> 유니코드 레인지 서브셋 WOFF2 적용

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-OPS-003/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

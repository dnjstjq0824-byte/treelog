# TB-SEO-002 — 동적 XML Sitemap Index, 도메인별 Sitemap, robots.txt, RSS 2.0 / Atom 1.0 / WebSub 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 5`

---

## 1. Work Order ID
- `TB-SEO-002`

## 2. 작업명 (Task Title)
- **동적 XML Sitemap Index, 도메인별 Sitemap, robots.txt, RSS 2.0 / Atom 1.0 / WebSub 구현**

## 3. 목적 (Purpose)
- 플랫폼 Sitemap Index(25,000건 단위 샤딩), 개별 블로그/개인 도메인별 /sitemap.xml, 동적 /robots.txt, RSS 2.0 및 Atom 1.0 피드를 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Search & SEO Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-SEO-001`

## 6. 구현 범위 (Implementation Scope)
- 스트리밍 XML 생성기, Redis 캐싱 및 발행/삭제 시 즉시 무효화, WebSub Hub Ping 전송기

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/seo/sitemap.go, rss.go`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/admin-console/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `blogs, posts, blog_settings` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-PRT-005` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `N/A (XML/Text Endpoints)` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) allow_search_indexing=false 블로그나 비공개/휴지통/임시조치 게시글은 Sitemap과 RSS에서 100% 제외.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/seo/sitemap.go, internal/modules/seo/rss.go, internal/modules/seo/robots.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-SEO-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- /sitemap.xml 및 /rss.xml이 표준 XML 스키마 검증을 통과하고 비공개 글이 단 1건도 포함되지 않아야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 대형 블로그 Sitemap 생성 부하 -> 커서 샤딩 및 Redis 10분 캐시 적용

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-SEO-002/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

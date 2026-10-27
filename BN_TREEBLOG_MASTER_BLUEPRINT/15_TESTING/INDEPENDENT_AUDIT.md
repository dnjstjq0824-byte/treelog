# INDEPENDENT_AUDIT.md — Independent Audit Agent 최종 독립 감사 프로토콜 (제20조 Step 9, 제25조, 제27조)

## 1. 독립 감사 원칙
- **권한 분리**: 코드를 작성한 에이전트는 자신의 Work Order를 최종 `VERIFIED`로 승인할 수 없다. 오직 코드 수정 권한이 없는 **Independent Audit Agent**만이 독립 검증 절차를 수행하고 최종 승인/반려를 판정한다.
- **Zero-Trust Verification**: 개발 에이전트가 제출한 보고서 문구만 믿지 않고, 감사 에이전트가 직접 클린 컨테이너 환경에서 빌드·마이그레이션·통합 테스트·Playwright E2E·보안 스캔을 재실행하여 출력 로그와 DB 상태를 대조한다.

## 2. 제27조 14대 금지사항 적발 감사 체크리스트 (Audit Gate)
하나라도 적발될 경우 전체 마일스톤은 즉시 `FAILED`로 반려된다.

| 감사 번호 | 제27조 금지 항목 | 감사 에이전트의 구체적 적발/검증 방법 |
| :--- | :--- | :--- |
| `AUD-01` | UI만 만든 뒤 전체 기능 완료 선언 | 모든 프론트엔드 이벤트 핸들러가 실제 API 클라이언트를 호출하고 DB 상태를 변경하는지 E2E 확인 |
| `AUD-02` | Mock API를 운영 백엔드로 보고 | 프로덕션/통합 테스트 코드 내 MSW, 하드코딩 JSON 스텁, 가짜 슬립(`time.Sleep`) 검색 |
| `AUD-03` | 실제 DB 미사용 저장 완료 선언 | API 호출 전후 PostgreSQL 테이블 카운트 및 컬럼 값 직접 `SELECT` 대조 |
| `AUD-04` | 테스트 없이 완료 보고 | Work Order별 매핑된 `TEST-*` 자동화 테스트 코드 존재 및 실제 실행 통과 로그 확인 |
| `AUD-05` | 가짜 브라우저 캡처 / 가짜 성능 수치 | CI/로컬에서 Playwright 및 Lighthouse/k6를 직접 재실행하여 산출물 해시 대조 |
| `AUD-06` | 존재하지 않는 Git 커밋 보고 | `git cat-file -t <SHA>` 및 `git show <SHA>`로 실제 커밋 존재 여부 검증 |
| `AUD-07` | 승인 없는 프로덕션 배포 | 배포 워크플로에 PO 수동 승인 게이트(`environment: production`) 존재 확인 |
| `AUD-08` | TreeScan 기존 코드 무단 수정 | TreeScan 저장소 변경 0건 확인 및 TreeBlog 단독 구동(`TEST-ZTR-001`) 확인 |
| `AUD-09` | 타사 디자인 및 소스코드 무단 복제 | 라이선스 스캔 및 `05_UXUI/DESIGN_SYSTEM.md` 고유 토큰 사용 확인 |
| `AUD-10` | 동의 없는 플랫폼 간 데이터 공유 | `user_consents` 테이블 동의 레코드 없이 프로필 동기화 API 호출 시 `403` 거부 확인 |
| `AUD-11` | 비공개 게시글의 검색 색인 노출 | `PRIVATE`/`MUTUAL_ONLY`/`NEIGHBOR_ONLY` 글이 OpenSearch, Sitemap, RSS, TreeScan 피드에 0건인지 쿼리 검증 |
| `AUD-12` | 사용자 입력 HTML 무검증 실행 | `bluemonday` 정화 우회 경로 유무 및 프론트엔드 무검증 `v-html` 사용 정적 분석 |
| `AUD-13` | 미검증 외부 라이브러리 의존성 추가 | `go.mod` 및 `package.json`을 허용 목록과 대조하고 `govulncheck` / `pnpm audit` 실행 |
| `AUD-14` | 미실행 검증의 허위 PASS 보고 | 실행하지 못했거나 외부 승인 대기 중인 항목이 `NOT_VERIFIED` / `BLOCKED`로 정직하게 분류되었는지 확인 |

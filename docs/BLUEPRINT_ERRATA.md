# TreeBlog 블루프린트 정오표 (Errata)

작성 2026-11-24. `BN_TREEBLOG_MASTER_BLUEPRINT/` 원문 158개 파일은 수정하지 않고
보존한다. 이 문서는 전수 점검에서 코드 실측값과 어긋남이 확인된 표기만 모았다.
각 항목은 원문 위치, 원문 표기, 실측값을 함께 기록한다.

실측 기준: 마이그레이션 `migrations/*.up.sql`, `internal/` Go 소스,
`apps/*/src`와 `apps/user-web/pages`, `04_API/openapi.yaml`,
`13_WORK_ORDERS/individual/TB-*.md`, `05_UXUI/SCREEN_INVENTORY.md`.

## E-01 — Work Order 수량 표기 불일치

| 위치 | 원문 표기 | 실측값 |
| --- | --- | --- |
| `START_HERE.md` Step 2-7 | "총 48개 Work Order" | `individual/TB-*.md` 49개 파일 존재 |
| `13_WORK_ORDERS/MASTER_WORK_ORDER.md` 제목 | "45+ 원자적 작업 분해" | 49개 |
| `13_WORK_ORDERS/MASTER_WORK_ORDER.md` 본문 | "49개" (선언) | 49개, 선언은 정확 |
| `13_WORK_ORDERS/MASTER_WORK_ORDER.md` 1절 2항 | "아래 45개 Work Order에 매핑" | 같은 문서의 49개 선언과 모순 |

조치: 이후 작업은 49개를 정본으로 삼는다. 45/48 표기는 구판 잔재로 무시한다.

## E-02 — DB 테이블 수 표기 불일치

| 위치 | 원문 표기 | 실측값 |
| --- | --- | --- |
| `START_HERE.md` Step 2-4 | "테이블 사전 58개 테이블" | `CREATE TABLE` 78문 (`0001`~`0011` 합산) |
| `13_WORK_ORDERS/MASTER_WORK_ORDER.md` 1절 2항 | "57개 DB 테이블" | 동일하게 78문과 불일치 |

조치: 테이블 사전·ERD와 마이그레이션의 1:1 대조표를 새로 작성할 때까지
78문을 기준으로 삼는다. 파티션 자식 테이블 포함 여부는 대조표 작성 시 확정한다.

## E-03 — API 엔드포인트 수 표기 불일치

| 위치 | 원문 표기 | 실측값 |
| --- | --- | --- |
| `START_HERE.md` Step 2-5 | "24개 태그 그룹 122개 엔드포인트" | 113 paths / 125 operations |
| `13_WORK_ORDERS/MASTER_WORK_ORDER.md` 1절 2항 | "24개 API 그룹(106개 엔드포인트)" | 동일하게 실측과 불일치 |

조치: `04_API/openapi.yaml` 실측값을 정본으로 삼는다. G-16에서 신규 17경로를
추가해 128 paths / 142 operations가 됐다(TB-API-001 허용 디렉터리 내 변경).
`packages/api-client` 코드젠 파이프라인과 CI 드리프트 게이트가 계약을 고정한다.

## E-04 — 화면 수 표기 불일치

| 위치 | 원문 표기 | 실측값 |
| --- | --- | --- |
| `05_UXUI/SCREEN_INVENTORY.md` 제목 | "전체 45개 화면" | `SCR-*` 행 46개 (EXP 10 + BLOG 13 + MEM 15 + ADM 8) |
| `05_UXUI/SCREEN_INVENTORY.md` 3행 | "46개 화면 … 45개 고유 화면 ID" | 고유 ID도 46개이며 "-1개 통합"의 근거가 문서에 없음 |
| `13_WORK_ORDERS/MASTER_WORK_ORDER.md` 1절 2항 | "45개 UX/UI 화면" | 46개와 불일치 |

조치: 46개 고유 화면 ID를 정본으로 삼는다.

## E-05 — 초기 디렉터리 구조 표기 불일치

| 위치 | 원문 표기 | 실측값 |
| --- | --- | --- |
| `START_HERE.md` Step 3-2 | 초기 디렉터리로 `backend/` 생성 | 저장소에 `backend/` 없음, `internal/` + `cmd/` 구조로 구현됨 |

조치: `internal/` + `cmd/` 구조를 정본으로 삼는다. `TB-ARCH-001`의
허용 디렉터리 해석 시 `backend/` 언급은 `internal/` + `cmd/`로 읽는다.

## E-06 — 에디터·프론트엔드 스택 버전 표기

| 위치 | 원문 표기 | 실측값 |
| --- | --- | --- |
| 블루프린트 `README.md` 2절 | "Tiptap v2 / ProseMirror" | `@tiptap/* ^3.32.3` (v3) 설치됨 |
| `01_A_TO_Z_BLUEPRINT/C_CONTENT_EDITOR.md` | Tiptap v2 전제 서술 | v3 기준으로 재해석 필요 |
| `01_A_TO_Z_BLUEPRINT/F_FRONTEND.md` | 코드젠·공용 패키지 기반 서술 | OpenAPI 코드젠 미도입, `editor-core` 65바이트 스텁 |

조치: 에디터 관련 구현·문서는 Tiptap v3 기준으로 작성한다.

## E-07 — Go 버전 하한 표기

| 위치 | 원문 표기 | 실측값 |
| --- | --- | --- |
| `START_HERE.md` Step 6-3, 블루프린트 `README.md` | "Go 1.22+" | `go.mod`: `go 1.23.0` |

조치: 하한 자체는 충족하므로 참고 표기로만 남긴다. 신규 문서는 1.23을 기준으로 한다.

## E-08 — 상태 문서의 구판 서술 (본 저장소 문서, 정정 완료)

`IMPLEMENTATION_STATUS.md`와 `README.md`의 "first two migrations",
"pages are scaffolds", "`TB-API-001`/`TB-WEB-001` NOT_STARTED" 서술은
2026-10-28 기준 구판이었으며 2026-11-24 개정에서 실측값으로 교체했다.

## E-09 — 구현기 마이그레이션 분기 (정보)

블루프린트 `03_DATABASE/migrations/`는 0010까지만 수록한다. 저장소 `migrations/`의
0011(인증 흐름), 0012(보호글), 0013(정산 컬럼)은 구현기에 추가된 것으로 원본과
비교 대상이 아니다. 이후 스키마 변경은 저장소 마이그레이션을 정본으로 삼는다.

## 적중하지 않은 점검 의견 (수정하지 않음)

- "테마 문서 8종 vs 시드 6종, 명칭 전부 상이"라는 1차 점검 의견은 오판으로 확인됐다.
  `07_BLOG_ENGINE/BLOG_THEMES.md` 5행은 6종(`forest-editorial`, `classic-neighbor`,
  `tech-monolith`, `visual-magazine`, `photo-atelier`, `paper-minimal`)을 명시하며
  `0010_seed_roles_themes_flags.up.sql`의 시드와 slug·수량이 정확히 일치한다.
- `WORK_ORDER_INDEX.csv`, `MASTER_TEST_MATRIX.csv`, `REQUIREMENTS_TRACEABILITY.csv`는
  전부 존재한다. 여전히 없는 것은 `MANIFEST.json`과 `FILE_CHECKSUMS.sha256`뿐이다.

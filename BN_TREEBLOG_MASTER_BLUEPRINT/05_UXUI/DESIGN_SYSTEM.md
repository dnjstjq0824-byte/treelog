# DESIGN_SYSTEM.md — TreeBlog 고유 디자인 시스템 (Forest & Paper Editorial UI)

## 1. 디자인 철학 (제3조 U항목 반영)
- **Naver Blog의 친숙함**: 이웃 새 글, 공감 리액션, 방명록, 카테고리 트리 등 한국 및 아시아권 사용자에게도 직관적인 관계형 동선 채택.
- **Tistory의 자유도**: 스킨 레이아웃(매거진/미니멀/기술문서/포토갤러리)과 CSS 디자인 토큰 변수를 통한 개성 표현.
- **Google Blogger의 간편함**: 복잡한 설정 없이 3번의 클릭만으로 블로그 개설 및 글 발행이 가능한 군더더기 없는 관리 UX.

## 2. 컬러 토큰 시스템 (Light & Dark Mode WCAG AA 준수)

| 토큰명 (CSS Variable) | Light Mode 값 | Dark Mode 값 | 용도 및 명도 대비율 |
| :--- | :--- | :--- | :--- |
| `--tb-bg-canvas` | `#FFFFFF` | `#0B0F19` | 기본 배경면 |
| `--tb-bg-elevated` | `#F8FAFC` | `#111827` | 카드, 사이드바, 모달 표면 |
| `--tb-bg-paper` | `#FAF8F5` | `#1E293B` | 장문 읽기 전용 웜페이퍼(Warm Paper) 테마 배경 |
| `--tb-text-primary` | `#0F172A` | `#F8FAFC` | 본문 및 제목 기본 텍스트 (대비율 15.8:1) |
| `--tb-text-secondary` | `#475569` | `#94A3B8` | 메타데이터, 작성일, 보조 설명 (대비율 7.2:1) |
| `--tb-brand-primary` | `#15803D` (Forest 700) | `#22C55E` (Forest 500) | 주요 CTA 버튼, 활성 링크, 브랜드 강조 |
| `--tb-status-danger` | `#DC2626` | `#F87171` | 삭제, 차단, 신고, 오류 상태 표시 |
| `--tb-status-warning` | `#D97706` | `#FBBF24` | 예약 대기, 임시조치 안내, DNS 미검증 경고 |

## 3. 타이포그래피 스케일 및 가독성 규칙
- **본문 기본 크기**: 데스크톱 `17px` (`1.0625rem`), 모바일 `16px` (`1rem`), 줄간격(`line-height`) `1.75` (장문 독서 피로도 최소화).
- **최적 글줄 길이 (Measure)**: 본문 단락 최대 너비는 `68ch ~ 74ch` (`680px ~ 740px`)로 제한하여 시선 이동 거리를 최적화한다.
- **제목 계층**:
  - 문서 제목(Post Title, `<h1>`): `clamp(1.75rem, 3vw, 2.5rem)`, `font-weight: 800`, `line-height: 1.25`
  - 본문 대제목(`<h2>`): `1.5rem` (`24px`), `font-weight: 700`, 상단 여백 `2.5rem`
  - 본문 중제목(`<h3>`): `1.25rem` (`20px`), `font-weight: 600`, 상단 여백 `2rem`

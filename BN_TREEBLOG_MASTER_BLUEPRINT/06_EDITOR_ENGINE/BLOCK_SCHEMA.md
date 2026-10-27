# BLOCK_SCHEMA.md — 에디터 블록 노드 및 마크(Mark) 스키마 명세 (제5조 35대 기능 매핑)

## 1. 허용 블록 노드(Block Nodes) 정의
| 노드 타입 (`type`) | 허용 속성 (`attrs`) | 렌더링 HTML 태그 및 기능 설명 |
| :--- | :--- | :--- |
| `doc` | `version: "1.0"` | 최상위 루트 문서 노드 |
| `heading` | `level: 2..6`, `id: string` | `<h2>`~`<h6>` (문서 제목이 `<h1>`이므로 본문 내 H1 입력 시 자동으로 H2로 정규화하여 SEO 단일 H1 원칙 준수, `id` 자동 부여로 목차 TOC 앵커 연결) |
| `paragraph` | `textAlign: left|center|right|justify` | 기본 본문 단락 `<p>` |
| `bulletList` / `orderedList` | `start: number` | 글머리표 `<ul>` 및 번호 목록 `<ol>` |
| `taskList` / `taskItem` | `checked: boolean` | 인터랙티브 체크리스트 |
| `blockquote` | `style: default|quote|accent` | 인용문 `<blockquote>` (출처 캡션 포함 가능) |
| `codeBlock` | `language: string`, `filename: string`, `showLineNumbers: boolean` | Shiki 구문 강조가 적용된 `<pre><code>` 블록 |
| `table` / `tableRow` / `tableHeader` / `tableCell` | `colspan`, `rowspan`, `colwidth`, `background` | 셀 병합, 열 너비 조절, 헤더 고정 및 모바일 가로 스크롤 래퍼가 적용된 표 |
| `figureImage` | `mediaAssetId`, `src`, `alt`(필수권장), `caption`, `alignment`(`left`/`center`/`right`/`wide`/`full`), `width` | `<figure><picture>...<img alt="..."></picture><figcaption>...</figcaption></figure>` |
| `imageGallery` | `layout: grid-2|grid-3|masonry|slider`, `items: Array<{mediaAssetId, alt, caption}>` | 다중 사진 콜라주 및 슬라이드 갤러리 블록 |
| `videoEmbed` | `provider: youtube|vimeo|globalvideo`, `embedUrl`, `aspectRatio`, `caption` | 샌드박스 처리된 반응형 동영상 임베드 블록 |
| `linkCard` | `url`, `title`, `description`, `thumbnailUrl`, `siteName` | 외부 링크 붙여넣기 시 생성되는 OpenGraph 북마크 카드 |
| `callout` | `tone: info|tip|warning|danger`, `icon: string` | 강조 알림 상자 블록 |
| `mathBlock` | `latex: string` | KaTeX 서버/클라이언트 수식 렌더링 블록 |
| `horizontalRule` | `style: solid|dashed|dots` | 구분선 `<hr>` |
| `paywallDivider` | `previewText: string` | 유료 멤버십/단건 유료글에서 **무료 공개 티저 구간과 유료 본문 구간을 가르는 경계선** (서버에서 이 노드 이후를 절삭) |

## 2. 허용 인라인 마크(Inline Marks) 정의
- `bold` (`<strong>`), `italic` (`<em>`), `underline` (`<u>`), `strike` (`<s>`), `code` (인라인 `<code>`), `highlight` (`<mark data-color="...">`), `subscript` (`<sub>`), `superscript` (`<sup>`), `link` (`<a href="..." target="_blank" rel="noopener noreferrer ugc">`, 프로토콜은 `http:`, `https:`, `mailto:`만 허용).

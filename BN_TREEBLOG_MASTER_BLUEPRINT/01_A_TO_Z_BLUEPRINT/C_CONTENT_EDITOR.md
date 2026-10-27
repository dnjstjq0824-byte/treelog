# C_CONTENT_EDITOR — Content Creation Engine (블로그 글쓰기 에디터 엔진)
> **부제**: 리치 텍스트 에디터, Markdown, 블록 편집, 사진·영상, 표, 코드, 링크, 임시저장, 자동저장, 예약 발행, 글 수정 이력
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-C_CONTENT_EDITOR`
> **적용 조항**: 제3조 (C_CONTENT_EDITOR 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- TreeBlog의 핵심 제품 경쟁력으로서, 네이버 스마트에디터 ONE의 직관적인 미디어·레이아웃 편집 편의성, 티스토리의 마크다운/코드 친화성, Notion/Tiptap의 현대적 슬래시(`/`) 블록 편집 경험을 결합한 독립 에디터 엔진을 구축한다.
- 네트워크 오류나 브라우저 크래시 상황에서도 작성 중인 원고가 단 한 글자도 유실되지 않는 무결성 자동저장 및 리비전 복원 체계를 제공한다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **개인/전문 블로거**: 슬래시 커맨드(`/image`, `/code`, `/table`, `/callout`, `/gallery`, `/toc`)와 마크다운 단축키(`## `, ```, `> `)를 혼합 사용하여 빠르게 글을 작성하고, 모바일에서도 사진 정렬과 캡션을 손쉽게 편집한다.
- **기업 블로그 작성자/편집자**: 작성자가 초안을 작성해 '검토 요청(`IN_REVIEW`)'으로 올리면 편집자가 리비전 Diff를 확인하고 수정 후 예약 발행(`SCHEDULED`)을 설정한다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-EDIT-001 (제5조 35대 필수 기능 완비)**:
  1~8) 리치 텍스트, 제목/본문 분리, 블록 드래그 이동, Bold/Italic/Underline/Strike/Highlight, H1~H6 정책(본문 내 H1은 문서 제목과 중복 방지를 위해 H2~H4 권장 및 자동 목차 생성), 글머리표/번호/체크리스트, 인용문(스타일 3종), 코드 블록(Shiki 기반 100+ 언어 구문 강조 및 복사 버튼).
  9~17) 표(행/열 추가·삭제·배경색·셀 병합), 이미지 업로드(드래그앤드롭/클립보드 붙여넣기), 이미지 정렬(좌/중앙/우/본문폭/전체폭) 및 리사이즈 핸들, 이미지 캡션 및 필수 `alt` 대체 텍스트 입력 UI, 이미지 갤러리(2~4열 그리드/슬라이드), 동영상 삽입, 외부 링크(자동 OG 프리뷰 카드 생성), 임베드(YouTube/Vimeo/X/CodePen/GitHub Gist), 구분선(3종).
  18~24) 실행 취소/다시 실행(History 200단계), 듀얼 자동저장(IndexedDB 3초 + 서버 15초), 수동 임시저장 목록(최대 100개), 디바이스별(PC/태블릿/모바일) 실시간 미리보기, 예약 발행(분 단위 타임존 지원), 게시글 수정, 리비전 수정 이력 조회 및 특정 버전 복원(Diff 비교).
  25~35) 태그(최대 30개, 자동완성), 카테고리 지정, 대표 썸네일 이미지 지정(미지정 시 첫 이미지 자동 선택 또는 텍스트 커버 생성), 사용자 정의 URL 슬러그(`slug`), SEO 메타 제목·설명 직접 편집, 6단계 공개 범위(`PUBLIC`, `UNLISTED`, `NEIGHBOR_ONLY`, `MUTUAL_ONLY`, `PAID_MEMBERS_ONLY`, `PRIVATE`), 댓글/공감/복사 허용 여부 설정, 발행/예약 상태 관리, Markdown 파일(`.md`) 가져오기·내보내기, 모바일 가상 키보드 상단 고정 툴바 대응, 오프라인 작성 후 온라인 복구 시 충돌 해결 모달 제공.
- **REQ-EDIT-002 (3중 데이터 저장 구조)**: 원본 `content_json` (ProseMirror JSON AST), 정화된 `content_html` (Sanitized HTML), 검색 색인용 `content_plain` (순수 텍스트)을 분리 생성·저장한다.

## 4. 데이터 구조 (Data Structures & Domain Models)
```json
{
  "type": "doc",
  "version": "1.0",
  "content": [
    {
      "type": "heading",
      "attrs": { "level": 2, "id": "sec-intro" },
      "content": [{ "type": "text", "text": "TreeBlog 에디터 아키텍처 소개" }]
    },
    {
      "type": "figureImage",
      "attrs": {
        "mediaAssetId": "018f2a10-7b11-7000-8000-000000000001",
        "src": "https://media.treeblog.example.com/variants/018f.../w1200.webp",
        "alt": "시스템 아키텍처 다이어그램",
        "caption": "그림 1. 에디터 3중 저장 파이프라인",
        "alignment": "wide",
        "width": 100
      }
    }
  ]
}
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- `packages/editor-core`: Tiptap v2 (`@tiptap/core`, `@tiptap/pm`) 기반 커스텀 Node/Mark 확장 패키지.
- **듀얼 자동저장 엔진 (`useAutosave`)**:
  - 로컬 계층: 타이핑 후 3초 디바운스로 브라우저 `IndexedDB` (`treeblog_drafts` 스토어)에 JSON AST와 `client_updated_at`, `base_revision_no` 저장.
  - 서버 계층: 변경사항 존재 시 15초 주기로 `PUT /api/v1/blogs/{blogId}/posts/{postId}/autosave` 호출.
  - 충돌 감지: 로컬 스냅샷의 `base_revision_no`보다 서버의 최신 `revision_no`가 크거나 다른 디바이스에서 수정된 경우 좌우 Diff 비교 모달을 띄워 사용자가 '내 버전 유지', '서버 버전 적용', '새 임시글로 사본 저장' 중 선택하도록 지원.

## 6. 백엔드 로직 (Backend Logic & Processing)
- `internal/modules/editor` & `internal/modules/post`:
  - **AST Validator**: 허용된 노드 타입(`paragraph`, `heading`, `bulletList`, `orderedList`, `taskList`, `blockquote`, `codeBlock`, `table`, `figureImage`, `imageGallery`, `videoEmbed`, `linkCard`, `callout`, `horizontalRule`, `mathBlock`)과 속성만 존재하는지 JSON 스키마 검증.
  - **Server-side HTML Serializer & Sanitizer**: AST를 시맨틱 HTML5로 변환 후 `bluemonday` 정책으로 `<script>`, `<iframe>`(허용된 도메인 샌드박스 외), `javascript:` URL, 인라인 이벤트 핸들러(`onload`, `onerror` 등)를 100% 제거.

## 7. API 계약 (API Contracts)
- `POST /api/v1/blogs/{blogId}/posts` (신규 초안 생성)
- `PUT /api/v1/blogs/{blogId}/posts/{postId}/autosave` (자동저장 갱신 — Optimistic Lock `expected_version` 포함)
- `POST /api/v1/blogs/{blogId}/posts/{postId}/publish` (즉시 발행 또는 예약 발행 설정)
- `GET /api/v1/blogs/{blogId}/posts/{postId}/revisions` (수정 이력 목록 조회)
- `POST /api/v1/blogs/{blogId}/posts/{postId}/revisions/{revisionId}/restore` (특정 버전으로 복원)
- `POST /api/v1/blogs/{blogId}/editor/link-preview` (SSRF 방어 적용 외부 URL OG 메타데이터 추출)
- `POST /api/v1/blogs/{blogId}/editor/markdown-import` / `GET /api/v1/blogs/{blogId}/posts/{postId}/markdown-export`

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- 주요 테이블: `posts`, `post_revisions`, `autosave_drafts`, `post_categories`, `post_tags`, `tags`, `post_media`, `scheduled_publications`.
- `post_revisions`는 게시글 발행 시 및 주요 수동 저장 시 생성되며, 스토리지 최적화를 위해 최근 50개 리비전은 전체 보관하고 이후 오래된 자동 생성 리비전은 배치 잡으로 압축/정리한다.

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **발행 중 이미지 변환 미완료 예외**: 에디터에서 고해상도 이미지를 올리고 즉시 발행할 경우, 원본 Presigned 업로드만 완료되면 임시 원본 URL(또는 블러 플레이스홀더)로 즉시 발행을 허용하고 백그라운드 워커가 WebP/AVIF 변환 완료 즉시 URL을 교체한다.
- **네트워크 단절 예외**: 오프라인 전환 시 에디터 상단에 `오프라인 모드 — 변경사항이 기기에 안전하게 보관 중입니다` 배지가 표시되며, 온라인 복귀 시 자동으로 서버와 동기화한다.

## 10. UX/UI 구조 (UX/UI Structure)
- 방해 없는 글쓰기(Distraction-free) 모드와 사이드바 설정 패널(발행 설정, SEO, 태그, 카테고리, 예약 시간) 토글 구조 제공 (`SCR-BLOG-013`).
- 자세한 블록 스키마와 워크플로는 `06_EDITOR_ENGINE/*` 문서에 상세 정의됨.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- Stored XSS 원천 차단: 클라이언트 제출 HTML 무시, 서버 AST 렌더링 + `bluemonday` 이중 검증.
- 링크 프리뷰 SSRF 차단: `SafeDialer`를 통해 사설 IP 및 로컬호스트 접근 원천 봉쇄.
- 비공개 초안(`DRAFT`) 미디어 보호: 아직 공개 발행되지 않은 초안에 첨부된 이미지는 서명된 단기 URL(Signed URL)로만 작성자에게 보인다.

## 12. 성능 및 확장성 (Performance & Scalability)
- 10,000단어 이상 장문 및 이미지 50장이 포함된 게시글에서도 타이핑 지연(Input Latency) < 16ms(60fps) 유지 (ProseMirror 가상 렌더링 및 이미지 Lazy Loading 적용).
- 자동저장 페이로드가 500KB 이상일 경우 Gzip/Brotli 압축 전송.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- Frontend: `@tiptap/vue-3`, `@tiptap/starter-kit`, `@tiptap/extension-table`, `@tiptap/extension-image`, `shiki`, `katex`, `idb-keyval`, `turndown`, `marked`.
- Backend: `github.com/microcosm-cc/bluemonday`, `github.com/yuin/goldmark`.

## 14. 개발 우선순위 (Development Priority)
- **Phase 3 (핵심 제품 경험)**: 블로그 기본 엔진 구축 직후 최우선 구현 (`TB-EDITOR-001` ~ `TB-EDITOR-003`, `TB-POST-001`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 선행: `B_BLOGGING_EXPERIENCE`, `V_VIDEO_MEDIA` (미디어 업로드 API).
- 후행: `O_OPENSEARCH` (본문 색인), `L_LIFECYCLE`.

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-EDIT-001`: `<img src=x onerror=alert(1)>` 및 `<a href="javascript:alert(1)">`가 포함된 JSON/Markdown 입력 시 서버 저장본 `content_html`에서 악성 속성이 100% 제거되는지 검증.
- `TEST-EDIT-002`: 브라우저 네트워크를 Offline으로 전환 후 본문 수정 → 페이지 새로고침 → IndexedDB 복구 프롬프트를 통해 작성 내용이 손실 없이 복원되는지 E2E 검증.
- `TEST-EDIT-003`: 리비전 v1 → v2 → v3 수정 후 v1으로 복원 시 v4 리비전이 생성되며 본문이 v1과 일치하는지 검증.

## 17. Acceptance Criteria (인수 기준)
1. 제5조의 35개 필수 에디터 기능이 데스크톱과 모바일 뷰포트에서 모두 정상 동작해야 한다.
2. 발행된 게시글의 `content_json`, `content_html`, `content_plain` 3개 컬럼이 DB에 모두 비어있지 않고 일관되게 저장되어야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- Playwright 에디터 35개 기능 자동화 테스트 통과 리포트, XSS 페이로드 정화 단위 테스트 로그, DB `posts` 및 `post_revisions` 레코드 덤프.

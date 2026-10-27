# EDITOR_ARCHITECTURE.md — Tiptap / ProseMirror 기반 글쓰기 에디터 엔진 아키텍처

## 1. 아키텍처 개요 (제5조)
TreeBlog 에디터(`packages/editor-core`)는 ProseMirror의 결정론적 상태 모델(Deterministic State Transaction Model)과 Tiptap v2의 모듈형 확장 구조를 기반으로 설계되었다.

```
[User Input: Typing / Slash Command (/) / Paste Markdown or HTML / Drag & Drop Media]
                                        │
                                        ▼
               [ProseMirror EditorState & Schema Validation Layer]
                - Custom Nodes: figureImage, imageGallery, videoEmbed,
                  linkCard, callout, codeBlockShiki, mathBlock, paywallDivider
                                        │
        ┌───────────────────────────────┴───────────────────────────────┐
        ▼ (Local 3s Debounce)                                           ▼ (Server 15s / Publish)
[IndexedDB Local Snapshot]                                  [Go Backend Editor/Post Module]
 - Offline Crash Recovery                                    1. JSON Schema AST Validator
 - Revision Conflict Detector                                2. Deterministic HTML Serializer
                                                             3. Bluemonday XSS Sanitizer
                                                             4. Plaintext Extractor (for OpenSearch)
```

## 2. 3중 데이터 표현 분리 저장 원칙
| 컬럼명 | 데이터 타입 | 용도 및 생성 주체 |
| :--- | :--- | :--- |
| `posts.content_json` | `JSONB` | 에디터 재편집 및 리비전 Diff 계산을 위한 **유일한 원본(Source of Truth)** |
| `posts.content_html` | `TEXT` | Go 서버가 `content_json`으로부터 렌더링 후 `bluemonday`로 정화한 **SSR 고속 출력용 안전 HTML** |
| `posts.content_plain` | `TEXT` | 모든 태그를 제거한 순수 본문 텍스트. **OpenSearch 전문 검색 색인, 요약문 생성, 단어 수/읽기 시간 산출용** |

# BLOG_EDITOR.md — 블로그 글쓰기 에디터 스튜디오 UX/UI 상세 명세 (SCR-BLOG-013)
> **부제**: 데스크톱 및 모바일 반응형 Tiptap 블록 에디터, 슬래시 메뉴, 미디어 툴바, 리비전 Diff, 발행 드로어 상세 설계
> **적용 조항**: 제4조, 제11조 및 `01_A_TO_Z_BLUEPRINT/U_UXUI.md` (11대 필수 설계 항목 전수 포함)

---

## [SCR-BLOG-013] 블로그 글쓰기 에디터 (`/studio/{blogId}/editor/{postId}`)
1. **페이지 목적**: Tiptap/ProseMirror 기반 35대 기능을 갖춘 핵심 블록 글쓰기 및 발행 스튜디오
2. **UI 컴포넌트 구조**: `EditorTopBar(자동저장 상태/미리보기/발행버튼), TitleSubtitleInputs, TiptapBlockCanvas, FloatingBubbleToolbar, SlashCommandMenu, MediaUploadDropzone, PublishSettingsDrawer(카테고리/태그/슬러그/SEO/공개범위/예약시간/CCL), RevisionHistoryDrawer, ConflictResolutionModal`
3. **사용자 액션**: 블록 작성/이동, 이미지/동영상/표/코드 삽입, 임시저장, 미리보기, 즉시/예약 발행, 버전 복원, 마크다운 입출력
4. **연동 API**: `EP-EDIT-001~006, EP-PUB-001~003, EP-MED-001~005`
5. **클라이언트 상태값 (State)**: `editorAst, dirtyFlag, autosaveStatus(Saved/Saving/Offline), publishConfig, revisions`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 툴바 `role=toolbar`, 단축키 안내(`Ctrl+B`, `/`), 이미지 삽입 시 `alt` 입력 유도 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: CSR/Auth Protected, `robots: noindex, nofollow`
11. **검증 테스트 항목**: `TEST-EDIT-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

---

## 부록: 에디터 반응형 뷰포트별 레이아웃 및 인터랙션 정책
1. **Desktop (1024px 이상)**:
   - 중앙 집중형 원고 캔버스(최대 본문 폭 `740px`, 와이드 블록 `1080px`, 전체폭 `100vw`) + 우측 접이식 발행/SEO/리비전 설정 드로어(`360px`).
   - 블록 좌측에 호버 시 나타나는 `+`(블록 추가) 및 `⋮⋮`(드래그 핸들 / 블록 옵션 메뉴) 제공.
2. **Mobile & Tablet (320px ~ 1023px)**:
   - 모바일 소프트 키보드(Virtual Keyboard)가 올라올 때 `visualViewport` API를 감지하여 **키보드 바로 위에 고정되는 가로 스크롤 플로팅 툴바(사진 추가, 제목 크기, 굵게, 목록, 링크, 실행취소)**를 배치한다.
   - 이미지 크기 및 정렬 변경은 드래그 핸들 대신 터치 친화적인 4단계 프리셋 버튼(`작게`, `본문폭`, `가득차게`, `정렬 변경`)으로 조작한다.

# AUTOSAVE.md — 듀얼 레이어 자동저장 및 네트워크 오류 복구 명세

## 1. 듀얼 레이어 자동저장 아키텍처
1. **Layer 1 — 브라우저 로컬 IndexedDB 즉시 백업 (3초 디바운스)**:
   - 사용자가 타이핑하거나 이미지를 추가할 때마다 마지막 입력 후 3초(또는 50타 입력) 시점에 `IndexedDB`의 `draft_snapshots` 스토어에 `{postId, baseRevisionNo, title, subtitle, contentJson, updatedAt}`을 비동기 저장한다.
   - 갑자기 노트북 배터리가 방전되거나 브라우저 탭이 강제 종료되어도 로컬 스냅샷이 100% 보존된다.
2. **Layer 2 — 서버 자동저장 API 동기화 (15초 주기)**:
   - `dirtyFlag == true`이고 `navigator.onLine == true`일 때 매 15초마다 `PUT /api/v1/blogs/{blogId}/posts/{postId}/autosave`를 호출하여 PostgreSQL `autosave_drafts` 테이블에 저장한다.

## 2. 오프라인 전환 및 버전 충돌(Conflict) 해결 알고리즘
- **시나리오 1 (오프라인 작성 후 온라인 복구)**:
  - 네트워크 연결이 끊기면 에디터 상단 상태 배지가 `오프라인 저장 중 (기기에 안전하게 보관됨)`으로 바뀌며 글쓰기는 중단 없이 계속 가능하다.
  - 네트워크가 복구(`window.addEventListener('online')`)되면 자동으로 서버 저장을 재시도한다.
- **시나리오 2 (멀티 디바이스 동시 수정 충돌 감지)**:
  - 데스크톱에서 작성하던 글을 닫지 않고 모바일에서 이어 수정하여 서버의 `posts.revision_no`가 올라간 경우, 데스크톱이 구 버전 `base_revision_no`로 저장을 시도하면 서버는 `409 EDIT_VERSION_CONFLICT`와 최신 서버 본문을 반환한다.
  - 에디터는 자동으로 덮어쓰지 않고 **충돌 해결 모달(`ConflictResolutionModal`)**을 띄워 좌측(내 기기 버전)과 우측(서버 최신 버전)의 문단별 Diff를 보여준 뒤 사용자가 1) 내 기기 버전으로 덮어쓰기, 2) 서버 최신 버전 불러오기, 3) 내 기기 버전을 별도 새 임시글로 분리 저장 중 하나를 선택하도록 한다.

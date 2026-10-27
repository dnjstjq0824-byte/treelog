# PUBLISH_WORKFLOW.md — 게시글 발행·검토·예약 발행 파이프라인 명세

## 1. 발행 전 사전 검증 체크리스트 (Pre-Publish Validation)
발행 버튼 클릭 시 백엔드 트랜잭션(`POST /api/v1/blogs/{blogId}/posts/{postId}/publish`)은 다음 순서로 검증 및 처리를 수행한다.
1. **권한 확인**: 요청자가 블로그의 `OWNER`, `ADMIN`, `EDITOR`이거나 본인 글에 대한 `AUTHOR`인지 확인한다. (만약 기업 블로그의 `require_approval_for_authors == true`이고 요청자가 `AUTHOR`이면 상태를 `IN_REVIEW`로 전이하고 편집자에게 검토 요청 알림을 발송한다.)
2. **슬러그(`slug`) 정규화 및 중복 해결**: 사용자가 슬러그를 비워둔 경우 제목으로부터 슬러그를 자동 생성하며, 동일 블로그 내 중복 슬러그 존재 시 `-2`, `-3` 접미사를 자동 부여한다. 기존 발행 글의 슬러그가 변경된 경우 구 슬러그를 `post_slug_history`에 저장하여 `301` 리다이렉트를 보장한다.
3. **미디어 참조 동기화**: 본문 AST 내 모든 `mediaAssetId`를 추출하여 `post_media` 테이블에 매핑하고 `media_assets.reference_count`를 갱신하여 고아 파일 가비지 컬렉션 대상에서 보호한다.
4. **대표 이미지 및 요약문 자동 보완**: `cover_media_id`가 미지정된 경우 본문 첫 번째 이미지를 대표 썸네일로 지정하며, `excerpt` 및 `seo_description`이 비어 있으면 `content_plain` 첫 160자를 추출하여 채운다.
5. **원자적 커밋 및 Outbox 발행**: `posts` 업데이트 + `post_revisions` 스냅샷 생성 + `search_index_jobs` 등록 + `outbox_events` 발행을 단일 PostgreSQL 트랜잭션으로 커밋한다.

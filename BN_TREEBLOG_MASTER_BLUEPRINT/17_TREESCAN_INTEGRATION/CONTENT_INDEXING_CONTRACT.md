# CONTENT_INDEXING_CONTRACT.md — TreeScan 검색 색인 동기화 및 비공개 노출 차단 계약 (제9조, 제18조)

## 1. 증분 색인 피드 (`GET /api/v1/treescan/sync/feed`) 프로토콜
TreeScan 색인 크롤러는 전체 게시글을 매번 다시 긁어갈 필요 없이, 단조 증가하는 시퀀스 커서(`sequence_cursor`)를 사용해 변경된 이벤트만 고속으로 가져간다.

### 1.1 공개 게시글 발행/수정 시 (`UPSERT`)
- 발생 조건: `post.status == 'PUBLISHED' && post.visibility == 'PUBLIC' && blog.status == 'ACTIVE' && blog_settings.allow_treescan_sync == true`
- 전달 데이터: 문서 고유 ID, 대표 Canonical URL, 제목, 부제목, 평문 본문(`content_plain`), 요약, 대표 썸네일 URL, 작성자명, 블로그명, 카테고리, 태그, 작성 언어, 발행/수정 시각, 객관적 품질 시그널(`word_count`, `quality_score`, `is_near_duplicate`).

### 1.2 비공개 전환·삭제·블로그 정지 시 (`DELETE_TOMBSTONE`)
- 발생 조건: 기존에 공개(`PUBLIC`)였던 게시글이 `PRIVATE`, `NEIGHBOR_ONLY`, `MUTUAL_ONLY`, `UNLISTED`, `UNPUBLISHED`, `TRASHED`, `QUARANTINED`로 바뀌거나 블로그가 정지/폐쇄된 경우.
- 전달 데이터 (본문 절대 미포함):
```json
{
  "sequence_cursor": 1048592,
  "sync_action": "DELETE_TOMBSTONE",
  "document_id": "treeblog:post:018f2a10-7b11-7000-8000-000000000101",
  "canonical_url": "https://tech-pioneer.treeblog.example.com/posts/go-opensearch-architecture",
  "deleted_at": "2026-10-27T15:10:00Z"
}
```
- 이를 통해 TreeBlog에서 비공개로 바뀐 글이 외부 TreeScan 검색 결과에 남아 있는 프라이버시 사고를 원천 차단한다.

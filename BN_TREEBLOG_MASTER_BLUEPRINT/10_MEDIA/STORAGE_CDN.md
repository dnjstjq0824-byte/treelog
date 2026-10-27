# STORAGE_CDN.md — 오브젝트 스토리지 격리, 수명주기 및 글로벌 CDN 캐시 정책

## 1. Dual-Bucket 물리 격리 아키텍처
1. **`treeblog-raw-private` (비공개 원본 버킷)**:
   - 퍼블릭 액세스 전면 차단(Block Public Access).
   - 업로드된 미검증 원본 파일, EXIF 제거 전 원본 사진, 블로그 내보내기/가져오기 백업 ZIP 파일 보관.
   - S3 Lifecycle Rule: 미완료 멀티파트 업로드(`AbortIncompleteMultipartUpload`)는 24시간 후 자동 파기.
2. **`treeblog-media-public` (공개 CDN 연동 버킷)**:
   - 오직 미디어 워커가 매직바이트 검사 및 EXIF 제거를 완료한 안전한 파생 이미지(`media_variants`)만 저장된다.
   - 파일 경로에 콘텐츠 해시가 포함되므로 CDN에서 `Cache-Control: public, max-age=31536000, immutable`로 1년간 캐싱하며, 게시글 영구 삭제 시 CDN Purge API로 즉시 무효화한다.

## 2. 고아 미디어 파일(Orphan Media) 자동 정리 배치
- 에디터에서 이미지를 업로드만 하고 글을 저장하지 않고 나간 경우 `media_assets.reference_count == 0` 상태로 남는다.
- 매일 04:00 UTC에 실행되는 `OrphanMediaGCWorker`가 `reference_count == 0 AND orphan_since < NOW() - INTERVAL '7 days'`인 에셋을 찾아 S3 오브젝트와 DB 레코드를 안전하게 정리하여 스토리지 과금 낭비를 방지한다.

# MEDIA_UPLOAD.md — S3 Presigned 업로드 세션 및 파일 보안 검증 명세 (제13조)

## 1. Presigned URL 2단계 업로드 프로토콜
API 서버의 네트워크 대역폭과 메모리를 보호하기 위해 모든 바이너리 파일 업로드는 클라이언트→S3 직접 업로드(Presigned PUT / Multipart)로 처리한다.
1. **세션 발급 (`POST /api/v1/blogs/{blogId}/media/upload-sessions`)**:
   - 입력: `filename`, `mime_type`, `size_bytes`, `sha256_hash`.
   - 검증: 허용 MIME(`image/jpeg`, `image/png`, `image/webp`, `image/gif`, `image/avif`, `image/heic`), 크기 상한(일반 이미지 25MB, GIF 50MB), 사용자 일일 업로드 쿼터(기본 2GB/일) 확인.
   - 중복 식별: 동일 `blog_id` + `sha256_hash`로 이미 `READY` 상태인 `media_assets`가 존재하면 업로드 없이 즉시 기존 에셋 정보를 반환(`deduplicated: true`)한다.
2. **업로드 완료 및 매직바이트 검사 (`POST .../complete`)**:
   - S3 Private 버킷에 올라간 오브젝트의 첫 512바이트를 읽어 실제 바이너리 시그니처(`FF D8 FF` JPEG, `89 50 4E 47` PNG, `52 49 46 46 ... 57 45 42 50` WebP 등)를 검증한다. 불일치 시 즉시 오브젝트를 삭제하고 `415 MEDIA_MAGIC_BYTE_MISMATCH`를 반환한다.

# V_VIDEO_MEDIA — Video & Media Engine (미디어 파일 처리 및 스토리지 엔진)
> **부제**: 사진, 이미지 편집, 썸네일, 이미지 최적화, GIF, 동영상 업로드·임베드, 미디어 보안, 파일 관리
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-V_VIDEO_MEDIA`
> **적용 조항**: 제3조 (V_VIDEO_MEDIA 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- 사진과 시각 자료가 풍부한 블로그 특성상, 대용량 이미지 업로드 속도와 조회 시 로딩 성능(WebP/AVIF 변환), 작성자 개인정보 보호(GPS EXIF 강제 제거), 그리고 스토리지 비용 최적화를 동시에 달성하는 미디어 파이프라인을 설계한다(제13조).

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **블로거**: 스마트폰으로 촬영한 12MB 고해상도 사진 20장을 에디터에 드래그앤드롭하면 Presigned URL을 통해 빠르게 업로드되고, 서버에서 자동으로 방향 보정(Auto-orient), GPS 위치정보 제거, 반응형 크기별 WebP/AVIF 변환 및 대표 썸네일 생성이 완료된다.
- **시스템 운영자**: 악성 실행 파일 위장 업로드나 스토리지 용량 남용(Quota Abuse)을 방어하고, 게시글에서 삭제된 고아 미디어 파일(Orphan Files)을 수명주기 정책에 따라 자동 정리한다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-MED-001 (S3 Presigned 업로드 세션 및 무결성 검증)**:
  - 클라이언트가 파일 메타데이터(파일명, 크기, MIME, SHA-256 해시)로 업로드 세션을 요청(`POST /api/v1/media/upload-sessions`)하면, 서버가 용량 제한(이미지 최대 25MB, GIF 최대 50MB, 동영상 확장 시 최대 500MB) 및 일일 업로드 쿼터를 검증한 뒤 **Private S3 Bucket(`treeblog-raw-private`)**으로의 단기 Presigned PUT/Multipart URL(유효기간 10분)을 발급한다.
  - 동일 블로그 내 동일 `sha256_hash` 파일이 이미 존재할 경우 재업로드 없이 기존 에셋을 즉시 참조하는 **중복 제거(Deduplication)**를 지원한다.
- **REQ-MED-002 (이미지 최적화·EXIF 개인정보 제거·파생본 생성)**:
  - 업로드 완료 콜백(`POST /api/v1/media/upload-sessions/{id}/complete`) 수신 시 미디어 워커가:
    1) 파일 헤더 Magic Bytes 및 이미지 해상도 상한(16384×16384px)을 검사하고 바이러스/악성 스트림 여부를 확인한다.
    2) EXIF의 Orientation 태그를 읽어 이미지를 정방향 회전시킨 뒤, **GPS 위도/경도, 카메라 일련번호, 촬영자 이름 등 모든 민감 EXIF/IPTC/XMP 메타데이터를 100% 영구 제거**한다.
    3) 반응형 파생본(`w320`, `w640`, `w1200`, `w1920`, `thumb_400x400`)을 **WebP** 및 **AVIF** 포맷으로 생성하여 **Public CDN Bucket(`treeblog-media-public`)**에 저장하고 블러 플레이스홀더(`blurhash` / LQIP)를 추출한다.
    4) 애니메이션 GIF는 용량 절감을 위해 애니메이션 WebP 및 루프 MP4 파생본으로 변환한다.
- **REQ-MED-003 (외부 동영상 임베드 우선 및 자체 동영상 트랜스코딩 확장 구조)**:
  - 초기에는 YouTube, Vimeo, Naver TV, Global Video 등 허용된 도메인 화이트리스트 기반 oEmbed / 샌드박스 iframe 임베드를 우선 제공한다.
  - 자체 동영상 업로드 및 HLS/DASH 트랜스코딩은 `media_processing_jobs` 테이블과 FFmpeg 워커 인터페이스로 설계하되, 초기 운영 비용 통제를 위해 Feature Flag(`FF_NATIVE_VIDEO_UPLOAD`)로 단계적 활성화한다.
- **REQ-MED-004 (원본/공개 권한 분리 및 고아 파일 수명주기)**:
  - 원본 파일(`treeblog-raw-private`)은 외부에 절대 직접 공개하지 않으며, 공개 발행된 게시글의 변환 파일만 CDN을 통해 서빙한다.
  - 업로드 후 7일이 지나도록 어떤 게시글(`post_media`)이나 프로필/테마에도 연결되지 않은 고아 에셋(Orphan Asset)은 가비지 컬렉션 워커가 스토리지에서 자동 삭제한다.

## 4. 데이터 구조 (Data Structures & Domain Models)
```go
type MediaAsset struct {
    ID            uuid.UUID       `json:"id"`
    BlogID        uuid.UUID       `json:"blog_id"`
    UploaderID    uuid.UUID       `json:"uploader_id"`
    MediaType     string          `json:"media_type"`     // IMAGE, ANIMATED_GIF, VIDEO, ARCHIVE
    MimeType      string          `json:"mime_type"`      // image/jpeg, image/png, image/webp
    OriginalSize  int64           `json:"original_size"`
    SHA256Hash    string          `json:"sha256_hash"`
    Width         int             `json:"width"`
    Height        int             `json:"height"`
    Blurhash      string          `json:"blurhash"`
    AltText       string          `json:"alt_text"`
    ExifStripped  bool            `json:"exif_stripped"`  // 반드시 true여야 공개 서빙 허용
    Status        string          `json:"status"`         // UPLOADING, PROCESSING, READY, QUARANTINED, DELETED
    Variants      []MediaVariant  `json:"variants"`
}
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- 에디터 내 이미지 크롭/회전/필터 경량 편집 모달, 업로드 진행률 오버레이, 반응형 `<picture><source type="image/avif"><source type="image/webp"><img loading="lazy" alt="..."></picture>` 자동 렌더링, Blurhash 스켈레톤 프리뷰.

## 6. 백엔드 로직 (Backend Logic & Processing)
- `internal/modules/media`:
  - S3 Presigner, Magic Byte Inspector, `bimg`/`govips` (libvips C 바인딩 또는 격리 실행) 기반 고속 이미지 프로세서, oEmbed 메타데이터 리졸버.

## 7. API 계약 (API Contracts)
- `POST /api/v1/blogs/{blogId}/media/upload-sessions` (Presigned 업로드 세션 생성)
- `POST /api/v1/blogs/{blogId}/media/upload-sessions/{sessionId}/complete` (업로드 완료 및 변환 큐 등록)
- `GET /api/v1/blogs/{blogId}/media` / `PATCH /api/v1/blogs/{blogId}/media/{assetId}` (블로그 미디어 라이브러리 조회 및 `alt` 텍스트 수정)
- `DELETE /api/v1/blogs/{blogId}/media/{assetId}` (미디어 삭제)
- `POST /api/v1/blogs/{blogId}/media/video-embed` (외부 동영상 URL 검증 및 oEmbed 추출)

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- 주요 테이블: `upload_sessions`, `media_assets`, `media_variants`, `media_processing_jobs`, `post_media`.

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **손상된 이미지 또는 이미지 폭탄(Decompression Bomb) 업로드 시**: 워커가 변환 단계에서 즉시 `status = 'QUARANTINED'`로 전환하고 원본 오브젝트를 삭제하며 에디터에 명확한 오류 토스트(`파일이 손상되었거나 허용된 해상도를 초과합니다`)를 전달한다.

## 10. UX/UI 구조 (UX/UI Structure)
- 에디터에서 이미지 선택 시 툴바에 `대체 텍스트(Alt Text) 입력` 버튼을 강조 표시하여 시각장애인 스크린리더 접근성과 이미지 SEO를 동시에 유도한다.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- **EXIF GPS 유출 원천 차단**: `exif_stripped == true`로 검증된 파생 파일만 공개 CDN 버킷에 기록된다.
- **Hotlink 및 대역폭 남용 방어**: CDN WAF에서 비정상적인 외부 도메인 대량 이미지 핫링킹 시 Referer 정책 또는 Signed Cookie/Rate Limit으로 방어한다.

## 12. 성능 및 확장성 (Performance & Scalability)
- `libvips` 스트리밍 파이프라인을 사용하여 10MB JPEG → 4종 WebP/AVIF 파생본 변환을 평균 350ms 이내, 메모리 50MB 미만으로 처리한다.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- `github.com/aws/aws-sdk-go-v2/service/s3`, `github.com/davidbyttow/govips/v2` (libvips), `github.com/buckket/go-blurhash`.

## 14. 개발 우선순위 (Development Priority)
- **Phase 3 (에디터 이미지 업로드/최적화/동영상 임베드)** (`TB-MEDIA-001`, `TB-MEDIA-002`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 선행: `H_HOSTING_INFRASTRUCTURE` (S3/MinIO), `J_JOBS_SCHEDULING`.
- 후행: `C_CONTENT_EDITOR`.

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-MED-001`: GPS 위도/경도 EXIF가 포함된 원본 JPEG 사진 업로드 후 생성된 공개 WebP/JPEG 파생본의 EXIF 태그를 전수 파싱했을 때 GPS 태그가 100% 제거되었는지 검증.
- `TEST-MED-002`: 허용되지 않은 스크립트 도메인(`evil.example.com`)의 동영상 iframe 임베드 요청 시 `400 UNSUPPORTED_VIDEO_PROVIDER`로 거부되는지 검증.

## 17. Acceptance Criteria (인수 기준)
1. 이미지 업로드, 형식/크기 검증, EXIF 개인정보 제거, 다중 해상도 WebP/AVIF 변환, 외부 영상 안전 임베드, 고아 파일 정리가 모두 동작해야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- EXIF 제거 전/후 메타데이터 비교 테스트 로그, `media_assets`/`media_variants` 생성 레코드 확인.

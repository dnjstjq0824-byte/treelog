# VIDEO_EMBED.md — 외부 동영상 안전 임베드 및 자체 비디오 확장 아키텍처 (제13조)

## 1. 초기 단계: 허용 목록 기반 외부 영상 안전 임베드
- **지원 프로바이더**: YouTube (`youtube.com`, `youtu.be`), Vimeo (`vimeo.com`), Naver TV (`tv.naver.com`), Booking Network Global Video.
- **보안 샌드박스 `<iframe>` 렌더링**:
  - 사용자가 임의의 `<iframe>` HTML을 직접 입력하는 것은 금지하며, 오직 동영상 URL만 입력받아 서버가 프로바이더 ID를 추출한 뒤 아래와 같이 제한된 샌드박스 iframe으로 렌더링한다.
  ```html
  <iframe
    src="https://www.youtube-nocookie.com/embed/{VIDEO_ID}"
    sandbox="allow-scripts allow-same-origin allow-presentation"
    referrerpolicy="strict-origin-when-cross-origin"
    loading="lazy"
    allowfullscreen
    title="동영상 플레이어">
  </iframe>
  ```

## 2. 확장 단계: 자체 동영상 업로드 및 HLS 트랜스코딩 (`FF_NATIVE_VIDEO_TRANSCODE`)
- 비용 및 운영 복잡도를 고려하여 `feature_flags`의 `FF_NATIVE_VIDEO_TRANSCODE`가 활성화된 환경(또는 기업/유료 프로 블로거 플랜)에서는 원본 MP4/MOV 업로드 후 FFmpeg 워커가 `480p`, `720p`, `1080p` HLS(`.m3u8` + `.ts`/`.m4s` 세그먼트) 및 포스터 썸네일로 트랜스코딩하여 스트리밍한다.

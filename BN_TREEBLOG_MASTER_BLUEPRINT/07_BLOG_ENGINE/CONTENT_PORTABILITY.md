# CONTENT_PORTABILITY.md — 외부 블로그 이전(Import) 및 콘텐츠 백업(Export) 상세 설계 (제17조)

## 1. 콘텐츠 내보내기 (Export Pipeline)
블로거는 언제든 자신의 블로그 데이터를 종속(Lock-in) 없이 다음 3가지 표준 포맷의 ZIP 아카이브로 내보낼 수 있다.
1. **Markdown Bundle (`MARKDOWN_ZIP`)**:
   - 디렉터리 구조: `posts/{YYYY}/{MM}/{slug}.md` (YAML Frontmatter에 `title`, `date`, `updated`, `categories`, `tags`, `visibility`, `slug`, `canonical_url` 포함) + `media/{asset_id}.{ext}` (원본 이미지 파일) + `metadata/comments.json`.
2. **Standalone HTML Archive (`HTML_ZIP`)**: 오프라인 브라우저에서 즉시 열람 가능한 정적 HTML 파일 및 로컬 상대경로 이미지 폴더.
3. **WordPress Extended RSS (`WXR_XML`)**: WordPress, Ghost, Tistory 등 타 플랫폼과 호환되는 표준 WXR 1.2 XML + 미디어 번들.

## 2. 허용된 외부 블로그 가져오기 (Import Pipeline)
1. **법적·기술적 원칙**:
   - 타사 로그인 세션을 가로채거나 비공개 API를 해킹하는 수집기는 절대 제공하지 않는다.
   - 사용자가 직접 업로드한 공식 백업 파일(WordPress WXR `.xml`, Markdown `.zip`, 표준 RSS XML)만 처리하며, 실행 전 `본인이 저작권 또는 적법한 이전 권한을 보유한 콘텐츠임을 확인합니다`(`user_consents.consent_type = 'CONTENT_IMPORT_OWNERSHIP'`) 동의를 필수로 기록한다.
2. **외부 이미지 이관 및 URL 매핑**:
   - 백업 파일 내 본문에 포함된 공개 이미지 URL들을 `SafeDialer`로 다운로드하여 TreeBlog S3 스토리지에 새로 저장하고 본문 내 이미지 주소를 TreeBlog CDN 주소로 자동 치환한다.
   - 기존 글의 고유 경로(`legacy_slug`)를 `post_slug_history` 테이블에 자동 등록하여 개인 도메인을 그대로 연결할 경우 기존 검색 유입 링크가 `301` 리다이렉트로 완벽히 이어지도록 보장한다.

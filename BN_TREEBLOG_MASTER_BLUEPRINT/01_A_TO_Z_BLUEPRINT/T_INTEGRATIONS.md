# T_INTEGRATIONS — Third-party Integration & Content Portability (외부 연동 및 데이터 이동성)
> **부제**: 외부 인증, 동영상 임베드, 외부 공유, RSS, Atom, WebSub 검토, 공개 API, 콘텐츠 내보내기·가져오기
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-T_INTEGRATIONS`
> **적용 조항**: 제3조 (T_INTEGRATIONS 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- 사용자의 콘텐츠 주권(Data Ownership & Portability)을 완벽히 보장하여 언제든 자신의 글을 표준 포맷(Markdown, HTML, 원본 이미지 ZIP, WordPress WXR)으로 내보내거나 외부 블로그에서 적법하게 가져올 수 있도록 한다(제17조).
- RSS 2.0, Atom 1.0, WebSub(PubSubHubbub), oEmbed 동영상 임베드, 소셜 공유, 개발자용 공개 읽기/쓰기 API를 통해 글로벌 웹 생태계와 매끄럽게 연결된다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **이주 블로거**: 기존 WordPress 또는 Tistory/Medium에서 백업한 WXR XML 또는 Markdown ZIP 파일을 TreeBlog에 업로드하여 기존 글, 카테고리, 태그, 발행일, 이미지와 함께 이전하고 기존 URL 매핑(`old_url` → `new_url`) 가이드를 받는다.
- **전문 블로거**: 자신의 블로그 전체 글을 매월 Markdown + Frontmatter + 원본 이미지가 포함된 ZIP 아카이브로 다운로드하여 개인 Git 저장소에 백업한다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-INT-001 (콘텐츠 내보내기 — Export, 제17조)**:
  - 블로그 전체 또는 선택 게시글에 대해 1) **Markdown (`.md` + YAML Frontmatter)**, 2) **Stand-alone HTML**, 3) **WordPress 호환 WXR (XML)** 포맷과 **원본/최적화 이미지 파일 폴더** 및 **게시글 메타데이터 JSON(`posts_metadata.json`: 작성일, 수정일, 카테고리, 태그, 슬러그, 댓글 목록)**을 하나의 ZIP 파일로 생성하여 비동기 다운로드 링크(Signed URL, 24시간 유효)를 제공한다.
- **REQ-INT-002 (허용된 외부 블로그 가져오기 — Import, 제17조)**:
  - 지원 포맷: WordPress WXR (`.xml`), Markdown 번들 (`.zip` / `.md`), 표준 RSS/Atom 피드 XML.
  - **윤리적·법적 준수 원칙**: 타사 서비스의 비공개 세션을 우회하거나 약관을 위반하는 무단 크롤러는 절대 구현하지 않는다. 사용자는 본인이 소유권을 가진 콘텐츠임을 확인하고 명시적 동의(`user_consents`)를 체크해야 가져오기가 실행된다.
  - 외부 이미지 안전 수집: 가져온 본문 내 공개 이미지 URL은 `SafeDialer`(SSRF 방어)를 통해 다운로드하여 TreeBlog 스토리지로 이관하고 본문 내 이미지 경로를 자동 치환한다. 실패한 항목은 상세 오류 리포트(`migration_jobs.error_report_json`)로 제공한다.
  - 기존 URL 매핑: 원본 글의 경로(`legacy_path`)를 `post_slug_history`에 매핑하여 개인 도메인 이전 시 기존 유입 링크가 깨지지 않고 `301` 리다이렉트되도록 지원한다.
- **REQ-INT-003 (RSS 2.0 / Atom 1.0 / JSON Feed 및 WebSub)**:
  - 블로그별 `/rss.xml`, `/atom.xml`, `/feed.json` 및 카테고리별 피드를 제공하며, 피드 출력 범위(전문 공개 vs 요약 공개)를 블로거가 선택할 수 있다.
  - WebSub Hub 알림 핑(Ping) 인터페이스를 지원하여 피드 리더기에 새 글이 즉시 전파되도록 한다.
- **REQ-INT-004 (외부 공유 및 Personal Access Token 공개 API)**:
  - X(Twitter), Facebook, LinkedIn, KakaoTalk, Line, 클립보드 URL 복사 공유 지원.
  - 블로거가 CI/CD나 외부 마크다운 편집기(Obsidian 등)에서 글을 자동 발행할 수 있도록 스코프가 제한된 Personal Access Token(`tb_pat_...`) 기반 Open API를 제공한다.

## 4. 데이터 구조 (Data Structures & Domain Models)
```go
type MigrationJob struct {
    ID              uuid.UUID       `json:"id"`
    BlogID          uuid.UUID       `json:"blog_id"`
    InitiatedBy     uuid.UUID       `json:"initiated_by"`
    Direction       string          `json:"direction"`        // "IMPORT", "EXPORT"
    Format          string          `json:"format"`           // "MARKDOWN_ZIP", "HTML_ZIP", "WXR_XML"
    Status          string          `json:"status"`           // "QUEUED", "PROCESSING", "COMPLETED", "PARTIAL_SUCCESS", "FAILED"
    TotalItems      int             `json:"total_items"`
    ProcessedItems  int             `json:"processed_items"`
    FailedItems     int             `json:"failed_items"`
    ArchiveMediaID  *uuid.UUID      `json:"archive_media_id,omitempty"`
    ErrorReportJSON json.RawMessage `json:"error_report_json,omitempty"`
    OwnershipConfirmed bool         `json:"ownership_confirmed"`
    CreatedAt       time.Time       `json:"created_at"`
}
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- 블로그 관리 센터 내 `데이터 백업(내보내기) 및 외부 블로그 가져오기` 위저드 화면: 소유권 확인 동의 체크박스, 파일 업로드, 실시간 진행률 프로그레스 바, 완료 후 URL 매핑 테이블 및 실패 항목 상세 리포트 뷰어.

## 6. 백엔드 로직 (Backend Logic & Processing)
- `internal/modules/post/portability.go`:
  - WXR XML 스트리밍 파서(`encoding/xml` 토큰 디코더 — 대용량 XML OOM 방지).
  - ZIP Bomb 방어: 가져오기 압축 해제 시 최대 파일 수(10,000개) 및 압축 해제 총량(2GB)을 초과하면 즉시 중단한다.

## 7. API 계약 (API Contracts)
- `POST /api/v1/blogs/{blogId}/exports` / `GET /api/v1/blogs/{blogId}/exports/{jobId}` (블로그 내보내기 생성 및 다운로드)
- `POST /api/v1/blogs/{blogId}/imports` / `GET /api/v1/blogs/{blogId}/imports/{jobId}` (외부 블로그 파일 가져오기 및 오류 리포트 조회)
- `GET /api/v1/public/blogs/{blogKey}/rss.xml` / `GET /api/v1/public/blogs/{blogKey}/atom.xml`

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- 주요 테이블: `migration_jobs`, `post_slug_history`, `user_consents`.

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **부분 실패(Partial Success) 처리**: 100개의 글을 가져오는 중 2개의 외부 이미지 URL이 `404`로 만료된 경우, 전체 작업을 롤백하지 않고 98개는 완전 성공, 2개는 본문만 가져온 뒤 `PARTIAL_SUCCESS` 상태로 완료하고 깨진 이미지 위치 목록을 `error_report_json`에 정확히 보고한다.

## 10. UX/UI 구조 (UX/UI Structure)
- 타 플랫폼에서 개인 도메인을 그대로 유지하며 TreeBlog로 이주하는 사용자를 위해 `기존 URL 구조 유지 및 301 리다이렉트 설정 가이드`를 단계별 체크리스트로 안내한다.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- XML 파싱 시 XXE(XML External Entity) 공격을 원천 차단하기 위해 외부 엔터티 해석을 비활성화한다.
- 내보내기 ZIP에는 해당 블로그의 데이터만 포함되며, 비공개 게시글 포함 여부는 블로그 소유자(`OWNER`)만 선택할 수 있다.

## 12. 성능 및 확장성 (Performance & Scalability)
- 모든 Import/Export 작업은 API 서버가 아닌 `queue:bulk` 워커에서 스트리밍 방식으로 처리하여 500MB 규모의 아카이브 처리 시에도 메모리 사용량을 150MB 이내로 억제한다.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- Go 표준 `archive/zip`, `encoding/xml`, `github.com/gorilla/feeds`.

## 14. 개발 우선순위 (Development Priority)
- **Phase 5 (RSS/Atom)** 및 **Phase 7 (Import/Export 데이터 이동성)** (`TB-SEO-002`, `TB-POST-003`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 선행: `C_CONTENT_EDITOR`, `V_VIDEO_MEDIA`, `J_JOBS_SCHEDULING`.

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-INT-001`: 게시글 10개(이미지, 태그, 카테고리 포함)를 Markdown ZIP으로 Export한 뒤 새 블로그에 다시 Import했을 때 제목, 본문, 태그, 발행일이 100% 보존되는지 Round-trip 검증.
- `TEST-INT-002`: 악성 XXE 페이로드가 포함된 WXR XML 업로드 시 파서가 외부 파일 접근 없이 안전하게 거부하는지 검증.

## 17. Acceptance Criteria (인수 기준)
1. Markdown/HTML/WXR 내보내기와 가져오기, URL 매핑, 오류 리포트, RSS/Atom 피드가 모두 정상 동작해야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- Export ZIP 구조 리스트, Round-trip Import/Export 통합 테스트 로그, RSS 2.0 XML W3C 유효성 검증 결과.

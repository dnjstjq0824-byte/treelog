# O_OPENSEARCH — OpenSearch & Discovery (자체 블로그 검색엔진 및 탐색)
> **부제**: 게시글 검색, 블로그 검색, 태그 검색, 자동완성, 랭킹, 필터, 인기 글, 콘텐츠 추천
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-O_OPENSEARCH`
> **적용 조항**: 제3조 (O_OPENSEARCH 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- PostgreSQL을 원본 데이터 저장소로 유지하면서, **OpenSearch 2.x 기반의 초고속 다국어 전문(Full-text) 검색 및 디스커버리 엔진**을 구축하여 플랫폼 전체 검색과 개별 블로그 내 검색, 실시간 자동완성, 오타 교정, 다차원 패싯 필터링을 제공한다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **방문자 / 독자**: 검색창에 검색어를 입력하는 즉시 자동완성과 추천 태그를 확인하고, 결과를 `관련도순`, `최신순`, `인기순` 및 `기간(1일/1주/1개월/1년/직접입력)`, `언어`, `카테고리`, `작성자` 필터로 좁혀서 탐색한다.
- **검색 품질 관리자**: 관리자 콘솔(`SCR-ADM-005`)에서 색인 지연 상태, 실패한 색인 작업, 무결과 검색어(Zero-result queries), 스팸 검색어를 모니터링하고 동의어 사전을 갱신한다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-SRC-001 (검색 대상 및 3대 전용 인덱스 분리 — 제10조)**:
  1) `treeblog_posts_v1`: 공개 발행된(`status = 'PUBLISHED' AND visibility = 'PUBLIC' AND blog_status = 'ACTIVE'`) 게시글의 제목, 부제목, 본문 평문(`content_plain`), 요약, 태그, 카테고리명, 블로그명, 작성자 표시명, 언어, 품질 점수, 반응 통계.
  2) `treeblog_blogs_v1`: 활성 공개 블로그의 제목, 핸들, 소개글, 주요 태그, 구독자 수, 최근 활동일.
  3) `treeblog_suggest_v1`: 검색어 자동완성(Completion Suggester + Edge N-gram), 인기 태그, 오타 교정용 사전.
- **REQ-SRC-002 (다국어 형태소 분석 및 오타 대응)**:
  - 한국어 `nori` 분석기(복합명사 분해 `mixed`), 일본어 `kuromoji`, 다국어 `icu_analyzer`를 멀티 필드로 구성한다.
  - 오타 대응을 위해 `fuzziness: "AUTO"` 및 초성/자모 분리 검색(예: 한글 자모 보조 필드), `did_you_mean` 추천 검색어를 제공한다.
- **REQ-SRC-003 (정렬 및 필터링)**:
  - 정렬: `relevance`(관련도 + 품질/최신성 부스팅), `latest`(`published_at DESC`), `popular`(조회/공감/댓글 가중 합산).
  - 필터: 기간(`date_from`, `date_to`), 언어(`language_code`), 플랫폼 주제 카테고리(`topic_slug`), 특정 블로그 내 검색(`blog_id`), 작성자(`author_id`), 태그(`tag_slug`).
- **REQ-SRC-004 (비동기 이벤트 기반 색인 및 비공개 절대 차단)**:
  - PostgreSQL `outbox_events` 및 `search_index_jobs`를 통해 비동기 색인한다.
  - 게시글이 비공개/삭제/제재 상태로 변경되는 즉시 OpenSearch에서 문서를 삭제하며, 검색 결과 반환 직전 API 레이어에서 비공개 캐시 누출 여부를 2차 검증한다.

## 4. 데이터 구조 (Data Structures & Domain Models)
```json
{
  "index": "treeblog_posts_v1",
  "id": "018f2a10-7b11-7000-8000-000000000101",
  "body": {
    "post_id": "018f2a10-7b11-7000-8000-000000000101",
    "blog_id": "018f2a10-7b11-7000-8000-000000000010",
    "blog_handle": "tech-pioneer",
    "blog_title": "Tech Pioneer Blog",
    "author_id": "018f2a10-7b11-7000-8000-000000000001",
    "author_display_name": "김아키",
    "title": "Go와 OpenSearch로 구축하는 글로벌 블로그 검색 아키텍처",
    "summary": "PostgreSQL Transactional Outbox와 OpenSearch Nori 형태소 분석기를 결합한 실전 설계...",
    "content_plain": "본문 전체 평문 텍스트...",
    "tags": ["golang", "opensearch", "architecture"],
    "topic_slug": "it-tech",
    "language_code": "ko",
    "quality_score": 88.5,
    "popularity_score": 42.1,
    "view_count": 1520,
    "reaction_count": 64,
    "comment_count": 18,
    "published_at": "2026-10-27T12:00:00Z"
  }
}
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- 통합 검색 결과 페이지(`SCR-EXP-007`: 게시글 탭 / 블로그 탭 / 태그 탭, 기간·언어·정렬 필터 바, 검색어 하이라이팅 표시), 태그별 검색(`SCR-EXP-008`), 개별 블로그 내부 검색(`SCR-BLOG-011`).
- 헤더 검색창 포커스 시 디바운스(150ms) 자동완성 드롭다운 및 최근 검색어(LocalStorage) 제공.

## 6. 백엔드 로직 (Backend Logic & Processing)
- `internal/modules/search`:
  - `QueryBuilder`: `function_score` 쿼리를 사용하여 텍스트 BM25 점수(`title^4`, `tags^3`, `summary^2`, `content_plain^1`)에 `quality_score` 필드값 팩터와 `published_at` 가우시안 시간 감쇠(Gaussian Decay)를 결합하여 스팸 문서를 억제하고 고품질 최신 글을 상위에 노출한다.
  - Zero-Downtime Reindexing: 인덱스 매핑 변경 시 새 버전 인덱스(`treeblog_posts_v2`)를 생성하고 색인 완료 후 Alias(`treeblog_posts_read`)를 원자적으로 스위칭한다.

## 7. API 계약 (API Contracts)
- `GET /api/v1/search/posts` (전체 또는 `blog_id` 지정 게시글 검색)
- `GET /api/v1/search/blogs` (블로그 검색)
- `GET /api/v1/search/tags` (태그 검색)
- `GET /api/v1/search/suggest?q={prefix}` (자동완성 및 추천 검색어)
- `POST /api/v1/search/click` (검색 품질 측정용 클릭 위치/문서 비콘 수집)
- `POST /api/v1/admin/search/reindex` (관리자 전체/부분 재색인 트리거)

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- PostgreSQL 원본 테이블 외에 검색 운영을 위한 `search_index_jobs`, `search_queries`, `search_click_events`, `content_quality_signals` 테이블을 운영한다. 상세 스키마는 `09_SEARCH_SEO/OPENSEARCH_SCHEMA.md` 참조.

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **악성 키워드 반복(Keyword Stuffing) 예외**: 동일 단어가 비정상적으로 반복된 문서는 BM25의 포화 파라미터(`b=0.75, k1=1.2`) 조정과 `content_quality_signals.spam_probability` 페널티를 통해 검색 상단 노출을 차단한다.

## 10. UX/UI 구조 (UX/UI Structure)
- 검색 결과가 없을 때 단순 빈 화면 대신 오타 교정 제안(`'오픈서치'로 검색하시겠습니까?`)과 인기 토픽/추천 글을 함께 노출한다.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- **비공개 콘텐츠 색인 원천 금지 (제9조, 제10조, 제27조)**: 인덱서 워커는 문서를 OpenSearch에 전송하기 직전 반드시 `post.status == 'PUBLISHED' && post.visibility == 'PUBLIC' && blog.status == 'ACTIVE'` 조건을 하드 체크한다. 조건 불만족 시 색인 대신 즉시 `DELETE` 요청을 전송한다.

## 12. 성능 및 확장성 (Performance & Scalability)
- 검색 API 응답 시간 P95 < 60ms. 자동완성(`suggest`) API 응답 시간 P95 < 25ms.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- `github.com/opensearch-project/opensearch-go/v4`.

## 14. 개발 우선순위 (Development Priority)
- **Phase 5 (SEO 및 OpenSearch 구축)** (`TB-SEARCH-001`, `TB-SEARCH-002`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 선행: `C_CONTENT_EDITOR`, `J_JOBS_SCHEDULING`, `K_CONTENT_QUALITY`.
- 후행: `R_RECOMMENDATIONS`, `Z_ZERO_TRUST_TREESCAN`.

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-SRC-001`: 공개 게시글 발행 후 2초 이내에 한글 형태소 부분 일치 검색어로 검색 결과에 노출되고, 비공개(`PRIVATE`)로 전환 후 1초 이내에 검색 결과에서 완전히 제거되는지 검증.
- `TEST-SRC-002`: 자동완성, 기간 필터, 언어 필터, 블로그 내 검색, 최신순/인기순/관련도순 정렬이 모두 정확한 결과를 반환하는지 검증.

## 17. Acceptance Criteria (인수 기준)
1. PostgreSQL을 유일한 원본으로 유지하면서 OpenSearch를 통해 게시글·블로그·태그 검색, 자동완성, 다중 필터가 완벽히 동작해야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- OpenSearch 인덱스 매핑 생성 로그, 검색/필터/비공개 제거 통합 테스트 실행 로그.

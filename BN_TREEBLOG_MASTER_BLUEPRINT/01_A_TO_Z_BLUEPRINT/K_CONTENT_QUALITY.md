# K_CONTENT_QUALITY — Knowledge & Content Quality (콘텐츠 품질 및 신뢰도 엔진)
> **부제**: 콘텐츠 품질, 원문 출처, 중복 콘텐츠, 저작권, 스팸 탐지, 자동 생성 콘텐츠 관리, 콘텐츠 신뢰도
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-K_CONTENT_QUALITY`
> **적용 조항**: 제3조 (K_CONTENT_QUALITY 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- 저품질 어뷰징 문서, 키워드 스터핑(Keyword Stuffing), 무단 복제물, 대량 자동 생성 스팸이 플랫폼 검색과 추천 피드를 오염시키는 것을 방지한다.
- 양질의 원본 콘텐츠가 정당한 노출과 수익을 얻도록 **콘텐츠 품질 점수(`content_quality_signals`)**와 출처·저작권 표시 체계를 구축한다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **성실한 크리에이터**: 직접 작성한 심층 리뷰나 기술 아티클에 CCL(Creative Commons License) 및 원문 출처를 명시하여 높은 콘텐츠 품질 점수를 부여받고 토픽 추천 피드에 노출된다.
- **Trust & Safety 시스템 / 모더레이터**: 단시간에 수백 개의 도박/불법 홍보 글을 생성하는 스팸 봇을 SimHash 지문과 빈도 분석으로 즉시 감지하여 검색 색인에서 격리(`QUARANTINED`)한다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-QLT-001 (중복 콘텐츠 및 도용 탐지 — SimHash / MinHash)**:
  - 게시글 발행 시 `content_plain`에 대해 64-bit **SimHash**와 문단 단위 **MinHash** 지문을 계산하여 `content_quality_signals` 테이블에 저장한다.
  - 플랫폼 내 기존 공개 게시글과 해밍 거리(Hamming Distance) ≤ 3 이내인 거의 동일한 복제 글이 타 블로그에서 반복 게시될 경우 `is_near_duplicate = true`로 플래그하고 추천/통합검색 랭킹에서 감점(Demotion) 처리한다(단, 작성자 본인이 적법하게 외부에서 이전해 온 글이나 명시적 Canonical이 지정된 신디케이션 글은 예외).
- **REQ-QLT-002 (콘텐츠 품질 시그널 산출)**:
  - 본문 길이 및 구조 충실도(제목/소제목/문단/이미지 alt 유무), 독자 체류 시간 및 완독률(Read Completion Ratio), 외부 스팸 링크 비율, 과도한 광고/어필리에이트 링크 밀도를 종합하여 `quality_score` (`0.0` ~ `100.0`)를 산출한다.
- **REQ-QLT-003 (원문 출처·저작권·CCL 설정 및 AI 생성 표시)**:
  - 게시글 단위로 저작권 라이선스(`ALL_RIGHTS_RESERVED`, `CC_BY`, `CC_BY_NC`, `CC_BY_ND`, `CC_BY_SA`, `CC_BY_NC_SA`, `CC_BY_NC_ND`) 및 원문 참조 출처 URL 목록을 지정할 수 있다.
  - 자동 생성/AI 보조 콘텐츠 투명성 라벨(`ai_disclosure_level`: `NONE`, `AI_ASSISTED`, `AI_GENERATED`) 설정을 지원하며, 대량 API 매크로를 통한 무검수 자동 생성 스팸을 속도 제한 및 품질 필터로 억제한다.

## 4. 데이터 구조 (Data Structures & Domain Models)
```go
type ContentQualitySignal struct {
    PostID              uuid.UUID `json:"post_id"`
    SimHash64           int64     `json:"simhash_64"`
    WordCount           int       `json:"word_count"`
    ImageCount          int       `json:"image_count"`
    ExternalLinkCount   int       `json:"external_link_count"`
    QualityScore        float64   `json:"quality_score"`        // 0.0 ~ 100.0
    SpamProbability     float64   `json:"spam_probability"`     // 0.0 ~ 1.0
    IsNearDuplicate     bool      `json:"is_near_duplicate"`
    OriginalSourcePostID *uuid.UUID `json:"original_source_post_id,omitempty"`
    IndexEligibility    string    `json:"index_eligibility"`    // ELIGIBLE, NOINDEX_LOW_QUALITY, QUARANTINED_SPAM
}
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- 에디터 발행 설정 패널 내 저작권(CCL) 선택기, 원문 출처 URL 입력란, AI 활용 여부 체크박스 제공.
- 게시글 하단에 선택된 CCL 아이콘 및 저작권 보호 안내 문구 자동 렌더링.

## 6. 백엔드 로직 (Backend Logic & Processing)
- `internal/modules/search/quality.go` & `internal/modules/moderation/spam.go`:
  - 발행 이벤트(`post.published.v1`) 수신 시 비동기 워커가 200ms 이내에 규칙 기반 스코어링 + SimHash 중복 검사를 수행하고(선택적 Python AI 분류 워커 연동 가능), `spam_probability >= 0.85`인 경우 자동으로 `QUARANTINED_SPAM`으로 분류하여 검색 색인을 보류하고 모더레이션 큐(`moderation_cases`)에 등록한다.

## 7. API 계약 (API Contracts)
- `GET /api/v1/admin/posts/{postId}/quality-signals` (관리자 품질/스팸 시그널 상세 조회)
- `POST /api/v1/admin/posts/{postId}/quality-override` (오탐지 판정 시 관리자의 수동 정상 복구 및 색인 허용)

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- 주요 테이블: `content_quality_signals`, `spam_detection_results`, `user_reports`, `moderation_cases`.

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **자동 스팸 오탐지(False Positive) 구제**: 자동 필터에 의해 검색/추천 노출이 제한된 경우 작성자 관리 센터에 사유를 투명하게 통지하고 원클릭 '재검토 요청(Appeal)' 버튼을 제공하여 인적 검토(Human-in-the-loop)로 즉시 복구할 수 있도록 보장한다(제16조 준수).

## 10. UX/UI 구조 (UX/UI Structure)
- 불펌(무단 복제) 방지를 원하는 블로거를 위해 블로그 설정에서 '우클릭/드래그 방지 가이드 스크립트 및 복사 시 출처 URL 자동 첨부' 옵션을 제공한다.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- 외부 악성 피싱/멀웨어 도메인 블랙리스트(Google Safe Browsing / 자체 차단 DB)에 포함된 URL이 본문에 삽입될 경우 즉시 게시를 차단하거나 경고 인터스티셜 페이지를 거치도록 보호한다.

## 12. 성능 및 확장성 (Performance & Scalability)
- SimHash 비교는 PostgreSQL `bit_count(simhash_64 # $1)` 연산 및 버킷 인덱싱을 활용하여 수백만 건의 게시글 중에서도 10ms 이내에 유사 문서를 탐색한다.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- `github.com/mfonda/simhash`, 자체 토크나이저 및 휴리스틱 스코어러.

## 14. 개발 우선순위 (Development Priority)
- **Phase 5 (검색 품질 연동)** 및 **Phase 7 (스팸/모더레이션 고도화)** (`TB-SEARCH-002`, `TB-SEC-002`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 선행: `C_CONTENT_EDITOR`, `O_OPENSEARCH`.
- 후행: `R_RECOMMENDATIONS`, `W_WORKFLOW_ADMIN`.

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-QLT-001`: 기존 공개 게시글과 98% 동일한 본문을 타 계정 블로그에서 발행 시 `is_near_duplicate = true`로 판정되고 추천 피드 후보군에서 제외되는지 검증.
- `TEST-QLT-002`: 스팸 키워드와 차단 도메인 링크가 다수 포함된 글 발행 시 `QUARANTINED_SPAM`으로 분류되어 공개 검색 결과에 노출되지 않는지 검증.

## 17. Acceptance Criteria (인수 기준)
1. 모든 발행 게시글에 대해 `content_quality_signals` 레코드가 생성되고 OpenSearch 랭킹 가중치에 실제로 반영되어야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- SimHash 중복 탐지 단위/통합 테스트 로그 및 `content_quality_signals` 테이블 조회 결과.

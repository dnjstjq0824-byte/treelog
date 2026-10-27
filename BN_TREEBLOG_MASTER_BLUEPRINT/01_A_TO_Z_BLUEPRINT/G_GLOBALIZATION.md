# G_GLOBALIZATION — Globalization (글로벌 다국어·시간대·지역화)
> **부제**: 다국어, 시간대, 국가별 운영, 언어별 콘텐츠, 번역, 글로벌 SEO, 다국어 검색 및 접근성
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-G_GLOBALIZATION`
> **적용 조항**: 제3조 (G_GLOBALIZATION 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- 한국어, 영어, 일본어를 비롯한 다국어 사용자와 크리에이터가 언어 장벽 없이 서비스를 이용하고, 다국어 콘텐츠가 글로벌 검색엔진과 OpenSearch에서 정확하게 색인·검색되도록 글로벌라이제이션(i18n/l10n) 아키텍처를 수립한다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **글로벌 블로거**: 자신의 기본 인터페이스 언어를 영어(`en-US`)로 설정하고 시간대를 `America/New_York`으로 지정하여 현지 시간 기준으로 예약 발행을 설정한다.
- **다국어 블로거**: 한국어 원본 게시글과 영어 번역 게시글을 연결(`translation_group_id`)하여 발행하면, 검색엔진에 `<link rel="alternate" hreflang="ko">`와 `hreflang="en"`이 자동 출력된다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-GLB-001 (UI 다국어 지원)**: 초기 기본 지원 로케일은 `ko-KR`(한국어), `en-US`(영어), `ja-JP`(일본어) 3개 언어를 완벽 지원하며, JSON 메시지 카탈로그 추가만으로 확장 가능한 구조(`@nuxtjs/i18n` + Go i18n 번들)를 갖춘다. RTL(아랍어/히브리어) 확장을 위해 CSS 물리 속성(`margin-left`) 대신 논리 속성(`margin-inline-start`)을 사용한다.
- **REQ-GLB-002 (시간대 표준화)**: DB와 API의 모든 시각은 **UTC (`TIMESTAMPTZ`, ISO 8601 `2026-10-27T14:30:00Z`)**로 저장·전송한다. 예약 발행 및 통계 집계만 사용자/블로그의 설정 타임존(IANA Timezone, 예: `Asia/Seoul`, `UTC`, `America/Los_Angeles`)을 기준으로 변환하여 처리한다.
- **REQ-GLB-003 (언어별 콘텐츠 및 다국어 SEO)**:
  - 각 게시글은 작성 언어(`language_code`, BCP 47)를 가지며, 게시 시 자동 언어 감지(CLD3/whatlanggo) 또는 작성자 지정 값을 저장한다.
  - 동일 글의 다국어 번역본 묶음(`translation_group_id`)이 존재할 경우 SSR 헤더에 상호 `hreflang` 링크와 `x-default`를 자동 생성한다.
- **REQ-GLB-004 (다국어 형태소 검색)**: OpenSearch 인덱스에서 한국어(`nori`), 일본어(`kuromoji`), 영어/다국어(`icu_analyzer` + `ngram`) 멀티 필드를 분리 색인하여 한·영·일 혼합 검색어에서도 높은 재현율(Recall)을 보장한다.

## 4. 데이터 구조 (Data Structures & Domain Models)
```go
type PostLocalizationMeta struct {
    PostID             uuid.UUID            `json:"post_id"`
    LanguageCode       string               `json:"language_code"`        // "ko", "en", "ja"
    TranslationGroupID *uuid.UUID           `json:"translation_group_id"` // 다국어 번역본 연결 키
    Alternates         []HreflangAlternate  `json:"alternates"`
}
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- `@nuxtjs/i18n`을 통해 브라우저 `Accept-Language` 감지 및 사용자 프로필 설정 로케일 동기화.
- 날짜/시간/숫자/통화 포맷팅은 하드코딩된 문자열 포맷 대신 브라우저 표준 `Intl.DateTimeFormat`, `Intl.NumberFormat`, `Intl.RelativeTimeFormat` 래퍼 유틸리티(`packages/shared-utils/intl.ts`)를 사용한다.

## 6. 백엔드 로직 (Backend Logic & Processing)
- `internal/pkg/i18n`: API 에러 메시지, 이메일 템플릿, 푸시 알림 문구를 수신자의 `preferred_locale`에 맞춰 렌더링.
- 유니코드 슬러그(Slug) 지원: 한글/일본어 제목으로부터 URL 슬러그 생성 시 NFC 정규화를 거친 유니코드 슬러그 또는 로마자 음차(Transliteration) 옵션을 모두 지원하며, URL 인코딩/디코딩 시 깨짐이 없도록 처리한다.

## 7. API 계약 (API Contracts)
- 모든 API 요청은 `Accept-Language` 헤더를 인식하며, 검색 API(`GET /api/v1/search/posts?lang=ko`)에서 언어별 필터링을 지원한다.
- `PATCH /api/v1/me/preferences` (사용자 인터페이스 언어 및 타임존 변경)

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- `users.preferred_locale`, `users.timezone`, `blogs.default_language`, `blogs.timezone`, `posts.language_code`, `posts.translation_group_id` 컬럼 정의.
- 모든 문자열 컬럼은 UTF-8 인코딩을 사용하며, 이모지(4바이트 UTF-8)를 완벽히 수용한다.

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **서머타임(DST) 전환 예외**: `America/New_York` 등 일광절약시간제 적용 지역의 예약 발행 시간 계산 시, 단순 오프셋(`-05:00`)이 아닌 IANA Timezone DB(`time.LoadLocation`)를 사용하여 DST 전환일에도 정확한 현지 시각에 발행되도록 보장한다.

## 10. UX/UI 구조 (UX/UI Structure)
- 글로벌 네비게이션 바 우측 상단(및 푸터)에 직관적인 언어/지역 전환 드롭다운 제공.
- 번역본이 존재하는 게시글 상단에는 `Available in: 한국어 | English | 日本語` 언어 전환 배지를 표시한다.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- 유니코드 동형이의어(Homograph) 피싱 공격 방어: 블로그 핸들(`handle`)은 순수 ASCII 영숫자/하이픈으로만 제한하고, 개인 커스텀 도메인(IDN/Punycode `xn--`) 등록 시 UI에 Punycode 원문을 함께 표시하여 사칭을 방지한다.

## 12. 성능 및 확장성 (Performance & Scalability)
- 로케일 JSON 번들은 현재 선택된 언어만 동적 로드(Lazy Load)하여 초기 JS 번들 크기를 증가시키지 않는다.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- Backend: `github.com/abadojack/whatlanggo` (고속 자연어 감지), `golang.org/x/text` (유니코드 NFC 정규화 및 다국어 처리).
- Frontend: `@nuxtjs/i18n`.

## 14. 개발 우선순위 (Development Priority)
- **Phase 1 (기초 시간대/UTF-8 설계)** 및 **Phase 9 (다국어 UI/검색/글로벌 운영 고도화)** (`TB-OPS-003`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 선행: `F_FRONTEND`, `O_OPENSEARCH`.
- 후행: `M_MONETIZATION` (다중 통화 표기), `S_SECURITY` (국가별 컴플라이언스).

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-GLB-001`: 한글, 영어, 일본어, 아랍어, 복합 이모지(`👨‍👩‍👧‍👦`)가 포함된 게시글 작성·발행·검색·Markdown 내보내기 시 인코딩 깨짐이 0건인지 검증.
- `TEST-GLB-002`: `America/Los_Angeles` 타임존 사용자가 현지 시각 기준 내일 오전 9시로 예약 발행 설정 시 DB에 정확한 UTC 시각으로 변환 저장되는지 검증.

## 17. Acceptance Criteria (인수 기준)
1. UI 언어를 `ko-KR`, `en-US`, `ja-JP`로 전환할 때 하드코딩된 미번역 키 노출이 없어야 한다.
2. 다국어 연결 게시글에서 `<link rel="alternate" hreflang="...">` 태그가 정확히 렌더링되어야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- 3개 국어 전환 UI 스크린샷, 다국어 검색 쿼리 응답 로그, `hreflang` HTML 출력 스냅샷.

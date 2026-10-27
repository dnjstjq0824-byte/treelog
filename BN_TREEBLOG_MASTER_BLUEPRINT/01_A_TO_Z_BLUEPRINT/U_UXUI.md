# U_UXUI — UX/UI Design System & Screen Specifications (전체 화면 설계 및 디자인 시스템)
> **부제**: 네이버 블로그의 친숙함, 티스토리의 자유도, Blogger의 간편함을 참고한 고유 디자인 및 전체 화면 명세
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-U_UXUI`
> **적용 조항**: 제3조 (U_UXUI 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- 네이버 블로그의 친숙한 커뮤니티 동선, 티스토리의 깔끔한 콘텐츠 집중형 레이아웃, Google Blogger의 단순 명쾌한 관리 흐름을 결합하여 **TreeBlog만의 고유한 디자인 시스템(Forest & Paper Editorial Design System)**과 **제4조 및 제11조의 전체 45개 화면 명세**를 수립한다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **Frontend Developer Agent / UI Designer**: `05_UXUI/`의 7개 상세 문서에 정의된 컬러/타이포그래피/간격 토큰과 45개 화면의 11대 필수 속성(목적, UI 컴포넌트, 액션, API, 상태값, 로딩/빈상태/오류상태, 접근성, SEO, 테스트)을 그대로 코드로 옮긴다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-UX-001 (제4조 최소 화면 37개 + 관리자 콘솔 8개 = 총 45개 전체 화면 명세 완비)**:
  1) **블로그 탐색 (10개 화면, `SCR-EXP-001` ~ `010`)**: 플랫폼 메인, 인기 블로그, 인기 게시글, 최신 게시글, 카테고리별 블로그, 주제별 추천 콘텐츠, 검색 결과, 태그별 검색, 추천 피드, 구독 중인 블로그 피드.
  2) **개인 블로그 (12개 화면, `SCR-BLOG-001` ~ `012` + 에디터 `013`)**: 블로그 메인, 프로필, 전체 게시글, 카테고리 목록, 게시글 상세, 사진 및 동영상 게시물, 공지 게시글, 방명록, 블로그 소개, 이웃 및 구독 목록, 블로그 내 검색, 블로그 신고 (+ 글쓰기 에디터).
  3) **회원 기능 (15개 화면, `SCR-MEM-001` ~ `015`)**: 회원가입, 로그인, 비밀번호 복구, 프로필 설정, 블로그 개설, 내 블로그 관리, 공감·좋아요 모달/내역, 댓글 컴포넌트/관리, 답글, 북마크, 구독 관리, 이웃 관리, 알림 센터, 차단 관리, 계정 탈퇴.
  4) **관리자 콘솔 (8개 화면, `SCR-ADM-001` ~ `008`)**: 통합 대시보드, 회원 관리, 블로그 관리, 게시글 관리, 커뮤니티/신고 관리, 검색 품질 관리, 광고/정산 관리, 시스템/감사로그 관리.
- **REQ-UX-002 (화면별 11대 필수 설계 속성 강제)**:
  - 모든 45개 화면은 예외 없이 1) 페이지 목적, 2) UI 컴포넌트 트리, 3) 사용자 액션, 4) 연동 API, 5) 상태값(State), 6) 로딩 화면(Skeleton), 7) 빈 상태(Empty State), 8) 오류 상태(Error State), 9) 접근성(ARIA/키보드), 10) SEO 정책, 11) 테스트 항목을 포함한다.
- **REQ-UX-003 (320px ~ 대형 데스크톱 반응형 및 접근성)**:
  - 최소 뷰포트 너비 `320px`에서 가로 스크롤 깨짐이 없어야 하며, 모든 터치 타깃은 최소 `44x44px`, 본문 대비율은 WCAG 2.1 AA(`4.5:1` 이상, 다크모드 포함)를 만족한다.

## 4. 데이터 구조 (Data Structures & Domain Models)
```json
{
  "designTokens": {
    "colors": {
      "brandPrimary": "#15803D",
      "brandPrimaryHover": "#166534",
      "surfaceLight": "#FFFFFF",
      "surfacePaper": "#FAF9F6",
      "surfaceDark": "#0F172A",
      "textPrimaryLight": "#111827",
      "textSecondaryLight": "#4B5563"
    },
    "typography": {
      "fontSans": "'Pretendard Variable', 'Inter', 'Noto Sans JP', system-ui, sans-serif",
      "fontSerif": "'Noto Serif KR', 'Merriweather', Georgia, serif",
      "fontMono": "'JetBrains Mono', 'Fira Code', monospace"
    }
  }
}
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- `packages/ui-kit`: 디자인 토큰이 내장된 Tailwind CSS 프리셋 및 32개 공통 원자/분자 컴포넌트(`TbButton`, `TbInput`, `TbModal`, `TbDrawer`, `TbSkeleton`, `TbEmptyState`, `TbErrorState`, `TbPagination`, `TbReactionBar`, `TbCommentThread` 등).

## 6. 백엔드 로직 (Backend Logic & Processing)
- BFF/API는 각 화면이 단일 또는 최소한의 API 호출로 렌더링될 수 있도록 페이지별 최적화된 DTO 응답 구조를 제공한다.

## 7. API 계약 (API Contracts)
- 45개 화면과 122개 API 엔드포인트 간의 1:N 매핑은 `05_UXUI/SCREEN_INVENTORY.md`에 전수 정의됨.

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- 사용자별 UI 설정(`users.preferred_locale`, 테마 다크모드 설정) 및 블로그별 스킨 설정(`blog_themes.config_json`)과 연동.

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- 모든 화면은 데이터가 0건일 때 단순 공백 대신 **행동 유도 버튼(CTA)이 포함된 Empty State**(예: 구독 피드가 비어 있을 때 `'인기 블로그 둘러보기' 버튼 노출`)를 렌더링한다.

## 10. UX/UI 구조 (UX/UI Structure)
- 전체 45개 화면의 상세 레이아웃과 11대 속성은 `05_UXUI/USER_WEB.md`, `BLOG_MANAGEMENT.md`, `BLOG_EDITOR.md`, `ADMIN_CONSOLE.md`, `DESIGN_SYSTEM.md`, `SCREEN_INVENTORY.md`, `RESPONSIVE_ACCESSIBILITY.md` 7개 문서에 완비되어 있다.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- 신고 모달(`SCR-BLOG-012`)이나 차단 확인창에서는 실수 클릭 방지를 위한 명확한 확인 단계를 거치며, 관리자 콘솔 UI는 권한 없는 메뉴를 아예 렌더링하지 않는다(백엔드 RBAC와 이중 검증).

## 12. 성능 및 확장성 (Performance & Scalability)
- 폰트 최적화: `font-display: swap` 및 서브셋 WOFF2 프리로드 적용. 아이콘은 인라인 SVG 스프라이트를 사용하여 추가 네트워크 요청을 제거한다.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- `tailwindcss`, `radix-vue` (또는 `headlessui` — 키보드/스크린리더 접근성 완벽 지원 프리미티브), `lucide-vue-next`.

## 14. 개발 우선순위 (Development Priority)
- **Phase 1 ~ Phase 7 전 과정에 걸쳐 구현** (`TB-WEB-001` ~ `TB-WEB-004`, `TB-EDITOR-001`, `TB-ADMIN-001`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 선행: `F_FRONTEND`.
- 후행: E2E 및 시각적/접근성 테스트(`Q_QUALITY_ASSURANCE`).

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-UX-001`: 45개 전체 화면 라우트에 대해 `axe-core` 자동 접근성 검사 실행 시 Critical/Serious 위반 0건 확인.
- `TEST-UX-002`: API 강제 실패(500) 및 빈 배열(`[]`) 반환 시 모든 화면에서 Error State(재시도 버튼 포함)와 Empty State가 정상 노출되는지 검증.

## 17. Acceptance Criteria (인수 기준)
1. 제4조의 37개 사용자 화면과 제11조의 8개 관리자 화면(총 45개)이 모두 11대 필수 속성에 맞춰 구현 및 검증되어야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- `05_UXUI/*` 7개 상세 명세서, Playwright 45개 화면 렌더링 및 접근성 리포트.

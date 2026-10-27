# F_FRONTEND — Frontend Architecture (프론트엔드 아키텍처)
> **부제**: 사용자 웹, 블로그 관리 페이지, 글쓰기 에디터, 관리자 콘솔, SSR, 렌더링, 컴포넌트, 반응형 UX/UI
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-F_FRONTEND`
> **적용 조항**: 제3조 (F_FRONTEND 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- Vue 3 + Nuxt 3 + TypeScript 지정 기술 스택을 기반으로 **검색엔진 최적화(SSR/ISR)가 필수인 공개 블로그 웹**과 **복잡한 상태 관리가 필요한 에디터·관리 콘솔**을 최적의 렌더링 전략으로 설계한다.
- 320px 초소형 모바일 뷰포트부터 2560px+ 대형 데스크톱 모니터까지 일관된 반응형 UX와 WCAG 2.1 AA 웹 접근성을 보장한다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **독자**: 모바일 브라우저나 인앱 브라우저(카카오톡, 인스타그램 등)에서 블로그 링크를 클릭했을 때 1.5초 이내에 레이아웃 시프트(CLS) 없이 완성된 본문을 읽는다.
- **블로거 / 플랫폼 관리자**: 데스크톱과 태블릿/모바일을 오가며 에디터로 글을 작성하고 차트/데이터 테이블이 포함된 관리 콘솔을 조작한다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-FE-001 (하이브리드 렌더링 정책 — Nuxt 3 Route Rules)**:
  - 공개 블로그 상세/목록/프로필 및 플랫폼 탐색 메인: **SSR + SWR Edge Caching** (비로그인 공개 트래픽 고속 응답 및 완벽한 크롤러 색인).
  - 비공개/이웃공개/미리보기 경로: **Dynamic SSR (No-Store)** (서버에서 세션 쿠키를 검증하여 비인가 시 본문 미전송).
  - 블로그 관리 센터(`apps/blog-console`) 및 관리자 콘솔(`apps/admin-console`): **CSR/SPA with Auth Guard** (초기 인증 확인 후 풍부한 클라이언트 인터랙션 제공).
- **REQ-FE-002 (타입 안전 API 계층)**: `04_API/openapi.yaml`로부터 자동 생성된 TypeScript 인터페이스와 Fetch 래퍼(`packages/api-client`)를 사용하여 프론트엔드-백엔드 간 스키마 불일치를 컴파일 타임에 차단한다.
- **REQ-FE-003 (반응형 브레이크포인트 정책 — 제4조)**:
  - `xs`: 320px ~ 639px (소형/일반 스마트폰, 단일 컬럼, 하단 액션바/드로어 네비게이션)
  - `sm`: 640px ~ 767px (대형 스마트폰/폴더블)
  - `md`: 768px ~ 1023px (태블릿, 2컬럼 축약형 사이드바)
  - `lg`: 1024px ~ 1279px (노트북/표준 데스크톱, 본문 + 사이드바 + 목차 TOC)
  - `xl` / `2xl`: 1280px ~ 1920px+ (와이드 데스크톱, 최대 본문 가독 폭 `72ch` 유지 및 여백 균형 배치)

## 4. 데이터 구조 (Data Structures & Domain Models)
```typescript
// 프론트엔드 공통 비동기 상태 래퍼 (Loading / Empty / Error 3대 상태 강제 처리)
export interface AsyncViewState<T> {
  status: 'idle' | 'loading' | 'success' | 'empty' | 'error';
  data: T | null;
  error: {
    code: string;        // e.g., "POST_NOT_FOUND", "FORBIDDEN_MUTUAL_ONLY"
    message: string;
    requestId?: string;
    retryable: boolean;
  } | null;
}
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- **디렉터리 및 상태 관리**:
  - Pinia 스토어 분리: `useAuthStore`, `useBlogContextStore`, `useEditorStore`, `useThemeStore`, `useNotificationStore`.
  - 모든 페이지와 데이터 위젯은 반드시 `AsyncStateBoundary` 컴포넌트를 사용하여 **스켈레톤 로딩(Loading)**, **가이드 액션이 포함된 빈 상태(Empty State)**, **재시도 버튼 및 에러 코드가 표시되는 오류 상태(Error State)**를 빠짐없이 렌더링한다.

## 6. 백엔드 로직 (Backend Logic & Processing)
- Nuxt 3 Nitro 서버는 내부 Docker 네트워크(`http://treeblog-api:8080`)를 통해 Go 백엔드 API를 호출하여 SSR 시 루프백 네트워크 지연을 2ms 미만으로 단축한다.
- SSR 요청 시 클라이언트의 `Host`, `X-Forwarded-For`, `User-Agent`, `Accept-Language`, 인증 쿠키를 Go API로 안전하게 포워딩한다.

## 7. API 계약 (API Contracts)
- 프론트엔드 BFF(Nitro) 및 클라이언트 스토어는 `04_API/openapi.yaml`의 모든 엔드포인트를 소비한다.

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- 클라이언트 로컬 스토리지 구조:
  - `IndexedDB (treeblog_local_db)`: 에디터 오프라인 자동저장 스냅샷(`draft_snapshots`), 업로드 재개용 청크 메타데이터.
  - `Cookie`: `tb_access_token` (`HttpOnly`, `Secure`, `SameSite=Lax`), `tb_locale`, `tb_color_mode`.

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **토큰 만료(401) 자동 갱신**: API 호출 중 `401 TOKEN_EXPIRED` 수신 시 인터셉터가 진행 중인 요청을 큐에 대기시키고 `POST /api/v1/auth/refresh`를 1회만 호출(Mutex Lock)하여 토큰을 갱신한 뒤 실패했던 요청들을 자동 재시도한다.
- **ChunkLoadError 대응**: 신규 배포 직후 구 버전 JS 청크 요청 실패 시 현재 작성 중인 에디터 상태를 IndexedDB에 즉시 백업한 뒤 소프트 리로드를 수행한다.

## 10. UX/UI 구조 (UX/UI Structure)
- `05_UXUI/DESIGN_SYSTEM.md` 및 `SCREEN_INVENTORY.md`에 정의된 45개 전체 화면과 디자인 토큰을 1:1 구현한다.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- `v-html` 디렉티브는 오직 서버 Sanitizer를 통과한 게시글 본문 렌더러 컴포넌트 내부에서만 제한적으로 허용하며, 일반 댓글·프로필·제목 등에서는 절대 사용하지 않고 텍스트 보간(`{{ }}`)만 사용한다.
- 모든 상태 변경 요청(POST/PUT/PATCH/DELETE)에 Double-Submit CSRF 토큰 헤더(`X-CSRF-Token`)를 자동 첨부한다.

## 12. 성능 및 확장성 (Performance & Scalability)
- **Core Web Vitals 목표**: LCP < 2.0초, INP < 150ms, CLS < 0.05.
- 모든 이미지에 `width`/`height` 또는 `aspect-ratio`를 명시하여 레이아웃 시프트를 방지하고, 에디터/차트 등 무거운 라이브러리는 `defineAsyncComponent`로 코드 스플리팅(Lazy Hydration)한다.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- `nuxt@^3.11`, `vue@^3.4`, `pinia@^2.1`, `vue-router@^4.3`, `@vueuse/core@^10.9`, `@nuxtjs/i18n@^8.3`, `tailwindcss@^3.4`, `lucide-vue-next`, `zod`.

## 14. 개발 우선순위 (Development Priority)
- **Phase 1 (공통 UI/인증)** → **Phase 3 (에디터 UI)** → **Phase 4 (사용자 웹/테마)** → **Phase 7 (어드민 콘솔)** (`TB-WEB-001` ~ `TB-WEB-004`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 선행: `A_ARCHITECTURE`, `U_UXUI`.
- 후행: 모든 도메인 화면 구현 Work Order.

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-FE-001`: 320px(iPhone SE), 768px(iPad), 1440px(Desktop) 3개 뷰포트에서 가로 스크롤바 깨짐 현상(Overflow)이 0건인지 Playwright 자동 검사.
- `TEST-FE-002`: `curl`로 공개 게시글 URL 요청 시 JS 실행 없이도 초기 HTML 응답 내에 본문(`article`), `<title>`, `<meta property="og:title">`, `application/ld+json`이 완전히 포함되어 있는지 검증.

## 17. Acceptance Criteria (인수 기준)
1. `pnpm typecheck` (`vue-tsc --noEmit`) 및 `pnpm build`가 경고/에러 0건으로 통과해야 한다.
2. 모든 45개 화면에서 Loading, Empty, Error 상태 컴포넌트가 실제로 구현되어 있어야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- 프론트엔드 프로덕션 빌드 로그, Lighthouse CI 리포트(Performance/Accessibility/SEO ≥ 90점), 뷰포트별 스냅샷.

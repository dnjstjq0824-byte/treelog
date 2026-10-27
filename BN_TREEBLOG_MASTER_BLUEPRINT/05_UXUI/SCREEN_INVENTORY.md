# SCREEN_INVENTORY.md — 전체 45개 화면 마스터 인벤토리 (제4조 & 제11조 전수 매핑)

> **총 화면 수**: **46개 화면** (블로그 탐색 10개 + 개인 블로그 및 에디터 13개 + 회원 기능 15개 + 관리자 콘솔 8개 - 1개 통합 = 45개 고유 화면 ID)
> **필수 명세 속성 (11대 항목)**: 모든 화면은 1) 페이지 목적, 2) UI 컴포넌트, 3) 사용자 액션, 4) API, 5) 상태값, 6) 로딩 화면, 7) 빈 상태, 8) 오류 상태, 9) 접근성, 10) SEO, 11) 테스트 항목을 완비한다.

| 화면 ID | 분류 | 화면명 | 라우트 경로 | 핵심 연동 API | SEO 정책 | 검증 테스트 ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `SCR-EXP-001` | 블로그 탐색 | **블로그 플랫폼 메인** | `/` | `탭 전환(추천/구독/최신), 토픽 필터 클릭, 게시글/블로그 카드 클릭, 검색창 입력` | Skip-to-content 링크, 탭 `role=tablist` 키보드 화살표 이동 | `SSR, Canonical `/`, WebSite + SearchAction JSON-LD` |
| `SCR-EXP-002` | 블로그 탐색 | **인기 블로그** | `/explore/popular-blogs` | `분야별 필터링, 기간(주간/월간) 선택, 즉시 팔로우/이웃 추가 클릭` | 팔로우 버튼 `aria-pressed` 상태 명시 | `SSR, Canonical `/explore/popular-blogs`, ItemList JSON-LD` |
| `SCR-EXP-003` | 블로그 탐색 | **인기 게시글** | `/explore/trending` | `기간/주제 전환, 북마크 토글, 게시글 상세 이동` | 순위 번호 스크린리더 읽기(`aria-label='1위 게시글'`) | `SSR + SWR(60s), Canonical `/explore/trending`` |
| `SCR-EXP-004` | 블로그 탐색 | **최신 게시글** | `/explore/latest` | `언어 필터 변경, 무한 스크롤 로드` | 피드 `role=feed` 및 `aria-busy` 상태 전달 | `SSR, Canonical `/explore/latest`` |
| `SCR-EXP-005` | 블로그 탐색 | **카테고리별 블로그** | `/explore/categories/{topicSlug}` | `상위/하위 주제 전환, 정렬(인기순/최신순) 변경` | 현재 선택된 카테고리 `aria-current=page` | `SSR, CollectionPage JSON-LD` |
| `SCR-EXP-006` | 블로그 탐색 | **주제별 추천 콘텐츠** | `/explore/collections/{collectionSlug}` | `컬렉션 북마크, 개별 포스트 열람, 소셜 공유` | 배너 이미지 대체 텍스트 제공 | `SSR, OG 대형 커버 이미지 생성` |
| `SCR-EXP-007` | 블로그 탐색 | **검색 결과** | `/search` | `검색어 입력, 자동완성 선택, 필터/정렬 변경, 결과 클릭` | 자동완성 `role=combobox`, `aria-expanded`, 검색 결과 수 `aria-live=polite` 알림 | `SSR, 검색 결과 페이지 `robots: noindex, follow` (검색 스팸 색인 방지)` |
| `SCR-EXP-008` | 블로그 탐색 | **태그별 검색** | `/tags/{tagSlug}` | `정렬(인기/최신) 변경, 연관 태그 이동` | 태그 칩 키보드 포커스 링 보장 | `SSR, Canonical `/tags/{tagSlug}`` |
| `SCR-EXP-009` | 블로그 탐색 | **추천 피드** | `/feed/for-you` | `'관심 없음' 피드백 제출, 북마크 추가, 무한 스크롤` | 피드백 메뉴 키보드 Escape 닫기 지원 | `Dynamic SSR, `noindex` (개인화 화면)` |
| `SCR-EXP-010` | 블로그 탐색 | **구독 중인 블로그 피드** | `/feed/following` | `이웃 그룹별 필터링(전체/상호이웃/즐겨찾기), 새 글 열람` | 필터 변경 시 포커스 유지 및 스크린리더 안내 | `Dynamic SSR, `Cache-Control: private, no-store`, `noindex`` |
| `SCR-BLOG-001` | 개인 블로그 | **블로그 메인** | `/@{handle}` | `카테고리 선택, 글 열람, 이웃 추가/구독 클릭, 다크모드 토글` | 시맨틱 랜드마크(`header`, `nav`, `main`, `aside`, `footer`) 준수 | `SSR, 대표 도메인 Canonical, `Blog` JSON-LD, RSS/Atom `<link rel=alternate>`` |
| `SCR-BLOG-002` | 개인 블로그 | **블로그 프로필** | `/@{handle}/profile` | `필진별 작성 글 필터링, 팔로우/이웃 신청` | 소셜 아이콘 버튼에 명확한 `aria-label` 부여 | `SSR, `Person` / `Organization` JSON-LD` |
| `SCR-BLOG-003` | 개인 블로그 | **전체 게시글 목록** | `/@{handle}/posts` | `뷰 스타일 전환, 정렬 변경, 페이지 이동` | 뷰 전환 버튼 그룹 `aria-pressed` 제공 | `SSR, `rel=next/prev` 또는 Canonical 정규화` |
| `SCR-BLOG-004` | 개인 블로그 | **카테고리별 게시글 목록** | `/@{handle}/categories/{categorySlug}` | `하위 카테고리 필터링, 목록 열람` | Breadcrumb `aria-label='Breadcrumb'` 및 현재 위치 표시 | `SSR, `BreadcrumbList` JSON-LD` |
| `SCR-BLOG-005` | 개인 블로그 | **게시글 상세** | `/@{handle}/posts/{slug}` | `공감 6종 클릭, 북마크, 공유, 댓글/대댓글 작성, 유료글 구매/멤버십 가입, 게시글 신고` | 본문 제목 계층(H1->H2->H3) 엄격 준수, 코드블록 복사 알림, 이미지 `alt` 보장 | `SSR, `BlogPosting` + `BreadcrumbList` JSON-LD, Canonical, OpenGraph/Twitter Card, 비공개 글 `X-Robots-Tag: noindex`` |
| `SCR-BLOG-006` | 개인 블로그 | **사진 및 동영상 게시물 (미디어 뷰)** | `/@{handle}/media` | `미디어 썸네일 클릭 시 고해상도 라이트박스 열기, 원문 게시글 이동` | 라이트박스 모달 포커스 트랩 및 키보드 좌우 화살표/ESC 지원 | `SSR, `ImageObject` / `VideoObject` 메타 연동` |
| `SCR-BLOG-007` | 개인 블로그 | **공지 게시글 목록** | `/@{handle}/notices` | `공지사항 열람` | 공지 배지 스크린리더 텍스트 제공 | `SSR, Canonical `/@{handle}/notices`` |
| `SCR-BLOG-008` | 개인 블로그 | **방명록 (Guestbook)** | `/@{handle}/guestbook` | `방명록 작성, 비밀글 설정, 블로그 주인의 답글 작성/삭제` | 비밀글 체크박스와 입력창 `label` 명시적 연결 | `SSR, `robots: noindex` (방명록 스팸 색인 방지 옵션)` |
| `SCR-BLOG-009` | 개인 블로그 | **블로그 소개 (About)** | `/@{handle}/about` | `멤버 프로필 클릭, 후원하기 모달 열기, 뉴스레터 구독` | 명확한 섹션 헤딩 제공 | `SSR, `AboutPage` JSON-LD` |
| `SCR-BLOG-010` | 개인 블로그 | **이웃 및 구독 목록** | `/@{handle}/neighbors` | `이웃 블로그 방문, 맞팔로우/이웃 추가` | 목록 리스트 키보드 탐색 보장 | `SSR, 비공개 설정 시 소유자에게만 노출` |
| `SCR-BLOG-011` | 개인 블로그 | **블로그 내 검색** | `/@{handle}/search` | `블로그 내 키워드 검색, 카테고리 한정 필터링` | 검색 입력창 `role=search` 랜드마크 지정 | `SSR, `robots: noindex, follow`` |
| `SCR-BLOG-012` | 개인 블로그 | **블로그/게시글 신고 모달 및 페이지** | `/@{handle}/report` | `신고 사유 선택, 증빙 URL 첨부, 저작권 권리자 확인 체크, 신고 제출` | 오류 발생 시 `aria-invalid` 및 에러 메시지로 포커스 이동 | ``robots: noindex, nofollow`` |
| `SCR-BLOG-013` | 개인 블로그 (스튜디오) | **블로그 글쓰기 에디터** | `/studio/{blogId}/editor/{postId}` | `블록 작성/이동, 이미지/동영상/표/코드 삽입, 임시저장, 미리보기, 즉시/예약 발행, 버전 복원, 마크다운 입출력` | 툴바 `role=toolbar`, 단축키 안내(`Ctrl+B`, `/`), 이미지 삽입 시 `alt` 입력 유도 | `CSR/Auth Protected, `robots: noindex, nofollow`` |
| `SCR-MEM-001` | 회원 기능 | **회원가입** | `/auth/register` | `약관 동의, 이메일 가입 제출, 6자리 인증코드 확인` | 비밀번호 조건 및 오류 메시지 `aria-describedby` 연결 | `SSR, Canonical `/auth/register`` |
| `SCR-MEM-002` | 회원 기능 | **로그인** | `/auth/login` | `로그인 요청, TOTP 6자리 또는 백업코드 입력, 소셜 로그인` | 로그인 실패 시 스크린리더 `role=alert` 즉시 고지 | `SSR, Canonical `/auth/login`` |
| `SCR-MEM-003` | 회원 기능 | **비밀번호 복구** | `/auth/password-reset` | `이메일 입력 후 링크 발송, 토큰 검증 후 새 비밀번호 저장` | 단계 전환 시 상단 헤딩으로 포커스 이동 | `SSR, `robots: noindex`` |
| `SCR-MEM-004` | 회원 기능 | **프로필 및 계정 보안 설정** | `/settings/account` | `프로필 저장, TOTP QR 스캔 활성화, 타 기기 원격 로그아웃, 소셜/OIDC 연결` | QR 코드 외 텍스트 시크릿 키 복사 버튼 함께 제공 | ``robots: noindex`` |
| `SCR-MEM-005` | 회원 기능 | **블로그 개설 위저드** | `/studio/new-blog` | `핸들 타이핑(디바운스 검사), 유형/테마 선택, 개설 완료` | 핸들 가용 여부 `aria-live=polite` 실시간 안내 | ``robots: noindex`` |
| `SCR-MEM-006` | 회원 기능 | **내 블로그 관리 센터** | `/studio/{blogId}/dashboard` | `글 상태 일괄 변경, 테마 커스터마이징, DNS 검증, 수익 정산 신청, 백업 ZIP 다운로드` | 복합 관리 탭 및 모달 키보드 접근성 100% 보장 | ``robots: noindex, nofollow`` |
| `SCR-MEM-007` | 회원 기능 | **공감·좋아요 내역 및 리액션 모달** | `/settings/activity/reactions` | `공감 유형별 필터링, 공감 취소` | 6종 공감 아이콘 각각에 명확한 텍스트 라벨 제공 | ``robots: noindex`` |
| `SCR-MEM-008` | 회원 기능 | **댓글 작성 및 내 댓글 관리** | `/settings/activity/comments` | `댓글 작성/수정/삭제, 블로그 주인의 승인 대기 댓글 일괄 승인/차단` | 댓글 입력창 글자수 카운터 스크린리더 안내 | ``robots: noindex`` |
| `SCR-MEM-009` | 회원 기능 | **답글(대댓글) 스레드 및 멘션 뷰** | `/settings/activity/replies` | `문맥 확인 후 즉시 답글 작성, 원문 게시글 이동` | 원댓글-답글 계층 관계를 들여쓰기 및 `aria-level`로 표현 | ``robots: noindex`` |
| `SCR-MEM-010` | 회원 기능 | **북마크 보관함** | `/me/bookmarks` | `폴더 생성/이동, 북마크 해제, 이어 읽기 클릭` | 읽기 진행률 `role=progressbar` 및 퍼센트 표시 | ``robots: noindex`` |
| `SCR-MEM-011` | 회원 기능 | **구독 및 유료 멤버십 관리** | `/me/subscriptions` | `블로그별 이메일/푸시 알림 토글, 유료 멤버십 해지 예약, 영수증 조회/환불 요청` | 토글 스위치 `role=switch` 및 `aria-checked` 명시 | ``robots: noindex`` |
| `SCR-MEM-012` | 회원 기능 | **이웃 및 상호이웃(서로이웃) 관리** | `/me/neighbors` | `상호이웃 신청 일괄 수락/거절, 이웃 그룹 이동, 즐겨찾기 별표 토글` | 일괄 선택 체크박스 전체 선택/해제 키보드 지원 | ``robots: noindex`` |
| `SCR-MEM-013` | 회원 기능 | **알림 센터 및 알림 설정** | `/me/notifications` | `알림 클릭 이동, 모두 읽음 표시, 야간 방해금지 시간 설정, 웹 푸시 권한 허용` | 새 실시간 알림 도착 시 `aria-live=polite` 배지 갱신 | ``robots: noindex`` |
| `SCR-MEM-014` | 회원 기능 | **차단 사용자 관리** | `/settings/blocked-users` | `차단 해제 클릭, 차단 일시/사유 확인` | 차단 해제 확인 모달 포커스 관리 | ``robots: noindex`` |
| `SCR-MEM-015` | 회원 기능 | **계정 탈퇴 및 데이터 정리** | `/settings/deactivate` | `데이터 백업 다운로드, 블로그 처리 방식 선택, 비밀번호+MFA 확인 후 탈퇴 신청` | 위험 작업 경고를 `role=alert`로 명확히 전달 | ``robots: noindex`` |
| `SCR-ADM-001` | 관리자 콘솔 | **통합 운영 대시보드** | `/admin/dashboard` | `긴급 신고 큐 바로가기, 장애 알람 상세 이동` | 차트 데이터 대체 요약 테이블 함께 제공 | `Internal Admin Only, `noindex`` |
| `SCR-ADM-002` | 관리자 콘솔 | **회원 통합 관리** | `/admin/users` | `회원 검색, 일시/영구 정지 처분(사유 필수), 플랫폼 관리자 역할 부여/회수` | 데이터 테이블 정렬 상태 `aria-sort` 명시 | `Internal Admin Only, `noindex`` |
| `SCR-ADM-003` | 관리자 콘솔 | **블로그 및 도메인 관리** | `/admin/blogs` | `스팸 의심 블로그 정지, 악용된 커스텀 도메인 강제 연결 해제` | 상태 배지 색상 외 텍스트/아이콘 병행 표시 | `Internal Admin Only, `noindex`` |
| `SCR-ADM-004` | 관리자 콘솔 | **게시글 및 저작권 신고 관리** | `/admin/posts` | `원문 및 신고 증빙 대조, 임시 게시중단/비공개/복원 실행` | 단축키를 통한 고속 심사 네비게이션 지원 | `Internal Admin Only, `noindex`` |
| `SCR-ADM-005` | 관리자 콘솔 | **커뮤니티 및 이의신청 관리** | `/admin/community` | `스팸 댓글 일괄 삭제, 이의신청 소명자료 검토 후 복원 승인 또는 기각 결정` | 이의신청 결정 전 확인 다이얼로그 제공 | `Internal Admin Only, `noindex`` |
| `SCR-ADM-006` | 관리자 콘솔 | **검색 색인 및 품질 관리** | `/admin/search` | `실패한 색인 잡 재시도, 전체/부분 재색인 트리거, 스팸 검색어 등록` | 재색인 실행 시 확인 모달 및 진행률 표시 | `Internal Admin Only, `noindex`` |
| `SCR-ADM-007` | 관리자 콘솔 | **광고 심사 및 정산 관리** | `/admin/monetization` | `광고 소재 승인/반려, 부정 수익 환수(Clawback), 정산 출금 승인/지급보류` | 금액 및 통화 단위 명확한 포맷팅 | `Internal Admin Only, `noindex`` |
| `SCR-ADM-008` | 관리자 콘솔 | **시스템 설정 및 감사 로그** | `/admin/system` | `Feature Flag 토글, 점검 공지 등록, 감사 로그 해시 체인 무결성 검증 및 검색` | JSON Diff 뷰어 고대비 모드 지원 | `Internal Admin Only, `noindex`` |

---

## 2. 전 화면 공통 로딩(Loading) · 빈 상태(Empty) · 오류 상태(Error) 표준 정책

모든 45개 화면은 `AsyncStateBoundary` 래퍼를 통해 아래 3가지 비정상/전이 상태를 100% 일관되게 처리한다.

1. **로딩 화면 (Loading State)**:
   - 전체 화면을 가리는 스피너 대신 실제 레이아웃 높이와 동일한 **Skeleton Shimmer UI (`TbSkeleton`)**를 렌더링하여 CLS(Cumulative Layout Shift)를 `0.00`으로 유지한다.
   - 스크린리더를 위해 컨테이너에 `aria-busy="true"` 및 `aria-live="polite"`를 설정한다.
2. **빈 상태 (Empty State)**:
   - 데이터가 0건일 때 일러스트 아이콘 + 명확한 안내 문구 + **즉시 실행 가능한 주요 행동 유도 버튼(Primary CTA)**을 표시한다.
   - 예: `SCR-EXP-010` (구독 피드 0건) → *"아직 구독하거나 이웃으로 추가한 블로그가 없습니다."* + `[인기 블로그 둘러보기 (/explore/popular-blogs)]` 버튼.
   - 예: `SCR-BLOG-001` (신규 개설 블로그 글 0건) → 블로그 주인에게는 `[첫 게시글 작성하기]` 버튼 노출, 일반 방문자에게는 *"아직 발행된 게시글이 없습니다."* 노출.
3. **오류 상태 (Error State)**:
   - HTTP 상태 코드 및 `error.code`에 따라 맞춤형 복구 UI를 렌더링한다.
   - `403 FORBIDDEN_MUTUAL_ONLY`: *"이 게시글은 작성자와 상호이웃(서로이웃)에게만 공개되어 있습니다."* + `[상호이웃 신청하기]` 버튼.
   - `403 FORBIDDEN_NEIGHBOR_ONLY`: *"이웃에게만 공개된 게시글입니다."* + `[이웃 추가하기]` 버튼.
   - `451 CONTENT_QUARANTINED`: *"권리 침해 신고 또는 운영 정책에 따라 임시 게시 중단된 콘텐츠입니다."* (작성자 본인에게는 `[이의신청(소명) 제출하기]` 버튼 추가 노출).
   - `5xx` / 네트워크 오류: `error.request_id` 표시 및 `[다시 시도 (Retry)]` 버튼 제공.

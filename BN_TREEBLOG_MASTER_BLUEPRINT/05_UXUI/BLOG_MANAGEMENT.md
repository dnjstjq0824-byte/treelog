# BLOG_MANAGEMENT.md — 회원 기능 및 내 블로그 관리 센터 화면 상세 설계 (15개 화면)
> **부제**: 회원 인증·프로필·블로그 개설·블로그 스튜디오·이웃·구독·알림·북마크·차단·탈퇴(SCR-MEM-001~015) 상세 명세
> **적용 조항**: 제4조, 제11조 및 `01_A_TO_Z_BLUEPRINT/U_UXUI.md` (11대 필수 설계 항목 전수 포함)

---

## [SCR-MEM-001] 회원가입 (`/auth/register`)
1. **페이지 목적**: 이메일 또는 소셜 계정을 통한 신규 회원가입 및 약관 동의
2. **UI 컴포넌트 구조**: `RegisterForm(Email/Password/Username), PasswordStrengthMeter, TermsConsentCheckboxes, OAuthButtons, EmailOtpVerificationStep`
3. **사용자 액션**: 약관 동의, 이메일 가입 제출, 6자리 인증코드 확인
4. **연동 API**: `EP-AUTH-001, EP-AUTH-002, EP-AUTH-008`
5. **클라이언트 상태값 (State)**: `step, email, passwordStrength, consents, otpCode`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 비밀번호 조건 및 오류 메시지 `aria-describedby` 연결 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: SSR, Canonical `/auth/register`
11. **검증 테스트 항목**: `TEST-USR-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-MEM-002] 로그인 (`/auth/login`)
1. **페이지 목적**: 이메일/비밀번호 로그인, MFA 2단계 인증 입력 및 소셜/OIDC 로그인
2. **UI 컴포넌트 구조**: `LoginForm, MfaTotpChallengeModal, SocialOAuthButtons, ForgotPasswordLink`
3. **사용자 액션**: 로그인 요청, TOTP 6자리 또는 백업코드 입력, 소셜 로그인
4. **연동 API**: `EP-AUTH-003, EP-AUTH-008, EP-AUTH-009`
5. **클라이언트 상태값 (State)**: `email, mfaRequired, lockRemainingSec`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 로그인 실패 시 스크린리더 `role=alert` 즉시 고지 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: SSR, Canonical `/auth/login`
11. **검증 테스트 항목**: `TEST-USR-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-MEM-003] 비밀번호 복구 (`/auth/password-reset`)
1. **페이지 목적**: 분실한 비밀번호 재설정 이메일 요청 및 새 비밀번호 변경
2. **UI 컴포넌트 구조**: `RequestResetEmailForm, SetNewPasswordForm, PasswordStrengthMeter`
3. **사용자 액션**: 이메일 입력 후 링크 발송, 토큰 검증 후 새 비밀번호 저장
4. **연동 API**: `EP-AUTH-006, EP-AUTH-007`
5. **클라이언트 상태값 (State)**: `step(Request/Reset/Done), tokenValid`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 단계 전환 시 상단 헤딩으로 포커스 이동 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: SSR, `robots: noindex`
11. **검증 테스트 항목**: `TEST-USR-003` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-MEM-004] 프로필 및 계정 보안 설정 (`/settings/account`)
1. **페이지 목적**: 내 공개 프로필, 인터페이스 언어/타임존, 비밀번호 변경, MFA, 로그인 기기 세션, 통합 계정 연동 관리
2. **UI 컴포넌트 구조**: `ProfileEditForm, LocaleTimezoneSelect, MfaSetupQrModal, ActiveSessionsTable, LinkedIdentitiesCard(TreeScan OIDC 포함)`
3. **사용자 액션**: 프로필 저장, TOTP QR 스캔 활성화, 타 기기 원격 로그아웃, 소셜/OIDC 연결
4. **연동 API**: `EP-USER-001~005, EP-AUTH-009, EP-TSC-003`
5. **클라이언트 상태값 (State)**: `user, profile, sessions, mfaSetupData`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: QR 코드 외 텍스트 시크릿 키 복사 버튼 함께 제공 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: `robots: noindex`
11. **검증 테스트 항목**: `TEST-USR-004` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-MEM-005] 블로그 개설 위저드 (`/studio/new-blog`)
1. **페이지 목적**: 30초 만에 핸들 중복 확인, 블로그 유형(개인/전문/기업), 초기 테마를 선택하여 블로그를 개설하는 화면
2. **UI 컴포넌트 구조**: `HandleCheckInput(실시간 가용성 배지), BlogTypeSelector, InitialThemePicker, CreateBlogSubmitButton`
3. **사용자 액션**: 핸들 타이핑(디바운스 검사), 유형/테마 선택, 개설 완료
4. **연동 API**: `EP-BCRT-001, EP-BCRT-002, EP-THM-001`
5. **클라이언트 상태값 (State)**: `handle, handleAvailable, blogType, selectedThemeSlug`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 핸들 가용 여부 `aria-live=polite` 실시간 안내 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: `robots: noindex`
11. **검증 테스트 항목**: `TEST-BLOG-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-MEM-006] 내 블로그 관리 센터 (`/studio/{blogId}/dashboard`)
1. **페이지 목적**: 블로그 운영 전반(글/카테고리/테마 스튜디오/커스텀 도메인/팀 멤버/통계/수익화/데이터 백업·이전) 통합 콘솔
2. **UI 컴포넌트 구조**: `StudioSidebarNav, OverviewStatsCards, PostManagementTable, CategoryTreeEditor, ThemeVisualStudio, CustomDomainManager, TeamMemberManager, AnalyticsCharts, MonetizationLedgerView, PortabilityWizard`
3. **사용자 액션**: 글 상태 일괄 변경, 테마 커스터마이징, DNS 검증, 수익 정산 신청, 백업 ZIP 다운로드
4. **연동 API**: `EP-BMGT-001~007, EP-DOM-001~005, EP-THM-002~004, EP-ANL-002~005, EP-PAY-001~006, EP-PRT-001~004`
5. **클라이언트 상태값 (State)**: `activeMenu, blog, stats, domainList, themeDraft`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 복합 관리 탭 및 모달 키보드 접근성 100% 보장 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: `robots: noindex, nofollow`
11. **검증 테스트 항목**: `TEST-BLOG-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-MEM-007] 공감·좋아요 내역 및 리액션 모달 (`/settings/activity/reactions`)
1. **페이지 목적**: 내가 공감한 게시글 모아보기 및 내 글에 공감한 사용자 목록 확인
2. **UI 컴포넌트 구조**: `ReactionTypeFilterPills(6종), ReactedPostList, PostReactorsModal`
3. **사용자 액션**: 공감 유형별 필터링, 공감 취소
4. **연동 API**: `EP-RCT-001, EP-RCT-002`
5. **클라이언트 상태값 (State)**: `selectedReactionType, items, cursor`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 6종 공감 아이콘 각각에 명확한 텍스트 라벨 제공 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: `robots: noindex`
11. **검증 테스트 항목**: `TEST-COM-003` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-MEM-008] 댓글 작성 및 내 댓글 관리 (`/settings/activity/comments`)
1. **페이지 목적**: 게시글 상세 내 댓글 컴포넌트 및 내가 작성한/수신한 댓글 통합 관리함
2. **UI 컴포넌트 구조**: `CommentComposer(비밀글/멘션), CommentFilterTabs(승인대기/공개/비밀/스팸), BatchModerationToolbar`
3. **사용자 액션**: 댓글 작성/수정/삭제, 블로그 주인의 승인 대기 댓글 일괄 승인/차단
4. **연동 API**: `EP-CMT-001~004`
5. **클라이언트 상태값 (State)**: `filterStatus, comments, selectedIds`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 댓글 입력창 글자수 카운터 스크린리더 안내 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: `robots: noindex`
11. **검증 테스트 항목**: `TEST-COM-004` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-MEM-009] 답글(대댓글) 스레드 및 멘션 뷰 (`/settings/activity/replies`)
1. **페이지 목적**: 내 댓글에 달린 답글 및 나를 멘션(`@username`)한 대댓글 모아보기 및 빠른 답글 작성
2. **UI 컴포넌트 구조**: `ParentCommentContextCard, ReplyThreadList, QuickInlineReplyBox`
3. **사용자 액션**: 문맥 확인 후 즉시 답글 작성, 원문 게시글 이동
4. **연동 API**: `EP-CMT-001, EP-CMT-002`
5. **클라이언트 상태값 (State)**: `replyThreads, activeReplyTargetId`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 원댓글-답글 계층 관계를 들여쓰기 및 `aria-level`로 표현 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: `robots: noindex`
11. **검증 테스트 항목**: `TEST-COM-004` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-MEM-010] 북마크 보관함 (`/me/bookmarks`)
1. **페이지 목적**: 저장해 둔 게시글을 폴더별로 정리하고 읽던 위치(`read_progress_pct`)부터 이어 읽는 화면
2. **UI 컴포넌트 구조**: `BookmarkFolderSidebar, BookmarkSearchInput, SavedPostCardList(읽기 진행률 바 표시)`
3. **사용자 액션**: 폴더 생성/이동, 북마크 해제, 이어 읽기 클릭
4. **연동 API**: `EP-BMK-001~003`
5. **클라이언트 상태값 (State)**: `selectedFolder, bookmarks, cursor`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 읽기 진행률 `role=progressbar` 및 퍼센트 표시 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: `robots: noindex`
11. **검증 테스트 항목**: `TEST-COM-005` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-MEM-011] 구독 및 유료 멤버십 관리 (`/me/subscriptions`)
1. **페이지 목적**: 무료 알림/이메일 뉴스레터 구독 블로그 목록 및 유료 멤버십 정기결제/단건 구매 내역 관리
2. **UI 컴포넌트 구조**: `SubscriptionTabs(무료구독 / 유료멤버십 / 결제내역), ChannelToggleRow(인앱/푸시/이메일), CancelMembershipModal`
3. **사용자 액션**: 블로그별 이메일/푸시 알림 토글, 유료 멤버십 해지 예약, 영수증 조회/환불 요청
4. **연동 API**: `EP-FOL-005, EP-PAY-002, EP-PAY-004`
5. **클라이언트 상태값 (State)**: `activeTab, subscriptions, orders`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 토글 스위치 `role=switch` 및 `aria-checked` 명시 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: `robots: noindex`
11. **검증 테스트 항목**: `TEST-MON-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-MEM-012] 이웃 및 상호이웃(서로이웃) 관리 (`/me/neighbors`)
1. **페이지 목적**: 이웃 그룹 폴더 정리, 수신/발신 상호이웃 신청 수락·거절, 상호이웃 끊기 통합 관리
2. **UI 컴포넌트 구조**: `NeighborGroupEditor, MutualRequestInboxTable(수락/거절 버튼), NeighborListTable(그룹이동/즐겨찾기/해제)`
3. **사용자 액션**: 상호이웃 신청 일괄 수락/거절, 이웃 그룹 이동, 즐겨찾기 별표 토글
4. **연동 API**: `EP-FOL-001~004`
5. **클라이언트 상태값 (State)**: `groups, pendingRequests, neighbors`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 일괄 선택 체크박스 전체 선택/해제 키보드 지원 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: `robots: noindex`
11. **검증 테스트 항목**: `TEST-COM-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-MEM-013] 알림 센터 및 알림 설정 (`/me/notifications`)
1. **페이지 목적**: Rollup 병합된 인앱 알림 확인, 전체 읽음 처리 및 채널별·방해금지 시간 설정
2. **UI 컴포넌트 구조**: `NotificationCategoryTabs, NotificationRollupList, QuietHoursConfigModal, PushPermissionPrompt`
3. **사용자 액션**: 알림 클릭 이동, 모두 읽음 표시, 야간 방해금지 시간 설정, 웹 푸시 권한 허용
4. **연동 API**: `EP-NOT-001~005`
5. **클라이언트 상태값 (State)**: `category, unreadCount, notifications, preferences`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 새 실시간 알림 도착 시 `aria-live=polite` 배지 갱신 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: `robots: noindex`
11. **검증 테스트 항목**: `TEST-NOT-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-MEM-014] 차단 사용자 관리 (`/settings/blocked-users`)
1. **페이지 목적**: 내가 차단한 계정 목록 확인, 차단 사유 메모 열람 및 차단 해제
2. **UI 컴포넌트 구조**: `BlockedUserListTable, UnblockConfirmModal`
3. **사용자 액션**: 차단 해제 클릭, 차단 일시/사유 확인
4. **연동 API**: `EP-MOD-003, EP-MOD-004`
5. **클라이언트 상태값 (State)**: `blockedUsers, cursor`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 차단 해제 확인 모달 포커스 관리 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: `robots: noindex`
11. **검증 테스트 항목**: `TEST-COM-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-MEM-015] 계정 탈퇴 및 데이터 정리 (`/settings/deactivate`)
1. **페이지 목적**: 소유 블로그 양도/폐쇄 확인, 미정산 수익금 안내, 데이터 백업 권장 후 계정 탈퇴(30일 유예) 신청
2. **UI 컴포넌트 구조**: `PreDeactivationChecklist(블로그/정산금 상태), BlogDispositionSelector, MfaAndPasswordConfirmForm`
3. **사용자 액션**: 데이터 백업 다운로드, 블로그 처리 방식 선택, 비밀번호+MFA 확인 후 탈퇴 신청
4. **연동 API**: `EP-USER-006, EP-PRT-001`
5. **클라이언트 상태값 (State)**: `ownedBlogs, unpaidBalance, confirmationText`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 위험 작업 경고를 `role=alert`로 명확히 전달 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: `robots: noindex`
11. **검증 테스트 항목**: `TEST-USR-006` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

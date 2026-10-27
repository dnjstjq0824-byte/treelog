# USER_WEB.md — 블로그 탐색 및 개인 블로그 사용자 웹 화면 상세 설계 (22개 화면)
> **부제**: 블로그 탐색(SCR-EXP-001~010) 및 개인 블로그(SCR-BLOG-001~012) 전체 화면 11대 속성 명세
> **적용 조항**: 제4조, 제11조 및 `01_A_TO_Z_BLUEPRINT/U_UXUI.md` (11대 필수 설계 항목 전수 포함)

---

## [SCR-EXP-001] 블로그 플랫폼 메인 (`/`)
1. **페이지 목적**: TreeBlog 플랫폼의 첫인상 및 인기·추천·구독 피드 허브
2. **UI 컴포넌트 구조**: `GlobalHeader, HeroTopicBanner, FeedTabSwitcher, TrendingPostGrid, PopularBlogCarousel, GlobalFooter`
3. **사용자 액션**: 탭 전환(추천/구독/최신), 토픽 필터 클릭, 게시글/블로그 카드 클릭, 검색창 입력
4. **연동 API**: `EP-REC-002, EP-REC-003, EP-REC-004`
5. **클라이언트 상태값 (State)**: `activeTab, selectedTopic, feedCursor, isLoading`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: Skip-to-content 링크, 탭 `role=tablist` 키보드 화살표 이동 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: SSR, Canonical `/`, WebSite + SearchAction JSON-LD
11. **검증 테스트 항목**: `TEST-UX-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-EXP-002] 인기 블로그 (`/explore/popular-blogs`)
1. **페이지 목적**: 활동성·완독률·품질 기반 분야별 대표 블로그 디렉터리 탐색
2. **UI 컴포넌트 구조**: `TopicFilterBar, PopularBlogLeaderboard, FollowActionButton`
3. **사용자 액션**: 분야별 필터링, 기간(주간/월간) 선택, 즉시 팔로우/이웃 추가 클릭
4. **연동 API**: `EP-REC-004, EP-FOL-001`
5. **클라이언트 상태값 (State)**: `topicSlug, period, blogList, followStateMap`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 팔로우 버튼 `aria-pressed` 상태 명시 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: SSR, Canonical `/explore/popular-blogs`, ItemList JSON-LD
11. **검증 테스트 항목**: `TEST-REC-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-EXP-003] 인기 게시글 (`/explore/trending`)
1. **페이지 목적**: 실시간/일간/주간 가장 주목받는 고품질 공개 게시글 차트
2. **UI 컴포넌트 구조**: `PeriodSelector(24h/7d/30d), TopicPills, RankedPostCardList`
3. **사용자 액션**: 기간/주제 전환, 북마크 토글, 게시글 상세 이동
4. **연동 API**: `EP-REC-003, EP-BMK-001`
5. **클라이언트 상태값 (State)**: `period, topicSlug, rankedPosts, cursor`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 순위 번호 스크린리더 읽기(`aria-label='1위 게시글'`) (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: SSR + SWR(60s), Canonical `/explore/trending`
11. **검증 테스트 항목**: `TEST-REC-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-EXP-004] 최신 게시글 (`/explore/latest`)
1. **페이지 목적**: 플랫폼 전체에서 방금 발행된 스팸 검증 통과 공개 글 스트림
2. **UI 컴포넌트 구조**: `LanguageFilterDropdown, LatestPostStream, InfiniteScrollSentinel`
3. **사용자 액션**: 언어 필터 변경, 무한 스크롤 로드
4. **연동 API**: `EP-SRC-001`
5. **클라이언트 상태값 (State)**: `lang, posts, nextCursor`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 피드 `role=feed` 및 `aria-busy` 상태 전달 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: SSR, Canonical `/explore/latest`
11. **검증 테스트 항목**: `TEST-UX-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-EXP-005] 카테고리별 블로그 (`/explore/categories/{topicSlug}`)
1. **페이지 목적**: IT·테크, 여행, 맛집, 경제, 창작 등 표준 대분류별 블로그/글 탐색
2. **UI 컴포넌트 구조**: `CategoryHeroHeader, SubtopicTabs, FeaturedBlogsRow, TopicPostList`
3. **사용자 액션**: 상위/하위 주제 전환, 정렬(인기순/최신순) 변경
4. **연동 API**: `EP-REC-004, EP-SRC-001`
5. **클라이언트 상태값 (State)**: `topicSlug, sortBy, blogs, posts`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 현재 선택된 카테고리 `aria-current=page` (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: SSR, CollectionPage JSON-LD
11. **검증 테스트 항목**: `TEST-UX-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-EXP-006] 주제별 추천 콘텐츠 (`/explore/collections/{collectionSlug}`)
1. **페이지 목적**: 에디토리얼 큐레이션 및 시즌별 테마 추천 포스트 모음
2. **UI 컴포넌트 구조**: `CollectionCoverBanner, CuratedPostGrid, ShareCollectionButton`
3. **사용자 액션**: 컬렉션 북마크, 개별 포스트 열람, 소셜 공유
4. **연동 API**: `EP-REC-002`
5. **클라이언트 상태값 (State)**: `collection, posts`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 배너 이미지 대체 텍스트 제공 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: SSR, OG 대형 커버 이미지 생성
11. **검증 테스트 항목**: `TEST-UX-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-EXP-007] 검색 결과 (`/search`)
1. **페이지 목적**: 키워드 기반 게시글·블로그·태그 통합 검색 및 다차원 필터링
2. **UI 컴포넌트 구조**: `SearchAutocompleteBar, SearchResultTabs(Posts/Blogs/Tags), FacetFilterSidebar(기간/언어/카테고리/정렬), HighlightedResultList, DidYouMeanBanner`
3. **사용자 액션**: 검색어 입력, 자동완성 선택, 필터/정렬 변경, 결과 클릭
4. **연동 API**: `EP-SRC-001, EP-SRC-002, EP-SRC-004, EP-SRC-005`
5. **클라이언트 상태값 (State)**: `query, tab, sort, dateRange, lang, results, totalEstimate`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 자동완성 `role=combobox`, `aria-expanded`, 검색 결과 수 `aria-live=polite` 알림 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: SSR, 검색 결과 페이지 `robots: noindex, follow` (검색 스팸 색인 방지)
11. **검증 테스트 항목**: `TEST-SRC-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-EXP-008] 태그별 검색 (`/tags/{tagSlug}`)
1. **페이지 목적**: 특정 해시태그가 부착된 전체 공개 게시글 및 연관 태그 탐색
2. **UI 컴포넌트 구조**: `TagHeaderCard, RelatedTagsCloud, TaggedPostList`
3. **사용자 액션**: 정렬(인기/최신) 변경, 연관 태그 이동
4. **연동 API**: `EP-SRC-003, EP-CAT-005`
5. **클라이언트 상태값 (State)**: `tagInfo, relatedTags, posts, cursor`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 태그 칩 키보드 포커스 링 보장 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: SSR, Canonical `/tags/{tagSlug}`
11. **검증 테스트 항목**: `TEST-SRC-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-EXP-009] 추천 피드 (`/feed/for-you`)
1. **페이지 목적**: 독자의 열람·공감·관심 태그 이력에 맞춘 개인화 추천 스트림
2. **UI 컴포넌트 구조**: `InterestTopicOnboardingModal, PersonalizedFeedCard, FeedbackMenu(관심없음/주제숨기기)`
3. **사용자 액션**: '관심 없음' 피드백 제출, 북마크 추가, 무한 스크롤
4. **연동 API**: `EP-REC-002, EP-BMK-001`
5. **클라이언트 상태값 (State)**: `feedItems, dismissedPostIds, cursor`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 피드백 메뉴 키보드 Escape 닫기 지원 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: Dynamic SSR, `noindex` (개인화 화면)
11. **검증 테스트 항목**: `TEST-REC-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-EXP-010] 구독 중인 블로그 피드 (`/feed/following`)
1. **페이지 목적**: 내가 팔로우·이웃·상호이웃·구독한 블로그들의 새 글 타임라인
2. **UI 컴포넌트 구조**: `NeighborGroupFilterTabs, UnreadBadgeFilter, FollowingFeedStream, EmptySubscriptionCTA`
3. **사용자 액션**: 이웃 그룹별 필터링(전체/상호이웃/즐겨찾기), 새 글 열람
4. **연동 API**: `EP-REC-001, EP-FOL-003`
5. **클라이언트 상태값 (State)**: `selectedGroupId, onlyMutual, posts, cursor`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 필터 변경 시 포커스 유지 및 스크린리더 안내 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: Dynamic SSR, `Cache-Control: private, no-store`, `noindex`
11. **검증 테스트 항목**: `TEST-REC-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-BLOG-001] 블로그 메인 (`/@{handle}`)
1. **페이지 목적**: 개별 블로그의 대표 홈 화면 (공지, 대표글, 최신글, 프로필, 위젯)
2. **UI 컴포넌트 구조**: `BlogCustomHeader, NoticePinnedBar, FeaturedPostsHero, PostListOrGrid, BlogSidebar(Profile/Categories/Tags/Neighbors/Search), BlogCustomFooter`
3. **사용자 액션**: 카테고리 선택, 글 열람, 이웃 추가/구독 클릭, 다크모드 토글
4. **연동 API**: `EP-BMGT-001, EP-PUB-006, EP-CAT-001, EP-THM-002`
5. **클라이언트 상태값 (State)**: `blog, themeConfig, pinnedPosts, recentPosts, categories`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 시맨틱 랜드마크(`header`, `nav`, `main`, `aside`, `footer`) 준수 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: SSR, 대표 도메인 Canonical, `Blog` JSON-LD, RSS/Atom `<link rel=alternate>`
11. **검증 테스트 항목**: `TEST-BLOG-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-BLOG-002] 블로그 프로필 (`/@{handle}/profile`)
1. **페이지 목적**: 블로그 운영자/필진 상세 프로필 및 작성 글·활동 요약
2. **UI 컴포넌트 구조**: `AuthorProfileCard, SocialLinksBar, AuthorStatsSummary, AuthorPostsTab`
3. **사용자 액션**: 필진별 작성 글 필터링, 팔로우/이웃 신청
4. **연동 API**: `EP-BMGT-001, EP-PUB-006`
5. **클라이언트 상태값 (State)**: `profile, stats, authorPosts`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 소셜 아이콘 버튼에 명확한 `aria-label` 부여 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: SSR, `Person` / `Organization` JSON-LD
11. **검증 테스트 항목**: `TEST-UX-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-BLOG-003] 전체 게시글 목록 (`/@{handle}/posts`)
1. **페이지 목적**: 블로그 내 열람 가능한 전체 게시글 목록 (카드/리스트/갤러리/목록형)
2. **UI 컴포넌트 구조**: `ViewStyleSwitcher(Card/List/Compact), SortDropdown, PostCollectionView, CursorPagination`
3. **사용자 액션**: 뷰 스타일 전환, 정렬 변경, 페이지 이동
4. **연동 API**: `EP-PUB-006`
5. **클라이언트 상태값 (State)**: `viewStyle, sort, posts, cursor`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 뷰 전환 버튼 그룹 `aria-pressed` 제공 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: SSR, `rel=next/prev` 또는 Canonical 정규화
11. **검증 테스트 항목**: `TEST-PUB-003` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-BLOG-004] 카테고리별 게시글 목록 (`/@{handle}/categories/{categorySlug}`)
1. **페이지 목적**: 특정 상위/하위 카테고리에 속한 게시글 목록 및 시리즈 탐색
2. **UI 컴포넌트 구조**: `CategoryBreadcrumb, SubcategoryPills, CategoryDescriptionBanner, CategoryPostList`
3. **사용자 액션**: 하위 카테고리 필터링, 목록 열람
4. **연동 API**: `EP-CAT-001, EP-PUB-006`
5. **클라이언트 상태값 (State)**: `currentCategory, subcategories, posts`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: Breadcrumb `aria-label='Breadcrumb'` 및 현재 위치 표시 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: SSR, `BreadcrumbList` JSON-LD
11. **검증 테스트 항목**: `TEST-BLOG-003` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-BLOG-005] 게시글 상세 (`/@{handle}/posts/{slug}`)
1. **페이지 목적**: 블로그 게시글 본문 열람, 목차(TOC), 공감, 댓글, 관련글, 수익화(광고/페이월) 핵심 화면
2. **UI 컴포넌트 구조**: `PostArticleHeader, VisibilityBadge, TableOfContentsSticky, SafePostHtmlRenderer, PaywallGateCard, CCLicenseNotice, AuthorBioBox, ReactionBar6Types, ShareToolbar, CommentThreadSection, RelatedPostsGrid, AdPlacementSlot`
3. **사용자 액션**: 공감 6종 클릭, 북마크, 공유, 댓글/대댓글 작성, 유료글 구매/멤버십 가입, 게시글 신고
4. **연동 API**: `EP-PUB-007, EP-CMT-001, EP-RCT-001, EP-REC-005, EP-ADS-004, EP-ANL-001`
5. **클라이언트 상태값 (State)**: `post, viewerRelation, reactionState, comments, relatedPosts`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 본문 제목 계층(H1->H2->H3) 엄격 준수, 코드블록 복사 알림, 이미지 `alt` 보장 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: SSR, `BlogPosting` + `BreadcrumbList` JSON-LD, Canonical, OpenGraph/Twitter Card, 비공개 글 `X-Robots-Tag: noindex`
11. **검증 테스트 항목**: `TEST-COM-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-BLOG-006] 사진 및 동영상 게시물 (미디어 뷰) (`/@{handle}/media`)
1. **페이지 목적**: 블로그 내 이미지·갤러리·동영상 중심 게시물 모아보기 및 라이트박스 뷰
2. **UI 컴포넌트 구조**: `MediaTypeFilter(All/Photos/Videos), MasonryMediaGrid, LightboxModalViewer`
3. **사용자 액션**: 미디어 썸네일 클릭 시 고해상도 라이트박스 열기, 원문 게시글 이동
4. **연동 API**: `EP-PUB-006`
5. **클라이언트 상태값 (State)**: `mediaFilter, mediaPosts, activeLightboxItem`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 라이트박스 모달 포커스 트랩 및 키보드 좌우 화살표/ESC 지원 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: SSR, `ImageObject` / `VideoObject` 메타 연동
11. **검증 테스트 항목**: `TEST-MED-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-BLOG-007] 공지 게시글 목록 (`/@{handle}/notices`)
1. **페이지 목적**: 블로그 운영자가 고정한 중요 공지사항 모아보기
2. **UI 컴포넌트 구조**: `NoticeBadgeList, NoticeArticlePreview`
3. **사용자 액션**: 공지사항 열람
4. **연동 API**: `EP-PUB-006`
5. **클라이언트 상태값 (State)**: `notices`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 공지 배지 스크린리더 텍스트 제공 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: SSR, Canonical `/@{handle}/notices`
11. **검증 테스트 항목**: `TEST-BLOG-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-BLOG-008] 방명록 (Guestbook) (`/@{handle}/guestbook`)
1. **페이지 목적**: 네이버 블로그 감성의 방문자 안부 남기기 및 주인 답글 소통 공간
2. **UI 컴포넌트 구조**: `GuestbookWriteForm(비밀글 체크박스 포함), GuestbookEntryList, OwnerReplyForm`
3. **사용자 액션**: 방명록 작성, 비밀글 설정, 블로그 주인의 답글 작성/삭제
4. **연동 API**: `EP-CMT-005`
5. **클라이언트 상태값 (State)**: `entries, newEntryText, isSecret, cursor`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 비밀글 체크박스와 입력창 `label` 명시적 연결 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: SSR, `robots: noindex` (방명록 스팸 색인 방지 옵션)
11. **검증 테스트 항목**: `TEST-BLOG-005` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-BLOG-009] 블로그 소개 (About) (`/@{handle}/about`)
1. **페이지 목적**: 블로그의 운영 취지, 연혁, 대표 시리즈,팀 멤버, 연락처/후원 안내 페이지
2. **UI 컴포넌트 구조**: `BlogAboutHero,RichDescriptionSection, TeamMembersGrid, MilestoneStats, SponsorshipCTA`
3. **사용자 액션**: 멤버 프로필 클릭, 후원하기 모달 열기, 뉴스레터 구독
4. **연동 API**: `EP-BMGT-001, EP-BMGT-004`
5. **클라이언트 상태값 (State)**: `blogDetails, members`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 명확한 섹션 헤딩 제공 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: SSR, `AboutPage` JSON-LD
11. **검증 테스트 항목**: `TEST-UX-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-BLOG-010] 이웃 및 구독 목록 (`/@{handle}/neighbors`)
1. **페이지 목적**: 해당 블로그가 추가한 이웃 목록 및 이 블로그를 구독하는 이웃 현황(공개 설정 시)
2. **UI 컴포넌트 구조**: `NeighborDirectionTabs(내가 추가한 이웃 / 나를 추가한 이웃), NeighborBlogCardList`
3. **사용자 액션**: 이웃 블로그 방문, 맞팔로우/이웃 추가
4. **연동 API**: `EP-FOL-001`
5. **클라이언트 상태값 (State)**: `activeTab, neighborList, cursor`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 목록 리스트 키보드 탐색 보장 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: SSR, 비공개 설정 시 소유자에게만 노출
11. **검증 테스트 항목**: `TEST-COM-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-BLOG-011] 블로그 내 검색 (`/@{handle}/search`)
1. **페이지 목적**: 현재 블로그 내부의 게시글·카테고리·태그만 좁혀서 검색하는 화면
2. **UI 컴포넌트 구조**: `ScopedSearchInput, CategoryFilterSelect, ScopedSearchResultList`
3. **사용자 액션**: 블로그 내 키워드 검색, 카테고리 한정 필터링
4. **연동 API**: `EP-SRC-001`
5. **클라이언트 상태값 (State)**: `blogId, query, categoryId, results`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 검색 입력창 `role=search` 랜드마크 지정 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: SSR, `robots: noindex, follow`
11. **검증 테스트 항목**: `TEST-SRC-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-BLOG-012] 블로그/게시글 신고 모달 및 페이지 (`/@{handle}/report`)
1. **페이지 목적**: 스팸, 저작권(DMCA), 명예훼손, 사칭, 피싱 콘텐츠 신고 접수 화면
2. **UI 컴포넌트 구조**: `ReportReasonRadioGroup, CopyrightOwnerDeclarationBox, EvidenceUrlInputs, ReportDescriptionTextarea`
3. **사용자 액션**: 신고 사유 선택, 증빙 URL 첨부, 저작권 권리자 확인 체크, 신고 제출
4. **연동 API**: `EP-MOD-001`
5. **클라이언트 상태값 (State)**: `targetType, targetId, reason, description, evidenceUrls, isSubmitting`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 오류 발생 시 `aria-invalid` 및 에러 메시지로 포커스 이동 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: `robots: noindex, nofollow`
11. **검증 테스트 항목**: `TEST-ADM-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

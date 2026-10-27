# ADMIN_CONSOLE.md — 플랫폼 통합 관리자 콘솔 UX/UI 상세 설계 (SCR-ADM-001 ~ 008)
> **부제**: 제11조 필수 관리자 모듈(회원·블로그·게시글·커뮤니티·검색·광고·시스템) 8개 전체 화면 명세
> **적용 조항**: 제4조, 제11조 및 `01_A_TO_Z_BLUEPRINT/U_UXUI.md` (11대 필수 설계 항목 전수 포함)

---

## [SCR-ADM-001] 통합 운영 대시보드 (`/admin/dashboard`)
1. **페이지 목적**: 실시간 가입·발행·트래픽·신고 큐 적체·검색 색인 상태·광고 수익·시스템 알람 종합 상황판
2. **UI 컴포넌트 구조**: `AdminKpiMetricRow, ModerationQueueAlertCard, SystemHealthStatusGrid, RealtimeTrafficChart`
3. **사용자 액션**: 긴급 신고 큐 바로가기, 장애 알람 상세 이동
4. **연동 API**: `EP-ADM-001, EP-ADM-008`
5. **클라이언트 상태값 (State)**: `kpiSummary, queueCounts, systemHealth`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 차트 데이터 대체 요약 테이블 함께 제공 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: Internal Admin Only, `noindex`
11. **검증 테스트 항목**: `TEST-ADM-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-ADM-002] 회원 통합 관리 (`/admin/users`)
1. **페이지 목적**: 전체 회원 검색, 계정 상태, 가입 통계, 정지/해제, 신고 누적 계정 검토 및 플랫폼 RBAC 권한 부여
2. **UI 컴포넌트 구조**: `UserSearchFilterBar, AdminUsersDataTable, UserDetailDrawer, SuspendUserModal, RoleAssignmentModal`
3. **사용자 액션**: 회원 검색, 일시/영구 정지 처분(사유 필수), 플랫폼 관리자 역할 부여/회수
4. **연동 API**: `EP-ADM-001`
5. **클라이언트 상태값 (State)**: `searchParams, users, selectedUser`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 데이터 테이블 정렬 상태 `aria-sort` 명시 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: Internal Admin Only, `noindex`
11. **검증 테스트 항목**: `TEST-ADM-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-ADM-003] 블로그 및 도메인 관리 (`/admin/blogs`)
1. **페이지 목적**: 전체 블로그 목록, 생성·삭제 이력, 비정상 스팸 활동 감지, 커스텀 도메인 연결 상태 및 블로그 제재
2. **UI 컴포넌트 구조**: `BlogFilterBar, AdminBlogsDataTable, CustomDomainStatusInspector, BlogSanctionModal`
3. **사용자 액션**: 스팸 의심 블로그 정지, 악용된 커스텀 도메인 강제 연결 해제
4. **연동 API**: `EP-ADM-002`
5. **클라이언트 상태값 (State)**: `blogs, domainFilter, selectedBlog`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 상태 배지 색상 외 텍스트/아이콘 병행 표시 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: Internal Admin Only, `noindex`
11. **검증 테스트 항목**: `TEST-ADM-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-ADM-004] 게시글 및 저작권 신고 관리 (`/admin/posts`)
1. **페이지 목적**: 전체 게시글 검색, 일반 신고·저작권(DMCA)·명예훼손 신고 검토, 임시 게시중단(QUARANTINE), 복원 및 제재 이력
2. **UI 컴포넌트 구조**: `PostReportQueueTable, PostSnapshotComparisonViewer, QuarantineActionModal, SanctionHistoryTimeline`
3. **사용자 액션**: 원문 및 신고 증빙 대조, 임시 게시중단/비공개/복원 실행
4. **연동 API**: `EP-ADM-003, EP-ADM-004`
5. **클라이언트 상태값 (State)**: `cases, selectedPost, actionReason`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 단축키를 통한 고속 심사 네비게이션 지원 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: Internal Admin Only, `noindex`
11. **검증 테스트 항목**: `TEST-ADM-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-ADM-005] 커뮤니티 및 이의신청 관리 (`/admin/community`)
1. **페이지 목적**: 악성/스팸 댓글 일괄 관리, 유해 게시물 격리, 플랫폼 전역 차단 및 제재 이의신청(Appeals) 재심사
2. **UI 컴포넌트 구조**: `SpamCommentBatchTable, AppealsReviewQueue, AppealDecisionModal(승인복원/기각)`
3. **사용자 액션**: 스팸 댓글 일괄 삭제, 이의신청 소명자료 검토 후 복원 승인 또는 기각 결정
4. **연동 API**: `EP-ADM-004, EP-ADM-005`
5. **클라이언트 상태값 (State)**: `activeTab(Comments/Appeals), appeals, selectedAppeal`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 이의신청 결정 전 확인 다이얼로그 제공 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: Internal Admin Only, `noindex`
11. **검증 테스트 항목**: `TEST-ADM-002` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-ADM-006] 검색 색인 및 품질 관리 (`/admin/search`)
1. **페이지 목적**: OpenSearch 색인 상태, 색인 오류(DLQ) 재처리, 인기/무결과 검색어 통계, 스팸 검색어 차단
2. **UI 컴포넌트 구조**: `ClusterHealthCards, IndexJobDlqTable, TopAndZeroResultQueriesTable, SpamQueryBlocklistManager`
3. **사용자 액션**: 실패한 색인 잡 재시도, 전체/부분 재색인 트리거, 스팸 검색어 등록
4. **연동 API**: `EP-ADM-006`
5. **클라이언트 상태값 (State)**: `clusterStats, failedJobs, blockedTerms`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 재색인 실행 시 확인 모달 및 진행률 표시 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: Internal Admin Only, `noindex`
11. **검증 테스트 항목**: `TEST-ADM-003` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-ADM-007] 광고 심사 및 정산 관리 (`/admin/monetization`)
1. **페이지 목적**: 광고주·캠페인·광고 소재 승인/거절 심사, 노출·수익 통계, 부정클릭(IVT) 탐지 및 크리에이터 정산 지급 승인/보류
2. **UI 컴포넌트 구조**: `AdCreativeReviewQueue, IvtFraudAlertTable, CreatorPayoutApprovalTable, LedgerReconciliationCard`
3. **사용자 액션**: 광고 소재 승인/반려, 부정 수익 환수(Clawback), 정산 출금 승인/지급보류
4. **연동 API**: `EP-ADM-007`
5. **클라이언트 상태값 (State)**: `pendingCreatives, pendingPayouts, ivtAlerts`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: 금액 및 통화 단위 명확한 포맷팅 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: Internal Admin Only, `noindex`
11. **검증 테스트 항목**: `TEST-ADM-004` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

## [SCR-ADM-008] 시스템 설정 및 감사 로그 (`/admin/system`)
1. **페이지 목적**: 글로벌 서비스 설정, 런타임 기능 플래그(Feature Flags), 운영 공지, 백업 상태, 워커 큐/API 모니터링, 불변 감사 로그
2. **UI 컴포넌트 구조**: `FeatureFlagsToggleTable, SystemSettingsForm, BackupStatusPanel, WorkerQueueMonitor, ImmutableAuditLogViewer(JSON Diff)`
3. **사용자 액션**: Feature Flag 토글, 점검 공지 등록, 감사 로그 해시 체인 무결성 검증 및 검색
4. **연동 API**: `EP-ADM-008`
5. **클라이언트 상태값 (State)**: `featureFlags, auditLogs, workerMetrics, backupStatus`
6. **로딩 화면 (Loading State)**: 레이아웃 고정형 `TbSkeleton` 컴포넌트 렌더링 (`aria-busy="true"`, CLS < 0.05 보장)
7. **빈 상태 (Empty State)**: 조건에 맞는 항목이 0건일 때 상황별 안내 문구 및 다음 행동 유도 CTA 버튼(`TbEmptyState`) 노출
8. **오류 상태 (Error State)**: 권한 부족(`403` 이웃/상호이웃 안내 및 신청 버튼), 삭제/미존재(`404/410`), 일시 장애(`5xx` 재시도 버튼 및 `request_id` 표시) 처리
9. **접근성 (Accessibility)**: JSON Diff 뷰어 고대비 모드 지원 (WCAG 2.1 AA 명도 대비 ≥ 4.5:1, 최소 터치 영역 44x44px, 키보드 Tab/Enter/Escape 완비)
10. **SEO 정책**: Internal Admin Only, `noindex`
11. **검증 테스트 항목**: `TEST-ADM-001` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)

# W_WORKFLOW_ADMIN — Workflow & Administration (통합 관리자 콘솔 및 모더레이션 워크플로)
> **부제**: 관리자 운영, 콘텐츠 신고, 제재, 이의신청, 운영 대시보드, 광고 심사, 사용자·블로그 관리
> **문서 식별자**: `BN-TREEBLOG-BLUEPRINT-W_WORKFLOW_ADMIN`
> **적용 조항**: 제3조 (W_WORKFLOW_ADMIN 항목) 및 제25조 상호 일관성 검증 기준 준수

---

## 1. 기능의 목적 (Purpose)
- 플랫폼 운영진이 수백만 명의 회원, 블로그, 게시글, 댓글, 검색 품질, 광고 캠페인, 정산, 시스템 장애를 **실제 DB 및 백엔드와 100% 연결된 통합 관리자 콘솔(`apps/admin-console`)**에서 투명하고 안전하게 운영할 수 있도록 설계한다(제11조).
- **신고 접수 → 자동/수동 검토 → 제재 조치 → 당사자 통지 → 이의신청(Appeal) → 재심사 및 복원**으로 이어지는 표준화된 Trust & Safety 워크플로를 제공한다.

## 2. 사용자 역할 및 사용 시나리오 (User Roles & Scenarios)
- **Trust & Safety Moderator (`TRUST_SAFETY_MOD`)**: 저작권 침해 또는 명예훼손 신고가 접수된 게시글을 검토하여 '임시 게시 중단(`QUARANTINED`)' 조치를 내리고, 작성자가 제출한 소명 자료(이의신청)를 심사하여 복원 또는 영구 삭제를 결정한다.
- **Ad Reviewer (`AD_REVIEWER`)**: 광고주가 등록한 신규 캠페인 소재의 랜딩 페이지와 이미지를 검수하여 승인(`APPROVED`) 또는 반려 사유를 등록한다.
- **Super Admin (`SUPER_ADMIN`)**: 기능 플래그(`feature_flags`)를 실시간 토글하고, 관리자 권한 변경 및 전체 감사 로그(`audit_logs`)를 감사한다.

## 3. 상세 기능 요구사항 (Detailed Functional Requirements)
- **REQ-ADM-001 (제11조 7대 관리자 모듈 전수 백엔드/DB 직결 구현)**:
  1) **회원 관리 (`SCR-ADM-002`)**: 회원 통합 검색(이메일/이름/ID/상태), 계정 상태 조회, 일/주/월 가입 통계, 계정 일시정지/영구정지 및 해제, 다중 신고 누적 계정 검토, 플랫폼 RBAC 권한 부여/회수.
  2) **블로그 관리 (`SCR-ADM-003`)**: 전체 블로그 목록/필터, 블로그 상태(`ACTIVE`/`SUSPENDED`/`CLOSED`), 생성·삭제 이력, 비정상 스팸 트래픽 탐지 목록, 커스텀 도메인 DNS/TLS 연결 상태 점검 및 강제 해제, 블로그 제재.
  3) **게시글 관리 (`SCR-ADM-004`)**: 전체 게시글 검색, 일반 신고 및 저작권(DMCA)/명예훼손 신고 큐, 강제 비공개 처리, 임시 게시 중단(`QUARANTINED`), 정상 복원, 제재 이력 타임라인 조회.
  4) **커뮤니티 관리 (`SCR-ADM-005`)**: 댓글 통합 조회, 악성/욕설 댓글 및 스팸 댓글 일괄 삭제, 유해 게시물 격리, 악성 사용자 플랫폼 전역 차단, **이의신청(`appeals`) 심사 및 처리**.
  5) **검색 관리 (`SCR-ADM-006`)**: OpenSearch 클러스터 및 색인 큐 상태, 색인 오류(Dead Letter) 재처리, 인기/무결과 검색어 통계, 검색 품질 시그널 조회, 스팸 검색어 블랙리스트 및 불량 문서 색인 제외.
  6) **광고 및 수익 관리 (`SCR-ADM-007`)**: 광고주 목록, 캠페인 예산/상태, 광고 소재 승인·거절 심사 큐, 노출/클릭/CTR 성과, 수익 배분 및 정산 지급 승인/보류, 부정 클릭(IVT) 탐지 리포트.
  7) **시스템 관리 (`SCR-ADM-008`)**: 글로벌 서비스 설정(`system_settings`), 런타임 기능 플래그(`feature_flags`), 플랫폼 전체 운영 공지 발송, 관리자 계정/권한 관리, 불변 감사 로그(`audit_logs`) 검색, DB/스토리지 백업 상태, 워커 큐/장애 현황 및 API 응답 지연 모니터링.
- **REQ-ADM-002 (신고 → 검토 → 조치 → 이의신청 표준 상태 머신)**:
  - 동일 게시글/댓글에 대한 다수 사용자의 중복 신고는 단일 `moderation_cases` 케이스로 자동 병합(`report_count` 증가 및 우선순위 상승)된다.
  - 관리자가 제재 조치(`moderation_actions`)를 실행할 때는 반드시 **구체적인 정책 위반 조항 코드와 사유**를 입력해야 하며, 이는 즉시 `audit_logs`에 기록되고 당사자에게 이의신청 링크가 포함된 알림/이메일이 발송된다.

## 4. 데이터 구조 (Data Structures & Domain Models)
```go
type ModerationCase struct {
    ID             uuid.UUID  `json:"id"`
    TargetType     string     `json:"target_type"`     // POST, COMMENT, BLOG, USER, AD_CREATIVE
    TargetID       uuid.UUID  `json:"target_id"`
    ReasonCategory string     `json:"reason_category"` // SPAM, COPYRIGHT, DEFAMATION, ADULT, PHISHING, HATE_SPEECH
    ReportCount    int        `json:"report_count"`
    PriorityScore  int        `json:"priority_score"`  // 1 ~ 100
    Status         string     `json:"status"`          // OPEN, IN_REVIEW, ACTION_TAKEN, DISMISSED, APPEALED, APPEAL_RESOLVED
    AssignedModID  *uuid.UUID `json:"assigned_mod_id,omitempty"`
    CreatedAt      time.Time  `json:"created_at"`
}
```

## 5. 프론트엔드 아키텍처 (Frontend Architecture)
- `apps/admin-console`: 고밀도 데이터 그리드(정렬, 다중 필터, 일괄 작업, CSV 내보내기), 신고 대상 원문/스냅샷 비교 뷰어, 제재 사유 템플릿 선택기, 감사 로그 Diff 모달.

## 6. 백엔드 로직 (Backend Logic & Processing)
- `internal/modules/admin` & `internal/modules/moderation`:
  - 모든 `/api/v1/admin/*` 엔드포인트는 `RequireAdminRole` 및 `AuditLogMiddleware`를 강제 통과한다.
  - 게시글 게시 중단(`QUARANTINED`) 실행 시 단일 트랜잭션으로 `posts.status` 변경 + `moderation_actions` 기록 + `audit_logs` 기록 + `outbox_events` (검색 색인 즉시 삭제 및 CDN 캐시 퍼지)를 원자적으로 수행한다.

## 7. API 계약 (API Contracts)
- `GET /api/v1/admin/users` / `POST /api/v1/admin/users/{userId}/suspend` / `POST /api/v1/admin/users/{userId}/unsuspend` / `PUT /api/v1/admin/users/{userId}/roles`
- `GET /api/v1/admin/blogs` / `POST /api/v1/admin/blogs/{blogId}/suspend`
- `GET /api/v1/admin/posts` / `POST /api/v1/admin/posts/{postId}/quarantine` / `POST /api/v1/admin/posts/{postId}/restore`
- `GET /api/v1/admin/moderation/cases` / `POST /api/v1/admin/moderation/cases/{caseId}/decide`
- `GET /api/v1/admin/moderation/appeals` / `POST /api/v1/admin/moderation/appeals/{appealId}/resolve`
- `GET /api/v1/admin/ads/creatives/pending` / `POST /api/v1/admin/ads/creatives/{creativeId}/review`
- `GET /api/v1/admin/system/settings` / `PUT /api/v1/admin/system/feature-flags/{key}` / `GET /api/v1/admin/system/audit-logs`

## 8. DB 테이블 및 관계 (DB Tables & Relationships)
- 주요 테이블: `user_reports`, `moderation_cases`, `moderation_actions`, `appeals`, `spam_detection_results`, `audit_logs`, `feature_flags`, `system_settings`.

## 9. 상태 전이 및 예외 상황 (State Transitions & Edge Cases)
- **제재 중 작성자의 임의 삭제/은폐 방지**: 법적 분쟁이나 불법 촬영물/피싱 증거 보전이 필요한 `QUARANTINED` 상태의 게시글은 일반 공개에서는 즉시 차단되지만, 작성자가 임의로 영구 삭제하여 증거를 인멸하지 못하도록 수정/삭제를 잠그고 보존 스냅샷을 유지한다.

## 10. UX/UI 구조 (UX/UI Structure)
- 디자인만 있는 빈 껍데기 대시보드를 엄격히 금지하며(제11조), 모든 통계 카드와 테이블 행 클릭 시 실제 DB 상세 레코드와 액션 모달이 동작한다.

## 11. 보안·개인정보 및 권한 관리 (Security, Privacy & RBAC)
- 관리자 권한 분리(Separation of Duties): `SUPPORT_AGENT`는 조회 및 케이스 에스컬레이션만 가능, `TRUST_SAFETY_MOD`는 콘텐츠/계정 제재 가능하나 금전 정산 승인 불가, `FINANCE_ADMIN`만 정산 지급 승인 가능, `SUPER_ADMIN`만 관리자 권한 부여 가능.

## 12. 성능 및 확장성 (Performance & Scalability)
- 관리자 콘솔의 대량 검색 쿼리는 전용 복합 인덱스 및 Read Replica를 사용하여 프로덕션 사용자 트래픽에 영향을 주지 않도록 격리한다.

## 13. 필요한 외부 라이브러리 (Required External Libraries)
- `@tanstack/vue-table` (고성능 관리자 데이터 그리드), `echarts` (운영/검색/광고 모니터링 차트).

## 14. 개발 우선순위 (Development Priority)
- **Phase 7 (관리자·보안·콘텐츠 운영)** (`TB-ADMIN-001` ~ `TB-ADMIN-003`).

## 15. 모듈 간 의존성 (Inter-Module Dependencies)
- 선행: `I_IDENTITY_AUTHORIZATION`, `B_BLOGGING_EXPERIENCE`, `C_CONTENT_EDITOR`, `O_OPENSEARCH`.
- 후행: `M_MONETIZATION` (광고/정산 어드민 연동).

## 16. 테스트 시나리오 (Test Scenarios)
- `TEST-ADM-001`: 일반 회원(`USER`) 또는 블로그 소유자(`BLOG_OWNER`) 토큰으로 `/api/v1/admin/*` 엔드포인트 호출 시 100% `403 FORBIDDEN` 반환 및 보안 감사 로그 기록 검증.
- `TEST-ADM-002`: 신고 접수 → 케이스 생성 → 관리자 `quarantine` 실행 → 공개 조회 `451/404` 전환 및 OpenSearch 제거 → 작성자 이의신청(`appeal`) → 관리자 승인 시 `PUBLISHED` 복원 및 재색인 전 과정 E2E 검증.

## 17. Acceptance Criteria (인수 기준)
1. 제11조의 7대 관리자 모듈 전체 메뉴가 실제 백엔드 API 및 PostgreSQL/OpenSearch와 연결되어 동작하고 모든 변경사항이 `audit_logs`에 기록되어야 한다.

## 18. 실제 완료 증빙 방법 (Verification Evidence)
- 관리자 콘솔 8개 화면 E2E 실행 로그 및 `moderation_cases` / `audit_logs` DB 상태 전이 캡처.

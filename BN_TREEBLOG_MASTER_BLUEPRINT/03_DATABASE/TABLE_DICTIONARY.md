# TABLE_DICTIONARY.md — 전체 57개 데이터베이스 테이블 상세 사전

본 문서는 `03_DATABASE/migrations/0001` ~ `0010`에 정의된 전체 테이블의 목적, PK, FK, UNIQUE, CHECK 제약조건, 삭제 정책을 총괄 정리한 데이터 사전이다.

---

## 1. Identity & Authorization 도메인 (9개 테이블)

| 번호 | 테이블명 | PK | 주요 FK 및 삭제 정책 | UNIQUE 및 CHECK 제약조건 | 테이블 역할 및 보존 정책 |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | `users` | `id (UUID)` | 없음 | `UNIQUE(email)`, `CHECK(status IN (...))`, `CHECK(trust_tier BETWEEN 0 AND 5)` | 계정 마스터. 탈퇴 신청 시 30일 유예(`DEACTIVATING`) 후 개인정보 마스킹 처리. |
| 2 | `user_profiles` | `user_id (UUID)` | `user_id -> users(id) CASCADE` | `UNIQUE(username)` | 공개 닉네임, 소개, 아바타, 소셜 링크. |
| 3 | `identities` | `id (UUID)` | `user_id -> users(id) CASCADE` | `UNIQUE(provider, provider_subject)` | 로컬 인증, 소셜 로그인, 향후 `TREESCAN_OIDC` 연동 매핑. |
| 4 | `sessions` | `id (UUID)` | `user_id -> users(id) CASCADE` | `UNIQUE(refresh_token_hash)` | 기기별 Refresh Token Rotation 및 세션 강제 만료 관리. |
| 5 | `roles` | `id (UUID)` | 없음 | `UNIQUE(code)` | 7대 플랫폼 전역 RBAC 역할 정의. |
| 6 | `permissions` | `id (UUID)` | 없음 | `UNIQUE(code)` | 세부 기능 단위 권한 코드 사전. |
| 7 | `role_permissions` | `(role_id, permission_id)` | `role_id CASCADE`, `permission_id CASCADE` | 복합 PK | 역할별 권한 매핑 테이블. |
| 8 | `user_roles` | `(user_id, role_id)` | `user_id CASCADE`, `role_id CASCADE` | 복합 PK | 사용자에게 부여된 플랫폼 전역 역할. |
| 9 | `user_consents` | `id (UUID)` | `user_id -> users(id) CASCADE` | `CHECK(consent_type IN (...))` | 약관, 개인정보, 마케팅, 외부 블로그 이전 소유권 확인, TreeScan 연동 동의 이력. |

## 2. Blog, Domains & Themes 도메인 (8개 테이블)

| 번호 | 테이블명 | PK | 주요 FK 및 삭제 정책 | UNIQUE 및 CHECK 제약조건 | 테이블 역할 및 보존 정책 |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 10 | `blogs` | `id (UUID)` | `owner_user_id -> users(id) RESTRICT` | `UNIQUE(handle)` 정규식 검증, `CHECK(blog_type)`, `CHECK(status)` | 개인/전문/기업/커뮤니티 블로그 마스터 엔터티. |
| 11 | `blog_members` | `id (UUID)` | `blog_id -> blogs(id) CASCADE`, `user_id -> users(id) CASCADE` | `UNIQUE(blog_id, user_id)`, `CHECK(role IN ('OWNER','ADMIN','EDITOR','AUTHOR','ANALYST'))` | 팀/기업 블로그 다중 멤버 권한 테이블. |
| 12 | `blog_settings` | `blog_id (UUID)` | `blog_id -> blogs(id) CASCADE` | `CHECK(posts_per_page BETWEEN 5 AND 50)` | 기본 공개 범위, 댓글 정책, 방명록/RSS/검색 허용 설정. |
| 13 | `blog_domains` | `id (UUID)` | `blog_id -> blogs(id) CASCADE` | `UNIQUE(domain_name)`, Partial Unique `(blog_id) WHERE is_primary = TRUE` | 서브도메인 및 개인 커스텀 도메인 DNS 검증·TLS 인증서 상태. |
| 14 | `themes` | `id (UUID)` | 없음 | `UNIQUE(slug)` | 플랫폼 공식 반응형 테마 카탈로그. |
| 15 | `blog_themes` | `blog_id (UUID)` | `blog_id CASCADE`, `theme_id -> themes(id) RESTRICT` | `CHECK(color_mode IN ('LIGHT','DARK','SYSTEM'))` | 블로그별 적용 테마, 디자인 토큰 오버라이드, 정화된 커스텀 CSS. |
| 16 | `blog_navigation` | `id (UUID)` | `blog_id -> blogs(id) CASCADE` | `CHECK(link_type IN (...))` | 블로그 상단 커스텀 메뉴 구성. |
| 17 | `blog_categories` | `id (UUID)` | `blog_id CASCADE`, `parent_id -> blog_categories(id) SET NULL` | `UNIQUE(blog_id, slug)`, `CHECK(list_style)`, `CHECK(visibility)` | 최대 2뎁스 계층형 카테고리 및 목록 레이아웃 설정. |

## 3. Content & Editor 도메인 (10개 테이블)

| 번호 | 테이블명 | PK | 주요 FK 및 삭제 정책 | UNIQUE 및 CHECK 제약조건 | 테이블 역할 및 보존 정책 |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 18 | `series` | `id (UUID)` | `blog_id -> blogs(id) CASCADE` | `UNIQUE(blog_id, slug)` | 연재물(시리즈) 묶음 관리. |
| 19 | `posts` | `id (UUID)` | `blog_id CASCADE`, `author_user_id RESTRICT`, `category_id SET NULL` | `UNIQUE(blog_id, slug)`, `CHECK(status)`, `CHECK(visibility)`, `CHECK(license_type)` | 게시글 마스터. `content_json`, `content_html`, `content_plain` 3중 저장. |
| 20 | `post_revisions` | `id (UUID)` | `post_id -> posts(id) CASCADE`, `editor_user_id RESTRICT` | `UNIQUE(post_id, revision_no)` | 게시글 수정 이력 스냅샷 및 버전 복원용 테이블. |
| 21 | `autosave_drafts` | `post_id (UUID)` | `post_id CASCADE`, `user_id CASCADE` | `base_revision_no >= 1` | 작성 중 실시간 서버 자동저장 작업본. |
| 22 | `post_categories` | `(post_id, category_id)` | `post_id CASCADE`, `category_id CASCADE` | 복합 PK | 다중 카테고리 매핑 지원. |
| 23 | `tags` | `id (UUID)` | 없음 | `UNIQUE(name)`, `UNIQUE(slug)` | 플랫폼 전역 정규화 해시태그 사전. |
| 24 | `post_tags` | `(post_id, tag_id)` | `post_id CASCADE`, `tag_id CASCADE` | 복합 PK | 게시글-태그 매핑 및 순서. |
| 25 | `post_visibility_rules` | `id (UUID)` | `post_id -> posts(id) CASCADE` | `CHECK(rule_type)` | 특정 이웃 그룹이나 멤버십 티어별 세부 열람 규칙. |
| 26 | `scheduled_publications`| `id (UUID)` | `post_id UNIQUE CASCADE`, `blog_id CASCADE` | `CHECK(status IN ('PENDING','EXECUTING','COMPLETED','CANCELED','FAILED'))` | 분 단위 정밀 예약 발행 스케줄 큐. |
| 27 | `post_slug_history` | `id (UUID)` | `blog_id CASCADE`, `post_id CASCADE` | `UNIQUE(blog_id, old_slug)` | 슬러그 변경 및 외부 블로그 이전 시 구 URL 301 리다이렉트 보장. |

## 4. Media 도메인 (5개 테이블)

| 번호 | 테이블명 | PK | 주요 FK 및 삭제 정책 | UNIQUE 및 CHECK 제약조건 | 테이블 역할 및 보존 정책 |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 28 | `upload_sessions` | `id (UUID)` | `blog_id CASCADE`, `user_id CASCADE` | `UNIQUE(storage_object_key)`, `CHECK(expected_size_bytes > 0)` | S3 Presigned URL 업로드 세션 및 만료 추적. |
| 29 | `media_assets` | `id (UUID)` | `blog_id CASCADE`, `uploader_user_id RESTRICT` | `CHECK(media_type)`, `CHECK(status)`, `exif_stripped` | 이미지/GIF/동영상/임베드 마스터. SHA-256 중복 식별 및 고아 참조 카운트. |
| 30 | `media_variants` | `id (UUID)` | `media_asset_id -> media_assets(id) CASCADE` | `UNIQUE(media_asset_id, variant_name, format)` | 해상도별(`W320`~`W1920`, `THUMB_400`) WebP/AVIF 변환 파일 및 CDN URL. |
| 31 | `media_processing_jobs` | `id (UUID)` | `media_asset_id -> media_assets(id) CASCADE` | `CHECK(job_type)`, `CHECK(status)` | 비동기 이미지 리사이징, EXIF 제거, 비디오 트랜스코딩 작업 상태. |
| 32 | `post_media` | `(post_id, media_asset_id, usage_role)` | `post_id CASCADE`, `media_asset_id RESTRICT` | 복합 PK | 게시글 내 사용된 미디어 참조 추적(고아 파일 GC 판단 기준). |

## 5. Community & Notifications 도메인 (10개 테이블)

| 번호 | 테이블명 | PK | 주요 FK 및 삭제 정책 | UNIQUE 및 CHECK 제약조건 | 테이블 역할 및 보존 정책 |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 33 | `neighbor_groups` | `id (UUID)` | `owner_user_id -> users(id) CASCADE` | `UNIQUE(owner_user_id, name)` | 사용자가 이웃 블로그를 분류하는 커스텀 그룹 폴더. |
| 34 | `follows` | `id (UUID)` | `follower_user_id CASCADE`, `target_blog_id CASCADE` | `UNIQUE(follower_user_id, target_blog_id)`, `CHECK(relation_type IN ('FOLLOW','NEIGHBOR','MUTUAL_NEIGHBOR'))` | 팔로우·이웃·상호이웃(서로이웃) 관계 및 승인 상태. |
| 35 | `subscriptions` | `id (UUID)` | `subscriber_user_id CASCADE`, `target_blog_id CASCADE` | `UNIQUE(subscriber_user_id, target_blog_id)` | 새 글 알림(인앱/푸시/이메일) 및 구독 티어 상태. |
| 36 | `reactions` | `id (UUID)` | `user_id -> users(id) CASCADE` | `UNIQUE(target_type, target_id, user_id)`, `CHECK(reaction_type)` | 게시글 6종 공감 및 댓글 좋아요. |
| 37 | `comments` | `id (UUID)` | `post_id CASCADE`, `blog_id CASCADE`, `parent_comment_id CASCADE` | `CHECK(depth IN (0, 1))`, `CHECK(status)` | 댓글·대댓글, 비밀댓글(`is_secret`), 승인제 상태 관리. |
| 38 | `bookmarks` | `id (UUID)` | `user_id CASCADE`, `post_id CASCADE` | `UNIQUE(user_id, post_id)` | 독자 개인 북마크 및 읽기 진행률(`read_progress_pct`). |
| 39 | `guestbook_entries` | `id (UUID)` | `blog_id CASCADE`, `author_user_id CASCADE` | `CHECK(status)` | 블로그 방명록 및 비밀 방명록. |
| 40 | `user_blocks` | `(blocker_user_id, blocked_user_id)` | `blocker_user_id CASCADE`, `blocked_user_id CASCADE` | `CHECK(blocker_user_id <> blocked_user_id)` | 사용자 간 차단 및 상호작용 원천 배제. |
| 41 | `notification_preferences`| `user_id (UUID)` | `user_id -> users(id) CASCADE` | 없음 | 알림 유형별/채널별 수신 동의 및 야간 방해금지 시간. |
| 42 | `notifications` | `(id, created_at)` | `recipient_user_id -> users(id) CASCADE` | 월별 Range Partition | Rollup 병합(`group_key`) 및 중복 방지(`dedup_key`) 적용 알림함. |

## 6. Search & Quality 도메인 (4개 테이블)

| 번호 | 테이블명 | PK | 주요 FK 및 삭제 정책 | UNIQUE 및 CHECK 제약조건 | 테이블 역할 및 보존 정책 |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 43 | `search_index_jobs` | `id (UUID)` | 없음 | `CHECK(target_index)`, `CHECK(operation IN ('UPSERT','DELETE'))` | OpenSearch 비동기 색인 동기화 작업 큐. |
| 44 | `search_queries` | `id (UUID)` | `blog_scope_id SET NULL` | `UNIQUE(normalized_query, language_code)` | 검색어 빈도, 무결과 검색어, 스팸 검색어 차단 관리. |
| 45 | `search_click_events` | `(id, created_at)` | `clicked_post_id -> posts(id) CASCADE` | 월별 Range Partition | 검색 결과 클릭 위치 및 체류시간 품질 측정 로그. |
| 46 | `content_quality_signals`| `post_id (UUID)` | `post_id -> posts(id) CASCADE` | `CHECK(quality_score BETWEEN 0 AND 100)` | 64비트 SimHash 지문, 중복 여부, 스팸 확률, 색인 적격 판정. |

## 7. Moderation & Trust/Safety 도메인 (5개 테이블)

| 번호 | 테이블명 | PK | 주요 FK 및 삭제 정책 | UNIQUE 및 CHECK 제약조건 | 테이블 역할 및 보존 정책 |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 47 | `moderation_cases` | `id (UUID)` | `assigned_moderator_id SET NULL` | `UNIQUE(target_type, target_id, reason_category)` | 동일 대상 신고 병합 케이스 및 우선순위 심사 큐. |
| 48 | `user_reports` | `id (UUID)` | `case_id -> moderation_cases(id) CASCADE` | `CHECK(reason_category)` | 개별 사용자가 제출한 신고 사유 및 증빙 URL. |
| 49 | `moderation_actions` | `id (UUID)` | `case_id RESTRICT`, `actor_moderator_id RESTRICT` | `CHECK(action_type)` | 모더레이터의 제재/복원/경고 결정 이력. |
| 50 | `appeals` | `id (UUID)` | `case_id RESTRICT`, `appellant_user_id CASCADE` | `CHECK(status)` | 제재 대상자의 이의신청(소명) 및 재심사 결과. |
| 51 | `spam_detection_results`| `id (UUID)` | 없음 | `CHECK(spam_score BETWEEN 0 AND 1)` | 자동 스팸 필터/AI 분류기의 판정 근거 로그. |

## 8. Monetization & Accounting 도메인 (9개 테이블)

| 번호 | 테이블명 | PK | 주요 FK 및 삭제 정책 | UNIQUE 및 CHECK 제약조건 | 테이블 역할 및 보존 정책 |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 52 | `ad_accounts` | `id (UUID)` | `owner_user_id -> users(id) RESTRICT` | `CHECK(prepaid_balance_minor >= 0)` | 광고주 계정 및 선불 충전 잔액. |
| 53 | `ad_campaigns` | `id (UUID)` | `ad_account_id -> ad_accounts(id) RESTRICT` | `CHECK(pricing_model IN ('CPM','CPC'))`, `CHECK(review_status)` | 광고 캠페인 예산, 타기팅, 소재 및 심사 상태. |
| 54 | `ad_placements` | `id (UUID)` | `blog_id -> blogs(id) CASCADE` | `UNIQUE(blog_id, slot_position)` | 블로그별 광고 슬롯 활성화 및 수익 배분율(`revenue_share_bps`). |
| 55 | `ad_events` | `(id, created_at)` | `campaign_id RESTRICT`, `placement_id RESTRICT` | 월별 Range Partition | 광고 노출(`IMPRESSION`)/클릭(`CLICK`), 과금액, 부정트래픽(IVT) 여부. |
| 56 | `memberships` | `id (UUID)` | `blog_id -> blogs(id) CASCADE` | `CHECK(monthly_price_minor > 0)` | 크리에이터 유료 멤버십 플랜 정의. |
| 57 | `orders` | `id (UUID)` | `buyer_user_id RESTRICT`, `recipient_blog_id RESTRICT` | `UNIQUE(idempotency_key)`, `CHECK(order_type)` | 멤버십 구독, 단건 유료글 구매, 독자 후원, 광고비 충전 주문. |
| 58 | `payments` | `id (UUID)` | `order_id -> orders(id) RESTRICT` | `UNIQUE(pg_transaction_id)` | PG사 승인/매입/환불/차지백 결제 트랜잭션 원본. |
| 59 | `payouts` | `id (UUID)` | `payee_user_id RESTRICT`, `blog_id RESTRICT` | `UNIQUE(idempotency_key)`, `CHECK(status)` | 블로거 수익금 정산 신청, 원천징수 세액, 지급 보류/완료 이력. |
| 60 | `ledger_entries` | `id (UUID)` | `owner_user_id RESTRICT`, `blog_id RESTRICT` | `UNIQUE(idempotency_key)`, `CHECK(direction IN ('DEBIT','CREDIT'))` | **불변 복식부기 회계 원장** (수정/삭제 불가, 법정 5년 이상 보존). |

## 9. Operations, Portability & TreeScan Integration 도메인 (7개 테이블)

| 번호 | 테이블명 | PK | 주요 FK 및 삭제 정책 | UNIQUE 및 CHECK 제약조건 | 테이블 역할 및 보존 정책 |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 61 | `audit_logs` | `(id, created_at)` | `actor_user_id -> users(id) RESTRICT` | 월별 Range Partition, 해시 체인(`prev_hash`, `entry_hash`) | 관리자 및 민감 보안 작업의 위변조 방지 감사 로그. |
| 62 | `feature_flags` | `key (VARCHAR)` | `updated_by SET NULL` | `CHECK(rollout_percentage BETWEEN 0 AND 100)` | 런타임 기능 토글 및 점진적 롤아웃 제어. |
| 63 | `system_settings` | `key (VARCHAR)` | `updated_by SET NULL` | 없음 | 플랫폼 전역 동적 설정 저장소. |
| 64 | `outbox_events` | `id (UUID)` | 없음 | `UNIQUE(idempotency_key)`, `CHECK(status)` | Transactional Outbox 비동기 이벤트 발행 큐. |
| 65 | `background_jobs` | `id (UUID)` | 없음 | `UNIQUE(idempotency_key)`, `CHECK(queue_name)` | 백그라운드 워커 작업 상태 및 Dead Letter 추적. |
| 66 | `migration_jobs` | `id (UUID)` | `blog_id CASCADE`, `initiated_by RESTRICT` | `CHECK(direction IN ('IMPORT','EXPORT'))` | 블로그 백업 내보내기 및 외부 블로그 가져오기 작업 이력. |
| 67 | `treescan_sync_state` | `id (UUID)` | `post_id UNIQUE CASCADE`, `blog_id CASCADE` | `CHECK(sync_action IN ('UPSERT','DELETE_TOMBSTONE'))` | 향후 TreeScan 검색엔진 색인 동기화 오프셋 및 삭제 툼스톤 추적. |

# INTEGRATION_ARCHITECTURE.md — TreeScan 및 Booking Network 생태계 연동 어댑터 아키텍처 (제1.2조, 제18조)

## 1. 느슨한 결합(Loosely Coupled) 어댑터 구조
TreeBlog 코어 모듈은 TreeScan을 직접 알지 못하며, 오직 `internal/modules/treescan/` 패키지의 어댑터와 `outbox_events` / `treescan_sync_state` 테이블을 통해서만 외부 생태계와 통신한다.

```
[TreeBlog Core Modules (post, blog, identity)]
                     │
                     ▼ (Transactional Outbox: post.published / post.visibility_changed)
        [treescan_sync_state Table]
         (Strict Gate: ONLY visibility == 'PUBLIC' OR 'DELETE_TOMBSTONE')
                     │
       ┌─────────────┴───────────────────────────────────────────┐
       ▼                                                         ▼
[1. Search Index Sync Adapter]                    [2. Booking Network Ecosystem Discovery API]
 - Pull: GET /api/v1/treescan/sync/feed            - GET /api/v1/treescan/discovery/posts
 - Push: CloudEvents Webhook (Optional)            - 연동 대상:
       │                                             • TreeScan Search (통합 블로그 탭)
       ▼                                             • Treewiki (백과사전 개념 연관 블로그 글)
[TreeScan Search Engine Indexer]                     • getforTrip / TripScanner (여행기·숙소/명소 리뷰)
                                                     • App Market (앱 사용기·기술 리뷰)
                                                     • OLMOA (제품 쇼핑 리뷰)
                                                     • Global Video (비디오 블로그 연동)
```

## 2. 제1.2조 9대 필수 통합 인터페이스 매핑
1. **Search API Adapter**: `GET /api/v1/treescan/discovery/posts` 및 검색 프록시 인터페이스.
2. **Search Index Synchronization**: `GET /api/v1/treescan/sync/feed` (커서 기반 증분 스트림 + 삭제 툼스톤).
3. **SSO / OIDC Integration**: `POST /api/v1/treescan/sso/link` (`identities.provider = 'TREESCAN_OIDC'`).
4. **Unified User Profile**: `user_consents.consent_type = 'TREESCAN_PROFILE_SYNC'` 동의자 한정 프로필 동기화.
5. **Content Discovery API**: 자매 플랫폼 엔터티 ID(`place_id`, `app_id`, `product_id`, `wiki_topic_id`) 기반 관련 블로그 글 조회.
6. **Blog Search Category**: 표준 토픽 분류 체계(`topic_code`) 공유.
7. **Cross-platform Content Linking**: 자매 플랫폼 URL 붙여넣기 시 공식 리치 프리뷰 블록 생성.
8. **Sitemap / Crawling Interface**: `/sitemap-index.xml` 및 `TreeScanBot` 전용 크롤링 모니터링(`EP-ANL-005`).
9. **Global Navigation Integration**: 헤더 내 Booking Network 패밀리 앱 스위처 위젯(설정 플래그로 ON/OFF).

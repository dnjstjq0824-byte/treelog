# POST_LIFECYCLE.md — 게시글 상태 머신 및 공개·비공개·삭제·복원 정책 (제5조, 제9조, 제16조)

## 1. 게시글 상태 전이 다이어그램
```
[DRAFT (초안)] ──(팀블로그 필진 검토요청)──► [IN_REVIEW (검토대기)]
      │                                              │
      ├──────────────(예약 시각 지정)────────────────┼──► [SCHEDULED (예약대기)]
      │                                              │            │ (스케줄러 정시 실행)
      └──────────────(즉시 발행 클릭)────────────────┴────────────▼
                                                        [PUBLISHED (발행됨)]
                                                        (visibility: PUBLIC / UNLISTED /
                                                         NEIGHBOR_ONLY / MUTUAL_ONLY /
                                                         PAID_MEMBERS_ONLY / PRIVATE)
                                                                  │
       ┌──────────────────────────┬───────────────────────────────┴──────────────────────┐
       ▼ (작성자 발행취소)        ▼ (신고/법적 임시조치)                                 ▼ (작성자 삭제)
[UNPUBLISHED]              [QUARANTINED (게시중단)]                              [TRASHED (휴지통 30일)]
                                  │ (이의신청 인용 시 복원)                              │ (30일 경과 또는 즉시파기)
                                  ▼                                                      ▼
                            [PUBLISHED]                                            [PURGED (영구삭제)]
```

## 2. 상태별 HTTP 응답 코드 및 검색엔진 제어 헤더
| 게시글 상태 / 가시성 | 일반 방문자 HTTP 응답 | 응답 헤더 (`X-Robots-Tag` / `Cache-Control`) | OpenSearch & TreeScan 색인 동작 |
| :--- | :--- | :--- | :--- |
| `PUBLISHED` + `PUBLIC` | `200 OK` | `index, follow` / `public, s-maxage=60, stale-while-revalidate=600` | 즉시 `UPSERT` 색인 |
| `PUBLISHED` + `UNLISTED` | `200 OK` | `noindex, nofollow` / `private, no-cache` | 색인 제외 (`DELETE`) |
| `PUBLISHED` + `NEIGHBOR_ONLY` / `MUTUAL_ONLY` | 권한자 `200 OK`, 비권한자 `403 Forbidden` | `noindex, nofollow` / `private, no-store` | 색인 절대 제외 (`DELETE`) |
| `PUBLISHED` + `PRIVATE` | 작성자/편집자 외 `404 Not Found` (존재 여부 은닉) | `noindex, nofollow` / `private, no-store` | 색인 절대 제외 (`DELETE`) |
| `QUARANTINED` (임시조치) | `451 Unavailable For Legal Reasons` | `noindex, nofollow` / `no-store` | 최우선 순위(`priority=100`) 즉시 색인 삭제 |
| `TRASHED` (휴지통) | `404 Not Found` | `noindex, nofollow` / `no-store` | 최우선 순위 즉시 색인 삭제 |
| `PURGED` (영구삭제) | `410 Gone` | `noindex, nofollow` / `public, max-age=3600` | 영구 삭제 툼스톤 유지 |

# SEARCH_API_ADAPTER.md — TreeScan 검색 API 어댑터 및 자매 플랫폼 디스커버리 명세

## 1. 자매 플랫폼 연동용 Content Discovery API (`GET /api/v1/treescan/discovery/posts`)
향후 TreeScan 통합 검색의 '블로그' 탭이나 getforTrip(여행지 상세), OLMOA(상품 상세), Treewiki(문서 하단)에서 관련 TreeBlog 공개 게시글을 실시간 조회할 때 사용하는 계약이다.
- **Query Parameters**:
  - `q`: 검색 키워드 (선택)
  - `topic`: 표준 토픽 슬러그 (`travel`, `it-tech`, `food`, `finance`, `shopping` 등)
  - `entity_ref`: Booking Network 공통 엔터티 참조 코드 (예: `getfortrip:place:seoul-namsan`, `treewiki:doc:golang`)
  - `lang`: 언어 코드 (`ko`, `en`, `ja`)
  - `sort`: `relevance` | `quality` | `latest`
  - `limit`: `1..50`
- **보안 및 중립성 보장**:
  - 서비스 간 mTLS 또는 `X-TreeBlog-Service-Signature` HMAC-SHA256 인증 필수.
  - 오직 `PUBLISHED` + `PUBLIC` + `allow_treescan_sync = true`인 게시글만 반환한다.

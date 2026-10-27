# CONTENT_INDEXING.md — Transactional Outbox 기반 실시간 검색 색인 파이프라인

## 1. 색인 동기화 및 삭제 전파 SLA
- **신규 발행 및 수정 (`UPSERT`)**: 게시글 발행 커밋 후 Outbox 워커를 통해 평시 **1.5초 이내(P95)**에 OpenSearch `treeblog_posts_v1`에 반영된다.
- **비공개 전환·휴지통 이동·관리자 제재 (`DELETE`)**: `critical` 우선순위 큐(`priority = 100`)를 통해 **0.5초 이내(P99)**에 OpenSearch에서 즉시 제거된다.
- **정합성 감사 배치 (Reconciliation Cron)**: 매일 새벽 03:00 UTC에 PostgreSQL의 공개 게시글 ID 집합과 OpenSearch 인덱스 문서 ID 집합을 비교하여 누락되거나 유령(Ghost) 상태로 남은 문서를 자동 보정한다.

# NOTIFICATIONS.md — 실시간 멀티채널 알림 및 스마트 병합(Rollup) 엔진 명세

## 1. 채널별 발송 파이프라인
- **In-App 알림**: PostgreSQL `notifications` 파티션 테이블 저장 + Redis Pub/Sub `notif:stream:{userId}` 채널로 발행하여 접속 중인 브라우저에 SSE(Server-Sent Events)로 0.5초 이내 실시간 전달.
- **Web Push 알림**: W3C Push API + VAPID 서명을 통해 브라우저가 백그라운드 상태일 때도 데스크톱/모바일 OS 네이티브 알림 표시.
- **Email 알림**: 새 글 뉴스레터, 상호이웃 신청, 보안 경고, 모더레이션 통지를 반응형 HTML 이메일 템플릿으로 발송.

## 2. Rollup 병합 및 쿨다운 디바운싱
- 동일 게시글(`post_id`)에 대해 읽지 않은(`is_read = false`) 공감 알림이 이미 존재하면 새 행을 INSERT하지 않고 기존 행의 `actor_count`를 `+1` 증가시키고 `actors_preview_json` 배열(최근 3명)과 `created_at`만 갱신한다.

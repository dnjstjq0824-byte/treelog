# CONFLICT_PREVENTION.md — 멀티 에이전트 파일 충돌 방지 및 계약 변경(RFC) 규칙

## 1. Git Merge 충돌 원천 차단 4대 규칙
1. **디렉터리 독점 소유권 (Directory Isolation)**: 각 에이전트는 `AGENT_ROLE_MATRIX.md`와 개별 Work Order의 `7. 수정 가능한 디렉터리`에 명시된 경로 외의 파일을 수정할 수 없다. CI 프리커밋 훅(`scripts/check-agent-ownership.sh`)이 허용되지 않은 경로 수정을 자동 차단한다.
2. **자동 생성 파일 수동 편집 금지**: `packages/api-client/` 및 `internal/db/sqlc/` 등 스키마로부터 코드젠되는 디렉터리는 수동으로 편집하지 않고 오직 생성 스크립트로만 갱신한다.
3. **모듈 간 인터페이스 통신 (`Service Port` & `Outbox`)**: 예를 들어 `Community Agent`가 게시글의 댓글 수를 증가시킬 때는 `internal/modules/post/` 파일을 직접 수정하지 않고, `PostService.IncrementCommentCount(ctx, postID, delta)` 인터페이스 메서드를 호출하거나 도메인 이벤트를 발행한다.
4. **DB 및 API 계약 변경 프로토콜 (Contract Change RFC)**:
   - 구현 도중 `03_DATABASE/migrations/` 또는 `04_API/openapi.yaml`의 수정이 불가피한 경우, 담당 에이전트는 임의로 수정하지 않고 `RFC-{번호}` 요청을 발행한다.
   - `Lead Architect Agent`와 `PostgreSQL Agent`가 영향받는 모든 Work Order 명세를 동시에 갱신한 후에만 스키마를 변경한다.

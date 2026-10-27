# AGENT_HANDOFF_PROTOCOL.md — 에이전트 간 작업 인계(Handoff) 및 상태 전이 프로토콜

## 1. Work Order 상태 전이 규칙
각 Work Order는 다음 6가지 상태만을 가질 수 있으며, 단계별 인계 조건을 충족해야 다음 에이전트가 작업을 시작할 수 있다.
1. `NOT_STARTED`: 선행 작업 대기 또는 착수 전.
2. `PARTIAL`: 담당 구현 에이전트가 코드를 작성 중이거나 일부 기능만 구현된 상태.
3. `IMPLEMENTED_UNVERIFIED`: 담당 구현 에이전트가 코드 작성 및 로컬 단위/통합 테스트를 마치고 Handoff Artifacts(Git Commit SHA, 테스트 로그)를 첨부하여 **QA / E2E Agent**에게 검증을 요청한 상태.
4. `VERIFIED`: **QA / E2E Agent**의 통합/E2E 검증과 **Independent Audit Agent**의 독립 감사를 모두 통과한 상태. **오직 `VERIFIED` 상태의 Work Order만 후속 Work Order의 선행조건(Prerequisites)을 충족한 것으로 인정한다.**
5. `FAILED`: 테스트 실패, API/DB 스키마 불일치, 또는 제27조 금지사항 위반이 적발되어 반려된 상태.
6. `BLOCKED`: 외부 승인(PG 실계약 등) 또는 선행 작업 미완료로 보류된 상태.

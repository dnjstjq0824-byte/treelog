# REPORTING_TEMPLATE.md — 에이전트 Work Order 완료 및 감사 보고서 표준 양식

모든 에이전트는 각 Work Order 완료 시 아래 마크다운/JSON 호환 양식을 빠짐없이 채워 제출해야 한다.

```markdown
# WORK ORDER COMPLETION REPORT — [TB-XXX-NNN]

- **Work Order ID**: `TB-XXX-NNN`
- **작업명**: (작업명 기재)
- **담당 에이전트**: (에이전트 역할명)
- **보고 상태**: `IMPLEMENTED_UNVERIFIED` | `VERIFIED` | `FAILED` | `BLOCKED`
- **Git Branch & Commit SHA**: `feat/TB-XXX-NNN` (`a1b2c3d4e5f6...`)
- **보고 일시 (UTC)**: `2026-10-27T15:00:00Z`

## 1. 변경된 파일 목록 및 소유권 준수 확인
- `path/to/file1.go` (Allowed Directory 내 포함 여부: YES)
- `path/to/file2.vue` (Allowed Directory 내 포함 여부: YES)
- Restricted Directory 침범 여부: `NONE (0 files)`

## 2. DB 및 API 계약 일치성 검증
- 관련 DB 테이블 마이그레이션 일치 여부: `MATCHED`
- OpenAPI 3.1.0 (`openapi.yaml`) 요청/응답 스키마 일치 여부: `MATCHED`

## 3. 실제 테스트 실행 결과 및 증빙 경로 (Mock 여부 명시 필수)
| 테스트 ID | 테스트 유형 (Unit / Real-DB Integration / Playwright E2E) | 실행 명령 | 결과 | 증빙 로그 경로 |
| :--- | :--- | :--- | :--- | :--- |
| `TEST-XXX-001` | Real-DB Integration (Postgres + OpenSearch + Redis) | `go test -v ./internal/modules/...` | `PASS` | `artifacts/evidence/TB-XXX-NNN/test.log` |

## 4. 미완료 사항, 알려진 제한점 또는 외부 승인 대기 항목
- (없을 경우 `NONE`, 외부 PG/법무 승인 대기 항목이 있을 경우 구체적 사유 기재)
```

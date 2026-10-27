# TB-MEDIA-001 — S3 Presigned 업로드 세션, 매직바이트 검증, 이미지 리사이징(WebP/AVIF) 및 EXIF 제거 워커 구현

> **문서 등급**: EXECUTABLE WORK ORDER SPECIFICATION (제19조 18대 필수 항목 완비)
> **소속 단계**: `Phase 3`

---

## 1. Work Order ID
- `TB-MEDIA-001`

## 2. 작업명 (Task Title)
- **S3 Presigned 업로드 세션, 매직바이트 검증, 이미지 리사이징(WebP/AVIF) 및 EXIF 제거 워커 구현**

## 3. 목적 (Purpose)
- Private S3 버킷 Presigned 업로드, 바이너리 시그니처 검증, libvips 기반 다중 해상도 WebP/AVIF 변환 및 GPS EXIF 개인정보 100% 제거 워커를 구현한다.

## 4. 담당 에이전트 유형 (Assigned Agent Role)
- **Media Processing Agent** (`14_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md` 참조)

## 5. 선행 작업 (Prerequisites)
- `TB-ARCH-002, TB-DB-002`

## 6. 구현 범위 (Implementation Scope)
- 업로드 세션 발급, SHA-256 중복 파일 재사용, EXIF 스트리핑, 5종 해상도 변환, Blurhash 추출, 7일 경과 고아 파일 GC 워커

## 7. 수정 가능한 디렉터리 (Allowed Directories)
- `internal/modules/media/`

## 8. 수정 제한 디렉터리 (Restricted Directories — 무단 수정 금지)
- `apps/` (스키마 또는 공통 API 계약 변경 필요 시 반드시 `Lead Architect Agent` RFC 승인 선행)

## 9. 관련 DB 테이블 (Related Database Tables)
- `upload_sessions, media_assets, media_variants, media_processing_jobs, post_media` (`03_DATABASE/TABLE_DICTIONARY.md` 참조)

## 10. 관련 API (Related API Endpoints)
- `EP-MED-001, EP-MED-002, EP-MED-003, EP-MED-004` (`04_API/openapi.yaml` 및 `API_CONTRACT.md` 참조)

## 11. 관련 UI (Related UX/UI Screens)
- `SCR-BLOG-006, SCR-BLOG-013` (`05_UXUI/SCREEN_INVENTORY.md` 참조)

## 12. 구현 상세 요구사항 (Detailed Implementation Requirements)
- 1) 이미지 폭/높이 16,384px 초과 시 디코딩 전 즉시 차단(Image Bomb 방어). 2) exif_stripped=true 확인 후에만 Public CDN 버킷으로 업로드.
- 모든 에러 처리는 `04_API/API_ERROR_MODEL.md`의 표준 에러 코드와 HTTP 상태 코드를 반환해야 한다.
- 하드코딩된 설정이나 비밀키를 금지하며 환경변수 및 의존성 주입(DI)을 사용한다.

## 13. 예상 파일 구조 (Expected File Structure)
- `internal/modules/media/upload.go, internal/modules/media/processor.go, internal/modules/media/gc_worker.go`

## 14. 테스트 시나리오 (Test Scenarios)
- 검증 테스트 케이스 ID: `TEST-MED-001, TEST-SEC-003` (`15_TESTING/MASTER_TEST_MATRIX.csv` 참조)
- 단위 테스트(Unit Test)와 실제 컨테이너 연동 통합 테스트(Integration Test)를 모두 작성·실행해야 한다.

## 15. Acceptance Criteria (인수 기준)
- GPS 메타데이터가 포함된 사진 업로드 후 생성된 모든 공개 파생본에 GPS 태그가 0바이트여야 함
- 린트(`golangci-lint` / `eslint`), 타입체크(`vue-tsc`), 빌드 및 연결된 테스트 케이스가 에러 0건으로 통과해야 한다.

## 16. 예상 위험 및 완화 대책 (Expected Risks & Mitigation)
- 대용량 GIF 변환 시 CPU 급증 -> media 전용 워커 큐 동시성 제한 적용

## 17. 최종 증빙 (Required Verification Evidence)
- Git Commit SHA, 변경 파일 목록(`git diff --stat`), 테스트 실행 표준 출력 로그(`artifacts/evidence/TB-MEDIA-001/test_output.log`), (UI 포함 시) Playwright 실행 트레이스 및 스크린샷.

## 18. 완료 선언 조건 (Definition of Done)
- 구현 에이전트가 자체 테스트 통과 후 상태를 `IMPLEMENTED_UNVERIFIED`로 보고하고, **QA / E2E Agent** 및 **Independent Audit Agent**가 실제 컨테이너/브라우저 환경에서 교차 검증하여 `VERIFIED` 판정을 내렸을 때만 완료로 선언한다(제20조, 제27조 준수).

# SECURITY_TEST_PLAN.md — 제16조 20대 보안 위협 침투 테스트 및 검증 계획서

## 1. 필수 자동화 보안 테스트 스위트 (`tests/security/penetration_test.go`)
1. **XSS Fuzzing Suite (`TEST-SEC-001`)**:
   - `<script>`, `<img src=x onerror=...>`, `<svg onload=...>`, `<a href="javascript:...">`, `<iframe srcdoc=...>`, CSS `expression()`, `url(javascript:...)` 등 50종의 공격 벡터를 에디터 JSON AST, 마크다운 가져오기, 댓글, 방명록, 프로필, 블로그 제목, 테마 커스텀 CSS에 주입한 뒤 저장된 HTML/CSS를 파싱하여 실행 가능한 스크립트 컨텍스트가 0건임을 검증한다.
2. **SSRF Isolation Suite (`TEST-SEC-002`)**:
   - 에디터 링크 프리뷰(`EP-EDIT-005`), 외부 동영상 임베드(`EP-MED-005`), 외부 블로그 가져오기(`EP-PRT-003`)에 `http://127.0.0.1:8080`, `http://[::1]`, `http://10.0.0.1`, `http://169.254.169.254/latest/meta-data/`, `file:///etc/passwd`, `gopher://`를 입력했을 때 소켓 연결 시도 전에 `400 SSRF_FORBIDDEN_HOST`로 즉시 거부되는지 검증한다.
3. **IDOR & Broken Access Control Matrix Suite (`TEST-COM-001`, `TEST-ADM-001`)**:
   - 사용자 A의 토큰으로 사용자 B의 블로그 설정 수정, 카테고리 삭제, 비공개 글 열람, 비밀 댓글 열람, 정산 원장 조회, 출금 신청, 관리자 API 호출을 전수 시도하여 100% `403`/`404`로 차단되는지 검증한다.
4. **Malicious File Upload & EXIF Privacy Suite (`TEST-MED-001`, `TEST-SEC-003`)**:
   - 확장자만 `.png`인 웹쉘/실행파일 업로드 거부 확인, 해상도 폭탄(30,000×30,000px) 거부 확인, 원본 JPEG의 GPS EXIF 태그가 공개 WebP/AVIF 파생본에서 100% 삭제되었는지 검증한다.
5. **CSRF & Auth Token Replay Suite (`TEST-IAM-001`)**:
   - 상태 변경 API에 `X-CSRF-Token` 누락/변조 시 `403 CSRF_TOKEN_MISMATCH` 반환 확인, 이미 사용된 Refresh Token 재전송 시 세션 패밀리 전체 폐기 확인.

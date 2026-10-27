# KNOWN_LIMITATIONS.md — 남은 설계 위험, 외부 승인 필요 사항 및 미실행 런타임 검증 명시 (제25조, 제26조, 제28조)

제25조 제18항 및 제28조 지침(*"실행하지 않은 검증을 PASS로 보고하지 않는다. 실제 운영 권한, 외부 서비스 약관, 법적 검토 및 프로덕션 배포에 필요한 승인은 임의로 완료 처리하지 않는다"*)에 따라 다음 항목을 투명하게 분리하여 보고한다.

---

## 1. 블루프린트 단계에서 실행하지 않은 검증 (Awaiting Code Implementation — `NOT_VERIFIED` at Runtime)
본 작업은 후속 개발 에이전트가 전체 플랫폼을 구현할 수 있도록 **마스터 아키텍처 및 작업지시 ZIP 패키지(`BN_TREEBLOG_MASTER_BLUEPRINT_V1_0.zip`)를 제작하는 단계**이다. 따라서:
- **정적/명세 검증 (실행 완료 — `PASS`)**: 139개 문서 생성, JSON/YAML/CSV 문법 검증, OpenAPI 3.1.0 스키마 참조 검증, SQL DDL 정적 검증, Work Order DAG 순환 참조 부재 검증, SHA-256 체크섬 및 ZIP 압축/해제 테스트는 실제 파이썬 검증 스크립트로 **100% 실행 완료(`VERIFIED`)**하였다.
- **런타임 애플리케이션 실행 검증 (후속 에이전트 수행 대상 — `NOT_EXECUTED_YET / AWAITING_IMPLEMENTATION`)**: 실제 Go 백엔드 바이너리 기동, 라이브 PostgreSQL/OpenSearch 컨테이너 연동 테스트, Playwright 브라우저 E2E 테스트, Lighthouse/k6 부하 테스트는 후속 개발 에이전트가 `START_HERE.md`의 **Step 6 ~ Step 9**에서 코드를 구현한 뒤 실행해야 하므로 `15_TESTING/MASTER_TEST_MATRIX.csv`에 허위로 `PASS`를 기재하지 않고 `AWAITING_IMPLEMENTATION`으로 정직하게 표기하였다.

## 2. 외부 승인 및 PO/법무/재무 의사결정이 필요한 사항 (`PENDING_EXTERNAL_APPROVAL`)
1. **실제 운영 기본 도메인 확정**: 명세서 내에서는 `treeblog.example.com` 및 환경변수 `PLATFORM_BASE_DOMAIN`으로 설계되어 있으며, 프로덕션 배포 전 PO가 최종 서비스 도메인명을 확정하고 DNS 레지스트라 권한을 연결해야 한다.
2. **PG사(결제대행사) 가맹점 계약 및 세무 정책 확정**: Stripe / Toss Payments 실운영 API 키 발급, 국가별 부가가치세(VAT) 및 크리에이터 원천징수 세율(한국 3.3% 사업소득세, 미국 W-8BEN/W-9 등)은 재무/법무팀의 승인이 필요하다.
3. **국가별 이용약관·개인정보처리방침·DMCA 대리인 법무 감수**: 기술적 준수 기능은 `11_SECURITY_MODERATION/COMPLIANCE_CHECKLIST.md`에 모두 설계되었으나, 실제 고시할 법률 문안은 변호사 감수가 필요하다.
4. **TreeScan 프로덕션 통합 시점 승인 (제22조)**: Phase 10에서 TreeBlog 측 통합 어댑터 개발 및 계약 테스트까지 완료하더라도, 실제 TreeScan 운영 클러스터와의 연결은 PO의 별도 승인 후에만 활성화한다.

## 3. 남은 기술적 설계 위험 및 완화 전략 (Residual Design Risks)
1. **커스텀 도메인 Let's Encrypt Rate Limit 위험**: 수많은 사용자가 동시에 잘못된 DNS로 TLS 발급을 시도하면 ACME Rate Limit에 걸릴 수 있다 → **완화**: 반드시 `_treeblog-verify` TXT 레코드와 CNAME/A 레코드가 Go 백엔드에서 먼저 `VERIFIED`로 확인된 도메인에 한해서만 Caddy On-Demand TLS가 ACME 요청을 보내도록 게이트를 구축함(`TB-BLOG-003`).
2. **대형 인플루언서 블로거의 새 글 발행 시 알림 Fan-out 부하**: 구독자 50만 명인 블로그가 글을 발행할 때 동기 알림을 생성하면 DB가 잠길 수 있다 → **완화**: 1,000명 단위 비동기 청크 잡(`queue:default`)으로 분할하여 백그라운드에서 점진 발송하도록 설계함(`TB-COMMUNITY-003`).
3. **자체 동영상 트랜스코딩 스토리지/대역폭 비용 급증 위험**: 무제한 4K 영상 직접 업로드를 초기부터 허용하면 인프라 비용이 급증할 수 있다 → **완화**: 초기에는 고효율 WebP/AVIF 이미지와 외부 영상 안전 임베드를 기본 제공하고, 자체 HLS 트랜스코딩은 `FF_NATIVE_VIDEO_TRANSCODE` 기능 플래그 및 쿼터 제한 하에 단계적으로 개방함(`TB-MEDIA-002`).

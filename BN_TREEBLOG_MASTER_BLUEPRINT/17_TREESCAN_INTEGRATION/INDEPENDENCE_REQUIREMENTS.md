# INDEPENDENCE_REQUIREMENTS.md — TreeBlog 100% 독립 운영 보장 및 비종속성 설계 명세 (제1조)

## 1. 독립 개발 및 단독 운영 절대 원칙
제1조에 따라 TreeBlog는 장기적으로 TreeScan 검색엔진의 블로그 서비스로 통합될 예정이나, **최초 개발부터 운영 안정화까지 TreeScan 내부에 일절 종속되지 않는다.**

| 시스템 구성 요소 | TreeBlog 단독 운영 구현 방식 (`FF_TREESCAN_INTEGRATION_ENABLED = false` 기본값) | TreeScan 부재 시 영향도 |
| :--- | :--- | :--- |
| **사용자 웹 & 블로그 렌더링** | 자체 Nuxt 3 SSR (`apps/user-web`) 및 자체 도메인 라우터로 100% 독립 구동 | **영향 없음 (0%)** |
| **회원가입 · 로그인 · 세션** | 자체 `users`, `identities (LOCAL/OAuth)`, `sessions` 및 자체 JWT/MFA로 독립 인증 | **영향 없음 (0%)** |
| **블로그 개설 · 글쓰기 · 발행** | 자체 PostgreSQL 16 (`blogs`, `posts`, `post_revisions`) 및 Tiptap 에디터로 독립 구동 | **영향 없음 (0%)** |
| **블로그 검색 · 추천 피드** | 자체 OpenSearch 2.x 클러스터(`treeblog_posts_v1`, `treeblog_blogs_v1`)로 독립 검색 제공 | **영향 없음 (0%)** |
| **미디어 저장 · 변환 · CDN** | 자체 S3 호환 스토리지 및 `libvips` 미디어 워커로 독립 처리 | **영향 없음 (0%)** |
| **광고 · 멤버십 · 정산 · 어드민** | 자체 수익화 모듈, 복식부기 원장(`ledger_entries`), `apps/admin-console`로 독립 운영 | **영향 없음 (0%)** |

## 2. 금지 사항 및 단계별 통합 원칙
1. **TreeScan 기존 코드·DB 수정 금지**: 본 프로젝트의 어떤 Work Order에서도 TreeScan의 기존 서버 코드나 데이터베이스를 수정하지 않는다.
2. **단계적 전환 순서**: `독립 개발 완료` → `실제 기능 검증(E2E/감사)` → `단독 운영 안정화` → `(PO 별도 승인 후) TreeScan Adapter 활성화`.

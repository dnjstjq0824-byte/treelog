-- Migration: 0010_seed_roles_themes_flags.up.sql
-- Purpose: Seed mandatory platform roles, official base themes, and initial feature flags (including TreeScan independence defaults)

INSERT INTO roles (code, name, description, is_system) VALUES
    ('USER', '일반 회원 및 블로거', '블로그 개설, 글 작성, 이웃/구독/댓글 활동 기본 권한', TRUE),
    ('ADVERTISER', '광고주', '광고 캠페인 및 소재 등록, 성과 리포트 조회 권한', TRUE),
    ('SUPPORT_AGENT', '고객지원 담당자', '신고 티켓 1차 열람 및 문의 응대 권한', TRUE),
    ('TRUST_SAFETY_MOD', '콘텐츠/커뮤니티 모더레이터', '신고 심사, 스팸/유해 게시글 임시조치 및 이의신청 처리 권한', TRUE),
    ('AD_REVIEWER', '광고 심사원', '광고 소재 승인/반려 및 부정 클릭 감사 권한', TRUE),
    ('FINANCE_ADMIN', '정산/재무 관리자', '크리에이터 정산 지급 승인, 보류 및 원장 대사 권한', TRUE),
    ('SUPER_ADMIN', '최고 시스템 관리자', '시스템 설정, Feature Flag, 관리자 권한 부여 및 감사 로그 전체 열람 권한', TRUE)
ON CONFLICT (code) DO NOTHING;

INSERT INTO themes (slug, name, description, layout_type, version, is_active) VALUES
    ('forest-editorial', 'Forest Editorial', '가독성 높은 타이포그래피와 넓은 여백을 제공하는 에디토리얼 기본 테마', 'EDITORIAL', '1.0.0', TRUE),
    ('classic-neighbor', 'Classic Neighbor', '좌측 프로필·카테고리 트리와 친숙한 블로그 홈 구조를 제공하는 클래식 테마', 'CLASSIC_BLOG', '1.0.0', TRUE),
    ('tech-monolith', 'Tech Monolith', '코드 블록, 다크모드, 우측 자동 목차(TOC)에 최적화된 개발/전문가 테마', 'TECH_DOCS', '1.0.0', TRUE),
    ('visual-magazine', 'Visual Magazine', '대형 커버 이미지와 그리드 카드로 매거진 경험을 제공하는 테마', 'MAGAZINE', '1.0.0', TRUE),
    ('photo-atelier', 'Photo Atelier', '사진가와 일러스트레이터를 위한 메이슨리/갤러리 중심 테마', 'PHOTO_GALLERY', '1.0.0', TRUE),
    ('paper-minimal', 'Paper Minimal', '불필요한 장식을 배제하고 글 자체에만 집중하는 미니멀 테마', 'MINIMAL', '1.0.0', TRUE)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO feature_flags (key, description, is_enabled, rollout_percentage) VALUES
    ('FF_CUSTOM_DOMAINS_ACME', '개인 소유 커스텀 도메인 연결 및 자동 Let''s Encrypt TLS 발급 활성화', TRUE, 100),
    ('FF_MONETIZATION_ADS', '자체 광고 네트워크 및 블로거 광고 수익 배분 모듈 활성화 (Phase 8)', FALSE, 0),
    ('FF_MONETIZATION_MEMBERSHIP', '유료 멤버십 및 유료 단건 포스트 결제 활성화 (Phase 8)', FALSE, 0),
    ('FF_NATIVE_VIDEO_TRANSCODE', '자체 원본 동영상 업로드 및 HLS 트랜스코딩 활성화 (기본값: 외부 임베드 우선)', FALSE, 0),
    ('FF_TREESCAN_INTEGRATION_ENABLED', 'TreeScan 검색엔진 색인 동기화 및 통합 OIDC SSO 활성화 (기본값: 독립 운영 FALSE)', FALSE, 0)
ON CONFLICT (key) DO NOTHING;

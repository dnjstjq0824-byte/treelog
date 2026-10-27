# BLOG_CREATION.md — 블로그 개설 및 멀티 블로그 운영 엔진 (제6조)

## 1. 블로그 개설 트랜잭션 및 예약어 정책
1. **핸들(`handle`) 규칙**:
   - 길이 3~30자, 영문 소문자(`a-z`), 숫자(`0-9`), 하이픈(`-`)만 허용하며 시작과 끝은 영숫자여야 한다(`^[a-z0-9][a-z0-9-]{1,28}[a-z0-9]$`).
   - **예약어 차단 목록 (Reserved Handles, 150개+)**: `admin`, `administrator`, `api`, `app`, `assets`, `auth`, `billing`, `blog`, `blogs`, `cdn`, `console`, `dashboard`, `dev`, `docs`, `edge`, `explore`, `feed`, `feeds`, `ftp`, `help`, `imap`, `internal`, `Legal`, `login`, `logout`, `mail`, `media`, `metrics`, `moderation`, `my`, `news`, `ns1`, `ns2`, `oauth`, `official`, `opensearch`, `payments`, `policy`, `pop`, `popular`, `post`, `posts`, `privacy`, `profile`, `public`, `register`, `report`, `root`, `rss`, `search`, `security`, `settings`, `signin`, `signup`, `sitemap`, `smtp`, ` ssl`, `staging`, `static`, `status`, `studio`, `לות`, `support`, `system`, `tag`, `tags`, `team`, `terms`, `test`, `theme`, `themes`, `treeblog`, `treescan`, `treewiki`, `trending`, `upload`, `user`, `users`, `video`, `web`, `webhook`, `www` 등 플랫폼 라우트 및 서브도메인과 충돌할 수 있는 모든 단어를 DB 및 코드 상수로 등록하여 개설을 차단한다.
2. **원자적 개설 파이프라인**:
   - `POST /api/v1/blogs` 호출 시 단일 DB 트랜잭션 내에서 `blogs` 생성 → `blog_members`(`OWNER`) 등록 → `blog_settings` 기본값 생성 → `blog_domains` 기본 서브도메인 등록 → `blog_themes` 선택한 초기 테마 매핑 → `blog_categories` 기본 '미분류' 카테고리 생성을 수행한다.

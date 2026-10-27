# BLOG_THEMES.md — 테마·스킨·레이아웃 커스터마이징 및 CSS 보안 샌드박스 (제7조)

## 1. 제7조 19대 필수 디자인 커스터마이징 기능 구조
`blog_themes` 테이블의 `tokens_override_json`, `layout_config_json`, `sidebar_widgets_json`, `custom_css_sanitized` 컬럼을 통해 다음 19개 기능을 완벽히 제어한다.
1. **기본 스킨 & 테마 선택**: 6종 공식 반응형 베이스 테마(`forest-editorial`, `classic-neighbor`, `tech-monolith`, `visual-magazine`, `photo-atelier`, `paper-minimal`) 즉시 교체.
2. **다크모드**: `LIGHT` 고정, `DARK` 고정, 방문자 OS 설정 연동(`SYSTEM`) 및 방문자 수동 토글 지원.
3. **색상 및 타이포그래피 설정**: 배경색, 강조색, 본문/제목 폰트 패밀리, 글자 크기(`15px`~`19px`), 줄간격(`1.6`~`2.0`), 본문 폭(`640px`~`900px`) 슬라이더 조절.
4. **프로필·헤더·카드·레이아웃 스타일**: 상단형/좌측형 프로필, 게시글 카드형(`CARD`)/목록형(`LIST`)/갤러리형(`GALLERY`)/컴팩트형(`COMPACT`) 전환.
5. **사이드바 위젯 및 메뉴·푸터 편집**: 프로필, 카테고리 트리, 최근글, 인기글, 태그 클라우드, 이웃 목록, 방문자 카운터 위젯의 노출 여부 및 순서 드래그 편집.
6. **실시간 모바일/데스크톱 미리보기 및 원클릭 기본값 복원(`POST /api/v1/blogs/{blogId}/theme/reset`)**.

## 2. 커스텀 CSS 보안 격리 정책
- 서버 사이드 코드(PHP, 템플릿 스크립트)나 사용자 임의 `<script>` 태그 삽입은 원천 불허한다.
- 고급 사용자를 위한 커스텀 CSS는 Go CSS 파서로 분석하여 1) 외부 `@import`, 2) `url()` 내 `javascript:`/`data:text/html` 스킴, 3) 플랫폼 필수 영역(`.tb-platform-header`, `.tb-ad-badge`, `.tb-report-btn`)을 가리거나 변조하는 셀렉터를 제거하고 `.tb-blog-theme-root` 스코프 내부로만 제한한다.

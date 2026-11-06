<template>
  <footer class="tb-site-footer">
    <div class="tb-container">
      <div class="tb-footer-grid">
        <div class="tb-footer-brand">
          <NuxtLink to="/" class="tb-brand" aria-label="TreeBlog 홈">
            <span class="tb-brand-mark" aria-hidden="true">
              <svg viewBox="0 0 32 32" fill="none"><path d="M16 5 L24 14 H19 V24 H13 V14 H8 Z" fill="currentColor" /></svg>
            </span>
            TreeBlog
          </NuxtLink>
          <p class="tb-footer-tagline">생각이 자라는 독립 블로그 플랫폼. 당신의 기록이 숲이 됩니다.</p>
        </div>
        <nav class="tb-footer-links" aria-label="푸터 메뉴">
          <div class="tb-footer-col">
            <span class="tb-footer-col-title">서비스</span>
            <NuxtLink to="/search">글 탐색</NuxtLink>
            <NuxtLink to="/onboarding">블로그 만들기</NuxtLink>
          </div>
          <div class="tb-footer-col">
            <span class="tb-footer-col-title">약관</span>
            <NuxtLink to="/terms">이용약관</NuxtLink>
            <NuxtLink to="/privacy">개인정보처리방침</NuxtLink>
          </div>
        </nav>
      </div>
      <div class="tb-footer-base">
        <span>© 2026 TreeBlog. All rights reserved.</span>
        <span>문의: support@treeblog.example.com</span>
      </div>
    </div>
  </footer>
</template>

export default defineNuxtConfig({
  compatibilityDate: '2025-07-01',
  devtools: { enabled: false },
  css: ['@treeblog/ui-kit/tokens.css', '@treeblog/ui-kit/components.css', '~/assets/main.css'],
  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8080',
      consoleUrl: process.env.NUXT_PUBLIC_CONSOLE_URL || 'http://localhost:5173',
      adminUrl: process.env.NUXT_PUBLIC_ADMIN_URL || 'http://localhost:5174'
    }
  },
  app: {
    head: {
      htmlAttrs: { lang: 'ko' },
      title: 'TreeBlog - 생각이 자라는 곳',
      meta: [
        { name: 'description', content: 'TreeBlog는 깊이 있는 기록을 위한 독립 블로그 플랫폼입니다. 나만의 블로그를 만들고, 이웃과 구독으로 이어지세요.' },
        { name: 'theme-color', content: '#ffffff' },
        { property: 'og:site_name', content: 'TreeBlog' },
        { property: 'og:type', content: 'website' },
        { name: 'twitter:card', content: 'summary' }
      ],
      link: [
        { rel: 'preconnect', href: 'https://cdn.jsdelivr.net' },
        {
          rel: 'stylesheet',
          href: 'https://cdn.jsdelivr.net/gh/orioncactus/pretendard@v1.3.9/dist/web/variable/pretendardvariable-dynamic-subset.min.css'
        },
        { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' }
      ]
    }
  }
})

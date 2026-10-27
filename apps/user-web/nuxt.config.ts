export default defineNuxtConfig({
  compatibilityDate: '2025-07-01',
  devtools: { enabled: false },
  css: ['~/assets/main.css'],
  runtimeConfig: {
    public: { apiBase: 'http://localhost:8080' }
  }
})

<template>
  <div class="tb-console">
    <aside class="tb-console-side">
      <RouterLink to="/" class="tb-console-brand">
        <span class="tb-console-brand-mark" aria-hidden="true">
          <svg viewBox="0 0 32 32" fill="none"><path d="M16 4 L26 10 V16 C26 23 21.5 27.5 16 29 C10.5 27.5 6 23 6 16 V10 Z" fill="currentColor" /></svg>
        </span>
        Admin
      </RouterLink>
      <nav class="tb-console-nav" aria-label="관리 메뉴">
        <RouterLink to="/">대시보드</RouterLink>
        <span class="tb-console-nav-label">플랫폼</span>
        <RouterLink to="/users">사용자</RouterLink>
        <RouterLink to="/blogs">블로그</RouterLink>
        <span class="tb-console-nav-label">신뢰와 안전</span>
        <RouterLink to="/moderation">모더레이션</RouterLink>
        <RouterLink to="/appeals">이의제기</RouterLink>
        <span class="tb-console-nav-label">재무</span>
        <RouterLink to="/payouts">정산 심사</RouterLink>
      </nav>
      <div style="flex: 1" />
      <div>
        <p style="font-size: 0.8rem; color: #9db8a9">{{ auth.user?.display_name }}</p>
        <button class="tb-btn tb-btn-ghost tb-btn-sm" style="color: #fff; padding-left: 0" @click="logout">
          로그아웃
        </button>
      </div>
    </aside>
    <main id="main-content" class="tb-console-main">
      <RouterView :key="route.fullPath" />
    </main>
  </div>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

async function logout(): Promise<void> {
  await auth.logout()
  await router.push('/login')
}
</script>

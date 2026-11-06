<template>
  <header class="tb-site-header">
    <div class="tb-container tb-site-header-inner">
      <NuxtLink to="/" class="tb-brand" aria-label="TreeBlog 홈">
        <span class="tb-brand-mark" aria-hidden="true">
          <svg viewBox="0 0 32 32" fill="none"><path d="M16 5 L24 14 H19 V24 H13 V14 H8 Z" fill="currentColor" /></svg>
        </span>
        TreeBlog
      </NuxtLink>

      <nav class="tb-site-nav" aria-label="주요 메뉴">
        <NuxtLink to="/">홈</NuxtLink>
        <NuxtLink to="/search">탐색</NuxtLink>
        <NuxtLink v-if="isLoggedIn" to="/me/bookmarks">서재</NuxtLink>
        <NuxtLink v-if="isLoggedIn" to="/me/notifications">알림</NuxtLink>
      </nav>

      <div class="tb-header-spacer" />

      <div class="tb-header-actions">
        <button class="tb-icon-btn" aria-label="검색" @click="open = true">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="11" cy="11" r="7" /><path d="M20 20l-3.5-3.5" /></svg>
        </button>
        <template v-if="isLoggedIn">
          <a class="tb-btn tb-btn-secondary tb-btn-sm" :href="consoleUrl">콘솔</a>
          <button class="tb-btn tb-btn-ghost tb-btn-sm" @click="logout">로그아웃</button>
        </template>
        <template v-else>
          <NuxtLink to="/login" class="tb-btn tb-btn-ghost tb-btn-sm">로그인</NuxtLink>
          <NuxtLink to="/register" class="tb-btn tb-btn-primary tb-btn-sm">시작하기</NuxtLink>
        </template>
      </div>
    </div>
    <SearchPalette v-model="open" />
  </header>
</template>

<script setup lang="ts">
const { isLoggedIn, logout } = useAuth()
const config = useRuntimeConfig()
const consoleUrl = config.public.consoleUrl as string
const open = ref(false)

function onKeydown(event: KeyboardEvent): void {
  if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
    event.preventDefault()
    open.value = !open.value
  }
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onUnmounted(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <div class="tb-auth-card">
    <h1 class="tb-auth-title">소셜 로그인</h1>
    <p v-if="state === 'working'" class="tb-hint">로그인 처리 중입니다…</p>
    <p v-else class="tb-error-text" role="alert">{{ message }}</p>
    <div v-if="state === 'failed'" class="tb-form-actions">
      <NuxtLink to="/login" class="tb-btn tb-btn-secondary tb-btn-block">로그인으로</NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
useSeoMeta({ title: '소셜 로그인 | TreeBlog', robots: 'noindex' })
definePageMeta({ layout: 'auth' })

const { finishOAuth } = useAuth()
const state = ref<'working' | 'failed'>('working')
const message = ref('')

onMounted(async () => {
  try {
    await finishOAuth()
  } catch (err) {
    state.value = 'failed'
    message.value = err instanceof Error ? err.message : '로그인에 실패했습니다.'
  }
})
</script>

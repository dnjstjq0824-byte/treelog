<template>
  <div class="tb-auth-card">
    <h1 class="tb-auth-title">결제가 취소되었습니다</h1>
    <p class="tb-hint">{{ message }}</p>
    <div class="tb-form-actions">
      <NuxtLink :to="returnTo" class="tb-btn tb-btn-primary tb-btn-block">블로그로 돌아가기</NuxtLink>
      <NuxtLink to="/" class="tb-btn tb-btn-secondary tb-btn-block">홈으로</NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
useSeoMeta({ title: '결제 취소 | TreeBlog', robots: 'noindex' })
definePageMeta({ layout: 'auth' })

const route = useRoute()
const message = ref('결제가 완료되지 않았습니다. 언제든 다시 시도할 수 있습니다.')
const returnTo = ref('/')

onMounted(() => {
  const text = typeof route.query.message === 'string' ? route.query.message : ''
  if (text) message.value = text
  const blog = typeof route.query.blog === 'string' ? route.query.blog : ''
  if (blog) returnTo.value = `/b/${blog}`
})
</script>

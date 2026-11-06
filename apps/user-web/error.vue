<template>
  <div class="tb-auth-card" style="text-align: center; margin: 48px auto">
    <p class="tb-article-kicker">{{ error?.statusCode || 404 }}</p>
    <h1 class="tb-auth-title">{{ title }}</h1>
    <p class="tb-hint">{{ description }}</p>
    <div class="tb-form-actions">
      <NuxtLink to="/" class="tb-btn tb-btn-primary tb-btn-block">홈으로</NuxtLink>
      <NuxtLink to="/search" class="tb-btn tb-btn-secondary tb-btn-block">글 둘러보기</NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{ error?: { statusCode?: number; message?: string } }>()

const is404 = computed(() => (props.error?.statusCode || 404) === 404)
const title = computed(() =>
  is404.value ? '페이지를 찾을 수 없습니다' : '일시적인 오류가 발생했습니다'
)
const description = computed(() =>
  is404.value
    ? '주소가 바뀌었거나 삭제된 페이지입니다.'
    : '잠시 후 다시 시도해 주세요. 문제가 계속되면 support@treeblog.example.com으로 알려 주세요.'
)

useSeoMeta({ title: '오류 | TreeBlog', robots: 'noindex' })
</script>

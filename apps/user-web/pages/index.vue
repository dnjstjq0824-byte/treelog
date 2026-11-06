<template>
  <div class="tb-hero">
    <p class="tb-hero-edition">
      <span class="tb-hero-edition-dot" aria-hidden="true" />
      <span>TreeBlog · 독립 블로그 플랫폼</span>
    </p>
    <h1 class="tb-hero-title">생각이 자라는 곳, <span class="tb-hero-title-accent">TreeBlog</span></h1>
    <p class="tb-hero-sub">
      독립적인 글을 쓰고, 읽고, 소유하세요. 당신의 기록이 숲이 됩니다.
    </p>
    <div class="tb-hero-actions">
      <NuxtLink to="/register" class="tb-btn tb-btn-primary tb-btn-lg">무료로 시작하기</NuxtLink>
      <NuxtLink to="/search" class="tb-hero-link">글 둘러보기 →</NuxtLink>
    </div>
  </div>

  <div class="tb-section-head">
    <h2 class="tb-section-title">지금 뜨는 글</h2>
    <NuxtLink to="/search" class="tb-nav-link">모두 보기</NuxtLink>
  </div>

  <div v-if="pending">
    <PostSkeleton v-for="index in 5" :key="index" />
  </div>
  <template v-else-if="posts.length > 0">
    <PostCard v-for="post in posts" :key="post.id" :post="post" />
    <div v-if="canShowMore" class="tb-load-more">
      <button class="tb-btn tb-btn-secondary" :disabled="loadingMore" @click="showMore">
        {{ loadingMore ? '불러오는 중…' : '더 보기' }}
      </button>
    </div>
  </template>
  <EmptyState
    v-else
    title="아직 발행된 글이 없어요"
    description="첫 번째 글을 쓸 주인공이 되어 보세요."
  >
    <template #action>
      <NuxtLink to="/register" class="tb-btn tb-btn-primary tb-btn-sm">블로그 시작하기</NuxtLink>
    </template>
  </EmptyState>

  <div class="tb-section-head" style="margin-top: 32px">
    <h2 class="tb-section-title">인기 블로그</h2>
  </div>
  <div v-if="blogsPending">
    <div v-for="index in 3" :key="index" class="tb-skeleton" style="height: 64px" />
  </div>
  <ul v-else-if="blogs.length > 0" class="tb-list">
    <li v-for="blog in blogs" :key="blog.id" class="tb-list-item">
      <div>
        <p class="tb-list-item-title">
          <NuxtLink :to="`/b/${blog.handle}`">{{ blog.title }}</NuxtLink>
        </p>
        <p class="tb-list-item-desc">@{{ blog.handle }} · 글 {{ formatCount(blog.post_count) }}</p>
      </div>
      <NuxtLink :to="`/b/${blog.handle}`" class="tb-btn tb-btn-ghost tb-btn-sm">방문</NuxtLink>
    </li>
  </ul>
</template>

<script setup lang="ts">
import { formatCount } from '@treeblog/ui-kit'
import type { SearchBlogItem, SearchPostItem } from '~/types/api'
import { api } from '~/utils/api'

useSeoMeta({
  title: 'TreeBlog — 생각이 자라는 독립 블로그 플랫폼',
  description:
    'TreeBlog에서 독립적인 글을 쓰고, 읽고, 소유하세요. 광고 없는 깨끗한 읽기 경험을 제공합니다.',
  ogTitle: 'TreeBlog — 생각이 자라는 곳',
  ogDescription: '독립적인 글쓰기와 읽기를 위한 블로그 플랫폼'
})

const posts = ref<SearchPostItem[]>([])
const blogs = ref<SearchBlogItem[]>([])
const pending = ref(true)
const blogsPending = ref(true)
const loadingMore = ref(false)
const limit = ref(10)
const canShowMore = ref(false)
const toast = useToast()

async function fetchTrending(nextLimit: number): Promise<void> {
  try {
    posts.value = await api<SearchPostItem[]>('/api/v1/discover/trending-posts', {
      auth: false,
      query: { limit: nextLimit }
    })
    limit.value = nextLimit
    canShowMore.value = posts.value.length >= nextLimit && nextLimit < 30
  } catch {
    if (nextLimit === 10) posts.value = []
    toast.push('피드를 불러오지 못했습니다', 'error')
  } finally {
    pending.value = false
    loadingMore.value = false
  }
}

async function showMore(): Promise<void> {
  if (loadingMore.value || !canShowMore.value) return
  loadingMore.value = true
  await fetchTrending(Math.min(30, limit.value + 10))
}

onMounted(async () => {
  await fetchTrending(10)
  try {
    blogs.value = await api<SearchBlogItem[]>('/api/v1/discover/popular-blogs', {
      auth: false,
      query: { limit: 5 }
    })
  } catch {
    blogs.value = []
  } finally {
    blogsPending.value = false
  }
})
</script>

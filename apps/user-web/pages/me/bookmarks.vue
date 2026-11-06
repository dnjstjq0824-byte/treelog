<template>
  <div>
    <h1 class="tb-section-title">내 서재</h1>
    <div v-if="pending">
      <div v-for="index in 4" :key="index" class="tb-skeleton" style="height: 72px" />
    </div>
    <template v-else-if="items.length > 0">
      <ul class="tb-list">
        <li v-for="item in items" :key="item.post_id" class="tb-list-item">
          <div>
            <p class="tb-list-item-title">
              <NuxtLink :to="`/b/${item.blog_handle}/${item.slug}`">{{ item.title }}</NuxtLink>
            </p>
            <p class="tb-list-item-desc">
              @{{ item.blog_handle }} · {{ item.folder_name }} · 읽음 {{ item.read_progress_pct }}%
            </p>
          </div>
          <button
            class="tb-btn tb-btn-ghost tb-btn-sm"
            @click="remove(item.post_id)"
          >
            제거
          </button>
        </li>
      </ul>
      <div v-if="canShowMore" class="tb-load-more">
        <button class="tb-btn tb-btn-secondary" :disabled="loadingMore" @click="showMore">
          {{ loadingMore ? '불러오는 중…' : '더 보기' }}
        </button>
      </div>
    </template>
    <EmptyState
      v-else
      title="저장한 글이 없습니다"
      description="마음에 드는 글을 서재에 모아 보세요."
    >
      <template #action>
        <NuxtLink to="/search" class="tb-btn tb-btn-secondary tb-btn-sm">글 둘러보기</NuxtLink>
      </template>
    </EmptyState>
  </div>
</template>

<script setup lang="ts">
import type { BookmarkItem } from '~/types/api'
import { api } from '~/utils/api'

useSeoMeta({ title: '내 서재 | TreeBlog', robots: 'noindex' })
definePageMeta({ middleware: 'auth' })

const items = ref<BookmarkItem[]>([])
const pending = ref(true)
const loadingMore = ref(false)
const limit = ref(30)
const canShowMore = ref(false)
const toast = useToast()

async function fetchPage(nextLimit: number): Promise<void> {
  try {
    items.value = await api<BookmarkItem[]>('/api/v1/me/bookmarks', {
      query: { limit: nextLimit }
    })
    limit.value = nextLimit
    canShowMore.value = items.value.length >= nextLimit && nextLimit < 100
  } catch {
    if (nextLimit === 30) items.value = []
    toast.push('서재를 불러오지 못했습니다', 'error')
  } finally {
    pending.value = false
    loadingMore.value = false
  }
}

async function showMore(): Promise<void> {
  if (!canShowMore.value || loadingMore.value) return
  loadingMore.value = true
  await fetchPage(Math.min(100, limit.value + 30))
}

async function remove(postId: string): Promise<void> {
  try {
    await api(`/api/v1/posts/${postId}/bookmark`, { method: 'DELETE' })
    items.value = items.value.filter((item) => item.post_id !== postId)
  } catch {
    toast.push('제거에 실패했습니다', 'error')
  }
}

onMounted(() => fetchPage(30))
</script>

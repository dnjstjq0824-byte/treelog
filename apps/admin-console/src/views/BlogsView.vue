<template>
  <div>
    <div class="tb-console-head">
      <h1 class="tb-console-title">블로그</h1>
      <select v-model="status" class="tb-select" style="max-width: 160px" @change="reload">
        <option value="">전체 상태</option>
        <option value="ACTIVE">활성</option>
        <option value="SUSPENDED">정지</option>
      </select>
    </div>
    <div v-if="pending">
      <div v-for="index in 6" :key="index" class="tb-skeleton" style="height: 64px" />
    </div>
    <ul v-else-if="items.length > 0" class="tb-list">
      <li v-for="item in items" :key="item.id" class="tb-list-item">
        <div>
          <p class="tb-list-item-title">{{ item.title }} <span class="tb-hint">@{{ item.handle }}</span></p>
          <p class="tb-list-item-desc">
            {{ item.status }} · 글 {{ item.post_count }}개 · {{ formatDate(item.created_at) }} 개설
          </p>
        </div>
        <span class="tb-badge">{{ item.status }}</span>
      </li>
    </ul>
    <div v-else class="tb-empty">
      <p class="tb-empty-title">블로그가 없습니다</p>
    </div>
    <div v-if="hasMore" class="tb-load-more">
      <button class="tb-btn tb-btn-secondary" :disabled="loadingMore" @click="loadMore">
        {{ loadingMore ? '불러오는 중…' : '더 보기' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { formatDate } from '@treeblog/ui-kit'
import { onMounted, ref } from 'vue'
import { api, friendlyMessage } from '../lib/api'
import type { AdminBlog } from '../lib/types'
import { useToastStore } from '../stores/toast'

const toast = useToastStore()
const items = ref<AdminBlog[]>([])
const status = ref('')
const offset = ref(0)
const hasMore = ref(false)
const pending = ref(true)
const loadingMore = ref(false)
const PAGE_SIZE = 30

async function fetchPage(nextOffset: number, append: boolean): Promise<void> {
  try {
    const page = await api<AdminBlog[]>('/api/v1/admin/blogs', {
      query: { limit: PAGE_SIZE, offset: nextOffset, status: status.value || undefined }
    })
    items.value = append ? [...items.value, ...page] : page
    offset.value = nextOffset
    hasMore.value = page.length >= PAGE_SIZE
  } catch (err) {
    if (!append) items.value = []
    toast.push(friendlyMessage(err), 'error')
  } finally {
    pending.value = false
    loadingMore.value = false
  }
}

async function reload(): Promise<void> {
  pending.value = true
  await fetchPage(0, false)
}

async function loadMore(): Promise<void> {
  if (!hasMore.value || loadingMore.value) return
  loadingMore.value = true
  await fetchPage(offset.value + PAGE_SIZE, true)
}

onMounted(() => fetchPage(0, false))
</script>

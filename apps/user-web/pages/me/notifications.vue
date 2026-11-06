<template>
  <div>
    <div class="tb-section-head">
      <h1 class="tb-section-title">알림</h1>
      <button
        v-if="items.some((item) => !item.is_read)"
        class="tb-btn tb-btn-ghost tb-btn-sm"
        @click="markAllRead"
      >
        모두 읽음
      </button>
    </div>
    <div v-if="pending">
      <div v-for="index in 5" :key="index" class="tb-skeleton" style="height: 56px" />
    </div>
    <template v-else-if="items.length > 0">
      <ul class="tb-list">
        <li
          v-for="item in items"
          :key="item.id"
          class="tb-list-item"
          :class="{ 'tb-list-item-unread': !item.is_read }"
        >
          <div>
            <p class="tb-list-item-title">{{ item.title }}</p>
            <p v-if="item.body" class="tb-list-item-desc">{{ item.body }}</p>
            <time class="tb-hint" :datetime="item.created_at">{{ timeAgo(item.created_at) }}</time>
          </div>
          <button
            v-if="!item.is_read"
            class="tb-btn tb-btn-ghost tb-btn-sm"
            @click="markRead(item.id)"
          >
            읽음
          </button>
        </li>
      </ul>
      <div v-if="hasMore" class="tb-load-more">
        <button class="tb-btn tb-btn-secondary" :disabled="loadingMore" @click="loadMore">
          {{ loadingMore ? '불러오는 중…' : '더 보기' }}
        </button>
      </div>
    </template>
    <EmptyState v-else title="알림이 없습니다" />
  </div>
</template>

<script setup lang="ts">
import { timeAgo } from '@treeblog/ui-kit'
import type { NotificationItem } from '~/types/api'
import { api, apiRaw } from '~/utils/api'

useSeoMeta({ title: '알림 | TreeBlog', robots: 'noindex' })
definePageMeta({ middleware: 'auth' })

const items = ref<NotificationItem[]>([])
const cursor = ref('')
const hasMore = ref(false)
const pending = ref(true)
const loadingMore = ref(false)
const toast = useToast()

async function fetchPage(append: boolean): Promise<void> {
  try {
    const envelope = await apiRaw<NotificationItem[]>('/api/v1/me/notifications/', {
      query: append && cursor.value ? { limit: 30, cursor: cursor.value } : { limit: 30 }
    })
    items.value = append ? [...items.value, ...envelope.data] : envelope.data
    cursor.value = envelope.next_cursor || ''
    hasMore.value = cursor.value !== ''
  } catch {
    if (!append) items.value = []
    toast.push('알림을 불러오지 못했습니다', 'error')
  } finally {
    pending.value = false
    loadingMore.value = false
  }
}

async function loadMore(): Promise<void> {
  if (!hasMore.value || loadingMore.value) return
  loadingMore.value = true
  await fetchPage(true)
}

async function markRead(id: string): Promise<void> {
  try {
    await api(`/api/v1/me/notifications/${id}/read`, { method: 'POST' })
    const target = items.value.find((item) => item.id === id)
    if (target) target.is_read = true
  } catch {
    toast.push('읽음 처리에 실패했습니다', 'error')
  }
}

async function markAllRead(): Promise<void> {
  try {
    await api('/api/v1/me/notifications/read-all', { method: 'POST' })
    items.value.forEach((item) => {
      item.is_read = true
    })
  } catch {
    toast.push('읽음 처리에 실패했습니다', 'error')
  }
}

onMounted(() => fetchPage(false))
</script>

<template>
  <div>
    <div class="tb-console-head">
      <h1 class="tb-console-title">모더레이션 케이스</h1>
      <select v-model="status" class="tb-select" style="max-width: 160px" @change="reload">
        <option value="">전체 상태</option>
        <option value="OPEN">열림</option>
        <option value="IN_REVIEW">검토 중</option>
        <option value="RESOLVED">해결됨</option>
      </select>
    </div>
    <div v-if="pending">
      <div v-for="index in 6" :key="index" class="tb-skeleton" style="height: 64px" />
    </div>
    <ul v-else-if="items.length > 0" class="tb-list">
      <li v-for="item in items" :key="item.id" class="tb-list-item">
        <div>
          <p class="tb-list-item-title">
            <RouterLink :to="`/moderation/${item.id}`">{{ item.target_type }} {{ item.target_id.slice(0, 8) }}</RouterLink>
          </p>
          <p class="tb-list-item-desc">
            {{ item.reason_category }} · 신고 {{ item.report_count }}건 ·
            우선순위 {{ item.priority_score }} · {{ formatDateTime(item.created_at) }}
          </p>
        </div>
        <span class="tb-badge">{{ item.status }}</span>
      </li>
    </ul>
    <div v-else class="tb-empty">
      <p class="tb-empty-title">케이스가 없습니다</p>
    </div>
    <div v-if="hasMore" class="tb-load-more">
      <button class="tb-btn tb-btn-secondary" :disabled="loadingMore" @click="loadMore">
        {{ loadingMore ? '불러오는 중…' : '더 보기' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { formatDateTime } from '@treeblog/ui-kit'
import { onMounted, ref } from 'vue'
import { api, friendlyMessage } from '../lib/api'
import type { ModerationCase } from '../lib/types'
import { useToastStore } from '../stores/toast'

const toast = useToastStore()
const items = ref<ModerationCase[]>([])
const status = ref('')
const offset = ref(0)
const hasMore = ref(false)
const pending = ref(true)
const loadingMore = ref(false)
const PAGE_SIZE = 30

async function fetchPage(nextOffset: number, append: boolean): Promise<void> {
  try {
    const page = await api<ModerationCase[]>('/api/v1/admin/moderation/cases', {
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

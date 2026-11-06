<template>
  <div>
    <div class="tb-console-head">
      <h1 class="tb-console-title">이의제기</h1>
      <select v-model="status" class="tb-select" style="max-width: 160px" @change="reload">
        <option value="">전체 상태</option>
        <option value="PENDING">대기 중</option>
        <option value="APPROVED_RESTORED">인용</option>
        <option value="REJECTED_UPHELD">기각</option>
      </select>
    </div>
    <div v-if="pending">
      <div v-for="index in 6" :key="index" class="tb-skeleton" style="height: 80px" />
    </div>
    <ul v-else-if="items.length > 0" class="tb-list">
      <li v-for="item in items" :key="item.id" class="tb-list-item">
        <div style="flex: 1">
          <p class="tb-list-item-title">
            <RouterLink :to="`/moderation/${item.case_id}`">케이스 {{ item.case_id.slice(0, 8) }}</RouterLink>
            <span class="tb-badge">{{ item.status }}</span>
          </p>
          <p class="tb-list-item-desc">{{ item.appeal_statement }}</p>
          <p class="tb-hint">{{ formatDateTime(item.created_at) }}</p>
          <div v-if="deciding === item.id" style="margin-top: 8px; display: flex; gap: 6px; flex-wrap: wrap">
            <input v-model="note" class="tb-input" style="flex: 1; min-width: 200px" placeholder="결정 사유 메모" maxlength="500" />
            <button class="tb-btn tb-btn-primary tb-btn-sm" :disabled="busy" @click="decide(item.id, 'APPROVED_RESTORED')">
              인용 (복원)
            </button>
            <button class="tb-btn tb-btn-secondary tb-btn-sm" :disabled="busy" @click="decide(item.id, 'REJECTED_UPHELD')">
              기각 (유지)
            </button>
          </div>
        </div>
        <button
          v-if="deciding !== item.id && item.status === 'PENDING'"
          class="tb-btn tb-btn-secondary tb-btn-sm"
          @click="deciding = item.id"
        >
          결정하기
        </button>
      </li>
    </ul>
    <div v-else class="tb-empty">
      <p class="tb-empty-title">이의제기가 없습니다</p>
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
import type { Appeal } from '../lib/types'
import { useToastStore } from '../stores/toast'

const toast = useToastStore()
const items = ref<Appeal[]>([])
const status = ref('')
const offset = ref(0)
const hasMore = ref(false)
const pending = ref(true)
const loadingMore = ref(false)
const deciding = ref('')
const note = ref('')
const busy = ref(false)
const PAGE_SIZE = 30

async function fetchPage(nextOffset: number, append: boolean): Promise<void> {
  try {
    const page = await api<Appeal[]>('/api/v1/admin/moderation/appeals', {
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

async function decide(appealId: string, decision: string): Promise<void> {
  busy.value = true
  try {
    const updated = await api<Appeal>(`/api/v1/admin/moderation/appeals/${appealId}/decision`, {
      method: 'POST',
      body: { decision, note: note.value.trim() }
    })
    const index = items.value.findIndex((item) => item.id === appealId)
    if (index >= 0) items.value[index] = updated
    deciding.value = ''
    note.value = ''
    toast.push('결정했습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  } finally {
    busy.value = false
  }
}

onMounted(() => fetchPage(0, false))
</script>

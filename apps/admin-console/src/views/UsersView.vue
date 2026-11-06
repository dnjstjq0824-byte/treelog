<template>
  <div>
    <div class="tb-console-head">
      <h1 class="tb-console-title">사용자</h1>
    </div>
    <form style="display: flex; gap: 8px; margin-bottom: 12px; flex-wrap: wrap" @submit.prevent="search">
      <input v-model="draft" class="tb-input" style="max-width: 280px" placeholder="이메일/이름 검색" />
      <select v-model="status" class="tb-select" style="max-width: 160px" @change="search">
        <option value="">전체 상태</option>
        <option value="ACTIVE">활성</option>
        <option value="SUSPENDED">정지</option>
        <option value="PENDING_VERIFICATION">인증 대기</option>
      </select>
      <button class="tb-btn tb-btn-secondary tb-btn-sm">검색</button>
    </form>
    <div v-if="pending">
      <div v-for="index in 6" :key="index" class="tb-skeleton" style="height: 64px" />
    </div>
    <ul v-else-if="items.length > 0" class="tb-list">
      <li v-for="item in items" :key="item.id" class="tb-list-item">
        <div>
          <p class="tb-list-item-title">{{ item.display_name }} <span class="tb-hint">@{{ item.username }}</span></p>
          <p class="tb-list-item-desc">
            {{ item.email }} · {{ item.status }} · 신뢰 {{ item.trust_tier }} ·
            {{ item.roles.join(', ') || '역할 없음' }}
          </p>
        </div>
        <button
          v-if="item.status === 'SUSPENDED'"
          class="tb-btn tb-btn-secondary tb-btn-sm"
          @click="unsuspend(item.id)"
        >
          정지 해제
        </button>
        <button v-else class="tb-btn tb-btn-danger tb-btn-sm" @click="suspend(item.id)">정지</button>
      </li>
    </ul>
    <div v-else class="tb-empty">
      <p class="tb-empty-title">사용자가 없습니다</p>
    </div>
    <div v-if="hasMore" class="tb-load-more">
      <button class="tb-btn tb-btn-secondary" :disabled="loadingMore" @click="loadMore">
        {{ loadingMore ? '불러오는 중…' : '더 보기' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, friendlyMessage } from '../lib/api'
import type { AdminUser } from '../lib/types'
import { useToastStore } from '../stores/toast'

const toast = useToastStore()
const items = ref<AdminUser[]>([])
const draft = ref('')
const query = ref('')
const status = ref('')
const offset = ref(0)
const hasMore = ref(false)
const pending = ref(true)
const loadingMore = ref(false)
const PAGE_SIZE = 30

async function fetchPage(nextOffset: number, append: boolean): Promise<void> {
  try {
    const page = await api<AdminUser[]>('/api/v1/admin/users', {
      query: {
        limit: PAGE_SIZE,
        offset: nextOffset,
        q: query.value || undefined,
        status: status.value || undefined
      }
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

async function search(): Promise<void> {
  query.value = draft.value.trim()
  pending.value = true
  await fetchPage(0, false)
}

async function loadMore(): Promise<void> {
  if (!hasMore.value || loadingMore.value) return
  loadingMore.value = true
  await fetchPage(offset.value + PAGE_SIZE, true)
}

async function suspend(userId: string): Promise<void> {
  if (!confirm('이 사용자를 정지할까요?')) return
  try {
    await api(`/api/v1/admin/users/${userId}/suspend`, { method: 'POST', body: {} })
    const target = items.value.find((item) => item.id === userId)
    if (target) target.status = 'SUSPENDED'
    toast.push('정지했습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  }
}

async function unsuspend(userId: string): Promise<void> {
  try {
    await api(`/api/v1/admin/users/${userId}/unsuspend`, { method: 'POST', body: {} })
    const target = items.value.find((item) => item.id === userId)
    if (target) target.status = 'ACTIVE'
    toast.push('정지를 해제했습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  }
}

onMounted(() => fetchPage(0, false))
</script>

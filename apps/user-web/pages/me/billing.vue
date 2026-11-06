<template>
  <div>
    <h1 class="tb-section-title">결제 내역</h1>
    <div v-if="pending">
      <div v-for="index in 4" :key="index" class="tb-skeleton" style="height: 52px" />
    </div>
    <template v-else-if="items.length > 0">
      <ul class="tb-list">
        <li v-for="item in items" :key="item.id" class="tb-list-item">
          <div>
            <p class="tb-list-item-title">{{ orderLabel(item.order_type) }}</p>
            <p class="tb-list-item-desc">
              {{ formatDate(item.created_at) }} ·
              {{ item.gross_amount_minor.toLocaleString() }}원 ·
              주문 {{ item.id.slice(0, 8) }}
            </p>
          </div>
          <span class="tb-badge" :class="item.status === 'PAID' ? 'tb-badge-success' : ''">
            {{ statusLabel(item.status) }}
          </span>
        </li>
      </ul>
      <div v-if="hasMore" class="tb-load-more">
        <button class="tb-btn tb-btn-secondary" :disabled="loadingMore" @click="loadMore">
          {{ loadingMore ? '불러오는 중…' : '더 보기' }}
        </button>
      </div>
    </template>
    <EmptyState v-else title="결제 내역이 없습니다" />
  </div>
</template>

<script setup lang="ts">
import { formatDate } from '@treeblog/ui-kit'
import type { Order } from '~/types/api'
import { api } from '~/utils/api'

useSeoMeta({ title: '결제 내역 | TreeBlog', robots: 'noindex' })
definePageMeta({ middleware: 'auth' })

const orderLabels: Record<string, string> = {
  MEMBERSHIP_SUB: '멤버십 구독',
  SINGLE_POST_UNLOCK: '글 열람',
  READER_DONATION: '후원',
  AD_PREPAID_TOPUP: '광고 충전'
}

const statusLabels: Record<string, string> = {
  PENDING: '결제 대기',
  PAID: '결제 완료',
  CANCELED: '취소됨'
}

function orderLabel(type: string): string {
  return orderLabels[type] || type
}

function statusLabel(status: string): string {
  return statusLabels[status] || status
}

const items = ref<Order[]>([])
const offset = ref(0)
const hasMore = ref(false)
const pending = ref(true)
const loadingMore = ref(false)
const toast = useToast()
const PAGE_SIZE = 30

async function fetchPage(nextOffset: number, append: boolean): Promise<void> {
  try {
    const page = await api<Order[]>('/api/v1/billing/orders', {
      query: { limit: PAGE_SIZE, offset: nextOffset }
    })
    items.value = append ? [...items.value, ...page] : page
    offset.value = nextOffset
    hasMore.value = page.length >= PAGE_SIZE
  } catch {
    if (!append) items.value = []
    toast.push('결제 내역을 불러오지 못했습니다', 'error')
  } finally {
    pending.value = false
    loadingMore.value = false
  }
}

async function loadMore(): Promise<void> {
  if (!hasMore.value || loadingMore.value) return
  loadingMore.value = true
  await fetchPage(offset.value + PAGE_SIZE, true)
}

onMounted(() => fetchPage(0, false))
</script>

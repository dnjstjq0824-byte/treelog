<template>
  <div>
    <div class="tb-console-head">
      <h1 class="tb-console-title">정산 심사</h1>
      <select v-model="status" class="tb-select" style="max-width: 160px" @change="reload">
        <option value="">전체 상태</option>
        <option value="PENDING">대기 중</option>
        <option value="APPROVED">승인됨</option>
        <option value="HOLD">보류</option>
        <option value="REJECTED">거절됨</option>
        <option value="PAID">지급 완료</option>
      </select>
    </div>
    <div v-if="pending">
      <div v-for="index in 6" :key="index" class="tb-skeleton" style="height: 80px" />
    </div>
    <ul v-else-if="items.length > 0" class="tb-list">
      <li v-for="item in items" :key="item.id" class="tb-list-item">
        <div style="flex: 1">
          <p class="tb-list-item-title">
            {{ item.net_amount_minor.toLocaleString() }}원
            <span class="tb-badge">{{ item.status }}</span>
          </p>
          <p class="tb-list-item-desc">
            총 {{ item.gross_amount_minor.toLocaleString() }}원 ·
            원천세 {{ item.withholding_tax_minor.toLocaleString() }}원 ·
            {{ item.bank_code }} {{ item.account_holder }} ·
            {{ formatDateTime(item.requested_at) }}
          </p>
          <code v-if="revealed[item.id]" class="tb-code">
            {{ revealed[item.id]?.bank_code }} {{ revealed[item.id]?.bank_account }} ({{ revealed[item.id]?.account_holder }})
          </code>
          <div v-if="deciding === item.id" style="margin-top: 8px; display: flex; gap: 6px; flex-wrap: wrap">
            <input v-model="reason" class="tb-input" style="flex: 1; min-width: 200px" placeholder="결정 사유" maxlength="500" />
            <button class="tb-btn tb-btn-primary tb-btn-sm" :disabled="busy" @click="decide(item.id, 'APPROVE')">
              승인
            </button>
            <button class="tb-btn tb-btn-secondary tb-btn-sm" :disabled="busy" @click="decide(item.id, 'HOLD')">
              보류
            </button>
            <button class="tb-btn tb-btn-danger tb-btn-sm" :disabled="busy" @click="decide(item.id, 'REJECT')">
              거절
            </button>
          </div>
        </div>
        <div style="display: flex; gap: 6px; flex-wrap: wrap; justify-content: flex-end">
          <button class="tb-btn tb-btn-ghost tb-btn-sm" @click="reveal(item.id)">계좌 확인</button>
          <button
            v-if="item.status === 'PENDING'"
            class="tb-btn tb-btn-secondary tb-btn-sm"
            @click="deciding = item.id"
          >
            심사
          </button>
          <button
            v-if="item.status === 'APPROVED'"
            class="tb-btn tb-btn-primary tb-btn-sm"
            @click="markPaid(item.id)"
          >
            지급 완료 처리
          </button>
        </div>
      </li>
    </ul>
    <div v-else class="tb-empty">
      <p class="tb-empty-title">정산 요청이 없습니다</p>
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
import type { Payout, PayoutAccount } from '../lib/types'
import { useToastStore } from '../stores/toast'

const toast = useToastStore()
const items = ref<Payout[]>([])
const revealed = ref<Record<string, PayoutAccount>>({})
const status = ref('')
const offset = ref(0)
const hasMore = ref(false)
const pending = ref(true)
const loadingMore = ref(false)
const deciding = ref('')
const reason = ref('')
const busy = ref(false)
const PAGE_SIZE = 30

async function fetchPage(nextOffset: number, append: boolean): Promise<void> {
  try {
    const page = await api<Payout[]>('/api/v1/admin/payouts/', {
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

async function reveal(payoutId: string): Promise<void> {
  try {
    revealed.value[payoutId] = await api<PayoutAccount>(`/api/v1/admin/payouts/${payoutId}/account`)
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  }
}

async function decide(payoutId: string, decision: string): Promise<void> {
  busy.value = true
  try {
    const updated = await api<Payout>(`/api/v1/admin/payouts/${payoutId}/decision`, {
      method: 'POST',
      body: { decision, reason: reason.value.trim() }
    })
    const index = items.value.findIndex((item) => item.id === payoutId)
    if (index >= 0) items.value[index] = updated
    deciding.value = ''
    reason.value = ''
    toast.push('심사했습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  } finally {
    busy.value = false
  }
}

async function markPaid(payoutId: string): Promise<void> {
  if (!confirm('지급 완료로 처리할까요?')) return
  try {
    const updated = await api<Payout>(`/api/v1/admin/payouts/${payoutId}/mark-paid`, {
      method: 'POST',
      body: {}
    })
    const index = items.value.findIndex((item) => item.id === payoutId)
    if (index >= 0) items.value[index] = updated
    toast.push('지급 완료 처리했습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  }
}

onMounted(() => fetchPage(0, false))
</script>

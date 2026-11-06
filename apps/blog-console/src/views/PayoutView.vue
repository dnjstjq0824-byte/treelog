<template>
  <div>
    <div class="tb-console-head">
      <h1 class="tb-console-title">정산 요청</h1>
    </div>
    <div style="max-width: 520px">
      <p class="tb-hint">정산 가능한 예치금 범위에서 출금을 요청합니다. 최소 10,000원부터 가능합니다.</p>
      <form style="margin-top: 12px" @submit.prevent="submit">
        <div class="tb-field">
          <label class="tb-label" for="payout-amount">요청 금액 (원)</label>
          <input id="payout-amount" v-model.number="form.gross_amount_minor" class="tb-input" type="number" min="10000" step="1000" required />
        </div>
        <div class="tb-field">
          <label class="tb-label" for="payout-bank">은행 코드 (3자리)</label>
          <input id="payout-bank" v-model="form.bank_code" class="tb-input" maxlength="3" minlength="3" required placeholder="예: 004" />
        </div>
        <div class="tb-field">
          <label class="tb-label" for="payout-account">계좌번호</label>
          <input id="payout-account" v-model="form.bank_account" class="tb-input" required autocomplete="off" />
        </div>
        <div class="tb-field">
          <label class="tb-label" for="payout-holder">예금주</label>
          <input id="payout-holder" v-model="form.account_holder" class="tb-input" required />
        </div>
        <div class="tb-form-actions">
          <button class="tb-btn tb-btn-primary" :disabled="busy">
            {{ busy ? '요청 중…' : '정산 요청' }}
          </button>
        </div>
      </form>
      <div v-if="result" class="tb-list-item" style="border: 1px solid var(--tb-border); border-radius: var(--tb-radius-lg); margin-top: 16px">
        <div>
          <p class="tb-list-item-title">요청 접수됨</p>
          <p class="tb-list-item-desc">
            실수령 {{ result.net_amount_minor.toLocaleString() }}원 (원천세 {{ result.withholding_tax_minor.toLocaleString() }}원) ·
            상태 {{ result.status }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import { api, friendlyMessage } from '../lib/api'
import type { Payout } from '../lib/types'
import { useToastStore } from '../stores/toast'

const route = useRoute()
const toast = useToastStore()
const blogId = computed(() => String(route.params.blogId || ''))

const form = ref({
  gross_amount_minor: 10000,
  bank_code: '',
  bank_account: '',
  account_holder: ''
})
const busy = ref(false)
const result = ref<Payout | null>(null)

async function submit(): Promise<void> {
  busy.value = true
  result.value = null
  try {
    result.value = await api<Payout>('/api/v1/billing/payouts', {
      method: 'POST',
      body: {
        blog_id: blogId.value,
        gross_amount_minor: form.value.gross_amount_minor,
        bank_code: form.value.bank_code.trim(),
        bank_account: form.value.bank_account.trim(),
        account_holder: form.value.account_holder.trim()
      }
    })
    toast.push('정산을 요청했습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  } finally {
    busy.value = false
  }
}
</script>

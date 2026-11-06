<template>
  <div class="tb-auth-card">
    <h1 class="tb-auth-title">결제 확인</h1>
    <p v-if="state === 'working'" class="tb-hint">결제를 승인하는 중입니다…</p>
    <template v-else-if="state === 'done'">
      <p class="tb-hint" style="color: var(--tb-success)">결제가 완료되었습니다. 멤버십이 활성화되었습니다.</p>
      <div class="tb-form-actions">
        <NuxtLink :to="returnTo" class="tb-btn tb-btn-primary tb-btn-block">돌아가기</NuxtLink>
      </div>
    </template>
    <template v-else>
      <p class="tb-error-text" role="alert">{{ message }}</p>
      <div class="tb-form-actions">
        <NuxtLink to="/" class="tb-btn tb-btn-secondary tb-btn-block">홈으로</NuxtLink>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { api, friendlyMessage } from '~/utils/api'

useSeoMeta({ title: '결제 확인 | TreeBlog', robots: 'noindex' })
definePageMeta({ layout: 'auth' })

const route = useRoute()
const state = ref<'working' | 'done' | 'failed'>('working')
const message = ref('')
const returnTo = ref('/')

onMounted(async () => {
  const paymentKey = typeof route.query.paymentKey === 'string' ? route.query.paymentKey : ''
  const orderId = typeof route.query.orderId === 'string' ? route.query.orderId : ''
  const amount = typeof route.query.amount === 'string' ? Number(route.query.amount) : NaN
  const blog = typeof route.query.blog === 'string' ? route.query.blog : ''
  if (blog) returnTo.value = `/b/${blog}`
  if (!paymentKey || !orderId || !Number.isFinite(amount)) {
    state.value = 'failed'
    message.value = '결제 정보가 올바르지 않습니다.'
    return
  }
  try {
    await api('/api/v1/billing/payments/confirm', {
      method: 'POST',
      body: { paymentKey, orderId, amount }
    })
    state.value = 'done'
  } catch (err) {
    state.value = 'failed'
    message.value = friendlyMessage(err)
  }
})
</script>

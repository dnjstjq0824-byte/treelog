<template>
  <div class="tb-auth-card">
    <h1 class="tb-auth-title">이메일 인증</h1>
    <p class="tb-hint">{{ email }} 주소로 보낸 6자리 코드를 입력하세요.</p>
    <form @submit.prevent="submit" style="margin-top: 16px">
      <div class="tb-field">
        <label class="tb-label" for="verify-code">인증 코드</label>
        <input
          id="verify-code"
          v-model="code"
          class="tb-input"
          inputmode="numeric"
          maxlength="6"
          minlength="6"
          required
          autocomplete="one-time-code"
          placeholder="6자리 숫자"
        />
      </div>
      <p v-if="error" class="tb-error-text" role="alert">{{ error }}</p>
      <div class="tb-form-actions">
        <button class="tb-btn tb-btn-primary tb-btn-block" :disabled="busy">
          {{ busy ? '인증 중…' : '인증하기' }}
        </button>
      </div>
    </form>
    <div class="tb-form-actions">
      <button class="tb-btn tb-btn-ghost tb-btn-sm" :disabled="resending" @click="resend">
        {{ resending ? '전송 중…' : '코드 다시 보내기' }}
      </button>
      <NuxtLink to="/login" class="tb-btn tb-btn-ghost tb-btn-sm">로그인으로</NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
import { api, friendlyMessage } from '~/utils/api'

useSeoMeta({ title: '이메일 인증 | TreeBlog', robots: 'noindex' })
definePageMeta({ layout: 'auth' })

const route = useRoute()
const toast = useToast()
const email = ref(typeof route.query.email === 'string' ? route.query.email : '')
const code = ref('')
const busy = ref(false)
const resending = ref(false)
const error = ref('')

async function submit(): Promise<void> {
  busy.value = true
  error.value = ''
  try {
    await api('/api/v1/auth/verify-email', {
      method: 'POST',
      auth: false,
      body: { email: email.value.trim(), code: code.value.trim() }
    })
    toast.push('이메일 인증이 완료되었습니다', 'success')
    await navigateTo('/login')
  } catch (err) {
    error.value = friendlyMessage(err)
  } finally {
    busy.value = false
  }
}

async function resend(): Promise<void> {
  resending.value = true
  try {
    await api('/api/v1/auth/verify-email/resend', {
      method: 'POST',
      auth: false,
      body: { email: email.value.trim() }
    })
    toast.push('인증 코드를 다시 보냈습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  } finally {
    resending.value = false
  }
}
</script>

<template>
  <div class="tb-auth-card">
    <h1 class="tb-auth-title">비밀번호 재설정</h1>
    <!-- Mode 2: set a new password from the emailed link (?token=…). -->
    <form v-if="token" @submit.prevent="confirmReset" style="margin-top: 16px">
      <div class="tb-field">
        <label class="tb-label" for="reset-new">새 비밀번호 (8~72자)</label>
        <input
          id="reset-new"
          v-model="newPassword"
          class="tb-input"
          type="password"
          required
          minlength="8"
          maxlength="72"
          autocomplete="new-password"
        />
      </div>
      <p v-if="error" class="tb-error-text" role="alert">{{ error }}</p>
      <div class="tb-form-actions">
        <button class="tb-btn tb-btn-primary tb-btn-block" :disabled="busy">
          {{ busy ? '변경 중…' : '비밀번호 변경' }}
        </button>
      </div>
    </form>
    <!-- Mode 1: request a reset email. -->
    <form v-else-if="!sent" @submit.prevent="requestReset" style="margin-top: 16px">
      <div class="tb-field">
        <label class="tb-label" for="reset-email">가입한 이메일</label>
        <input
          id="reset-email"
          v-model="email"
          class="tb-input"
          type="email"
          required
          autocomplete="email"
        />
      </div>
      <p v-if="error" class="tb-error-text" role="alert">{{ error }}</p>
      <div class="tb-form-actions">
        <button class="tb-btn tb-btn-primary tb-btn-block" :disabled="busy">
          {{ busy ? '전송 중…' : '재설정 메일 보내기' }}
        </button>
      </div>
    </form>
    <template v-else>
      <p class="tb-hint">가입된 이메일이라면 재설정 메일을 보냈습니다. 메일함을 확인해 주세요.</p>
      <div class="tb-form-actions">
        <NuxtLink to="/login" class="tb-btn tb-btn-secondary tb-btn-block">로그인으로</NuxtLink>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { api, friendlyMessage } from '~/utils/api'

useSeoMeta({ title: '비밀번호 재설정 | TreeBlog', robots: 'noindex' })
definePageMeta({ layout: 'auth', middleware: 'guest' })

const route = useRoute()
const toast = useToast()
const token = typeof route.query.token === 'string' ? route.query.token : ''
const email = ref('')
const newPassword = ref('')
const busy = ref(false)
const sent = ref(false)
const error = ref('')

async function requestReset(): Promise<void> {
  busy.value = true
  error.value = ''
  try {
    await api('/api/v1/auth/password/forgot', {
      method: 'POST',
      body: { email: email.value.trim() },
      auth: false
    })
    sent.value = true
  } catch (err) {
    error.value = friendlyMessage(err)
  } finally {
    busy.value = false
  }
}

async function confirmReset(): Promise<void> {
  busy.value = true
  error.value = ''
  try {
    await api('/api/v1/auth/password/reset', {
      method: 'POST',
      body: { token, new_password: newPassword.value },
      auth: false
    })
    toast.push('비밀번호를 변경했습니다. 로그인해 주세요', 'success')
    await navigateTo('/login')
  } catch (err) {
    error.value = friendlyMessage(err)
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div class="tb-auth-card">
    <h1 class="tb-auth-title">로그인</h1>
    <p class="tb-hint">TreeBlog 계정으로 로그인하세요.</p>
    <form @submit.prevent="submit" style="margin-top: 16px">
      <div class="tb-field">
        <label class="tb-label" for="login-email">이메일</label>
        <input
          id="login-email"
          v-model="email"
          class="tb-input"
          type="email"
          required
          autocomplete="email"
        />
      </div>
      <div class="tb-field">
        <label class="tb-label" for="login-password">비밀번호</label>
        <input
          id="login-password"
          v-model="password"
          class="tb-input"
          type="password"
          required
          autocomplete="current-password"
        />
      </div>
      <p v-if="error" class="tb-error-text" role="alert">{{ error }}</p>
      <div class="tb-form-actions">
        <button class="tb-btn tb-btn-primary tb-btn-block" :disabled="busy">
          {{ busy ? '로그인 중…' : '로그인' }}
        </button>
      </div>
    </form>
    <div class="tb-oauth-group">
      <span class="tb-hint">소셜 계정으로 계속하기</span>
      <div style="display: flex; gap: 8px; margin-top: 8px; flex-wrap: wrap">
        <button v-for="p in providers" :key="p.id" class="tb-btn tb-btn-secondary tb-btn-sm" @click="oauth(p.id)">
          {{ p.label }}
        </button>
      </div>
    </div>
    <p class="tb-hint" style="margin-top: 16px">
      <NuxtLink to="/register">회원가입</NuxtLink> · <NuxtLink to="/reset-password">비밀번호 찾기</NuxtLink>
    </p>
  </div>
</template>

<script setup lang="ts">
import { friendlyMessage } from '~/utils/api'

useSeoMeta({ title: '로그인 | TreeBlog' })
definePageMeta({ layout: 'auth', middleware: 'guest' })

const { login, startOAuth } = useAuth()
const email = ref('')
const password = ref('')
const busy = ref(false)
const error = ref('')

const providers = [
  { id: 'google', label: 'Google' },
  { id: 'kakao', label: '카카오' },
  { id: 'naver', label: '네이버' },
  { id: 'github', label: 'GitHub' },
  { id: 'apple', label: 'Apple' }
]

async function submit(): Promise<void> {
  busy.value = true
  error.value = ''
  try {
    await login(email.value.trim(), password.value)
  } catch (err) {
    error.value = friendlyMessage(err)
  } finally {
    busy.value = false
  }
}

async function oauth(provider: string): Promise<void> {
  await startOAuth(provider)
}
</script>

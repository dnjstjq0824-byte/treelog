<template>
  <div class="tb-login-wrap">
    <div class="tb-login-card">
      <h1 class="tb-console-title">TreeBlog Admin</h1>
      <p class="tb-hint">운영자 계정으로 로그인하세요.</p>
      <form style="margin-top: 16px" @submit.prevent="submit">
        <div class="tb-field">
          <label class="tb-label" for="admin-email">이메일</label>
          <input id="admin-email" v-model="email" class="tb-input" type="email" required autocomplete="email" />
        </div>
        <div class="tb-field">
          <label class="tb-label" for="admin-password">비밀번호</label>
          <input
            id="admin-password"
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
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { friendlyMessage } from '../lib/api'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const email = ref('')
const password = ref('')
const busy = ref(false)
const error = ref('')

async function submit(): Promise<void> {
  busy.value = true
  error.value = ''
  try {
    await auth.login(email.value.trim(), password.value)
    const next = typeof route.query.next === 'string' ? route.query.next : '/'
    await router.push(next)
  } catch (err) {
    error.value = friendlyMessage(err)
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div class="tb-auth-card">
    <h1 class="tb-auth-title">회원가입</h1>
    <p class="tb-hint">이메일 인증 후 블로그를 만들 수 있습니다.</p>
    <form @submit.prevent="submit" style="margin-top: 16px">
      <div class="tb-field">
        <label class="tb-label" for="reg-email">이메일</label>
        <input
          id="reg-email"
          v-model="email"
          class="tb-input"
          type="email"
          required
          autocomplete="email"
        />
      </div>
      <div class="tb-field">
        <label class="tb-label" for="reg-username">사용자 이름 (영문, 4~32자)</label>
        <input
          id="reg-username"
          v-model="username"
          class="tb-input"
          required
          minlength="4"
          maxlength="32"
          autocomplete="username"
        />
      </div>
      <div class="tb-field">
        <label class="tb-label" for="reg-display">표시 이름</label>
        <input id="reg-display" v-model="displayName" class="tb-input" required maxlength="50" />
      </div>
      <div class="tb-field">
        <label class="tb-label" for="reg-password">비밀번호 (8~72자)</label>
        <input
          id="reg-password"
          v-model="password"
          class="tb-input"
          type="password"
          required
          minlength="8"
          maxlength="72"
          autocomplete="new-password"
        />
      </div>
      <label class="tb-check">
        <input v-model="agreeTerms" type="checkbox" required />
        <span><NuxtLink to="/terms">이용약관</NuxtLink>에 동의합니다 (필수)</span>
      </label>
      <label class="tb-check">
        <input v-model="agreePrivacy" type="checkbox" required />
        <span><NuxtLink to="/privacy">개인정보처리방침</NuxtLink>에 동의합니다 (필수)</span>
      </label>
      <p v-if="error" class="tb-error-text" role="alert">{{ error }}</p>
      <div class="tb-form-actions">
        <button class="tb-btn tb-btn-primary tb-btn-block" :disabled="busy">
          {{ busy ? '가입 중…' : '가입하기' }}
        </button>
      </div>
    </form>
    <p class="tb-hint" style="margin-top: 16px">
      이미 계정이 있나요? <NuxtLink to="/login">로그인</NuxtLink>
    </p>
  </div>
</template>

<script setup lang="ts">
import { friendlyMessage } from '~/utils/api'

useSeoMeta({ title: '회원가입 | TreeBlog' })
definePageMeta({ layout: 'auth', middleware: 'guest' })

const { register } = useAuth()
const email = ref('')
const username = ref('')
const displayName = ref('')
const password = ref('')
const agreeTerms = ref(false)
const agreePrivacy = ref(false)
const busy = ref(false)
const error = ref('')

async function submit(): Promise<void> {
  busy.value = true
  error.value = ''
  try {
    await register({
      email: email.value.trim(),
      username: username.value.trim(),
      display_name: displayName.value.trim(),
      password: password.value
    })
  } catch (err) {
    error.value = friendlyMessage(err)
  } finally {
    busy.value = false
  }
}
</script>

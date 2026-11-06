<template>
  <div style="max-width: 560px; margin: 0 auto">
    <h1 class="tb-section-title">블로그 만들기</h1>
    <ol class="tb-steps">
      <li :class="{ active: step === 1 }">1. 기본 정보</li>
      <li :class="{ active: step === 2 }">2. 핸들 선택</li>
      <li :class="{ active: step === 3 }">3. 개설 완료</li>
    </ol>

    <form v-if="step === 1" @submit.prevent="step = 2">
      <div class="tb-field">
        <label class="tb-label" for="ob-title">블로그 이름</label>
        <input id="ob-title" v-model="title" class="tb-input" required maxlength="120" />
      </div>
      <div class="tb-field">
        <label class="tb-label" for="ob-desc">한 줄 소개</label>
        <input id="ob-desc" v-model="description" class="tb-input" maxlength="200" />
      </div>
      <div class="tb-form-actions">
        <button class="tb-btn tb-btn-primary" :disabled="!title.trim()">다음</button>
      </div>
    </form>

    <form v-else-if="step === 2" @submit.prevent="create">
      <div class="tb-field">
        <label class="tb-label" for="ob-handle">핸들 (영문 소문자, 숫자, 하이픈, 3~30자)</label>
        <input
          id="ob-handle"
          v-model="handle"
          class="tb-input"
          required
          minlength="3"
          maxlength="30"
          pattern="[a-z0-9][a-z0-9-]{1,28}[a-z0-9]"
          autocomplete="off"
          @input="onHandleInput"
        />
        <p class="tb-hint">treeblog.com/b/{{ handle || 'your-handle' }}</p>
        <p v-if="handleMessage" class="tb-hint" :style="{ color: handleAvailable ? 'var(--tb-success)' : 'var(--tb-error)' }">
          {{ handleMessage }}
        </p>
      </div>
      <p v-if="error" class="tb-error-text" role="alert">{{ error }}</p>
      <div class="tb-form-actions">
        <button type="button" class="tb-btn tb-btn-ghost" @click="step = 1">이전</button>
        <button class="tb-btn tb-btn-primary" :disabled="busy || !handleAvailable">
          {{ busy ? '개설 중…' : '블로그 개설' }}
        </button>
      </div>
    </form>

    <template v-else>
      <p class="tb-hint" style="color: var(--tb-success)">블로그가 개설되었습니다.</p>
      <div class="tb-form-actions">
        <NuxtLink :to="`/b/${createdHandle}`" class="tb-btn tb-btn-primary">내 블로그 보기</NuxtLink>
        <a :href="consoleUrl" class="tb-btn tb-btn-secondary">콘솔에서 글 쓰기</a>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import type { Blog, CheckHandle } from '~/types/api'
import { api, friendlyMessage } from '~/utils/api'

useSeoMeta({ title: '블로그 만들기 | TreeBlog' })
definePageMeta({ middleware: 'auth' })

const config = useRuntimeConfig()
const consoleUrl = config.public.consoleUrl as string
const step = ref(1)
const title = ref('')
const description = ref('')
const handle = ref('')
const handleAvailable = ref(false)
const handleMessage = ref('')
const busy = ref(false)
const error = ref('')
const createdHandle = ref('')
let timer: ReturnType<typeof setTimeout> | null = null

function onHandleInput(): void {
  handleAvailable.value = false
  handleMessage.value = ''
  if (timer) clearTimeout(timer)
  const value = handle.value.trim().toLowerCase()
  if (value.length < 3) return
  timer = setTimeout(async () => {
    try {
      const result = await api<CheckHandle>('/api/v1/blogs/check-handle', {
        auth: false,
        query: { handle: value }
      })
      handleAvailable.value = result.available
      handleMessage.value = result.available ? '사용 가능한 핸들입니다' : '이미 사용 중이거나 유효하지 않은 핸들입니다'
    } catch {
      handleAvailable.value = false
      handleMessage.value = '핸들을 확인할 수 없습니다'
    }
  }, 300)
}

async function create(): Promise<void> {
  busy.value = true
  error.value = ''
  try {
    const blog = await api<Blog>('/api/v1/blogs/', {
      method: 'POST',
      body: {
        handle: handle.value.trim().toLowerCase(),
        title: title.value.trim(),
        description: description.value.trim()
      }
    })
    createdHandle.value = blog.handle
    step.value = 3
  } catch (err) {
    error.value = friendlyMessage(err)
  } finally {
    busy.value = false
  }
}
</script>

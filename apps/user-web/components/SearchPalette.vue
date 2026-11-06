<template>
  <div v-if="modelValue" class="tb-modal-backdrop" @click.self="close">
    <div class="tb-modal" role="dialog" aria-modal="true" aria-label="검색" style="max-width: 560px">
      <form @submit.prevent="submit">
        <label class="tb-label" for="tb-search-input">TreeBlog 검색</label>
        <input
          id="tb-search-input"
          ref="inputRef"
          v-model="query"
          class="tb-input"
          type="search"
          placeholder="글, 블로그, 태그 검색"
          autocomplete="off"
          style="margin-top: 8px"
          @input="onInput"
        />
      </form>
      <div v-if="suggestions.length > 0" style="margin-top: 12px">
        <button
          v-for="suggestion in suggestions"
          :key="suggestion"
          class="tb-chip"
          style="margin: 0 6px 6px 0"
          @click="pick(suggestion)"
        >
          {{ suggestion }}
        </button>
      </div>
      <p v-else-if="query.trim().length >= 2 && !loading" class="tb-hint" style="margin-top: 12px">
        Enter를 눌러 검색하세요
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { api } from '~/utils/api'

const props = defineProps<{ modelValue: boolean }>()
const emit = defineEmits<{ (event: 'update:modelValue', value: boolean): void }>()

const query = ref('')
const suggestions = ref<string[]>([])
const loading = ref(false)
const inputRef = ref<HTMLInputElement | null>(null)
let timer: ReturnType<typeof setTimeout> | null = null

function close(): void {
  emit('update:modelValue', false)
}

function onInput(): void {
  if (timer) clearTimeout(timer)
  const value = query.value.trim()
  if (value.length < 2) {
    suggestions.value = []
    return
  }
  timer = setTimeout(async () => {
    loading.value = true
    try {
      suggestions.value = await api<string[]>('/api/v1/search/suggest', {
        query: { q: value },
        auth: false
      })
    } catch {
      suggestions.value = []
    } finally {
      loading.value = false
    }
  }, 220)
}

async function submit(): Promise<void> {
  const value = query.value.trim()
  if (!value) return
  close()
  await navigateTo({ path: '/search', query: { q: value } })
}

async function pick(suggestion: string): Promise<void> {
  query.value = suggestion
  await submit()
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      query.value = ''
      suggestions.value = []
      nextTick(() => inputRef.value?.focus())
    }
  }
)

function onEscape(event: KeyboardEvent): void {
  if (event.key === 'Escape' && props.modelValue) close()
}

onMounted(() => window.addEventListener('keydown', onEscape))
onUnmounted(() => window.removeEventListener('keydown', onEscape))
</script>

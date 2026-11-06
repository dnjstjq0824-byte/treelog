<template>
  <div>
    <div class="tb-console-head">
      <h1 class="tb-console-title">테마</h1>
    </div>
    <div v-if="pending">
      <div class="tb-skeleton" style="height: 300px" />
    </div>
    <div v-else>
      <h2 class="tb-section-title" style="margin-bottom: 8px">현재 테마</h2>
      <div v-if="current" class="tb-list-item" style="border: 1px solid var(--tb-border); border-radius: var(--tb-radius-lg); margin-bottom: 20px">
        <div>
          <p class="tb-list-item-title">{{ current.theme_name }} <span class="tb-hint">({{ current.theme_slug }})</span></p>
          <p class="tb-list-item-desc">레이아웃 {{ current.active_layout }} · 다크모드 {{ current.dark_mode_preference }}</p>
        </div>
        <button class="tb-btn tb-btn-ghost tb-btn-sm" @click="reset">기본값으로 초기화</button>
      </div>

      <h2 class="tb-section-title" style="margin-bottom: 8px">테마 카탈로그</h2>
      <ul class="tb-list">
        <li v-for="theme in catalog" :key="theme.id" class="tb-list-item">
          <div>
            <p class="tb-list-item-title">{{ theme.name }}</p>
            <p class="tb-list-item-desc">{{ theme.description }} · {{ theme.layout_type }} · v{{ theme.version }}</p>
          </div>
          <button
            class="tb-btn tb-btn-secondary tb-btn-sm"
            :disabled="busy || current?.theme_id === theme.id"
            @click="applyTheme(theme.id)"
          >
            {{ current?.theme_id === theme.id ? '사용 중' : '적용' }}
          </button>
        </li>
      </ul>

      <h2 class="tb-section-title" style="margin: 20px 0 8px">사용자 CSS</h2>
      <form style="max-width: 640px" @submit.prevent="saveCss">
        <div class="tb-field">
          <label class="tb-label" for="theme-css">Custom CSS</label>
          <textarea id="theme-css" v-model="customCss" class="tb-textarea tb-code" rows="8" spellcheck="false" />
        </div>
        <div class="tb-field">
          <label class="tb-label" for="theme-dark">다크모드</label>
          <select id="theme-dark" v-model="darkMode" class="tb-select">
            <option value="SYSTEM">시스템 설정 따르기</option>
            <option value="LIGHT">라이트 고정</option>
            <option value="DARK">다크 고정</option>
          </select>
        </div>
        <div class="tb-form-actions">
          <button class="tb-btn tb-btn-primary" :disabled="busy">
            {{ busy ? '저장 중…' : '저장' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { api, friendlyMessage } from '../lib/api'
import type { BlogThemeConfig, ThemeCatalogItem } from '../lib/types'
import { useToastStore } from '../stores/toast'

const route = useRoute()
const toast = useToastStore()
const blogId = computed(() => String(route.params.blogId || ''))

const catalog = ref<ThemeCatalogItem[]>([])
const current = ref<BlogThemeConfig | null>(null)
const customCss = ref('')
const darkMode = ref('SYSTEM')
const pending = ref(true)
const busy = ref(false)

onMounted(async () => {
  try {
    const [themes, config] = await Promise.all([
      api<ThemeCatalogItem[]>('/api/v1/themes'),
      api<BlogThemeConfig>(`/api/v1/blogs/${blogId.value}/theme/`)
    ])
    catalog.value = themes
    current.value = config
    customCss.value = config.custom_css
    darkMode.value = config.dark_mode_preference
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  } finally {
    pending.value = false
  }
})

async function applyTheme(themeId: string): Promise<void> {
  busy.value = true
  try {
    current.value = await api<BlogThemeConfig>(`/api/v1/blogs/${blogId.value}/theme/`, {
      method: 'PUT',
      body: { theme_id: themeId }
    })
    customCss.value = current.value.custom_css
    darkMode.value = current.value.dark_mode_preference
    toast.push('테마를 적용했습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  } finally {
    busy.value = false
  }
}

async function saveCss(): Promise<void> {
  busy.value = true
  try {
    current.value = await api<BlogThemeConfig>(`/api/v1/blogs/${blogId.value}/theme/`, {
      method: 'PUT',
      body: { custom_css: customCss.value, dark_mode_preference: darkMode.value }
    })
    toast.push('저장했습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  } finally {
    busy.value = false
  }
}

async function reset(): Promise<void> {
  if (!confirm('테마 설정을 기본값으로 초기화할까요?')) return
  busy.value = true
  try {
    current.value = await api<BlogThemeConfig>(`/api/v1/blogs/${blogId.value}/theme/reset`, {
      method: 'POST',
      body: {}
    })
    customCss.value = current.value.custom_css
    darkMode.value = current.value.dark_mode_preference
    toast.push('초기화했습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div>
    <div class="tb-console-head">
      <h1 class="tb-console-title">블로그 설정</h1>
    </div>
    <div v-if="pending">
      <div class="tb-skeleton" style="height: 400px" />
    </div>
    <form v-else-if="settings" style="max-width: 560px" @submit.prevent="save">
      <div class="tb-field">
        <label class="tb-label" for="set-visibility">기본 공개 범위</label>
        <select id="set-visibility" v-model="settings.default_post_visibility" class="tb-select">
          <option v-for="option in VISIBILITIES" :key="option.id" :value="option.id">
            {{ option.label }}
          </option>
        </select>
      </div>
      <div class="tb-field">
        <label class="tb-label" for="set-comments">기본 댓글 정책</label>
        <select id="set-comments" v-model="settings.default_allow_comments" class="tb-select">
          <option v-for="option in COMMENT_POLICIES" :key="option.id" :value="option.id">
            {{ option.label }}
          </option>
        </select>
      </div>
      <div class="tb-field">
        <label class="tb-label" for="set-per-page">페이지당 글 수 (5~50)</label>
        <input
          id="set-per-page"
          v-model.number="settings.posts_per_page"
          class="tb-input"
          type="number"
          min="5"
          max="50"
        />
      </div>
      <label v-for="flag in flags" :key="flag.id" class="tb-check">
        <input v-model="(settings as unknown as Record<string, boolean>)[flag.id]" type="checkbox" />
        <span>{{ flag.label }}</span>
      </label>
      <div class="tb-form-actions">
        <button class="tb-btn tb-btn-primary" :disabled="busy">
          {{ busy ? '저장 중…' : '저장' }}
        </button>
      </div>
    </form>
    <div v-else class="tb-empty">
      <p class="tb-empty-title">설정을 불러오지 못했습니다</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { api, friendlyMessage } from '../lib/api'
import { COMMENT_POLICIES, VISIBILITIES, type BlogSettings } from '../lib/types'
import { useToastStore } from '../stores/toast'

const route = useRoute()
const toast = useToastStore()
const blogId = computed(() => String(route.params.blogId || ''))

const settings = ref<BlogSettings | null>(null)
const pending = ref(true)
const busy = ref(false)

const flags = [
  { id: 'allow_guestbook', label: '방명록 허용' },
  { id: 'allow_mutual_requests', label: '맞이웃 신청 허용' },
  { id: 'allow_rss_full_content', label: 'RSS 전문 공개' },
  { id: 'allow_search_indexing', label: '검색엔진 색인 허용' },
  { id: 'allow_treescan_sync', label: 'TreeScan 동기화 허용' },
  { id: 'anti_copy_protection', label: '복사 방지' },
  { id: 'require_approval_for_authors', label: '필자 글 승인제' },
  { id: 'show_visitor_counter', label: '방문자 수 표시' },
  { id: 'adult_content_warning', label: '성인 콘텐츠 경고' }
] as const

onMounted(async () => {
  try {
    settings.value = await api<BlogSettings>(`/api/v1/blogs/${blogId.value}/settings`)
  } catch (err) {
    settings.value = null
    toast.push(friendlyMessage(err), 'error')
  } finally {
    pending.value = false
  }
})

async function save(): Promise<void> {
  if (!settings.value) return
  busy.value = true
  try {
    await api(`/api/v1/blogs/${blogId.value}/settings`, {
      method: 'PATCH',
      body: settings.value
    })
    toast.push('저장했습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  } finally {
    busy.value = false
  }
}
</script>

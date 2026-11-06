<template>
  <div>
    <div class="tb-console-head">
      <h1 class="tb-console-title">시리즈</h1>
      <button class="tb-btn tb-btn-primary tb-btn-sm" @click="showCreate = true">새 시리즈</button>
    </div>

    <div v-if="showCreate" class="tb-modal-backdrop" @click.self="showCreate = false">
      <div class="tb-modal" role="dialog" aria-modal="true" aria-label="새 시리즈">
        <h2 class="tb-modal-title">새 시리즈</h2>
        <div class="tb-field">
          <label class="tb-label" for="series-title">제목</label>
          <input id="series-title" v-model="form.title" class="tb-input" required maxlength="120" />
        </div>
        <div class="tb-field">
          <label class="tb-label" for="series-slug">슬러그</label>
          <input id="series-slug" v-model="form.slug" class="tb-input" required maxlength="120" />
        </div>
        <div class="tb-field">
          <label class="tb-label" for="series-desc">설명</label>
          <input id="series-desc" v-model="form.description" class="tb-input" maxlength="300" />
        </div>
        <div class="tb-modal-actions">
          <button class="tb-btn tb-btn-ghost" @click="showCreate = false">취소</button>
          <button class="tb-btn tb-btn-primary" :disabled="busy || !form.title.trim() || !form.slug.trim()" @click="create">
            {{ busy ? '생성 중…' : '생성' }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="pending">
      <div v-for="index in 4" :key="index" class="tb-skeleton" style="height: 56px" />
    </div>
    <ul v-else-if="items.length > 0" class="tb-list">
      <li v-for="item in items" :key="item.id" class="tb-list-item">
        <div>
          <p class="tb-list-item-title">{{ item.title }}</p>
          <p class="tb-list-item-desc">{{ item.description || '설명 없음' }} · 글 {{ item.post_count }}개</p>
        </div>
        <span v-if="item.is_completed" class="tb-badge">완결</span>
      </li>
    </ul>
    <div v-else class="tb-empty">
      <p class="tb-empty-title">시리즈가 없습니다</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { api, friendlyMessage } from '../lib/api'
import type { Series } from '../lib/types'
import { useToastStore } from '../stores/toast'

const route = useRoute()
const toast = useToastStore()
const blogId = computed(() => String(route.params.blogId || ''))

const items = ref<Series[]>([])
const pending = ref(true)
const busy = ref(false)
const showCreate = ref(false)
const form = ref({ title: '', slug: '', description: '' })

async function load(): Promise<void> {
  pending.value = true
  try {
    items.value = await api<Series[]>(`/api/v1/blogs/${blogId.value}/series`)
  } catch (err) {
    items.value = []
    toast.push(friendlyMessage(err), 'error')
  } finally {
    pending.value = false
  }
}

async function create(): Promise<void> {
  busy.value = true
  try {
    const created = await api<Series>(`/api/v1/blogs/${blogId.value}/series`, {
      method: 'POST',
      body: {
        title: form.value.title.trim(),
        slug: form.value.slug.trim(),
        description: form.value.description.trim()
      }
    })
    items.value = [...items.value, created]
    showCreate.value = false
    form.value = { title: '', slug: '', description: '' }
    toast.push('시리즈를 만들었습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  } finally {
    busy.value = false
  }
}

onMounted(load)
</script>

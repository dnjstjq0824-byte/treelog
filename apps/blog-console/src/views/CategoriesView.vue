<template>
  <div>
    <div class="tb-console-head">
      <h1 class="tb-console-title">카테고리</h1>
      <button class="tb-btn tb-btn-primary tb-btn-sm" @click="showCreate = true">새 카테고리</button>
    </div>

    <div v-if="showCreate" class="tb-modal-backdrop" @click.self="showCreate = false">
      <div class="tb-modal" role="dialog" aria-modal="true" aria-label="새 카테고리">
        <h2 class="tb-modal-title">새 카테고리</h2>
        <div class="tb-field">
          <label class="tb-label" for="cat-name">이름</label>
          <input id="cat-name" v-model="form.name" class="tb-input" required maxlength="60" />
        </div>
        <div class="tb-field">
          <label class="tb-label" for="cat-slug">슬러그</label>
          <input id="cat-slug" v-model="form.slug" class="tb-input" required maxlength="80" />
        </div>
        <div class="tb-field">
          <label class="tb-label" for="cat-desc">설명</label>
          <input id="cat-desc" v-model="form.description" class="tb-input" maxlength="200" />
        </div>
        <div class="tb-modal-actions">
          <button class="tb-btn tb-btn-ghost" @click="showCreate = false">취소</button>
          <button class="tb-btn tb-btn-primary" :disabled="busy || !form.name.trim() || !form.slug.trim()" @click="create">
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
          <p class="tb-list-item-title">{{ item.name }} <span class="tb-hint">/{{ item.slug }}</span></p>
          <p class="tb-list-item-desc">글 {{ item.post_count }}개 · 순서 {{ item.sort_order }}</p>
        </div>
        <button class="tb-btn tb-btn-ghost tb-btn-sm" @click="remove(item.id)">삭제</button>
      </li>
    </ul>
    <div v-else class="tb-empty">
      <p class="tb-empty-title">카테고리가 없습니다</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { api, friendlyMessage } from '../lib/api'
import type { BlogCategory } from '../lib/types'
import { useToastStore } from '../stores/toast'

const route = useRoute()
const toast = useToastStore()
const blogId = computed(() => String(route.params.blogId || ''))

const items = ref<BlogCategory[]>([])
const pending = ref(true)
const busy = ref(false)
const showCreate = ref(false)
const form = ref({ name: '', slug: '', description: '' })

async function load(): Promise<void> {
  pending.value = true
  try {
    items.value = await api<BlogCategory[]>(`/api/v1/blogs/${blogId.value}/categories`)
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
    const created = await api<BlogCategory>(`/api/v1/blogs/${blogId.value}/categories`, {
      method: 'POST',
      body: {
        name: form.value.name.trim(),
        slug: form.value.slug.trim(),
        description: form.value.description.trim(),
        visibility: 'PUBLIC',
        sort_order: items.value.length
      }
    })
    items.value = [...items.value, created]
    showCreate.value = false
    form.value = { name: '', slug: '', description: '' }
    toast.push('카테고리를 만들었습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  } finally {
    busy.value = false
  }
}

async function remove(categoryId: string): Promise<void> {
  if (!confirm('이 카테고리를 삭제할까요? 속한 글은 미분류가 됩니다.')) return
  try {
    await api(`/api/v1/blogs/${blogId.value}/categories/${categoryId}`, { method: 'DELETE' })
    items.value = items.value.filter((item) => item.id !== categoryId)
    toast.push('삭제했습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  }
}

onMounted(load)
</script>

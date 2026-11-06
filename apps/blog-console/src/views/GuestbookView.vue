<template>
  <div>
    <div class="tb-console-head">
      <h1 class="tb-console-title">방명록</h1>
    </div>
    <div v-if="pending">
      <div v-for="index in 4" :key="index" class="tb-skeleton" style="height: 64px" />
    </div>
    <ul v-else-if="items.length > 0" class="tb-list">
      <li v-for="item in items" :key="item.id" class="tb-list-item">
        <div>
          <p class="tb-list-item-title">
            {{ item.author_name || item.author_username || '익명' }}
            <span v-if="item.is_secret" class="tb-badge">비밀글</span>
          </p>
          <p class="tb-list-item-desc">{{ item.content_text }}</p>
          <p class="tb-hint">{{ formatDateTime(item.created_at) }}</p>
        </div>
      </li>
    </ul>
    <div v-else class="tb-empty">
      <p class="tb-empty-title">방명록이 비어 있습니다</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { formatDateTime } from '@treeblog/ui-kit'
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { api, friendlyMessage } from '../lib/api'
import type { GuestbookEntry } from '../lib/types'
import { useToastStore } from '../stores/toast'

const route = useRoute()
const toast = useToastStore()
const blogId = computed(() => String(route.params.blogId || ''))

const items = ref<GuestbookEntry[]>([])
const pending = ref(true)

onMounted(async () => {
  try {
    items.value = await api<GuestbookEntry[]>(`/api/v1/blogs/${blogId.value}/guestbook`)
  } catch (err) {
    items.value = []
    toast.push(friendlyMessage(err), 'error')
  } finally {
    pending.value = false
  }
})
</script>

<template>
  <div>
    <div class="tb-console-head">
      <h1 class="tb-console-title">글 관리</h1>
      <RouterLink :to="`/b/${blogId}/posts/new`" class="tb-btn tb-btn-primary tb-btn-sm">새 글 쓰기</RouterLink>
    </div>
    <div class="tb-tabs" role="tablist" aria-label="상태 필터">
      <button
        v-for="option in POST_STATUSES"
        :key="option.id"
        role="tab"
        class="tb-tab"
        :aria-selected="status === option.id"
        @click="changeStatus(option.id)"
      >
        {{ option.label }}
      </button>
    </div>
    <div v-if="pending" style="margin-top: 12px">
      <div v-for="index in 6" :key="index" class="tb-skeleton" style="height: 64px" />
    </div>
    <ul v-else-if="items.length > 0" class="tb-list">
      <li v-for="item in items" :key="item.id" class="tb-list-item">
        <div>
          <p class="tb-list-item-title">
            <RouterLink :to="`/b/${blogId}/posts/${item.id}`">{{ item.title || '(제목 없음)' }}</RouterLink>
          </p>
          <p class="tb-list-item-desc">
            {{ statusLabel(item.status) }} · {{ item.visibility }} ·
            조회 {{ item.view_count }} · 댓글 {{ item.comment_count }} ·
            {{ formatDate(item.published_at || item.created_at) }}
          </p>
        </div>
        <div style="display: flex; gap: 6px">
          <RouterLink :to="`/b/${blogId}/posts/${item.id}`" class="tb-btn tb-btn-secondary tb-btn-sm">편집</RouterLink>
          <button
            v-if="item.status === 'PUBLISHED'"
            class="tb-btn tb-btn-ghost tb-btn-sm"
            @click="unpublish(item.id)"
          >
            발행 취소
          </button>
          <button class="tb-btn tb-btn-ghost tb-btn-sm" @click="trash(item.id)">삭제</button>
        </div>
      </li>
    </ul>
    <div v-else class="tb-empty">
      <p class="tb-empty-title">글이 없습니다</p>
      <p class="tb-empty-desc">새 글을 작성해 보세요.</p>
    </div>
    <div v-if="hasMore" class="tb-load-more">
      <button class="tb-btn tb-btn-secondary" :disabled="loadingMore" @click="loadMore">
        {{ loadingMore ? '불러오는 중…' : '더 보기' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { formatDate } from '@treeblog/ui-kit'
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api, friendlyMessage } from '../lib/api'
import { POST_STATUSES, type PostSummary } from '../lib/types'
import { useToastStore } from '../stores/toast'

const route = useRoute()
const toast = useToastStore()
const blogId = computed(() => String(route.params.blogId || ''))

const items = ref<PostSummary[]>([])
const status = ref('')
const offset = ref(0)
const hasMore = ref(false)
const pending = ref(true)
const loadingMore = ref(false)
const PAGE_SIZE = 20

const statusLabels: Record<string, string> = {
  DRAFT: '초안',
  IN_REVIEW: '검토 중',
  SCHEDULED: '예약됨',
  PUBLISHED: '발행됨',
  UNPUBLISHED: '발행 취소'
}

function statusLabel(value: string): string {
  return statusLabels[value] || value
}

async function fetchPage(nextOffset: number, append: boolean): Promise<void> {
  try {
    const page = await api<PostSummary[]>(`/api/v1/blogs/${blogId.value}/posts/`, {
      query: { limit: PAGE_SIZE, offset: nextOffset, status: status.value || undefined }
    })
    items.value = append ? [...items.value, ...page] : page
    offset.value = nextOffset
    hasMore.value = page.length >= PAGE_SIZE
  } catch (err) {
    if (!append) items.value = []
    toast.push(friendlyMessage(err), 'error')
  } finally {
    pending.value = false
    loadingMore.value = false
  }
}

async function changeStatus(next: string): Promise<void> {
  if (status.value === next) return
  status.value = next
  pending.value = true
  await fetchPage(0, false)
}

async function loadMore(): Promise<void> {
  if (!hasMore.value || loadingMore.value) return
  loadingMore.value = true
  await fetchPage(offset.value + PAGE_SIZE, true)
}

async function unpublish(postId: string): Promise<void> {
  if (!confirm('발행을 취소할까요?')) return
  try {
    await api(`/api/v1/blogs/${blogId.value}/posts/${postId}/unpublish`, { method: 'POST', body: {} })
    const target = items.value.find((item) => item.id === postId)
    if (target) target.status = 'UNPUBLISHED'
    toast.push('발행을 취소했습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  }
}

async function trash(postId: string): Promise<void> {
  if (!confirm('이 글을 휴지통으로 이동할까요?')) return
  try {
    await api(`/api/v1/blogs/${blogId.value}/posts/${postId}/`, { method: 'DELETE' })
    items.value = items.value.filter((item) => item.id !== postId)
    toast.push('휴지통으로 이동했습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  }
}

watch(blogId, () => {
  pending.value = true
  void fetchPage(0, false)
})

onMounted(() => fetchPage(0, false))
</script>

<template>
  <div>
    <h1 class="tb-section-title">구독 관리</h1>
    <div v-if="pending">
      <div v-for="index in 3" :key="index" class="tb-skeleton" style="height: 64px" />
    </div>
    <template v-else-if="items.length > 0">
      <ul class="tb-list">
        <li v-for="item in items" :key="item.blog_id" class="tb-list-item">
          <div>
            <p class="tb-list-item-title">
              <NuxtLink :to="`/b/${item.handle}`">{{ item.title }}</NuxtLink>
            </p>
            <p class="tb-list-item-desc">
              {{ item.tier === 'PAID' ? '유료 멤버십' : '무료 구독' }}
            </p>
          </div>
          <button
            class="tb-btn tb-btn-secondary tb-btn-sm"
            @click="unsubscribe(item.blog_id)"
          >
            구독 취소
          </button>
        </li>
      </ul>
    </template>
    <EmptyState v-else title="구독 중인 블로그가 없습니다" />
  </div>
</template>

<script setup lang="ts">
import type { SubscriptionItem } from '~/types/api'
import { api } from '~/utils/api'

useSeoMeta({ title: '구독 관리 | TreeBlog', robots: 'noindex' })
definePageMeta({ middleware: 'auth' })

const items = ref<SubscriptionItem[]>([])
const pending = ref(true)
const toast = useToast()

async function load(): Promise<void> {
  pending.value = true
  try {
    items.value = await api<SubscriptionItem[]>('/api/v1/me/subscriptions', {
      query: { limit: 100 }
    })
  } catch {
    items.value = []
    toast.push('구독 목록을 불러오지 못했습니다', 'error')
  } finally {
    pending.value = false
  }
}

async function unsubscribe(blogId: string): Promise<void> {
  if (!confirm('구독을 취소할까요?')) return
  try {
    await api(`/api/v1/blogs/${blogId}/subscribe`, { method: 'DELETE' })
    items.value = items.value.filter((item) => item.blog_id !== blogId)
    toast.push('구독을 취소했습니다', 'success')
  } catch {
    toast.push('구독 취소에 실패했습니다', 'error')
  }
}

onMounted(load)
</script>

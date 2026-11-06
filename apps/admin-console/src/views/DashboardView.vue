<template>
  <div>
    <div class="tb-console-head">
      <h1 class="tb-console-title">대시보드</h1>
      <button class="tb-btn tb-btn-secondary tb-btn-sm" :disabled="pending" @click="load">새로고침</button>
    </div>
    <div v-if="pending">
      <div v-for="index in 6" :key="index" class="tb-skeleton" style="height: 90px" />
    </div>
    <div v-else-if="stats" class="tb-stat-grid">
      <div class="tb-stat-card">
        <p class="tb-stat-value">{{ stats.users_total.toLocaleString() }}</p>
        <p class="tb-stat-label">전체 사용자 (활성 {{ stats.users_active.toLocaleString() }})</p>
      </div>
      <div class="tb-stat-card">
        <p class="tb-stat-value">{{ stats.users_suspended.toLocaleString() }}</p>
        <p class="tb-stat-label">정지된 사용자</p>
      </div>
      <div class="tb-stat-card">
        <p class="tb-stat-value">{{ stats.blogs_total.toLocaleString() }}</p>
        <p class="tb-stat-label">전체 블로그</p>
      </div>
      <div class="tb-stat-card">
        <p class="tb-stat-value">{{ stats.posts_published.toLocaleString() }}</p>
        <p class="tb-stat-label">발행된 글 (오늘 {{ stats.posts_today.toLocaleString() }})</p>
      </div>
      <div class="tb-stat-card">
        <p class="tb-stat-value">{{ stats.comments_total.toLocaleString() }}</p>
        <p class="tb-stat-label">전체 댓글</p>
      </div>
      <div class="tb-stat-card">
        <p class="tb-stat-value">{{ stats.media_ready.toLocaleString() }}</p>
        <p class="tb-stat-label">처리된 미디어</p>
      </div>
      <div class="tb-stat-card">
        <p class="tb-stat-value">{{ stats.cases_open.toLocaleString() }}</p>
        <p class="tb-stat-label">열린 모더레이션 케이스</p>
      </div>
      <div class="tb-stat-card">
        <p class="tb-stat-value">{{ stats.appeals_pending.toLocaleString() }}</p>
        <p class="tb-stat-label">대기 중인 이의제기</p>
      </div>
      <div class="tb-stat-card">
        <p class="tb-stat-value">{{ stats.orders_paid_today.toLocaleString() }}</p>
        <p class="tb-stat-label">오늘 결제 건수</p>
      </div>
      <div class="tb-stat-card">
        <p class="tb-stat-value">{{ stats.revenue_today_minor.toLocaleString() }}원</p>
        <p class="tb-stat-label">오늘 매출</p>
      </div>
    </div>
    <div v-else class="tb-empty">
      <p class="tb-empty-title">통계를 불러오지 못했습니다</p>
      <p class="tb-empty-desc">SUPER_ADMIN 권한이 필요합니다.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, friendlyMessage } from '../lib/api'
import type { PlatformStats } from '../lib/types'
import { useToastStore } from '../stores/toast'

const toast = useToastStore()
const stats = ref<PlatformStats | null>(null)
const pending = ref(true)

async function load(): Promise<void> {
  pending.value = true
  try {
    stats.value = await api<PlatformStats>('/api/v1/admin/stats')
  } catch (err) {
    stats.value = null
    toast.push(friendlyMessage(err), 'error')
  } finally {
    pending.value = false
  }
}

onMounted(load)
</script>

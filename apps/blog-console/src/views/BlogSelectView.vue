<template>
  <div>
    <div class="tb-console-head">
      <h1 class="tb-console-title">내 블로그</h1>
    </div>
    <div v-if="pending">
      <div v-for="index in 3" :key="index" class="tb-skeleton" style="height: 72px" />
    </div>
    <ul v-else-if="blogStore.blogs.length > 0" class="tb-list">
      <li v-for="blog in blogStore.blogs" :key="blog.id" class="tb-list-item">
        <div>
          <p class="tb-list-item-title">
            <RouterLink :to="`/b/${blog.id}/posts`">{{ blog.title }}</RouterLink>
          </p>
          <p class="tb-list-item-desc">
            @{{ blog.handle }} · 글 {{ blog.post_count }} · 구독자 {{ blog.subscriber_count }}
          </p>
        </div>
        <RouterLink :to="`/b/${blog.id}/posts`" class="tb-btn tb-btn-secondary tb-btn-sm">관리</RouterLink>
      </li>
    </ul>
    <div v-else class="tb-empty">
      <p class="tb-empty-title">아직 블로그가 없습니다</p>
      <p class="tb-empty-desc">웹사이트에서 블로그를 먼저 개설하세요.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { friendlyMessage } from '../lib/api'
import { useBlogStore } from '../stores/blog'
import { useToastStore } from '../stores/toast'

const blogStore = useBlogStore()
const toast = useToastStore()
const pending = ref(true)

onMounted(async () => {
  try {
    await blogStore.load()
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  } finally {
    pending.value = false
  }
})
</script>

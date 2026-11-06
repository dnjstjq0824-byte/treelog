<template>
  <div class="tb-console">
    <aside class="tb-console-side">
      <RouterLink to="/" class="tb-console-brand">
        <span class="tb-console-brand-mark" aria-hidden="true">
          <svg viewBox="0 0 32 32" fill="none"><path d="M16 5 L24 14 H19 V24 H13 V14 H8 Z" fill="currentColor" /></svg>
        </span>
        Studio
      </RouterLink>
      <div v-if="blogStore.blogs.length > 0" class="tb-blog-switcher">
        <label class="tb-console-nav-label" for="blog-switch">블로그</label>
        <select id="blog-switch" :value="blogId" @change="switchBlog">
          <option v-for="blog in blogStore.blogs" :key="blog.id" :value="blog.id">
            {{ blog.title }}
          </option>
        </select>
      </div>
      <nav class="tb-console-nav" aria-label="콘솔 메뉴">
        <RouterLink to="/">내 블로그</RouterLink>
        <template v-if="blogId">
          <span class="tb-console-nav-label">콘텐츠</span>
          <RouterLink :to="`/b/${blogId}/posts`">글 관리</RouterLink>
          <RouterLink :to="`/b/${blogId}/posts/new`">새 글 쓰기</RouterLink>
          <RouterLink :to="`/b/${blogId}/media`">미디어</RouterLink>
          <RouterLink :to="`/b/${blogId}/categories`">카테고리</RouterLink>
          <RouterLink :to="`/b/${blogId}/series`">시리즈</RouterLink>
          <RouterLink :to="`/b/${blogId}/guestbook`">방명록</RouterLink>
          <span class="tb-console-nav-label">수익</span>
          <RouterLink :to="`/b/${blogId}/memberships`">멤버십</RouterLink>
          <RouterLink :to="`/b/${blogId}/payout`">정산</RouterLink>
          <span class="tb-console-nav-label">설정</span>
          <RouterLink :to="`/b/${blogId}/members`">멤버</RouterLink>
          <RouterLink :to="`/b/${blogId}/theme`">테마</RouterLink>
          <RouterLink :to="`/b/${blogId}/domains`">도메인</RouterLink>
          <RouterLink :to="`/b/${blogId}/settings`">블로그 설정</RouterLink>
        </template>
      </nav>
      <div style="flex: 1" />
      <div>
        <p style="font-size: 0.8rem; color: #9db8a9">{{ auth.user?.display_name }}</p>
        <button class="tb-btn tb-btn-ghost tb-btn-sm" style="color: #fff; padding-left: 0" @click="logout">
          로그아웃
        </button>
      </div>
    </aside>
    <main id="main-content" class="tb-console-main">
      <RouterView :key="route.fullPath" />
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useBlogStore } from '../stores/blog'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const blogStore = useBlogStore()

const blogId = computed(() => {
  const param = route.params.blogId
  if (typeof param === 'string' && param) return param
  return blogStore.currentId
})

async function switchBlog(event: Event): Promise<void> {
  const id = (event.target as HTMLSelectElement).value
  blogStore.select(id)
  await router.push(`/b/${id}/posts`)
}

async function logout(): Promise<void> {
  await auth.logout()
  await router.push('/login')
}
</script>

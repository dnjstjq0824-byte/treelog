<template>
  <div>
    <form @submit.prevent="submit" role="search" aria-label="글 검색">
      <div class="tb-field">
        <label class="tb-label" for="search-q">검색</label>
        <input
          id="search-q"
          v-model="draft"
          class="tb-input"
          type="search"
          placeholder="글, 블로그, 태그를 검색하세요"
          autocomplete="off"
        />
      </div>
      <div class="tb-tabs" role="tablist" aria-label="검색 범위" style="margin-top: 12px">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          role="tab"
          class="tb-tab"
          :aria-selected="scope === tab.id"
          @click="scope = tab.id"
        >
          {{ tab.label }}
        </button>
      </div>
      <div class="tb-form-actions">
        <button class="tb-btn tb-btn-primary" type="submit">검색</button>
      </div>
    </form>

    <template v-if="scope === 'posts'">
      <div v-if="result" class="tb-section-head" style="margin-top: 24px">
        <h2 class="tb-section-title">“{{ result.query }}” 검색 결과</h2>
        <span class="tb-hint">총 {{ result.total }}건</span>
      </div>
      <div v-if="pending">
        <PostSkeleton v-for="index in 4" :key="index" />
      </div>
      <template v-else-if="result && result.items.length > 0">
        <PostCard
          v-for="(post, index) in result.items"
          :key="post.id"
          :post="post"
          @click="logClick(post.id, index)"
        />
        <div v-if="hasMore" class="tb-load-more">
          <button class="tb-btn tb-btn-secondary" :disabled="loadingMore" @click="loadMore">
            {{ loadingMore ? '불러오는 중…' : '더 보기' }}
          </button>
        </div>
      </template>
      <EmptyState
        v-else-if="submitted"
        title="검색 결과가 없습니다"
        description="다른 키워드로 검색해 보세요."
      />
      <EmptyState
        v-else
        title="무엇을 찾고 계세요?"
        description="키워드를 입력하고 Enter를 눌러 검색하세요."
      />
    </template>

    <template v-else-if="scope === 'blogs'">
      <div v-if="pending">
        <div v-for="index in 4" :key="index" class="tb-skeleton" style="height: 64px" />
      </div>
      <ul v-else-if="blogItems.length > 0" class="tb-list">
        <li v-for="blog in blogItems" :key="blog.id" class="tb-list-item">
          <div>
            <p class="tb-list-item-title">
              <NuxtLink :to="`/b/${blog.handle}`">{{ blog.title }}</NuxtLink>
            </p>
            <p class="tb-list-item-desc">@{{ blog.handle }} · 글 {{ formatCount(blog.post_count) }}</p>
          </div>
          <NuxtLink :to="`/b/${blog.handle}`" class="tb-btn tb-btn-ghost tb-btn-sm">방문</NuxtLink>
        </li>
      </ul>
      <EmptyState v-else-if="submitted" title="검색 결과가 없습니다" />
      <EmptyState v-else title="블로그를 검색해 보세요" />
    </template>

    <template v-else>
      <div v-if="pending">
        <div v-for="index in 4" :key="index" class="tb-skeleton" style="height: 52px" />
      </div>
      <div v-else-if="tagItems.length > 0" class="tb-chip-row" style="margin-top: 16px">
        <span v-for="tag in tagItems" :key="tag.id" class="tb-chip">
          #{{ tag.name }} ({{ formatCount(tag.post_count) }})
        </span>
      </div>
      <EmptyState v-else-if="submitted" title="검색 결과가 없습니다" />
      <EmptyState v-else title="태그를 검색해 보세요" />
    </template>
  </div>
</template>

<script setup lang="ts">
import { formatCount } from '@treeblog/ui-kit'
import type { SearchBlogItem, SearchPostsData, SearchTagItem } from '~/types/api'
import { api } from '~/utils/api'

useSeoMeta({ title: '검색 | TreeBlog' })

const tabs = [
  { id: 'posts', label: '글' },
  { id: 'blogs', label: '블로그' },
  { id: 'tags', label: '태그' }
] as const

type Scope = (typeof tabs)[number]['id']

const route = useRoute()
const draft = ref(typeof route.query.q === 'string' ? route.query.q : '')
const scope = ref<Scope>('posts')
const result = ref<SearchPostsData | null>(null)
const blogItems = ref<SearchBlogItem[]>([])
const tagItems = ref<SearchTagItem[]>([])
const pending = ref(false)
const loadingMore = ref(false)
const submitted = ref(false)
const toast = useToast()

const hasMore = computed(() => {
  if (!result.value) return false
  return result.value.offset + result.value.items.length < result.value.total
})

async function runPosts(query: string, offset: number, append: boolean): Promise<void> {
  const data = await api<SearchPostsData>('/api/v1/search/posts', {
    auth: false,
    query: { q: query, limit: 20, offset }
  })
  result.value = append && result.value
    ? { ...data, items: [...result.value.items, ...data.items] }
    : data
}

async function run(query: string): Promise<void> {
  if (!query.trim()) return
  pending.value = true
  submitted.value = true
  try {
    if (scope.value === 'posts') {
      await runPosts(query.trim(), 0, false)
    } else if (scope.value === 'blogs') {
      blogItems.value = await api<SearchBlogItem[]>('/api/v1/search/blogs', {
        auth: false,
        query: { q: query.trim(), limit: 20 }
      })
    } else {
      tagItems.value = await api<SearchTagItem[]>('/api/v1/search/tags', {
        auth: false,
        query: { q: query.trim(), limit: 20 }
      })
    }
  } catch {
    toast.push('검색에 실패했습니다', 'error')
  } finally {
    pending.value = false
  }
}

async function submit(): Promise<void> {
  await navigateTo({ path: '/search', query: draft.value.trim() ? { q: draft.value.trim() } : {} })
  await run(draft.value)
}

async function loadMore(): Promise<void> {
  if (!result.value || loadingMore.value || !hasMore.value) return
  loadingMore.value = true
  try {
    await runPosts(result.value.query, result.value.offset + result.value.items.length, true)
  } catch {
    toast.push('검색에 실패했습니다', 'error')
  } finally {
    loadingMore.value = false
  }
}

function logClick(postId: string, index: number): void {
  if (!result.value) return
  void api('/api/v1/search/click', {
    method: 'POST',
    auth: false,
    body: {
      query_text: result.value.query,
      clicked_post_id: postId,
      rank_position: index + 1,
      session_hash: 'web',
      dwell_time_sec: 0
    }
  }).catch(() => {})
}

watch(scope, () => {
  if (draft.value.trim()) void run(draft.value)
})

watch(
  () => route.query.q,
  (value) => {
    const q = typeof value === 'string' ? value : ''
    draft.value = q
    if (q) void run(q)
  }
)

onMounted(() => {
  const q = typeof route.query.q === 'string' ? route.query.q : ''
  if (q) void run(q)
})
</script>

<template>
  <div v-if="pending" class="tb-article">
    <div class="tb-skeleton" style="width: 40%; height: 14px" />
    <div class="tb-skeleton" style="width: 90%; height: 36px" />
    <div class="tb-skeleton" style="width: 100%; height: 200px" />
  </div>
  <div v-else-if="failed">
    <EmptyState title="블로그를 찾을 수 없습니다" description="주소를 다시 확인해 주세요.">
      <template #action>
        <NuxtLink to="/" class="tb-btn tb-btn-secondary tb-btn-sm">홈으로</NuxtLink>
      </template>
    </EmptyState>
  </div>
  <div v-else-if="blog">
    <div class="tb-article-head">
      <p class="tb-article-kicker">@{{ blog.handle }}</p>
      <h1 class="tb-article-title">{{ blog.title }}</h1>
      <p v-if="blog.description" class="tb-hint">{{ blog.description }}</p>
      <div class="tb-post-row-meta" style="margin-top: 12px">
        <span>구독자 {{ formatCount(blog.subscriber_count) }}</span>
        <span>·</span>
        <span>글 {{ formatCount(blog.post_count) }}</span>
      </div>
      <div v-if="isLoggedIn" style="margin-top: 12px; display: flex; gap: 8px; flex-wrap: wrap">
        <button
          class="tb-btn tb-btn-primary tb-btn-sm"
          :disabled="relationBusy"
          @click="toggleSubscribe"
        >
          {{ subscribed ? '구독 중' : '구독하기' }}
        </button>
        <button
          class="tb-btn tb-btn-secondary tb-btn-sm"
          :disabled="relationBusy"
          @click="toggleFollow"
        >
          {{ following ? '이웃 ✓' : '이웃 추가' }}
        </button>
      </div>
    </div>

    <section v-if="tiers.length > 0" aria-label="멤버십" style="margin: 20px 0">
      <h2 class="tb-section-title">멤버십</h2>
      <ul class="tb-list">
        <li v-for="tier in tiers" :key="tier.id" class="tb-list-item">
          <div>
            <p class="tb-list-item-title">{{ tier.name }}</p>
            <p class="tb-list-item-desc">
              {{ tier.description }} · 월 {{ tier.monthly_price_minor.toLocaleString() }}원
            </p>
          </div>
          <button
            v-if="isLoggedIn"
            class="tb-btn tb-btn-primary tb-btn-sm"
            :disabled="checkoutBusy"
            @click="subscribeTier(blog.id, blog.handle, tier)"
          >
            가입하기
          </button>
          <NuxtLink v-else to="/login" class="tb-btn tb-btn-secondary tb-btn-sm">로그인 후 가입</NuxtLink>
        </li>
      </ul>
    </section>

    <div v-if="postsPending">
      <PostSkeleton v-for="index in 4" :key="index" />
    </div>
    <template v-else-if="posts.length > 0">
      <PostCard
        v-for="post in posts"
        :key="post.id"
        :post="post"
        :blog-handle="blog.handle"
        :blog-title="blog.title"
      />
      <div v-if="canShowMore" class="tb-load-more">
        <button class="tb-btn tb-btn-secondary" :disabled="loadingMore" @click="showMore">
          {{ loadingMore ? '불러오는 중…' : '더 보기' }}
        </button>
      </div>
    </template>
    <EmptyState v-else title="아직 발행된 글이 없습니다" />
  </div>
</template>

<script setup lang="ts">
import { formatCount } from '@treeblog/ui-kit'
import type { Blog, FollowedBlog, MembershipTier, PostSummary, SubscriptionItem } from '~/types/api'
import { api } from '~/utils/api'

const route = useRoute()
const handle = computed(() => String(route.params.handle || ''))
const { isLoggedIn } = useAuth()
const { busy: checkoutBusy, subscribeTier } = useCheckout()
const toast = useToast()

const blog = ref<Blog | null>(null)
const posts = ref<PostSummary[]>([])
const tiers = ref<MembershipTier[]>([])
const pending = ref(true)
const postsPending = ref(true)
const failed = ref(false)
const following = ref(false)
const subscribed = ref(false)
const relationBusy = ref(false)
const limit = ref(10)
const canShowMore = ref(false)
const loadingMore = ref(false)

useSeoMeta({
  title: () => `${blog.value?.title || '블로그'} | TreeBlog`,
  description: () => blog.value?.description || 'TreeBlog 블로그'
})

async function loadPosts(nextLimit: number): Promise<void> {
  try {
    posts.value = await api<PostSummary[]>(`/api/v1/public/blogs/${handle.value}/posts/`, {
      auth: false,
      query: { limit: nextLimit }
    })
    limit.value = nextLimit
    canShowMore.value = posts.value.length >= nextLimit && nextLimit < 50
  } catch {
    if (nextLimit === 10) posts.value = []
  } finally {
    postsPending.value = false
    loadingMore.value = false
  }
}

async function loadRelations(): Promise<void> {
  if (!isLoggedIn.value || !blog.value) return
  try {
    const [follows, subscriptions] = await Promise.all([
      api<FollowedBlog[]>('/api/v1/me/follows', { query: { limit: 100 } }),
      api<SubscriptionItem[]>('/api/v1/me/subscriptions', { query: { limit: 100 } })
    ])
    following.value = follows.some((item) => item.blog_id === blog.value?.id)
    subscribed.value = subscriptions.some((item) => item.blog_id === blog.value?.id)
  } catch {
    // Relation state is best-effort; buttons still work.
  }
}

async function toggleFollow(): Promise<void> {
  if (!blog.value) return
  relationBusy.value = true
  try {
    await api(`/api/v1/blogs/${blog.value.id}/follow`, {
      method: following.value ? 'DELETE' : 'POST'
    })
    following.value = !following.value
    toast.push(following.value ? '이웃으로 추가했습니다' : '이웃을 해제했습니다', 'success')
  } catch {
    toast.push('요청에 실패했습니다', 'error')
  } finally {
    relationBusy.value = false
  }
}

async function toggleSubscribe(): Promise<void> {
  if (!blog.value) return
  relationBusy.value = true
  try {
    await api(`/api/v1/blogs/${blog.value.id}/subscribe`, {
      method: subscribed.value ? 'DELETE' : 'POST'
    })
    subscribed.value = !subscribed.value
    toast.push(subscribed.value ? '구독했습니다' : '구독을 취소했습니다', 'success')
  } catch {
    toast.push('요청에 실패했습니다', 'error')
  } finally {
    relationBusy.value = false
  }
}

async function showMore(): Promise<void> {
  if (!canShowMore.value || loadingMore.value) return
  loadingMore.value = true
  await loadPosts(Math.min(50, limit.value + 10))
}

onMounted(async () => {
  pending.value = true
  failed.value = false
  try {
    blog.value = await api<Blog>(`/api/v1/blogs/${handle.value}`, { auth: false })
  } catch {
    failed.value = true
    pending.value = false
    return
  }
  pending.value = false
  await Promise.all([
    loadPosts(10),
    loadRelations(),
    (async () => {
      try {
        tiers.value = await api<MembershipTier[]>(`/api/v1/blogs/${blog.value?.id}/memberships/`, {
          auth: false
        })
      } catch {
        tiers.value = []
      }
    })()
  ])
})
</script>

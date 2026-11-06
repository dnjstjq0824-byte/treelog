<template>
  <div v-if="pending" class="tb-article">
    <div class="tb-skeleton" style="width: 50%; height: 14px" />
    <div class="tb-skeleton" style="width: 95%; height: 38px" />
    <div class="tb-skeleton" style="width: 100%; height: 400px" />
  </div>
  <div v-else-if="failed || !post">
    <EmptyState title="글을 찾을 수 없습니다" description="삭제되었거나 주소가 잘못되었습니다.">
      <template #action>
        <NuxtLink :to="`/b/${handle}`" class="tb-btn tb-btn-secondary tb-btn-sm">블로그로</NuxtLink>
      </template>
    </EmptyState>
  </div>
  <div v-else>
    <article class="tb-article">
      <div class="tb-article-head">
        <p class="tb-article-kicker">
          <NuxtLink :to="`/b/${handle}`">{{ blogTitle }}</NuxtLink>
        </p>
        <h1 class="tb-article-title">{{ post.title }}</h1>
        <div class="tb-post-row-meta" style="margin-top: 12px">
          <span>{{ post.author_name }}</span>
          <span>·</span>
          <time :datetime="post.published_at || post.created_at">
            {{ formatDate(post.published_at || post.created_at) }}
          </time>
          <span>·</span>
          <span>{{ readingMinutes(post.word_count) }}</span>
          <span>·</span>
          <span>조회 {{ formatCount(post.view_count) }}</span>
        </div>
      </div>
      <img
        v-if="post.cover_image_url"
        :src="post.cover_image_url"
        alt=""
        class="tb-article-cover"
      />
      <div class="tb-article-body" v-html="post.content_html" />
      <div v-if="post.tags?.length" style="margin-bottom: 12px">
        <span v-for="tag in post.tags" :key="tag" class="tb-badge" style="margin-right: 6px">
          #{{ tag }}
        </span>
      </div>
      <div v-if="post.is_paywalled" class="tb-paywall">
        <p class="tb-paywall-title">유료 글입니다</p>
        <p class="tb-paywall-desc">
          {{ (post.paywall_price_minor || 0).toLocaleString() }}원으로 이 글을 열람할 수 있습니다.
        </p>
        <button
          v-if="isLoggedIn"
          class="tb-btn tb-btn-primary"
          :disabled="checkoutBusy"
          @click="unlock"
        >
          {{ checkoutBusy ? '결제 준비 중…' : '열람하기' }}
        </button>
        <NuxtLink v-else to="/login" class="tb-btn tb-btn-primary">로그인 후 열람</NuxtLink>
      </div>
      <div class="tb-article-actions">
        <div v-if="post.allow_reactions" style="display: flex; gap: 6px; flex-wrap: wrap">
          <button
            v-for="kind in reactionKinds"
            :key="kind.id"
            class="tb-chip"
            :aria-pressed="reactions.mine === kind.id"
            :disabled="reactionBusy"
            :title="kind.label"
            @click="react(kind.id)"
          >
            {{ kind.emoji }} {{ reactions.counts[kind.id] || 0 }}
          </button>
        </div>
        <button class="tb-btn tb-btn-secondary tb-btn-sm" :disabled="bookmarkBusy" @click="toggleBookmark">
          {{ bookmarked ? '서재에 저장됨' : '서재에 저장' }}
        </button>
        <button class="tb-btn tb-btn-ghost tb-btn-sm" @click="share">공유</button>
        <button v-if="isLoggedIn" class="tb-btn tb-btn-ghost tb-btn-sm" @click="showReport = true">
          신고
        </button>
      </div>
    </article>

    <div v-if="showReport" class="tb-modal-backdrop" @click.self="showReport = false">
      <div class="tb-modal" role="dialog" aria-modal="true" aria-label="신고하기">
        <h2 class="tb-modal-title">신고하기</h2>
        <form @submit.prevent="submitReport">
          <div class="tb-field">
            <label class="tb-label" for="report-reason">사유</label>
            <select id="report-reason" v-model="reportReason" class="tb-select">
              <option value="SPAM">스팸/광고</option>
              <option value="HATE_SPEECH">혐오 표현</option>
              <option value="DEFAMATION">명예훼손</option>
              <option value="PRIVACY_LEAK">개인정보 노출</option>
              <option value="ADULT_ILLEGAL">불법/음란물</option>
              <option value="PHISHING_MALWARE">피싱/악성코드</option>
              <option value="IMPERSONATION">사칭</option>
              <option value="COPYRIGHT_DMCA">저작권 침해</option>
            </select>
          </div>
          <div class="tb-field">
            <label class="tb-label" for="report-detail">상세 설명</label>
            <textarea
              id="report-detail"
              v-model="reportDetail"
              class="tb-textarea"
              rows="3"
              maxlength="1000"
            />
          </div>
          <div class="tb-modal-actions">
            <button type="button" class="tb-btn tb-btn-ghost" @click="showReport = false">취소</button>
            <button class="tb-btn tb-btn-primary" :disabled="reportBusy">
              {{ reportBusy ? '접수 중…' : '신고 접수' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <section v-if="related.length > 0" aria-label="관련 글" style="margin-top: 28px">
      <h2 class="tb-section-title">관련 글</h2>
      <PostCard v-for="item in related" :key="item.id" :post="item" />
    </section>

    <section aria-label="댓글" style="margin-top: 32px">
      <h2 class="tb-section-title">댓글 {{ formatCount(post.comment_count) }}</h2>
      <form v-if="isLoggedIn && post.allow_comments !== 'DISABLED'" style="margin: 12px 0 20px" @submit.prevent="submitComment">
        <label class="tb-label" for="comment-body">댓글 쓰기</label>
        <textarea
          id="comment-body"
          v-model="commentDraft"
          class="tb-textarea"
          rows="3"
          maxlength="3000"
          placeholder="따뜻한 댓글을 남겨 주세요"
          style="margin-top: 6px"
        />
        <label class="tb-check">
          <input v-model="commentSecret" type="checkbox" />
          <span>비밀 댓글</span>
        </label>
        <div class="tb-form-actions">
          <button class="tb-btn tb-btn-primary tb-btn-sm" :disabled="commentBusy || !commentDraft.trim()">
            {{ commentBusy ? '등록 중…' : '댓글 등록' }}
          </button>
        </div>
      </form>
      <p v-else-if="!isLoggedIn" class="tb-hint">
        댓글을 쓰려면 <NuxtLink to="/login">로그인</NuxtLink>이 필요합니다.
      </p>
      <p v-else-if="post.allow_comments === 'DISABLED'" class="tb-hint">
        댓글이 비활성화된 글입니다.
      </p>
      <ul class="tb-comment-list">
        <li v-for="comment in comments" :key="comment.id" class="tb-comment">
          <div class="tb-comment-head">
            <strong>{{ comment.author_name }}</strong>
            <time :datetime="comment.created_at">{{ timeAgo(comment.created_at) }}</time>
            <span v-if="comment.is_pinned" class="tb-badge">고정</span>
            <button
              v-if="isLoggedIn && comment.author_user_id === user?.id"
              class="tb-btn tb-btn-ghost tb-btn-sm"
              @click="deleteComment(comment.id)"
            >
              삭제
            </button>
          </div>
          <p class="tb-comment-body">{{ comment.content_text }}</p>
        </li>
      </ul>
      <EmptyState v-if="comments.length === 0" title="첫 댓글을 남겨 보세요" />
    </section>
  </div>
</template>

<script setup lang="ts">
import { formatCount, formatDate, readingMinutes, timeAgo } from '@treeblog/ui-kit'
import type { Blog, BookmarkItem, CommentItem, PostDetail, ReactionSummary, SearchPostItem } from '~/types/api'
import { api } from '~/utils/api'

const reactionKinds = [
  { id: 'LIKE', label: '좋아요', emoji: '♥' },
  { id: 'LOVE', label: '사랑해요', emoji: '💕' },
  { id: 'HELPFUL', label: '유용해요', emoji: '👍' },
  { id: 'FUNNY', label: '재밌어요', emoji: '😄' },
  { id: 'TOUCHING', label: '감동이에요', emoji: '🥲' },
  { id: 'INSIGHTFUL', label: '통찰적이에요', emoji: '💡' }
]

const route = useRoute()
const handle = computed(() => String(route.params.handle || ''))
const slug = computed(() => String(route.params.slug || ''))
const { isLoggedIn, user } = useAuth()
const { busy: checkoutBusy, unlockPost } = useCheckout()
const toast = useToast()

const post = ref<PostDetail | null>(null)
const blogTitle = ref('')
const pending = ref(true)
const failed = ref(false)
const reactions = ref<ReactionSummary>({ counts: {}, total: 0, mine: null })
const reactionBusy = ref(false)
const bookmarked = ref(false)
const bookmarkBusy = ref(false)
const comments = ref<CommentItem[]>([])
const commentDraft = ref('')
const commentSecret = ref(false)
const commentBusy = ref(false)
const related = ref<SearchPostItem[]>([])
const showReport = ref(false)
const reportReason = ref('SPAM')
const reportDetail = ref('')
const reportBusy = ref(false)

useSeoMeta({
  title: () => (post.value ? `${post.value.title} | TreeBlog` : '글 | TreeBlog'),
  description: () => post.value?.excerpt || 'TreeBlog 글',
  ogTitle: () => post.value?.title || 'TreeBlog',
  ogImage: () => post.value?.cover_image_url || '/og-default.png'
})

async function load(): Promise<void> {
  pending.value = true
  failed.value = false
  try {
    post.value = await api<PostDetail>(`/api/v1/public/blogs/${handle.value}/posts/${slug.value}`, {
      auth: isLoggedIn.value
    })
    const postId = post.value.id
    const settled = await Promise.allSettled([
      api<Blog>(`/api/v1/blogs/${handle.value}`, { auth: false }),
      api<ReactionSummary>(`/api/v1/posts/${postId}/reactions`, { auth: isLoggedIn.value }),
      api<CommentItem[]>(`/api/v1/posts/${postId}/comments`, {
        auth: isLoggedIn.value,
        query: { limit: 50 }
      }),
      api<SearchPostItem[]>(`/api/v1/public/posts/${postId}/related`, {
        auth: false,
        query: { limit: 3 }
      })
    ])
    if (settled[0].status === 'fulfilled') blogTitle.value = settled[0].value.title
    if (settled[1].status === 'fulfilled') reactions.value = settled[1].value
    if (settled[2].status === 'fulfilled') comments.value = settled[2].value
    if (settled[3].status === 'fulfilled') related.value = settled[3].value
    if (isLoggedIn.value) {
      try {
        const bookmarks = await api<BookmarkItem[]>('/api/v1/me/bookmarks', { query: { limit: 100 } })
        bookmarked.value = bookmarks.some((item) => item.post_id === postId)
      } catch {
        bookmarked.value = false
      }
    }
  } catch {
    failed.value = true
  } finally {
    pending.value = false
  }
}

async function react(kind: string): Promise<void> {
  if (!post.value) return
  if (!isLoggedIn.value) {
    await navigateTo('/login')
    return
  }
  reactionBusy.value = true
  try {
    if (reactions.value.mine === kind) {
      reactions.value = await api<ReactionSummary>(`/api/v1/posts/${post.value.id}/reactions`, {
        method: 'DELETE'
      })
    } else {
      reactions.value = await api<ReactionSummary>(`/api/v1/posts/${post.value.id}/reactions`, {
        method: 'POST',
        body: { reaction_type: kind }
      })
    }
  } catch {
    toast.push('리액션 처리에 실패했습니다', 'error')
  } finally {
    reactionBusy.value = false
  }
}

async function toggleBookmark(): Promise<void> {
  if (!post.value) return
  if (!isLoggedIn.value) {
    await navigateTo('/login')
    return
  }
  bookmarkBusy.value = true
  try {
    await api(`/api/v1/posts/${post.value.id}/bookmark`, {
      method: bookmarked.value ? 'DELETE' : 'POST',
      body: bookmarked.value ? undefined : {}
    })
    bookmarked.value = !bookmarked.value
    toast.push(bookmarked.value ? '서재에 저장했습니다' : '서재에서 제거했습니다', 'success')
  } catch {
    toast.push('서재 저장에 실패했습니다', 'error')
  } finally {
    bookmarkBusy.value = false
  }
}

async function submitComment(): Promise<void> {
  if (!post.value) return
  const content = commentDraft.value.trim()
  if (!content) return
  commentBusy.value = true
  try {
    const created = await api<CommentItem>(`/api/v1/posts/${post.value.id}/comments`, {
      method: 'POST',
      body: { content_text: content, is_secret: commentSecret.value }
    })
    comments.value = [...comments.value, created]
    if (post.value) post.value.comment_count += 1
    commentDraft.value = ''
    commentSecret.value = false
    toast.push('댓글을 등록했습니다', 'success')
  } catch {
    toast.push('댓글 등록에 실패했습니다', 'error')
  } finally {
    commentBusy.value = false
  }
}

async function deleteComment(commentId: string): Promise<void> {
  if (!confirm('댓글을 삭제할까요?')) return
  try {
    await api(`/api/v1/comments/${commentId}/`, { method: 'DELETE' })
    comments.value = comments.value.filter((comment) => comment.id !== commentId)
    if (post.value) post.value.comment_count = Math.max(0, post.value.comment_count - 1)
  } catch {
    toast.push('댓글 삭제에 실패했습니다', 'error')
  }
}

async function share(): Promise<void> {
  const url = window.location.href
  try {
    await navigator.clipboard.writeText(url)
    toast.push('링크를 복사했습니다', 'success')
  } catch {
    toast.push(url, 'info')
  }
}

async function unlock(): Promise<void> {
  if (!post.value) return
  await unlockPost(
    post.value.blog_id,
    handle.value,
    post.value.id,
    post.value.paywall_price_minor || 0,
    post.value.title
  )
}

async function submitReport(): Promise<void> {
  if (!post.value) return
  reportBusy.value = true
  try {
    await api('/api/v1/reports/', {
      method: 'POST',
      body: {
        target_type: 'POST',
        target_id: post.value.id,
        reason_category: reportReason.value,
        description: reportDetail.value.trim()
      }
    })
    showReport.value = false
    reportDetail.value = ''
    toast.push('신고가 접수되었습니다', 'success')
  } catch {
    toast.push('신고 접수에 실패했습니다', 'error')
  } finally {
    reportBusy.value = false
  }
}

onMounted(load)
</script>

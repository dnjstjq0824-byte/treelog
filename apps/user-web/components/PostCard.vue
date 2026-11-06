<template>
  <article class="tb-post-row">
    <div class="tb-post-row-main">
      <div class="tb-post-row-meta">
        <span class="tb-avatar" style="width: 22px; height: 22px; font-size: 0.65rem" aria-hidden="true">
          {{ initials(post.author_name) }}
        </span>
        <span>{{ post.author_name }}</span>
        <span aria-hidden="true">·</span>
        <span>{{ blogTitle }}</span>
        <span aria-hidden="true">·</span>
        <time :datetime="stamp">{{ timeAgo(stamp) }}</time>
      </div>
      <h3 class="tb-post-row-title">
        <NuxtLink :to="`/b/${blogHandle}/${post.slug}`">{{ post.title }}</NuxtLink>
      </h3>
      <p v-if="post.excerpt" class="tb-post-row-excerpt">{{ post.excerpt }}</p>
      <div class="tb-post-row-meta">
        <span v-if="post.category_name" class="tb-badge">{{ post.category_name }}</span>
        <span>조회 {{ formatCount(post.view_count) }}</span>
        <span>댓글 {{ formatCount(post.comment_count) }}</span>
        <span>{{ readingMinutes(post.word_count) }}</span>
      </div>
    </div>
    <NuxtLink
      v-if="post.cover_image_url"
      :to="`/b/${blogHandle}/${post.slug}`"
      :aria-label="`${post.title} 글 읽기`"
      tabindex="-1"
    >
      <img class="tb-post-row-thumb" :src="post.cover_image_url" :alt="''" loading="lazy" />
    </NuxtLink>
  </article>
</template>

<script setup lang="ts">
import { formatCount, initials, readingMinutes, timeAgo } from '@treeblog/ui-kit'
import type { PostSummary, SearchPostItem } from '~/types/api'

const props = defineProps<{
  post: PostSummary | SearchPostItem
  blogHandle?: string
  blogTitle?: string
}>()

const stamp = computed(() => {
  if (props.post.published_at) return props.post.published_at
  if ('created_at' in props.post) return props.post.created_at
  return ''
})

const blogHandle = computed(() => {
  if (props.blogHandle) return props.blogHandle
  if ('blog_handle' in props.post && props.post.blog_handle) return props.post.blog_handle
  return ''
})

const blogTitle = computed(() => {
  if (props.blogTitle) return props.blogTitle
  if ('blog_title' in props.post && props.post.blog_title) return props.post.blog_title
  return ''
})
</script>

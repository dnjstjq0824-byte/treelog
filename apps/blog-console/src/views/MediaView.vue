<template>
  <div>
    <div class="tb-console-head">
      <h1 class="tb-console-title">미디어</h1>
      <div style="display: flex; gap: 8px">
        <button class="tb-btn tb-btn-secondary tb-btn-sm" :disabled="uploading" @click="pickFile">
          {{ uploading ? '업로드 중…' : '이미지 업로드' }}
        </button>
        <button class="tb-btn tb-btn-secondary tb-btn-sm" @click="showEmbed = true">동영상 임베드</button>
        <input ref="fileInput" type="file" accept="image/*,video/*" hidden multiple @change="onFilesPicked" />
      </div>
    </div>

    <div v-if="showEmbed" class="tb-modal-backdrop" @click.self="showEmbed = false">
      <div class="tb-modal" role="dialog" aria-modal="true" aria-label="동영상 임베드">
        <h2 class="tb-modal-title">동영상 임베드</h2>
        <div class="tb-field">
          <label class="tb-label" for="embed-url">동영상 URL (YouTube 등)</label>
          <input id="embed-url" v-model="embedUrl" class="tb-input" placeholder="https://" />
        </div>
        <div class="tb-modal-actions">
          <button class="tb-btn tb-btn-ghost" @click="showEmbed = false">취소</button>
          <button class="tb-btn tb-btn-primary" :disabled="embedBusy" @click="createEmbed">
            {{ embedBusy ? '등록 중…' : '등록' }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="editing" class="tb-modal-backdrop" @click.self="editing = null">
      <div class="tb-modal" role="dialog" aria-modal="true" aria-label="메타데이터 편집">
        <h2 class="tb-modal-title">메타데이터 편집</h2>
        <div class="tb-field">
          <label class="tb-label" for="meta-alt">대체 텍스트</label>
          <input id="meta-alt" v-model="editing.alt_text" class="tb-input" maxlength="200" />
        </div>
        <div class="tb-field">
          <label class="tb-label" for="meta-caption">캡션</label>
          <input id="meta-caption" v-model="editing.caption" class="tb-input" maxlength="300" />
        </div>
        <div class="tb-modal-actions">
          <button class="tb-btn tb-btn-ghost" @click="editing = null">취소</button>
          <button class="tb-btn tb-btn-primary" :disabled="metaBusy" @click="saveMeta">
            {{ metaBusy ? '저장 중…' : '저장' }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="pending">
      <div class="tb-skeleton" style="height: 300px" />
    </div>
    <template v-else-if="items.length > 0">
      <div class="tb-media-grid">
        <div v-for="asset in items" :key="asset.id" class="tb-media-card">
          <img
            v-if="asset.media_type === 'IMAGE'"
            :src="publicMediaURL(asset.id)"
            :alt="asset.alt_text || '미디어'"
            loading="lazy"
          />
          <div v-else style="aspect-ratio: 4/3; display: flex; align-items: center; justify-content: center; background: var(--tb-bg-muted)">
            <span class="tb-badge">{{ asset.embed_provider || asset.media_type }}</span>
          </div>
          <div class="tb-media-card-body">
            <p :title="asset.mime_type">{{ asset.mime_type }} · {{ asset.status }}</p>
            <p class="tb-hint">{{ Math.round(asset.original_size_bytes / 1024) }} KB</p>
            <div style="display: flex; gap: 6px; margin-top: 6px">
              <button class="tb-btn tb-btn-ghost tb-btn-sm" @click="copyUrl(asset.id)">URL 복사</button>
              <button class="tb-btn tb-btn-ghost tb-btn-sm" @click="startEdit(asset)">편집</button>
            </div>
          </div>
        </div>
      </div>
      <div v-if="hasMore" class="tb-load-more">
        <button class="tb-btn tb-btn-secondary" :disabled="loadingMore" @click="loadMore">
          {{ loadingMore ? '불러오는 중…' : '더 보기' }}
        </button>
      </div>
    </template>
    <div v-else class="tb-empty">
      <p class="tb-empty-title">미디어가 없습니다</p>
      <p class="tb-empty-desc">이미지를 업로드해 보세요.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { api, apiRaw, friendlyMessage } from '../lib/api'
import { publicMediaURL, uploadImage } from '../lib/media'
import type { MediaAsset } from '../lib/types'
import { useToastStore } from '../stores/toast'

const route = useRoute()
const toast = useToastStore()
const blogId = computed(() => String(route.params.blogId || ''))

const items = ref<MediaAsset[]>([])
const cursor = ref('')
const hasMore = ref(false)
const pending = ref(true)
const loadingMore = ref(false)
const uploading = ref(false)
const showEmbed = ref(false)
const embedUrl = ref('')
const embedBusy = ref(false)
const editing = ref<MediaAsset | null>(null)
const metaBusy = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

async function fetchPage(append: boolean): Promise<void> {
  try {
    const envelope = await apiRaw<MediaAsset[]>(`/api/v1/blogs/${blogId.value}/media/`, {
      query: append && cursor.value ? { limit: 30, cursor: cursor.value } : { limit: 30 }
    })
    items.value = append ? [...items.value, ...envelope.data] : envelope.data
    cursor.value = envelope.next_cursor || ''
    hasMore.value = cursor.value !== ''
  } catch (err) {
    if (!append) items.value = []
    toast.push(friendlyMessage(err), 'error')
  } finally {
    pending.value = false
    loadingMore.value = false
  }
}

async function loadMore(): Promise<void> {
  if (!hasMore.value || loadingMore.value) return
  loadingMore.value = true
  await fetchPage(true)
}

function pickFile(): void {
  fileInput.value?.click()
}

async function onFilesPicked(event: Event): Promise<void> {
  const files = [...((event.target as HTMLInputElement).files || [])]
  if (files.length === 0) return
  uploading.value = true
  try {
    for (const file of files) {
      const asset = await uploadImage(blogId.value, file)
      items.value = [asset, ...items.value]
    }
    toast.push(`${files.length}개 업로드했습니다`, 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  } finally {
    uploading.value = false
    if (fileInput.value) fileInput.value.value = ''
  }
}

async function createEmbed(): Promise<void> {
  const url = embedUrl.value.trim()
  if (!url) return
  embedBusy.value = true
  try {
    const asset = await api<MediaAsset>(`/api/v1/blogs/${blogId.value}/media/video-embed`, {
      method: 'POST',
      body: { video_url: url }
    })
    items.value = [asset, ...items.value]
    showEmbed.value = false
    embedUrl.value = ''
    toast.push('동영상을 등록했습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  } finally {
    embedBusy.value = false
  }
}

function startEdit(asset: MediaAsset): void {
  editing.value = { ...asset }
}

async function saveMeta(): Promise<void> {
  if (!editing.value) return
  metaBusy.value = true
  try {
    const updated = await api<MediaAsset>(
      `/api/v1/blogs/${blogId.value}/media/${editing.value.id}`,
      {
        method: 'PATCH',
        body: { alt_text: editing.value.alt_text, caption: editing.value.caption }
      }
    )
    const index = items.value.findIndex((item) => item.id === updated.id)
    if (index >= 0) items.value[index] = updated
    editing.value = null
    toast.push('저장했습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  } finally {
    metaBusy.value = false
  }
}

async function copyUrl(assetId: string): Promise<void> {
  try {
    await navigator.clipboard.writeText(publicMediaURL(assetId))
    toast.push('URL을 복사했습니다', 'success')
  } catch {
    toast.push(publicMediaURL(assetId), 'info')
  }
}

onMounted(() => fetchPage(false))
</script>

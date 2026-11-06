<template>
  <div>
    <div v-if="pending">
      <div class="tb-skeleton" style="width: 60%; height: 40px" />
      <div class="tb-skeleton" style="height: 400px" />
    </div>
    <div v-else-if="failed" class="tb-empty">
      <p class="tb-empty-title">글을 불러오지 못했습니다</p>
      <RouterLink :to="`/b/${blogId}/posts`" class="tb-btn tb-btn-secondary tb-btn-sm">목록으로</RouterLink>
    </div>
    <div v-else>
      <div class="tb-console-head">
        <h1 class="tb-console-title">{{ isNew ? '새 글 쓰기' : '글 편집' }}</h1>
        <div style="display: flex; gap: 8px; align-items: center">
          <span class="tb-hint">{{ saveState }}</span>
          <button class="tb-btn tb-btn-secondary tb-btn-sm" :disabled="publishing" @click="saveNow">
            저장
          </button>
          <button class="tb-btn tb-btn-primary tb-btn-sm" :disabled="publishing" @click="showPublish = true">
            {{ publishing ? '발행 중…' : '발행 설정' }}
          </button>
        </div>
      </div>

      <div v-if="conflict" class="tb-conflict" role="alert">
        <p><strong>동시 편집 충돌.</strong> 다른 기기에서 먼저 저장했습니다 (리비전 {{ conflict.server_revision_no }}).</p>
        <div style="display: flex; gap: 8px">
          <button class="tb-btn tb-btn-secondary tb-btn-sm" @click="applyServerVersion">
            서버 버전 적용
          </button>
          <button class="tb-btn tb-btn-primary tb-btn-sm" @click="keepMyVersion">
            내 버전으로 덮어쓰기
          </button>
        </div>
      </div>

      <div class="tb-editor-wrap">
        <div>
          <input
            v-model="title"
            class="tb-editor-title-input"
            placeholder="제목을 입력하세요"
            aria-label="제목"
            @input="scheduleSave"
          />
          <input
            v-model="subtitle"
            class="tb-input"
            placeholder="부제 (선택)"
            aria-label="부제"
            style="margin-bottom: 12px"
            @input="scheduleSave"
          />
          <div v-if="editor" class="tb-editor-toolbar" role="toolbar" aria-label="서식">
            <button :class="{ 'is-active': editor.isActive('bold') }" @click="editor.chain().focus().toggleBold().run()"><strong>B</strong></button>
            <button :class="{ 'is-active': editor.isActive('italic') }" @click="editor.chain().focus().toggleItalic().run()"><em>I</em></button>
            <button :class="{ 'is-active': editor.isActive('strike') }" @click="editor.chain().focus().toggleStrike().run()"><s>S</s></button>
            <button :class="{ 'is-active': editor.isActive('heading', { level: 2 }) }" @click="editor.chain().focus().toggleHeading({ level: 2 }).run()">H2</button>
            <button :class="{ 'is-active': editor.isActive('heading', { level: 3 }) }" @click="editor.chain().focus().toggleHeading({ level: 3 }).run()">H3</button>
            <button :class="{ 'is-active': editor.isActive('bulletList') }" @click="editor.chain().focus().toggleBulletList().run()">• 목록</button>
            <button :class="{ 'is-active': editor.isActive('orderedList') }" @click="editor.chain().focus().toggleOrderedList().run()">1. 목록</button>
            <button :class="{ 'is-active': editor.isActive('blockquote') }" @click="editor.chain().focus().toggleBlockquote().run()">인용</button>
            <button :class="{ 'is-active': editor.isActive('codeBlock') }" @click="editor.chain().focus().toggleCodeBlock().run()">코드</button>
            <button @click="insertLink">링크</button>
            <button :disabled="uploading" @click="pickImage">{{ uploading ? '업로드 중…' : '이미지' }}</button>
            <button @click="editor.chain().focus().setHorizontalRule().run()">구분선</button>
            <input
              ref="fileInput"
              type="file"
              accept="image/*"
              hidden
              @change="onFilePicked"
            />
          </div>
          <div class="tb-editor-body">
            <EditorContent :editor="editor" />
          </div>
          <div class="tb-side-card" style="margin-top: 16px">
            <h3>리비전 기록</h3>
            <ul v-if="revisions.length > 0" class="tb-list">
              <li v-for="revision in revisions" :key="revision.id" class="tb-list-item">
                <div>
                  <p class="tb-list-item-title">#{{ revision.revision_no }} {{ revision.title || '(제목 없음)' }}</p>
                  <p class="tb-list-item-desc">
                    {{ formatDateTime(revision.created_at) }}
                    <span v-if="revision.is_published_snapshot"> · 발행 스냅샷</span>
                  </p>
                </div>
                <button class="tb-btn tb-btn-ghost tb-btn-sm" @click="restoreRevision(revision.id)">
                  복원
                </button>
              </li>
            </ul>
            <p v-else class="tb-hint">아직 리비전이 없습니다.</p>
          </div>
        </div>

        <aside>
          <div class="tb-side-card">
            <h3>태그</h3>
            <input
              v-model="tagsText"
              class="tb-input"
              placeholder="쉼표로 구분 (예: 에세이, 일상)"
            />
          </div>
          <div class="tb-side-card">
            <h3>링크 미리보기</h3>
            <div style="display: flex; gap: 6px">
              <input v-model="previewUrl" class="tb-input" placeholder="https://" />
              <button class="tb-btn tb-btn-secondary tb-btn-sm" @click="fetchPreview">확인</button>
            </div>
            <div v-if="preview" style="margin-top: 8px">
              <p class="tb-list-item-title">{{ preview.title || preview.url }}</p>
              <p class="tb-list-item-desc">{{ preview.description }}</p>
            </div>
          </div>
          <div class="tb-side-card">
            <h3>Markdown 가져오기</h3>
            <textarea v-model="markdownDraft" class="tb-textarea" rows="4" placeholder="Markdown 붙여넣기" />
            <div class="tb-form-actions">
              <button class="tb-btn tb-btn-secondary tb-btn-sm" @click="importMarkdown">에디터에 삽입</button>
              <button class="tb-btn tb-btn-ghost tb-btn-sm" @click="exportMarkdown">현재 글 → MD</button>
            </div>
            <pre v-if="markdownOut" class="tb-code" style="margin-top: 8px; white-space: pre-wrap">{{ markdownOut }}</pre>
          </div>
        </aside>
      </div>

      <div v-if="showPublish" class="tb-modal-backdrop" @click.self="showPublish = false">
        <div class="tb-modal" role="dialog" aria-modal="true" aria-label="발행 설정" style="max-width: 560px">
          <h2 class="tb-modal-title">발행 설정</h2>
          <div class="tb-field">
            <label class="tb-label" for="pub-slug">슬러그 (비우면 자동 생성)</label>
            <input id="pub-slug" v-model="publish.slug" class="tb-input" placeholder="my-post-slug" />
          </div>
          <div class="tb-field">
            <label class="tb-label" for="pub-category">카테고리</label>
            <select id="pub-category" v-model="publish.category_id" class="tb-select">
              <option :value="null">선택 안 함</option>
              <option v-for="category in categories" :key="category.id" :value="category.id">
                {{ category.name }}
              </option>
            </select>
          </div>
          <div class="tb-field">
            <label class="tb-label" for="pub-visibility">공개 범위</label>
            <select id="pub-visibility" v-model="publish.visibility" class="tb-select">
              <option v-for="option in VISIBILITIES" :key="option.id" :value="option.id">
                {{ option.label }}
              </option>
            </select>
          </div>
          <div class="tb-field">
            <label class="tb-label" for="pub-comments">댓글 정책</label>
            <select id="pub-comments" v-model="publish.allow_comments" class="tb-select">
              <option v-for="option in COMMENT_POLICIES" :key="option.id" :value="option.id">
                {{ option.label }}
              </option>
            </select>
          </div>
          <div class="tb-field">
            <label class="tb-label" for="pub-excerpt">요약 (비우면 자동 생성)</label>
            <textarea id="pub-excerpt" v-model="publish.excerpt" class="tb-textarea" rows="2" maxlength="600" />
          </div>
          <div class="tb-field">
            <label class="tb-label" for="pub-license">라이선스</label>
            <select id="pub-license" v-model="publish.license_type" class="tb-select">
              <option v-for="option in LICENSES" :key="option.id" :value="option.id">
                {{ option.label }}
              </option>
            </select>
          </div>
          <div class="tb-field">
            <label class="tb-label" for="pub-ai">AI 사용 고지</label>
            <select id="pub-ai" v-model="publish.ai_disclosure_level" class="tb-select">
              <option v-for="option in AI_LEVELS" :key="option.id" :value="option.id">
                {{ option.label }}
              </option>
            </select>
          </div>
          <div class="tb-field">
            <label class="tb-label" for="pub-paywall">유료 가격 (원, 0이면 무료)</label>
            <input id="pub-paywall" v-model.number="publish.paywall_price_minor" class="tb-input" type="number" min="0" step="100" />
          </div>
          <div class="tb-field">
            <label class="tb-label" for="pub-scheduled">예약 발행 (비우면 즉시)</label>
            <input id="pub-scheduled" v-model="publish.scheduled_for" class="tb-input" type="datetime-local" />
          </div>
          <label class="tb-check">
            <input v-model="publish.allow_reactions" type="checkbox" />
            <span>리액션 허용</span>
          </label>
          <label class="tb-check">
            <input v-model="publish.is_notice" type="checkbox" />
            <span>공지로 등록</span>
          </label>
          <label class="tb-check">
            <input v-model="publish.is_pinned" type="checkbox" />
            <span>상단 고정</span>
          </label>
          <div class="tb-modal-actions">
            <button class="tb-btn tb-btn-ghost" @click="showPublish = false">취소</button>
            <button class="tb-btn tb-btn-primary" :disabled="publishing" @click="publishNow">
              {{ publishing ? '발행 중…' : '발행하기' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Image from '@tiptap/extension-image'
import Link from '@tiptap/extension-link'
import Placeholder from '@tiptap/extension-placeholder'
import StarterKit from '@tiptap/starter-kit'
import { EditorContent, useEditor } from '@tiptap/vue-3'
import { formatDateTime } from '@treeblog/ui-kit'
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ApiError, api, friendlyMessage } from '../lib/api'
import { publicMediaURL, uploadImage } from '../lib/media'
import {
  AI_LEVELS,
  COMMENT_POLICIES,
  LICENSES,
  VISIBILITIES,
  type AutosaveConflict,
  type BlogCategory,
  type DraftCreated,
  type PostDetail,
  type RevisionDetail,
  type RevisionSummary
} from '../lib/types'
import { useToastStore } from '../stores/toast'

const route = useRoute()
const router = useRouter()
const toast = useToastStore()
const blogId = computed(() => String(route.params.blogId || ''))
const postIdParam = computed(() => String(route.params.postId || ''))
const isNew = computed(() => route.name === 'post-new')

const pending = ref(true)
const failed = ref(false)
const postId = ref('')
const baseRevision = ref(1)
const title = ref('')
const subtitle = ref('')
const tagsText = ref('')
const saveState = ref('저장 대기')
const publishing = ref(false)
const uploading = ref(false)
const showPublish = ref(false)
const conflict = ref<AutosaveConflict | null>(null)
const revisions = ref<RevisionSummary[]>([])
const categories = ref<BlogCategory[]>([])
const previewUrl = ref('')
const preview = ref<{ url: string; title: string; description: string } | null>(null)
const markdownDraft = ref('')
const markdownOut = ref('')
const fileInput = ref<HTMLInputElement | null>(null)

const publish = ref({
  slug: '',
  category_id: null as string | null,
  visibility: 'PUBLIC',
  allow_comments: 'ALL_ALLOWED',
  excerpt: '',
  license_type: 'ALL_RIGHTS_RESERVED',
  ai_disclosure_level: 'NONE',
  paywall_price_minor: 0,
  scheduled_for: '',
  allow_reactions: true,
  is_notice: false,
  is_pinned: false
})

const editor = useEditor({
  extensions: [
    StarterKit,
    Image.configure({ inline: false }),
    Link.configure({ openOnClick: false }),
    Placeholder.configure({ placeholder: '이야기를 시작하세요…' })
  ],
  content: { type: 'doc', content: [] },
  onUpdate: () => scheduleSave()
})

let saveTimer: ReturnType<typeof setTimeout> | null = null

function scheduleSave(): void {
  if (!postId.value || conflict.value) return
  saveState.value = '편집 중…'
  if (saveTimer) clearTimeout(saveTimer)
  saveTimer = setTimeout(() => void saveNow(), 2000)
}

async function saveNow(): Promise<void> {
  if (!postId.value || !editor.value || conflict.value) return
  if (saveTimer) {
    clearTimeout(saveTimer)
    saveTimer = null
  }
  saveState.value = '저장 중…'
  try {
    await api(`/api/v1/blogs/${blogId.value}/posts/${postId.value}/autosave`, {
      method: 'PUT',
      body: {
        base_revision_no: baseRevision.value,
        title: title.value,
        subtitle: subtitle.value,
        content_json: editor.value.getJSON(),
        client_time: new Date().toISOString()
      }
    })
    baseRevision.value += 1
    saveState.value = `저장됨 ${new Date().toLocaleTimeString('ko-KR')}`
  } catch (err) {
    if (err instanceof ApiError && err.status === 409 && err.data) {
      conflict.value = err.data as AutosaveConflict
      saveState.value = '충돌 발생'
    } else {
      saveState.value = '저장 실패'
      toast.push(friendlyMessage(err), 'error')
    }
  }
}

function applyServerVersion(): void {
  if (!conflict.value || !editor.value) return
  title.value = conflict.value.server_title
  subtitle.value = conflict.value.server_subtitle
  editor.value.commands.setContent(conflict.value.server_content as never)
  baseRevision.value = conflict.value.server_revision_no
  conflict.value = null
  saveState.value = '서버 버전 적용됨'
}

async function keepMyVersion(): Promise<void> {
  if (!conflict.value) return
  baseRevision.value = conflict.value.server_revision_no
  conflict.value = null
  await saveNow()
}

function pickImage(): void {
  fileInput.value?.click()
}

async function onFilePicked(event: Event): Promise<void> {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file || !editor.value) return
  uploading.value = true
  try {
    const asset = await uploadImage(blogId.value, file)
    editor.value.chain().focus().setImage({ src: publicMediaURL(asset.id), alt: file.name }).run()
    scheduleSave()
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  } finally {
    uploading.value = false
    if (fileInput.value) fileInput.value.value = ''
  }
}

function insertLink(): void {
  if (!editor.value) return
  const url = window.prompt('링크 URL을 입력하세요', 'https://')
  if (!url) return
  editor.value.chain().focus().setLink({ href: url }).run()
  scheduleSave()
}

async function fetchPreview(): Promise<void> {
  const url = previewUrl.value.trim()
  if (!url) return
  try {
    preview.value = await api(`/api/v1/blogs/${blogId.value}/editor/link-preview`, {
      method: 'POST',
      body: { url }
    })
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  }
}

async function importMarkdown(): Promise<void> {
  const markdown = markdownDraft.value
  if (!markdown.trim() || !editor.value) return
  try {
    const result = await api<{ ast: unknown }>(`/api/v1/blogs/${blogId.value}/editor/markdown-convert`, {
      method: 'POST',
      body: { direction: 'TO_AST', markdown }
    })
    editor.value.commands.setContent(result.ast as never)
    scheduleSave()
    toast.push('Markdown을 가져왔습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  }
}

async function exportMarkdown(): Promise<void> {
  if (!editor.value) return
  try {
    const result = await api<{ markdown: string }>(`/api/v1/blogs/${blogId.value}/editor/markdown-convert`, {
      method: 'POST',
      body: { direction: 'TO_MARKDOWN', ast: JSON.stringify(editor.value.getJSON()) }
    })
    markdownOut.value = result.markdown
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  }
}

async function loadRevisions(): Promise<void> {
  try {
    revisions.value = await api<RevisionSummary[]>(
      `/api/v1/blogs/${blogId.value}/posts/${postId.value}/revisions`
    )
  } catch {
    revisions.value = []
  }
}

async function restoreRevision(revisionId: string): Promise<void> {
  if (!confirm('이 리비전으로 복원할까요? 현재 내용은 새 리비전으로 남습니다.')) return
  try {
    const detail = await api<RevisionDetail>(
      `/api/v1/blogs/${blogId.value}/posts/${postId.value}/revisions/${revisionId}/restore`,
      { method: 'POST', body: {} }
    )
    title.value = detail.title
    subtitle.value = detail.subtitle
    editor.value?.commands.setContent(detail.content_json as never)
    baseRevision.value = detail.revision_no + 1
    await loadRevisions()
    toast.push('리비전을 복원했습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  }
}

function parseTags(): string[] {
  return [...new Set(tagsText.value.split(',').map((tag) => tag.trim()).filter(Boolean))]
}

async function publishNow(): Promise<void> {
  if (!editor.value) return
  if (!title.value.trim()) {
    toast.push('제목을 입력하세요', 'error')
    return
  }
  publishing.value = true
  try {
    await api(`/api/v1/blogs/${blogId.value}/posts/${postId.value}/publish`, {
      method: 'POST',
      body: {
        title: title.value.trim(),
        subtitle: subtitle.value.trim(),
        slug: publish.value.slug.trim() || undefined,
        category_id: publish.value.category_id,
        content_json: editor.value.getJSON(),
        content_html: editor.value.getHTML(),
        content_plain: editor.value.getText(),
        excerpt: publish.value.excerpt.trim() || undefined,
        tags: parseTags(),
        visibility: publish.value.visibility,
        allow_comments: publish.value.allow_comments,
        allow_reactions: publish.value.allow_reactions,
        is_notice: publish.value.is_notice,
        is_pinned: publish.value.is_pinned,
        license_type: publish.value.license_type,
        ai_disclosure_level: publish.value.ai_disclosure_level,
        paywall_price_minor: publish.value.paywall_price_minor > 0 ? publish.value.paywall_price_minor : undefined,
        scheduled_for: publish.value.scheduled_for ? new Date(publish.value.scheduled_for).toISOString() : undefined
      }
    })
    toast.push('발행했습니다', 'success')
    await router.push(`/b/${blogId.value}/posts`)
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  } finally {
    publishing.value = false
  }
}

onMounted(async () => {
  try {
    categories.value = await api<BlogCategory[]>(`/api/v1/blogs/${blogId.value}/categories`)
  } catch {
    categories.value = []
  }
  try {
    if (isNew.value) {
      const draft = await api<DraftCreated>(`/api/v1/blogs/${blogId.value}/posts`, {
        method: 'POST',
        body: { title: '' }
      })
      postId.value = draft.id
      baseRevision.value = draft.revision_no
      await router.replace(`/b/${blogId.value}/posts/${draft.id}`)
    } else {
      postId.value = postIdParam.value
      const post = await api<PostDetail>(`/api/v1/blogs/${blogId.value}/posts/${postId.value}/`)
      title.value = post.title
      subtitle.value = post.subtitle
      tagsText.value = (post.tags || []).join(', ')
      baseRevision.value = post.revision_no
      publish.value.slug = post.slug
      publish.value.category_id = post.category_id || null
      publish.value.visibility = post.visibility
      publish.value.allow_comments = post.allow_comments
      publish.value.excerpt = post.excerpt
      publish.value.license_type = post.license_type
      publish.value.ai_disclosure_level = post.ai_disclosure_level
      publish.value.paywall_price_minor = post.paywall_price_minor || 0
      publish.value.allow_reactions = post.allow_reactions
      publish.value.is_notice = post.is_notice
      publish.value.is_pinned = post.is_pinned
      editor.value?.commands.setContent((post.content_json || { type: 'doc', content: [] }) as never)
      await loadRevisions()
    }
  } catch (err) {
    failed.value = true
    toast.push(friendlyMessage(err), 'error')
  } finally {
    pending.value = false
  }
})

onBeforeUnmount(() => {
  if (saveTimer) clearTimeout(saveTimer)
  editor.value?.destroy()
})
</script>

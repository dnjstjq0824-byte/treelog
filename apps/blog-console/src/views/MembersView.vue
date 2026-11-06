<template>
  <div>
    <div class="tb-console-head">
      <h1 class="tb-console-title">멤버</h1>
      <button class="tb-btn tb-btn-primary tb-btn-sm" @click="showInvite = true">멤버 초대</button>
    </div>

    <div v-if="showInvite" class="tb-modal-backdrop" @click.self="showInvite = false">
      <div class="tb-modal" role="dialog" aria-modal="true" aria-label="멤버 초대">
        <h2 class="tb-modal-title">멤버 초대</h2>
        <div class="tb-field">
          <label class="tb-label" for="invite-email">이메일</label>
          <input id="invite-email" v-model="email" class="tb-input" type="email" required />
        </div>
        <div class="tb-field">
          <label class="tb-label" for="invite-role">역할</label>
          <select id="invite-role" v-model="role" class="tb-select">
            <option value="AUTHOR">필자</option>
            <option value="EDITOR">편집자</option>
            <option value="ADMIN">관리자</option>
          </select>
        </div>
        <div class="tb-modal-actions">
          <button class="tb-btn tb-btn-ghost" @click="showInvite = false">취소</button>
          <button class="tb-btn tb-btn-primary" :disabled="busy || !email.trim()" @click="invite">
            {{ busy ? '초대 중…' : '초대' }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="pending">
      <div v-for="index in 4" :key="index" class="tb-skeleton" style="height: 56px" />
    </div>
    <ul v-else-if="items.length > 0" class="tb-list">
      <li v-for="item in items" :key="item.id" class="tb-list-item">
        <div>
          <p class="tb-list-item-title">{{ item.display_name || item.username || item.user_id }}</p>
          <p class="tb-list-item-desc">{{ roleLabel(item.role) }} · {{ inviteLabel(item.invite_status) }}</p>
        </div>
        <button
          v-if="item.role !== 'OWNER'"
          class="tb-btn tb-btn-ghost tb-btn-sm"
          @click="remove(item.user_id)"
        >
          제거
        </button>
      </li>
    </ul>
    <div v-else class="tb-empty">
      <p class="tb-empty-title">멤버가 없습니다</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { api, friendlyMessage } from '../lib/api'
import type { BlogMember } from '../lib/types'
import { useToastStore } from '../stores/toast'

const route = useRoute()
const toast = useToastStore()
const blogId = computed(() => String(route.params.blogId || ''))

const items = ref<BlogMember[]>([])
const pending = ref(true)
const busy = ref(false)
const showInvite = ref(false)
const email = ref('')
const role = ref('AUTHOR')

const roleLabels: Record<string, string> = {
  OWNER: '소유자',
  ADMIN: '관리자',
  EDITOR: '편집자',
  AUTHOR: '필자'
}

const inviteLabels: Record<string, string> = {
  ACCEPTED: '참여 중',
  PENDING: '초대 대기',
  DECLINED: '거절됨'
}

function roleLabel(value: string): string {
  return roleLabels[value] || value
}

function inviteLabel(value: string): string {
  return inviteLabels[value] || value
}

async function load(): Promise<void> {
  pending.value = true
  try {
    items.value = await api<BlogMember[]>(`/api/v1/blogs/${blogId.value}/members`)
  } catch (err) {
    items.value = []
    toast.push(friendlyMessage(err), 'error')
  } finally {
    pending.value = false
  }
}

async function invite(): Promise<void> {
  busy.value = true
  try {
    const member = await api<BlogMember>(`/api/v1/blogs/${blogId.value}/members`, {
      method: 'POST',
      body: { email: email.value.trim(), role: role.value }
    })
    items.value = [...items.value, member]
    showInvite.value = false
    email.value = ''
    toast.push('초대했습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  } finally {
    busy.value = false
  }
}

async function remove(userId: string): Promise<void> {
  if (!confirm('이 멤버를 제거할까요?')) return
  try {
    await api(`/api/v1/blogs/${blogId.value}/members/${userId}`, { method: 'DELETE' })
    items.value = items.value.filter((item) => item.user_id !== userId)
    toast.push('제거했습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  }
}

onMounted(load)
</script>

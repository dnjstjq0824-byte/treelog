<template>
  <div>
    <a href="#main-content" class="tb-skip-link">본문으로 건너뛰기</a>
    <RouterView />
    <div class="tb-toast-region" role="status" aria-live="polite">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        :class="['tb-toast', toast.kind === 'error' ? 'tb-toast-error' : toast.kind === 'success' ? 'tb-toast-success' : '']"
      >
        <span>{{ toast.message }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useToastStore } from './stores/toast'

const { toasts } = storeToRefs(useToastStore())
</script>

<template>
  <div class="tb-empty">
    <p class="tb-empty-title">{{ title }}</p>
    <p v-if="description" class="tb-empty-desc">{{ description }}</p>
    <div v-if="$slots.action" style="margin-top: 8px">
      <slot name="action" />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>

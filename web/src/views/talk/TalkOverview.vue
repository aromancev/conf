<template>
  <div class="layout">
    <div class="description" :class="{ empty: !talk.description }">{{ talk.description || "No description" }}</div>
  </div>
</template>

<script setup lang="ts">
import { Talk } from "@/api/models/talk"

defineProps<{
  talk: Talk
}>()
</script>

<style scoped lang="sass">
@use '@/css/theme'

.layout
  width: 100%
  max-width: theme.$content-width
  display: flex
  flex-direction: row
  justify-content: center
  padding: 30px

.description
  flex-grow: 1
  margin: 10px
  padding: 20px
  white-space: pre-wrap
  outline: 1px solid var(--color-outline)
  border-radius: 4px
  &.empty
    color: var(--color-font-disabled)
</style>

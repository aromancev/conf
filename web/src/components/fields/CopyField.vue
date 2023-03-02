<template>
  <div class="field">
    {{ value }}
    <span class="material-icons copy" @click="copy">content_copy</span>
  </div>
</template>

<script setup lang="ts">
import { notificationStore } from "@/api/models/notifications"

const props = defineProps<{
  value: string
}>()

function copy() {
  navigator.clipboard.writeText(props.value)
  notificationStore.info("copied to clipboard")
}
</script>

<style scoped lang="sass">
@use '@/css/theme'

.field
  position: relative
  background-color: var(--color-fade-background)
  padding: 10px
  padding-right: 40px
  border-radius: 4px
  margin: 10px
  display: flex
  flex-direction: row
  justify-content: center
  align-items: center
  font-size: 12px

.copy
  @include theme.clickable

  position: absolute
  margin: 10px
  right: 0
  font-size: 20px
</style>

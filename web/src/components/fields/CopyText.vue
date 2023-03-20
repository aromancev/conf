<template>
  <div class="link copy-text" @click="copy">{{ value }} <span class="copy-icon material-icons">content_copy</span></div>
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

.copy-text
  cursor: pointer
  position: relative

.copy-icon
  font-size: 12px
</style>

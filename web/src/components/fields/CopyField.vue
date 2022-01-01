<template>
  <div class="field">
    {{ value }}
    <span class="material-icons copy" @click="copy">content_copy</span>
    <transition name="bubble">
      <div v-if="bubble" class="copy-bubble">copied</div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue"

const props = defineProps<{
  value: string
}>()

const bubble = ref(false)
let timer = 0

function copy() {
  navigator.clipboard.writeText(props.value)

  clearTimeout(timer)
  bubble.value = true
  timer = window.setTimeout(() => {
    bubble.value = false
  }, 1000)
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

.copy-bubble
    cursor: default
    position: absolute
    right: 20px
    bottom: 50px
    transform: translateX(50%)

.bubble-enter-active,
.bubble-leave-active
  transition: opacity .2s, bottom .2s

.bubble-enter-from,
.bubble-leave-to
  opacity: 0

.bubble-enter-from
  bottom: 10px
</style>

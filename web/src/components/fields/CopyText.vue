<template>
  <div class="copy-text" @click="copy">
    {{ value }}
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

.copy-text
  cursor: pointer
  position: relative

.copy-bubble
  font-size: 10px
  cursor: default
  position: absolute
  left: 50%
  bottom: 35px
  transform: translateX(-50%)

.bubble-enter-active,
.bubble-leave-active
  transition: opacity .2s, bottom .2s

.bubble-enter-from,
.bubble-leave-to
  opacity: 0

.bubble-enter-from
  bottom: 10px
</style>

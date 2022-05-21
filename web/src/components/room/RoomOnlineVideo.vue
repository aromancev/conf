<template>
  <div
    ref="container"
    class="container"
    :style="isFullscreen && !isInterfaceVisible ? { cursor: 'none' } : {}"
    @dblclick="toggleFullscreen"
    @mousemove="onMouseMove"
  >
    <video class="video" :srcObject="src" autoplay muted></video>
    <div v-if="isInterfaceVisible" class="interface">
      <div class="bottom-panel">
        <div class="bottom-right-panel">
          <div v-if="!isFullscreen" class="material-icons panel-btn" @click="toggleFullscreen">fullscreen</div>
          <div v-if="isFullscreen" class="material-icons panel-btn" @click="toggleFullscreen">fullscreen_exit</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, defineProps, onMounted, onUnmounted } from "vue"
import { debounce } from "@/platform/sync"

defineProps<{
  src?: MediaStream
}>()

const container = ref<HTMLElement>()
const isFullscreen = ref<boolean>(false)
const isInterfaceVisible = ref<boolean>(false)

onMounted(() => {
  document.addEventListener("fullscreenchange", onFullscreen)
  onFullscreen()
})

onUnmounted(() => {
  document.removeEventListener("fullscreenchange", onFullscreen)
})

const hideInterfaceDebounced = debounce(() => {
  isInterfaceVisible.value = false
}, 3000)

function onMouseMove() {
  isInterfaceVisible.value = true
  hideInterfaceDebounced()
}

function onFullscreen() {
  isFullscreen.value = document.fullscreenElement ? true : false
}

function toggleFullscreen() {
  if (!container.value) {
    return
  }
  if (document.fullscreenElement) {
    document.exitFullscreen()
  } else {
    container.value.requestFullscreen()
  }
}
</script>

<style lang="sass" scoped>
@use '@/css/theme'

.container
  width: 100%
  height: 100%

.interface
  position: absolute
  top: 0
  left: 0
  height: 100%
  width: 100%
  background-image: linear-gradient(180deg, transparent 85%, black 95%)

.bottom-panel
  position: absolute
  bottom: 0
  left: 0
  width: 100%
  display: flex
  flex-direction: row
  align-items: center
  justify-content: flex-start

.bottom-right-panel
  margin-left: auto

.panel-btn
  @include theme.clickable

  padding: 10px

.video
  object-fit: contain
  width: 100%
  height: 100%
</style>

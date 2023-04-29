<template>
  <div
    ref="container"
    class="container"
    :style="isFullscreen && !isInterfaceVisible ? { cursor: 'none' } : {}"
    @dblclick="toggleFullscreen"
    @mousemove="onMouseMove"
  >
    <video ref="video" class="video"></video>
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
import { ref, onMounted, onUnmounted, watch, inject } from "vue"
import { debounce } from "@/platform/sync"
import { Track } from "./live"

type Attacher = {
  attach(trackId: string, el: HTMLMediaElement): void
}

const props = defineProps<{
  track?: Track
  disableControls?: boolean
}>()

const attacher = inject<Attacher>("attacher")
const container = ref<HTMLElement>()
const video = ref<HTMLMediaElement>()
const isFullscreen = ref<boolean>(false)
const isInterfaceVisible = ref<boolean>(false)

watch(
  () => props.track,
  () => {
    attachTrack()
  },
)

onMounted(() => {
  document.addEventListener("fullscreenchange", onFullscreen)
  onFullscreen()
  attachTrack()
})

onUnmounted(() => {
  document.removeEventListener("fullscreenchange", onFullscreen)
})

const hideInterfaceDebounced = debounce(() => {
  isInterfaceVisible.value = false
}, 3000)

function onMouseMove() {
  if (props.disableControls) {
    return
  }
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

function attachTrack() {
  if (!props.track) {
    return
  }
  if (!video.value) {
    throw new Error("Failed to attach track.")
  }
  if (!attacher) {
    throw new Error("Track attacher not provided.")
  }

  attacher.attach(props.track.id, video.value)
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

  color: white
  padding: 10px

.video
  object-fit: contain
  width: 100%
  height: 100%
</style>

<template>
  <div
    ref="container"
    class="container"
    :style="state.isFullscreen && !state.isInterfaceVisible ? { cursor: 'none' } : {}"
    @mousemove="onMouseMove"
  >
    <video ref="video" class="video" muted :style="{ display: controllerState.isActive ? undefined : 'none' }"></video>
    <div v-if="state.isInterfaceVisible" class="interface">
      <div class="free-screen" @dblclick="toggleFullscreen" @click="emit('togglePlay')"></div>
      <div class="bottom-panel">
        <div ref="timeline" class="timeline" @click="onRewind">
          <div class="buffer" :style="{ width: (props.buffer / props.duration) * 100 + '%' }"></div>
          <div class="progress" :style="{ width: (state.progress / props.duration) * 100 + '%' }"></div>
        </div>
        <div class="bottom-tools">
          <div class="bottom-left-panel">
            <div class="material-icons panel-btn" @click="emit('togglePlay')">
              {{ isPlaying ? "pause" : "play_arrow" }}
            </div>
          </div>
          <div class="bottom-right-panel">
            <div class="material-icons panel-btn" @click="toggleFullscreen">
              {{ state.isFullscreen ? "fullscreen_exit" : "fullscreen" }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, defineProps, onMounted, onUnmounted, reactive, watch } from "vue"
import { debounce } from "@/platform/sync"
import { Media } from "./aggregators/media"
import { MediaController } from "./media-controller"

const props = defineProps<{
  media?: Media
  isPlaying: boolean
  duration: number
  delta: number
  unpausedAt: number
  buffer: number
  disableControlls?: boolean
}>()

const emit = defineEmits<{
  (e: "togglePlay"): void
  (e: "rewind", p: number): void
}>()

const state = reactive({
  isFullscreen: false,
  isInterfaceVisible: !props.disableControlls,
  progress: 0,
})
const container = ref<HTMLElement>()
const timeline = ref<HTMLElement>()
const video = ref<HTMLElement>()
const controller = new MediaController()
const controllerState = controller.state()
let progressInterval: ReturnType<typeof setInterval> = 0

watch(
  [() => props.media, () => props.isPlaying, () => props.delta, () => props.unpausedAt, video],
  () => {
    controller.update({
      media: props.media,
      element: video.value,
      isPlaying: props.isPlaying,
      delta: props.delta,
      unpausedAt: props.unpausedAt,
    })
  },
  {
    immediate: true,
    deep: true,
  },
)

watch([() => props.isPlaying, () => state.isInterfaceVisible], () => {
  clearInterval(progressInterval)
  if (props.isPlaying && state.isInterfaceVisible) {
    iterate()
    progressInterval = setInterval(() => iterate(), 100)
  }
})

onMounted(() => {
  document.addEventListener("fullscreenchange", onFullscreen)
  onFullscreen()
})

onUnmounted(() => {
  document.removeEventListener("fullscreenchange", onFullscreen)
  controller.close()
})

const hideInterfaceDebounced = debounce(() => {
  state.isInterfaceVisible = false
}, 3000)

function iterate(): void {
  state.progress = Date.now() - props.unpausedAt + props.delta
}

function onMouseMove() {
  if (props.disableControlls) {
    return
  }
  state.isInterfaceVisible = true
  hideInterfaceDebounced()
}

function onFullscreen() {
  state.isFullscreen = document.fullscreenElement ? true : false
}

function toggleFullscreen() {
  if (!container.value || props.disableControlls) {
    return
  }
  if (document.fullscreenElement) {
    document.exitFullscreen()
  } else {
    container.value.requestFullscreen()
  }
}

function onRewind(event: MouseEvent) {
  if (!timeline.value) {
    throw new Error("Timeline element not found.")
  }
  const rect = timeline.value.getBoundingClientRect()
  const progresss = (event.clientX - rect.left) / rect.width
  const progressMs = props.duration * progresss
  state.progress = progressMs
  emit("rewind", progressMs)
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
  background-image: linear-gradient(0, black 0, transparent 50px)

.free-screen
  position: absolute
  width: 100%
  height: 100%

.bottom-panel
  position: absolute
  bottom: 0
  left: 0
  width: 100%

.timeline
  position: relative
  cursor: pointer
  width: 100%
  height: 10px

.buffer
  bottom: 0
  left: 0
  position: absolute
  height: 2px
  width: 100%
  background-color: grey

.progress
  bottom: 0
  left: 0
  position: absolute
  height: 2px
  width: 100%
  background-color: var(--color-highlight-background)
  transition: width 50ms linear

.bottom-tools
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

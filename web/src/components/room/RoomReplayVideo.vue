<template>
  <div
    ref="container"
    class="container"
    :style="state.isFullscreen && !state.isInterfaceVisible ? { cursor: 'none' } : {}"
    @mousemove="onMouseMove"
  >
    <video
      ref="video"
      class="video"
      muted
      :style="{ display: controllerState.isActive ? undefined : 'none', objectFit: fit || 'contain' }"
    ></video>
    <div v-if="state.isInterfaceVisible" class="interface">
      <div class="free-screen" @dblclick="toggleFullscreen" @click="emit('togglePlay')"></div>
      <div class="bottom-panel">
        <div
          ref="timeline"
          class="timeline"
          @click="onRewind"
          @mousemove="updateHighlight"
          @mouseenter="updateHighlight"
          @mouseleave="state.timePointerVisible = false"
        >
          <div class="outline"></div>
          <div class="buffer" :style="{ width: `${((props.buffer || 0) / props.duration) * 100}%` }"></div>
          <div class="highlight" :style="{ width: `${state.highlight * 100}%` }"></div>
          <div class="progress" :style="{ width: `${(state.progress / props.duration) * 100}%` }"></div>
          <div v-if="state.timePointerVisible" class="timeline-pointer" :style="{ left: `${state.highlight * 100}%` }">
            {{ formatTime(props.duration * state.highlight) }}
          </div>
        </div>
        <div class="bottom-tools">
          <div class="bottom-left-panel">
            <div class="material-icons panel-btn" @click="emit('togglePlay')">
              {{ isPlaying ? "pause" : state.progress < duration ? "play_arrow" : "replay" }}
            </div>
            <div class="time">{{ formatTime(state.progress) }} / {{ formatTime(props.duration) }}</div>
          </div>
          <div class="bottom-right-panel">
            <div class="material-icons panel-btn" @click="toggleFullscreen">
              {{ state.isFullscreen ? "fullscreen_exit" : "fullscreen" }}
            </div>
          </div>
        </div>
      </div>
    </div>
    <div v-if="isBuffering && !hideLoader" class="loader-box">
      <PageLoader></PageLoader>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, reactive, watch } from "vue"
import { debounce } from "@/platform/sync"
import { TrackRecord } from "./aggregators/record"
import { MediaController } from "./media-controller"
import PageLoader from "@/components/PageLoader.vue"
import { Progress } from "./replay"

const props = defineProps<{
  record?: TrackRecord
  isPlaying: boolean
  isBuffering: boolean
  duration: number
  progress: Progress
  buffer?: number
  disableControlls?: boolean
  fit?: "contain" | "cover"
  hideLoader?: boolean
}>()

const emit = defineEmits<{
  (e: "togglePlay"): void
  (e: "rewind", ms: number): void
  (e: "buffer", bufferMs: number, durationMs: number): void
}>()

const state = reactive({
  isFullscreen: false,
  isInterfaceVisible: !props.disableControlls,
  progress: 0,
  highlight: 0,
  timePointerVisible: false,
})
const container = ref<HTMLElement>()
const timeline = ref<HTMLElement>()
const video = ref<HTMLMediaElement>()
const controller = new MediaController({
  media: () => props.record,
  element: video,
  isPlaying: () => props.isPlaying,
  isBuffering: () => props.isBuffering,
  progress: () => props.progress,
})
controller.onBuffer = (bufferMs: number, durationMs: number): void => {
  emit("buffer", bufferMs, durationMs)
}
const controllerState = controller.state()
let progressInterval = 0

watch([() => props.isPlaying, () => props.isBuffering, () => state.isInterfaceVisible], () => {
  clearInterval(progressInterval)
  if (props.isPlaying && !props.isBuffering && state.isInterfaceVisible) {
    updateProgress()
    progressInterval = window.setInterval(() => updateProgress(), 100)
  }
})

watch(
  () => props.progress.value,
  () => {
    updateProgress()
  },
)

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

function updateProgress(): void {
  state.progress = progressForNow(props.progress)
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
  emit("rewind", progressMs)
}

function updateHighlight(event: MouseEvent) {
  if (!timeline.value) {
    throw new Error("Timeline element not found.")
  }
  state.timePointerVisible = true
  const rect = timeline.value.getBoundingClientRect()
  state.highlight = (event.clientX - rect.left) / rect.width
}

function progressForNow(progress: Progress): number {
  if (!progress.increasingSince) {
    return progress.value
  }
  return Date.now() - progress.increasingSince + progress.value
}

function formatTime(ms: number): string {
  const seconds = Math.floor(ms / 1000)
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = seconds % 60
  return `${minutes}:${remainingSeconds.toString().padStart(2, "0")}`
}
</script>

<style lang="sass" scoped>
@use '@/css/theme'

.container
  width: 100%
  height: 100%
  display: flex
  align-items: center
  justify-content: center

.interface
  position: absolute
  top: 0
  left: 0
  height: 100%
  width: 100%

.free-screen
  position: absolute
  width: 100%
  height: 100%

.bottom-panel
  position: absolute
  bottom: 0
  left: 0
  width: 100%
  padding: 0 15px

.timeline
  position: relative
  cursor: pointer
  height: 10px
  &:hover > .progress, &:hover > .highlight
    height: 5px

.progress
  bottom: 0
  left: 0
  position: absolute
  height: 2px
  background-color: var(--color-highlight-background)
  transition: width 50ms linear, height 100ms ease-in

.highlight
  bottom: 0
  left: 0
  position: absolute
  height: 0px
  background-color: lightgrey
  transition: height 100ms ease-in

.buffer
  bottom: 0
  left: 0
  max-width: 100%
  position: absolute
  height: 2px
  background-color: grey
  transition: width 100ms linear

.outline
  bottom: 0
  left: 0
  position: absolute
  width: 100%
  height: 2px
  background-color: white
  opacity: 0.15

.timeline-pointer
  top: -25px
  left: 0
  transform: translateX(-50%)
  position: absolute
  padding: 5px 7px
  background-color: black
  color: white
  border-radius: 5px
  font-size: 12px

.timeline-pointer::before
  content: ""
  position: absolute
  bottom: -5px
  left: 50%
  transform: translateX(-50%)
  border-width: 5px 5px 0
  border-style: solid
  border-color: black transparent transparent

.bottom-tools
  width: 100%
  display: flex
  flex-direction: row
  align-items: center
  justify-content: flex-start

.bottom-right-panel
  margin-left: auto

.bottom-left-panel
  display: flex
  flex-direction: row
  align-items: center

.time
  font-size: 13px

.panel-btn
  @include theme.clickable

  color: white
  padding: 10px

.video
  width: 100%
  height: 100%

.loader-box
  position: absolute
</style>

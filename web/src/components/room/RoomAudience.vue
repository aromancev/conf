<template>
  <div class="audience" @mousemove="select" @mouseleave="deselect">
    <div class="selected">{{ selected?.profile.name || "" }}</div>
    <div class="divider"></div>
    <div class="canvas">
      <canvas ref="audience" :style="{ display: isLoading ? 'none' : 'block' }"></canvas>
      <canvas ref="statuses"></canvas>
      <canvas ref="selection"></canvas>
      <router-link
        v-if="selected?.profile.handle"
        class="profile-link"
        :to="route.profile(selected.profile.handle, 'overview')"
        target="_blank"
      ></router-link>
      <PageLoader v-if="isLoading"></PageLoader>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from "vue"
import { Throttler } from "@/platform/sync"
import { Peer, Status } from "./aggregators/peers"
import { route } from "@/router"
import { Renderer } from "./audience"
import PageLoader from "@/components/PageLoader.vue"

const ANIMATE_IF_LESS = 1000
const AUTO_RESIZE_INTERVAL = 1000

const props = defineProps<{
  userId: string
  isLoading: boolean
  isPlaying: boolean
  selfReactions?: boolean
  peers: Map<string, Peer>
  statuses: Map<string, Status>
}>()

const audience = ref<HTMLCanvasElement>()
const statuses = ref<HTMLCanvasElement>()
const selection = ref<HTMLCanvasElement>()
const selected = ref(null as Peer | null)
const updatePeers = new Throttler({
  delayMs: 300,
})
const updateStatuses = new Throttler({
  delayMs: 200,
})

let renderer = null as Renderer | null
let resizeIntervalId = 0
let animateIntervalId = 0
let nextAnimationId = 0

watch(
  () => props.isPlaying,
  (isPlaying: boolean) => {
    if (isPlaying) {
      startAnimation()
    } else {
      stopAnimation()
    }
  },
)

watch(
  () => props.peers,
  () => {
    updatePeers.do()
  },
  { deep: true, immediate: false },
)

watch(
  () => props.statuses,
  () => {
    updateStatuses.do()
    if (props.isPlaying) {
      startAnimation()
    }
  },
  { deep: true, immediate: false },
)

defineExpose({
  resize,
})

onMounted(() => {
  if (!audience.value || !selection.value || !statuses.value) {
    console.error("not created")
    return
  }

  const audCtx = audience.value.getContext("2d")
  const selectionCtx = selection.value.getContext("2d")
  const statesCtx = statuses.value.getContext("2d")
  if (!audCtx || !selectionCtx || !statesCtx) {
    throw new Error("Failed to get canvas context.")
  }
  renderer = new Renderer({
    userId: props.userId,
    peers: props.peers,
    statuses: props.statuses,
    selfReactions: props.selfReactions || false,
    context: {
      audience: audCtx,
      selection: selectionCtx,
      statuses: statesCtx,
    },
    width: audience.value.width,
    height: audience.value.height,
  })
  if (props.isPlaying) {
    renderer.play()
  }
  updatePeers.func = () => {
    renderer?.updatePeers()
  }
  updateStatuses.func = () => {
    renderer?.updateStatuses()
  }
  updatePeers.do()
  updateStatuses.do()

  clearInterval(resizeIntervalId)
  resizeIntervalId = window.setInterval(resize, AUTO_RESIZE_INTERVAL)
  window.addEventListener("resize", resize)
  resize()
})

onUnmounted(() => {
  clearInterval(resizeIntervalId)
  clearInterval(animateIntervalId)
  window.removeEventListener("resize", resize)
  window.cancelAnimationFrame(nextAnimationId)
  renderer?.close()
  renderer = null
})

function startAnimation() {
  // If too many or no statuses, don't animate for performance.
  if (props.statuses.size > ANIMATE_IF_LESS || props.statuses.size === 0) {
    stopAnimation()
    return
  }

  // Already animating.
  if (nextAnimationId) {
    return
  }

  renderer?.play()
  nextAnimationId = window.requestAnimationFrame(animate)
}

function stopAnimation() {
  // Already stopped.
  if (!nextAnimationId) {
    return
  }
  renderer?.pause()
  window.cancelAnimationFrame(nextAnimationId)
  nextAnimationId = 0
}

function animate() {
  renderer?.animate()
  nextAnimationId = window.requestAnimationFrame(animate)
}

function resize() {
  if (document.fullscreenElement) {
    return
  }
  if (!audience.value || !selection.value || !statuses.value) {
    return
  }

  const dpr = window.devicePixelRatio || 1
  const width = audience.value.offsetWidth * dpr
  const height = audience.value.offsetHeight * dpr
  if (audience.value.width === width && audience.value.height === height) {
    return
  }

  audience.value.width = width
  audience.value.height = height
  selection.value.width = width
  selection.value.height = height
  statuses.value.width = width
  statuses.value.height = height
  renderer?.resize(width, height)
}

function select(ev: MouseEvent) {
  if (!renderer) {
    return
  }
  const dpr = window.devicePixelRatio || 1
  const rect = (ev.target as HTMLElement).getBoundingClientRect()
  const userId = renderer.hover((ev.clientX - rect.left) * dpr, (ev.clientY - rect.top) * dpr)
  if (userId === (selected.value?.userId || null)) {
    return
  }
  if (userId) {
    selected.value = props.peers.get(userId) || null
    renderer.select(userId)
  } else {
    selected.value = null
    renderer.select("")
  }
}

function deselect() {
  if (!renderer) {
    return
  }
  renderer.select("")
}
</script>

<style scoped lang="sass">
.audience
  display: flex
  flex-direction: column
  background-color: transparent
  overflow: hidden

.selected
  margin: 10px
  height: 1em
  text-align: center

.divider
  height: 1px
  background: linear-gradient(to right, transparent 0, var(--color-highlight-background) 50%, transparent)

.canvas
  position: relative
  flex: 1
  display: flex
  justify-content: center

canvas
  position: absolute
  top: 0
  left: 0
  width: 100%
  height: 100%
  cursor: default

.loader
  height: 100%
  z-index: 100

.profile-link
  position: absolute
  top: 0
  left: 0
  width: 100%
  height: 100%
</style>

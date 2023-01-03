<template>
  <div class="panel">
    <HoldableButton
      class="btn-switch clap-btn"
      :class="{ active: state.isClapping }"
      @down="clapStart"
      @up="clapStop"
      @single="clapSingle"
    >
      <div class="clap-animation" :class="{ active: state.isClapping }"></div>
      <div v-if="state.isClapWaveDisplayed" class="clap-wave"></div>
    </HoldableButton>
  </div>
</template>

<script setup lang="ts">
import { onUnmounted, reactive } from "vue"
import { Reaction } from "@/api/room/schema"
import { sleep, repeat } from "@/platform/sync"
import { Sound } from "./actor"
import HoldableButton from "@/components/HoldableButton.vue"
import reactionsSoundURL from "/static/room/reactions.webm"

const emit = defineEmits<{
  (e: "reaction", reaction: Reaction): void
}>()

interface State {
  isClapping: boolean
  isClapWaveDisplayed: boolean
}

const CLAP_SOUND_DELAY_MS = 250
const CLAP_INTERVAL_MS = 300
const CLAP_SINGLE_DURATION_MS = 1000
const state = reactive<State>({
  isClapping: false,
  isClapWaveDisplayed: false,
})
const sounds = new Sound(reactionsSoundURL, {
  clap: [
    [2650, 200],
    [2850, 200],
    [3100, 200],
    [3325, 200],
    [3550, 200],
  ],
})
let clapCtrl = new AbortController()

async function clapStart() {
  clapCtrl.abort()
  clapCtrl = new AbortController()

  state.isClapping = true
  clap()
  repeat(clapCtrl.signal, CLAP_INTERVAL_MS, clap)
  emit("reaction", {
    clap: {
      isStarting: true,
    },
  })
}

function clapStop() {
  clapCtrl.abort()
  state.isClapping = false
  emit("reaction", {
    clap: {
      isStarting: false,
    },
  })
}

async function clapSingle() {
  if (state.isClapping) {
    return
  }

  clapCtrl.abort()
  clapCtrl = new AbortController()

  emit("reaction", {
    clap: {
      isStarting: true,
    },
  })

  clap()
  try {
    state.isClapping = true
    await sleep(clapCtrl.signal, CLAP_INTERVAL_MS)
    state.isClapping = false
    await sleep(clapCtrl.signal, CLAP_SINGLE_DURATION_MS - CLAP_INTERVAL_MS)
  } catch (e) {
    state.isClapping = false
  }

  emit("reaction", {
    clap: {
      isStarting: false,
    },
  })

  clapCtrl.abort()
}

async function clap(): Promise<void> {
  try {
    await sleep(clapCtrl.signal, CLAP_SOUND_DELAY_MS)
    sounds.play(clapCtrl.signal, "clap", 0)
  } catch (e) {
    return
  }

  state.isClapWaveDisplayed = true
  setTimeout(() => (state.isClapWaveDisplayed = false), 200)
}

onUnmounted(() => sounds.close())
</script>

<style lang="sass" scoped>
@use '@/css/theme'

.panel
  display: flex
  justify-content: center
  align-items: center

@keyframes sprite
  100%
    background-position: -3072px

.clap-btn
  position: relative
  margin: 10px
  width: 66px
  height: 66px
  display: flex
  align-items: center
  justify-content: center
  &:active:not([disabled]):not(:disabled)
    @include theme.shadow-xs
    background: var(--color-concave)
  &.active
    @include theme.shadow-xs
    background: var(--color-concave)

.clap-animation
  position: absolute
  width: 256px
  height: 256px
  background: url(/static/room/reactions.webp) 0 0 no-repeat
  animation: none
  transform: scale(0.15) translate(-5%, -10%)
  &.active
    animation: sprite steps(12) 300ms infinite both

@keyframes wave
  0%
    transform: scale(0.3)
    opacity: 0.5

  100%
    transform: scale(1.2)
    opacity: 0

.clap-wave
  animation: 250ms ease-in 0s 1 wave
  position: absolute
  width: 100%
  height: 100%
  border-radius: 50%
  border: 1px solid white
</style>

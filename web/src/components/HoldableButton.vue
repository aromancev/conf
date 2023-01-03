<template>
  <div @mousedown="onDown" @mouseup="onUp" @mouseleave="onLeave">
    <slot></slot>
  </div>
</template>
<script setup lang="ts">
import { sleep } from "@/platform/sync"

const emit = defineEmits<{
  (e: "single"): void
  (e: "up"): void
  (e: "down"): void
}>()

const HOLD_DELAY_MS = 200
let ctrl = new AbortController()
let downAt = 0

async function onDown() {
  ctrl.abort()
  ctrl = new AbortController()

  downAt = Date.now()
  try {
    await sleep(ctrl.signal, HOLD_DELAY_MS)
  } catch (e) {
    return
  }
  emit("down")
}

function onUp() {
  ctrl.abort()
  if (downAt === 0) {
    return
  }
  if (Date.now() - downAt > HOLD_DELAY_MS) {
    emit("up")
  } else {
    emit("single")
  }
  downAt = 0
}

function onLeave() {
  ctrl.abort()
  if (downAt !== 0 && Date.now() - downAt > HOLD_DELAY_MS) {
    emit("up")
  }
  downAt = 0
}
</script>

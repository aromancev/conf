<template>
  <audio ref="audio"></audio>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, inject } from "vue"
import { Track } from "./live"

type Attacher = {
  attach(trackId: string, el: HTMLMediaElement): void
}

const props = defineProps<{
  track?: Track
}>()

const attacher = inject<Attacher>("attacher")
const audio = ref<HTMLVideoElement>()

watch(
  () => props.track,
  () => {
    attachTrack()
  },
)

onMounted(() => {
  attachTrack()
})

function attachTrack() {
  if (!props.track) {
    return
  }
  if (!audio.value) {
    throw new Error("Failed to attach track.")
  }
  if (!attacher) {
    throw new Error("Track attacher not provided.")
  }

  attacher.attach(props.track.id, audio.value)
}
</script>

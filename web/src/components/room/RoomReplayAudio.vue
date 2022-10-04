<template>
  <audio ref="audio"></audio>
</template>

<script setup lang="ts">
import { ref, onUnmounted } from "vue"
import { Media } from "./aggregators/media"
import { MediaController } from "./media-controller"
import { Progress } from "./replay"

const emit = defineEmits<{
  (e: "buffer", ms: number): void
}>()

const props = defineProps<{
  media?: Media
  isPlaying: boolean
  isBuffering: boolean
  duration: number
  progress: Progress
}>()

const audio = ref<HTMLElement>()
const controller = new MediaController({
  media: () => props.media,
  element: audio,
  isPlaying: () => props.isPlaying,
  isBuffering: () => props.isBuffering,
  progress: () => props.progress,
})
controller.onBuffer = (ms) => {
  emit("buffer", ms)
}

onUnmounted(() => {
  controller.close()
})
</script>

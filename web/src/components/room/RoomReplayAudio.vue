<template>
  <audio ref="audio"></audio>
</template>

<script setup lang="ts">
import { ref, onUnmounted } from "vue"
import { Media } from "./aggregators/media"
import { MediaController } from "./media-controller"

const props = defineProps<{
  media?: Media
  isPlaying: boolean
  duration: number
  delta: number
  unpausedAt: number
}>()

const audio = ref<HTMLElement>()
const controller = new MediaController({
  media: () => props.media,
  element: audio,
  isPlaying: () => props.isPlaying,
  unpausedAt: () => props.unpausedAt,
  delta: () => props.delta,
})

onUnmounted(() => {
  controller.close()
})
</script>

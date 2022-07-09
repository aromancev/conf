<template>
  <audio ref="audio"></audio>
</template>

<script setup lang="ts">
import { ref, defineProps, watch } from "vue"
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
const controller = new MediaController()

watch(
  [() => props.media, () => props.isPlaying, () => props.delta, () => props.unpausedAt, audio],
  () => {
    controller.update({
      media: props.media,
      element: audio.value,
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
</script>

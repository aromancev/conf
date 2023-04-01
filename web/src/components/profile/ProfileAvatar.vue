<template>
  <img :src="source" />
</template>

<script setup lang="ts">
import { genAvatar } from "@/platform/gen"
import { ref, watch } from "vue"

const props = defineProps<{
  size: number
  userId: string
  src: string
}>()

const source = ref<string>("")

watch(
  [() => props.size, () => props.userId, () => props.src],
  async () => {
    if (props.src.length) {
      source.value = props.src
    } else {
      const generated = await genAvatar(props.userId, props.size)
      // Check again to make sure it wasn't update while we were generating avatar.
      if (!props.src.length) {
        source.value = generated
      }
    }
  },
  { immediate: true },
)
</script>

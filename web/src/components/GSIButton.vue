<template>
  <div ref="button" class="gsi"></div>
</template>

<script setup lang="ts">
import { ref, watch, onUpdated, onMounted } from "vue"
import { gsiRenderButton } from "./gsi"
import { styleStore } from "@/api/models/style"

const button = ref<HTMLElement>()

watch(() => styleStore.state.theme, render)
onMounted(render)
onUpdated(render)

function render() {
  if (!button.value) {
    throw new Error("Failed to render GSI button.")
  }
  gsiRenderButton(button.value, styleStore.state.theme === "light" ? "filled_black" : "outline")
}
</script>

<style scoped lang="sass">
.gsi
  display: flex
  align-content: center
  justify-content: left
  height: 50px
</style>

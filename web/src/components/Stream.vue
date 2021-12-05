<template>
  <video
    ref="video"
    v-if="stream"
    v-bind:class="{ mirrored: mirrored }"
    v-bind:muted="muted"
    v-bind:srcObject="stream"
    autoplay
  />
  <div v-else>loading stream</div>
</template>

<script lang="ts">
import { defineComponent } from "vue"

export default defineComponent({
  name: "Stream",
  props: {
    stream: MediaStream,
    mirrored: Boolean,
    muted: Boolean,
  },
})
</script>

<style scoped lang="scss">
.mirrored {
  transform: rotateY(180deg);
}
</style>

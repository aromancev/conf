<template>
  <div class="d-flex justify-content-center">
    <div class="content">
      <input class="input px-3 py-2" v-model="mediaId">
      <button class="btn px-3 py-1" @click="go">go</button>
      <video ref="videoPlayer" controls></video>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue"
import { MediaPlayer } from "dashjs"

export default defineComponent({
  name: "Stream",
  components: {},
  data() {
    return {
      mediaId: "",
    }
  },
  methods: {
    go() {
      const url = `/api/media/v1/${this.mediaId}/manifest.mpd`
      const player = MediaPlayer().create()
      player.initialize(this.$refs.videoPlayer as HTMLElement, url, false)
    },
  },
})
</script>

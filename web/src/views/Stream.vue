<template>
  <div class="container">
    <div class="row">
      <h1>Stream</h1>
      <video ref="videoPlayer" controls></video>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue"
import { MediaPlayer } from "dashjs"
// import "shaka-player"

export default defineComponent({
  name: "Stream",
  components: {},
  data() {
    return {
    }
  },
  mounted() {
    const url = "/api/media/v1/test_mkv.mpd"
    // const player = new shaka.Player(this.$refs.videoPlayer as HTMLMediaElement)
    // player.load(url)
    const player = MediaPlayer().create()
    player.initialize(this.$refs.videoPlayer as HTMLElement, url, false)
  },
})
</script>

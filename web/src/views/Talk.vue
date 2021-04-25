<template>
  <div class="container">
    <div class="row">
      <h1></h1>
      <h3>Local Video</h3>
      <Stream
        v-bind:stream="localStream"
        v-bind:mirrored="true"
        v-bind:muted="true"
      />

      <h3>Remote Video</h3>
      <Stream
        v-for="stream in remoteStreams"
        v-bind:key="stream.id"
        v-bind:stream="stream"
        width="150"
      />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue"
import Stream from "@/components/Stream.vue"
import { Client, LocalStream, RemoteStream } from "ion-sdk-js"
import { Signal } from "@/rtc"

export default defineComponent({
  name: "Talk",
  components: {
    Stream,
  },
  data() {
    return {
      localStream: null as MediaStream | null,
      remoteStreams: {} as { [key: string]: MediaStream },
    }
  },

  async created() {
    let protocol = "wss"
    if (process.env.NODE_ENV == "development") {
      protocol = "ws"
    }
    const signal = new Signal(
      `${protocol}://${window.location.hostname}/api/rtc/v1/ws`,
    )
    const client = new Client(signal)
    signal.onopen = async () => {
      client.ontrack = (track: MediaStreamTrack, stream: RemoteStream) => {
        if (track.kind !== "video") {
          return
        }

        stream.preferLayer("low")
        this.remoteStreams[stream.id] = stream
        stream.onremovetrack = () => {
          delete this.remoteStreams[stream.id]
        }
      }

      const local = await LocalStream.getUserMedia({
        codec: "vp8",
        resolution: "vga",
        video: true,
        audio: false,
      })

      await client.join("test session")
      client.publish(local)
      this.localStream = local
    }
  },
})
</script>

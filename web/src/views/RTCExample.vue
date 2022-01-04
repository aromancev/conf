<template>
  <div></div>
</template>

<script lang="ts">
import { defineComponent } from "vue"
import { Client, LocalStream, RemoteStream } from "ion-sdk-js"
import { client, userStore, confaClient, Talk, talkClient } from "@/api"

enum Dialog {
  None = "",
  Error = "error",
}

export default defineComponent({
  name: "RTCExample",
  data() {
    return {
      Dialog,
      talk: null as Talk | null,
      localStream: null as MediaStream | null,
      remoteStreams: {} as { [key: string]: MediaStream },
      modal: Dialog.None,
    }
  },

  computed: {
    userId() {
      return userStore.getState().id
    },
  },

  async created() {
    this.talk = await this.fetchTalk()
    if (this.talk === null) {
      alert("talk not found")
      return
    }

    const rtc = await client.rtc(this.talk.roomId)
    const sfu = new Client(rtc)
    rtc.onopen = async () => {
      if (!this.talk) {
        throw new Error("Talk not set")
      }
      await sfu.join(this.talk.roomId, this.userId)
      const local = await LocalStream.getUserMedia({
        codec: "vp8",
        resolution: "vga",
        simulcast: true,
        video: true,
        audio: false,
      })
      this.localStream = local
      sfu.publish(local)
    }
    sfu.ontrack = (track: MediaStreamTrack, stream: RemoteStream) => {
      if (track.kind !== "video") {
        return
      }
      this.remoteStreams[stream.id] = stream
      stream.onremovetrack = () => {
        delete this.remoteStreams[stream.id]
      }
    }
  },

  methods: {
    async fetchTalk(): Promise<Talk | null> {
      const confaHanle = this.$route.params.confa as string
      const talkHandle = this.$route.params.talk as string

      if (talkHandle !== "new") {
        try {
          return await talkClient.fetchOne({
            handle: talkHandle,
          })
        } catch (e) {
          this.modal = Dialog.Error
        }
      }

      try {
        const conf = await confaClient.fetchOne({
          handle: confaHanle,
        })
        if (conf === null) {
          return null
        }
        const tlk = await talkClient.create({handle: confaHanle}, {})
        this.$router.replace("/" + confaHanle + "/" + tlk.handle)
        return tlk
      } catch (e) {
        this.modal = Dialog.Error
      }

      return null
    },
  },
})
</script>

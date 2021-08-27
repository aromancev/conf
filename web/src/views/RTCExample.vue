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
      <div
        v-if="userId === talk?.ownerId"
        class="btn px-3 py-2"
        v-on:click="start"
      >
        Start
      </div>

      <InternalError
        v-if="modal === Dialog.Error"
        v-on:click="modal = Dialog.None"
      />
    </div>
  </div>
</template>

<script lang="ts">
import InternalError from "@/components/modals/InternalError.vue"
import { defineComponent } from "vue"
import Stream from "@/components/Stream.vue"
import { Client, LocalStream, RemoteStream } from "ion-sdk-js"
import { client, userStore, confa, Talk, talk } from "@/api"

enum Dialog {
  None = "",
  Error = "error",
}

export default defineComponent({
  name: "RTCExample",
  components: {
    Stream,
    InternalError,
  },

  computed: {
    userId() {
      return userStore.getState().id
    },
  },
  data() {
    return {
      Dialog,
      talk: null as Talk | null,
      localStream: null as MediaStream | null,
      remoteStreams: {} as { [key: string]: MediaStream },
      modal: Dialog.None,
    }
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
    async start() {
      if (!this.talk) {
        return
      }
      try {
        await talk.start(this.talk.id)
      } catch (e) {
        this.modal = Dialog.Error
      }
    },

    async fetchTalk(): Promise<Talk | null> {
      const confaHanle = this.$route.params.confa as string
      const talkHandle = this.$route.params.talk as string

      if (talkHandle !== "new") {
        try {
          return await talk.fetchOne({
            handle: talkHandle,
          })
        } catch (e) {
          this.modal = Dialog.Error
        }
      }

      try {
        const conf = await confa.fetchOne({
          handle: confaHanle,
        })
        if (conf === null) {
          return null
        }
        const tlk = await talk.create(conf.id)
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

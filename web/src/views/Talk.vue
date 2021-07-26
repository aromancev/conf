<template>
  <div class="container">
    <div class="row">
      <!-- <h1></h1>
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
      /> -->
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
// import { Client, LocalStream, RemoteStream } from "ion-sdk-js"
// import { Signal } from "@/api/rtc"
import { userStore, confa, Talk, talk } from "@/api"

enum Dialog {
  None = "",
  Error = "error",
}

export default defineComponent({
  name: "Talk",
  components: {
    // Stream,
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

  async beforeCreate() {
    const confaHanle = this.$route.params.confa as string
    const talkHandle = this.$route.params.talk as string

    if (talkHandle !== "new") {
      try {
        this.talk = await talk.fetchOne({
          handle: talkHandle,
        })
      } catch (e) {
        this.modal = Dialog.Error
      }
      if (this.talk === null) {
        alert("talk not found")
      }
      return
    }

    try {
      const conf = await confa.fetchOne({
        handle: confaHanle,
      })
      if (conf === null) {
        alert("confa not found")
        return
      }
      this.talk = await talk.create(conf.id)
      this.$router.replace("/" + confaHanle + "/" + this.talk.handle)
    } catch (e) {
      this.modal = Dialog.Error
    }
  },

  async created() {
    // const signal = new Signal()
    // const client = new Client(signal)
    // const uid = Math.random().toString()
    // signal.onopen = async () => {
    //   await client.join("test session", uid)
    //   const local = await LocalStream.getUserMedia({
    //     codec: "vp8",
    //     resolution: "vga",
    //     simulcast: true,
    //     video: true,
    //     audio: false,
    //   })
    //   this.localStream = local
    //   client.publish(local)
    // }
    // client.ontrack = (track: MediaStreamTrack, stream: RemoteStream) => {
    //   if (track.kind !== "video") {
    //     return
    //   }
    //   this.remoteStreams[stream.id] = stream
    //   stream.onremovetrack = () => {
    //     delete this.remoteStreams[stream.id]
    //   }
    // }
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
  },
})
</script>

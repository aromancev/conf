<template>
  <div class="container">
    <div class="row">
      <audience ref="audience" />
      <messages ref="messages" :userId="userId" :emitter="rtc" />

      <div class="btn px-3 py-1" @click="shareCamera">Share camera</div>
      <div class="btn px-3 py-1" @click="shareScreen">Share screen</div>
      <div class="btn px-3 py-1" @click="unshareCamera">UNShare camera</div>
      <div class="btn px-3 py-1" @click="unshareScreen">UNShare screen</div>

      <video :srcObject="remoteView.screen" autoplay muted />
      <video :srcObject="remoteView.camera" autoplay muted />

      <InternalError
        v-if="modal === Dialog.Error"
        v-on:click="modal = Dialog.None"
      />
    </div>
  </div>
</template>

<script lang="ts">
import InternalError from "@/components/modals/InternalError.vue"
import Audience from "@/components/room/audience.vue"
import Messages from "@/components/room/messages.vue"
import { defineComponent } from "vue"
import { Client, LocalStream, RemoteStream } from "ion-sdk-js"
import { EventType, PayloadPeerState } from "@/api/models"
import { userStore, RTC, Event, client, event, State, Hint, Track } from "@/api"
import { RecordProcessor, BufferedProcessor } from "@/components/room"
import { Record } from "@/components/room/record"

enum Dialog {
  None = "",
  Error = "error",
}

interface RemoteView {
  camera?: RemoteStream
  screen?: RemoteStream
}

function trackId(s: MediaStream): string {
  const tracks = s.getTracks()
  if (tracks.length < 1) {
    return ""
  }
  return tracks[0].id
}

export default defineComponent({
  name: "LiveRoom",
  components: {
    Audience,
    Messages,
    InternalError,
  },
  props: {
    roomId: String,
  },

  data() {
    return {
      Dialog,
      modal: Dialog.None,
      rtc: null as RTC | null,
      sfu: null as Client | null,
      localCamera: null as LocalStream | null,
      localScreen: null as LocalStream | null,
      localCameraLoading: false,
      localScreenLoading: false,
      streamsByTrackId: {} as { [key: string]: RemoteStream },
      tracksById: {} as { [key: string]: Track },
      remoteCamera: null as RemoteStream | null,
      state: { tracks: {} } as State,
    }
  },

  computed: {
    userId() {
      return userStore.getState().id
    },
    remoteView(): RemoteView {
      const view = {} as RemoteView
      for (const id in this.streamsByTrackId) {
        const track = this.tracksById[id]
        if (!track) {
          continue
        }
        switch (track.hint) {
          case Hint.Camera:
            view.camera = this.streamsByTrackId[id]
            break
          case Hint.Screen:
            view.screen = this.streamsByTrackId[id]
            break
        }
      }
      return view
    },
  },

  watch: {
    async roomId(val: string) {
      const roomId = val

      const processors = [
        this.$refs.audience,
        this.$refs.messages,
        this,
      ] as RecordProcessor[]
      const buffered = new BufferedProcessor(processors, 500)

      const rtc = await client.rtc(roomId)
      const sfu = new Client(rtc)

      rtc.onevent = (event: Event) => {
        buffered.put([event], true)
      }
      rtc.onopen = async () => {
        await sfu.join(roomId, this.userId)
        this.sfu = sfu
        this.rtc = rtc
      }
      sfu.ontrack = (track: MediaStreamTrack, stream: RemoteStream) => {
        if (track.kind !== "video") {
          return
        }

        const id = trackId(stream)
        this.streamsByTrackId[id] = stream
        stream.onremovetrack = () => {
          delete this.streamsByTrackId[id]
        }
      }

      const iter = event.fetch({ roomId: val })
      const events = await iter.next({ count: 500, seconds: 2 * 60 * 60 })
      buffered.flush()
      buffered.put(events, false)
      buffered.autoflush = true
    },
  },

  methods: {
    async shareCamera() {
      if (!this.sfu || this.localCameraLoading) {
        return
      }

      try {
        this.localCameraLoading = true
        this.localCamera = await LocalStream.getUserMedia({
          codec: "vp8",
          resolution: "vga",
          simulcast: true,
          video: true,
          audio: false,
        })
        this.state.tracks[trackId(this.localCamera)] = { hint: Hint.Camera }
        await this.rtc?.state(this.state)
        this.sfu.publish(this.localCamera)
      } finally {
        this.localCameraLoading = false
      }
    },
    unshareCamera() {
      if (!this.localCamera) {
        return
      }
      delete this.state.tracks[trackId(this.localCamera)]
      this.localCamera.unpublish()
      this.localCamera = null
    },
    async shareScreen() {
      if (!this.sfu || this.localScreenLoading) {
        return
      }

      try {
        this.localScreenLoading = true
        this.localScreen = await LocalStream.getDisplayMedia({
          codec: "vp8",
          resolution: "hd",
          simulcast: true,
          video: {
            width: { ideal: 2560 },
            height: { ideal: 1440 },
            frameRate: {
              ideal: 15,
              max: 30,
            },
          },
          audio: false,
        })
        for (const t of this.localScreen.getTracks()) {
          t.onended = () => {
            this.unshareScreen()
          }
        }
        this.state.tracks[trackId(this.localScreen)] = { hint: Hint.Screen }
        await this.rtc?.state(this.state)
        this.sfu.publish(this.localScreen)
      } finally {
        this.localScreenLoading = false
      }
    },
    unshareScreen() {
      if (!this.localScreen) {
        return
      }
      delete this.state.tracks[trackId(this.localScreen)]
      this.localScreen.unpublish()
      this.localScreen = null
    },
    processRecords(records: Record[]): void {
      for (const record of records) {
        if (record.event.payload.type !== EventType.PeerState) {
          continue
        }
        const payload = record.event.payload.payload as PayloadPeerState
        if (!payload.tracks) {
          continue
        }
        Object.assign(this.tracksById, payload.tracks)
      }
    },
  },
})
</script>

<style scoped lang="sass">
@use '@/css/theme'

.audience
  @include theme.shadow-inset-s

  position: absolute
  right: 0
  width: 200px
  height: 300px

.messages
  @include theme.shadow-inset-s

  position: absolute
  right: 0
  margin-top: 300px
  width: 200px
  height: 300px

.camera
  // width: 500px
  // height: 500px
  border: 1px solid red
</style>

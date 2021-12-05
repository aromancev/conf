<template>
  <div class="content">
    <div class="room">
      <div class="video-content">
        <div class="videos">
          <div class="screen video-container">
            <video
              v-if="remoteView.screen"
              class="screen-video"
              :srcObject="remoteView.screen"
              autoplay
              muted
            />
            <div v-else class="video-off">
              <div class="video-off-icon material-icons">
                desktop_access_disabled
              </div>
            </div>
          </div>
          <div class="camera video-container">
            <video
              v-if="remoteView.screen"
              class="camera-video"
              :srcObject="localCamera"
              autoplay
              muted
            />
            <div v-else class="video-off">
              <div class="video-off-icon material-icons">videocam_off</div>
            </div>
          </div>
        </div>
      </div>

      <audience ref="audience" />
    </div>
    <div class="controls">
      <div
        v-if="sidePanel !== SidePanel.None"
        class="ctrl-btn btn-switch px-3 py-3 material-icons"
        @click="switchSidePanel(SidePanel.None)"
      >
        close
      </div>
      <div
        class="ctrl-btn btn-switch px-3 py-3 material-icons"
        :class="{ pressed: sidePanel === SidePanel.Chat }"
        @click="switchSidePanel(SidePanel.Chat)"
      >
        chat
      </div>
    </div>
    <div class="side-panel" :class="{ opened: sidePanel !== SidePanel.None }">
      <messages ref="messages" :userId="userId" :emitter="rtc" />
    </div>
  </div>

  <InternalError
    v-if="modal === Dialog.Error"
    v-on:click="modal = Dialog.None"
  />
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

enum SidePanel {
  None = "",
  Chat = "chat",
}

interface RemoteView {
  camera?: RemoteStream
  screen?: RemoteStream
}

interface Resizer {
  resize(): void
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
      SidePanel,
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
      sidePanel: SidePanel.None,
    }
  },

  async mounted() {
    this.localCamera = await LocalStream.getUserMedia({
      codec: "vp8",
      resolution: "vga",
      simulcast: true,
      video: true,
      audio: false,
    })
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
    switchSidePanel(panel: SidePanel) {
      if (this.sidePanel === panel) {
        panel = SidePanel.None
      }
      this.sidePanel = panel
      this.$nextTick(() => {
        ;(this.$refs.audience as Resizer).resize()
      })
    },
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

.content
  width: 100%
  height: 100%

  display: flex
  flex-direction: row
  padding: 20px

.room
  flex: 1
  display: flex
  flex-direction: column

.videos
  display: flex
  flex-direction: row
  justify-content: center
  align-items: flex-start
  max-width: 1000px
  width: 100%

.video-container
  overflow: hidden
  position: relative

.video-content
  display: flex
  flex-direction: row
  justify-content: center

video
  position: absolute
  left: 50%
  top: 50%
  transform: translate(-50%, -50%)

.video-off
  top: 0
  left: 0
  position: absolute
  width: 100%
  height: 100%
  background: var(--color-background)
  cursor: default
  display: flex
  align-items: center
  justify-content: center
  user-select: none
  -webkit-tap-highlight-color: rgba(0,0,0,0)

.video-off-icon
  font-size: 50px
  color: var(--color-highlight-background)

.screen-video
  max-height: 100%
  max-width: 100%
  width: 100%

.camera-video
  height: 100%

.screen
  @include theme.shadow-l

  flex: 3
  border-radius: 4px
  background: black
  margin: 10px
  padding-top: 50%

.camera
  @include theme.shadow-m

  flex: 1
  border-radius: 4px
  background: black
  margin: 10px
  padding-top: 20%

.audience
  flex: 1
  border-radius: 4px
  margin: 10px

.controls
  display: flex
  flex-direction: column
  align-items: center
  justify-content: flex-end
  width: 60px
  margin: 30px

.ctrl-btn
  border-radius: 50%
  margin: 10px

.side-panel
  display: none
  flex-direction: column
  width: 450px
  &.opened
    display: flex

.messages
  @include theme.shadow-inset-m

  border-radius: 4px
  flex: 1
  margin: 10px
</style>

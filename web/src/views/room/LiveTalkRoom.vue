<template>
  <div class="content">
    <div class="room">
      <div class="video-content">
        <div class="videos">
          <div class="screen video-container">
            <video v-if="remote.screen" class="screen-video" :srcObject="remote.screen" autoplay muted></video>
            <div v-else class="video-off">
              <div class="video-off-icon material-icons">desktop_access_disabled</div>
            </div>
          </div>
          <div class="camera video-container">
            <video
              v-if="remote.camera"
              class="camera-video"
              :class="{ local: local.camera }"
              :srcObject="remote.camera"
              autoplay
              muted
            ></video>
            <div v-else class="video-off">
              <div class="video-off-icon material-icons">videocam_off</div>
            </div>
          </div>
        </div>
      </div>

      <RoomAudience ref="audience" />
    </div>
    <div class="controls">
      <div class="controls-top">
        <div
          class="ctrl-btn btn-switch material-icons"
          :class="{ active: local.screen }"
          :disabled="publishing ? true : null"
          @click="switchScreen"
        >
          {{ local.screen ? "desktop_windows" : "desktop_access_disabled" }}
        </div>
        <div
          class="ctrl-btn btn-switch material-icons"
          :class="{ active: local.camera }"
          :disabled="publishing ? true : null"
          @click="switchCamera"
        >
          {{ local.camera ? "videocam" : "videocam_off" }}
        </div>
        <div
          class="ctrl-btn btn-switch material-icons"
          :class="{ active: local.mic }"
          :disabled="publishing ? true : null"
          @click="switchMic"
        >
          {{ local.mic ? "mic" : "mic_off" }}
        </div>
      </div>
      <div class="controls-bottom">
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
    </div>
    <div v-if="sidePanel !== SidePanel.None" class="side-panel">
      <RoomMessages :user-id="user.id" :messages="messages" @message="sendMessage" />
    </div>
  </div>

  <audio v-for="stream in remote.audios" :key="stream.id" :srcObject="stream" autoplay></audio>

  <InternalError v-if="modal === Modal.Error" @click="modal = Modal.None" />
</template>

<script setup lang="ts">
import { reactive, computed, ref, watch, nextTick } from "vue"
import { Client, LocalStream, RemoteStream, Constraints } from "ion-sdk-js"
import { EventType, PayloadPeerState } from "@/api/models"
import { userStore, RTC, Event, client, eventClient, State, Hint, Track } from "@/api"
import { RecordProcessor, BufferedProcessor } from "@/components/room"
import { Record } from "@/components/room/record"
import { MessageProcessor, Message } from "@/components/room/messages"
import InternalError from "@/components/modals/InternalError.vue"
import RoomAudience from "@/components/room/RoomAudience.vue"
import RoomMessages from "@/components/room/RoomMessages.vue"

enum Modal {
  None = "",
  Error = "error",
}

enum SidePanel {
  None = "",
  Chat = "chat",
}

interface Remote {
  camera?: MediaStream
  screen?: MediaStream
  audios: MediaStream[]
}

interface Local {
  camera: LocalStream | null
  screen: LocalStream | null
  mic: LocalStream | null
  test: string
}

interface Resizer {
  resize(): void
}

const user = userStore.getState()

const props = defineProps<{
  roomId: string
}>()

const modal = ref(Modal.None)
const local = reactive<Local>({
  camera: null,
  screen: null,
  mic: null,
  test: "123",
}) as Local
const publishing = ref(false)
const sidePanel = ref(SidePanel.None)
const audience = ref<RecordProcessor & Resizer>()

const streamsByTrackId = {} as { [key: string]: RemoteStream }
const tracksById = {} as { [key: string]: Track }
const state = { tracks: {} } as State

let messageProcessor = null as MessageProcessor | null
let rtcClient = null as RTC | null
let sfuClient = null as Client | null

const messages = computed((): Message[] => {
  if (!messageProcessor) {
    return []
  }
  return messageProcessor.messages()
})

const remote = computed((): Remote => {
  const view = {
    audios: [],
  } as Remote

  for (const id in streamsByTrackId) {
    const track = tracksById[id]
    if (!track) {
      continue
    }
    switch (track.hint) {
      case Hint.Camera:
        view.camera = streamsByTrackId[id]
        break
      case Hint.Screen:
        view.screen = streamsByTrackId[id]
        break
      case Hint.UserAudio:
        view.audios.push(streamsByTrackId[id])
        break
    }
  }
  if (local.camera) {
    view.camera = local.camera
  }
  if (local.screen) {
    view.screen = local.screen
  }
  return view
})

watch(
  () => props.roomId,
  async (value: string) => {
    const roomId = value

    const rtc = await client.rtc(roomId)
    const sfu = new Client(rtc)

    messageProcessor = new MessageProcessor(rtc)

    const processors = [audience.value, messageProcessor, { processRecords }] as RecordProcessor[]
    const buffered = new BufferedProcessor(processors, 500)

    rtc.onevent = (event: Event) => {
      buffered.put([event], true)
    }
    rtc.onopen = async () => {
      await sfu.join(roomId, user.id)
      sfuClient = sfu
      rtcClient = rtc
    }
    sfu.ontrack = (track: MediaStreamTrack, stream: RemoteStream) => {
      if (track.kind !== "video" && track.kind !== "audio") {
        return
      }

      const id = trackId(stream)
      streamsByTrackId[id] = stream
      stream.onremovetrack = () => {
        delete streamsByTrackId[id]
      }
    }

    const iter = eventClient.fetch({ roomId: value })
    const events = await iter.next({ count: 500, seconds: 2 * 60 * 60 })
    buffered.flush()
    buffered.put(events, false)
    buffered.autoflush = true
  },
  { immediate: true },
)

function trackId(s: MediaStream): string {
  const tracks = s.getTracks()
  if (tracks.length < 1) {
    return ""
  }
  return tracks[0].id
}

function sendMessage(message: string): void {
  messageProcessor?.send(user.id, message)
}

function switchSidePanel(panel: SidePanel) {
  if (sidePanel.value === panel) {
    panel = SidePanel.None
    return
  }
  sidePanel.value = panel
  nextTick(() => {
    if (!audience.value) {
      return
    }
    audience.value.resize()
  })
}

function switchCamera() {
  if (local.camera) {
    unshareCamera()
  } else {
    shareCamera()
  }
}

function switchScreen() {
  if (local.screen) {
    unshareScreen()
  } else {
    shareScreen()
  }
}

function switchMic() {
  if (local.mic) {
    unshareMic()
  } else {
    shareMic()
  }
}

async function shareCamera() {
  local.camera = await share(async () => {
    return await LocalStream.getUserMedia({
      codec: "vp8",
      resolution: "vga",
      simulcast: true,
      video: true,
      audio: false,
    })
  }, Hint.Camera)
}

function unshareCamera() {
  unshare(local.camera)
  local.camera = null
}

async function shareScreen() {
  local.screen = await share(async () => {
    const stream = await LocalStream.getDisplayMedia({
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
    for (const t of stream.getTracks()) {
      t.onended = () => {
        unshareScreen()
      }
    }
    return stream
  }, Hint.Screen)
}

function unshareScreen() {
  unshare(local.screen)
  local.screen = null
}

async function shareMic() {
  local.mic = await share(() => {
    return LocalStream.getUserMedia({
      video: false,
      audio: true,
    } as Constraints)
  }, Hint.UserAudio)
}

function unshareMic() {
  unshare(local.mic)
  local.mic = null
}

async function share(fetch: () => Promise<LocalStream>, hint: Hint): Promise<LocalStream | null> {
  if (!rtcClient || !sfuClient || publishing.value) {
    return null
  }

  try {
    publishing.value = true
    const stream = await fetch()
    state.tracks[trackId(stream)] = { hint: hint }
    await rtcClient.state(state)
    sfuClient.publish(stream)
    return stream
  } catch (e) {
    console.warn("Failed to share media.")
    return null
  } finally {
    publishing.value = false
  }
}

function unshare(stream: LocalStream | null | undefined) {
  if (!stream) {
    return
  }
  delete state.tracks[trackId(stream)]
  stream.unpublish()
  for (const t of stream.getTracks()) {
    t.stop()
  }
}

function processRecords(records: Record[]): void {
  for (const record of records) {
    if (record.event.payload.type !== EventType.PeerState) {
      continue
    }
    const payload = record.event.payload.payload as PayloadPeerState
    if (!payload.tracks) {
      continue
    }
    Object.assign(tracksById, payload.tracks)
  }
}
</script>

<style scoped lang="sass">
@use '@/css/theme'

.content
  width: 100%
  height: 100%

  display: flex
  flex-direction: row
  padding: 30px

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
  &.local
    transform: scale(-1, 1)
    left: 0
    top: 0
    max-height: 100%
    max-width: 100%
    width: 100%

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
  margin: 0 10px
  padding-top: 50%

.camera
  @include theme.shadow-m

  flex: 1
  border-radius: 4px
  background: black
  margin: 0 10px
  padding-top: 20%

.audience
  flex: 1
  border-radius: 4px

.controls
  display: flex
  flex-direction: column
  align-items: center
  justify-content: flex-start
  width: 60px
  margin: 0 20px

.controls-bottom
  margin-top: auto

.ctrl-btn
  border-radius: 50%
  margin: 11px
  padding: 0.6em
  &.active
    margin: 10px
    border: 1px solid var(--color-highlight-background)

.side-panel
  display: flex
  flex-direction: column
  width: 450px
  max-height: 100%
  overflow: hidden

.messages
  @include theme.shadow-inset-m

  border-radius: 4px
  flex: 1
  max-height: 100%
</style>

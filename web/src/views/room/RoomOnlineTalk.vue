<template>
  <div class="content">
    <div class="room">
      <div class="video-content">
        <div class="videos">
          <div class="screen video-container">
            <video
              v-if="local.screen || remote.screen"
              class="screen-video"
              :srcObject="local.screen || remote.screen"
              autoplay
              muted
            ></video>
            <div v-else class="video-off">
              <div class="video-off-icon material-icons">desktop_access_disabled</div>
            </div>
          </div>
          <div class="camera video-container">
            <video
              v-if="local.camera || remote.camera"
              class="camera-video"
              :class="{ local: local.camera }"
              :srcObject="local.camera || remote.camera"
              autoplay
              muted
            ></video>
            <div v-else class="video-off">
              <div class="video-off-icon material-icons">videocam_off</div>
            </div>
          </div>
        </div>
      </div>

      <RoomAudience ref="audience" :loading="loading" />
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
          class="ctrl-btn btn-switch material-icons"
          @click="switchSidePanel(SidePanel.None)"
        >
          close
        </div>
        <div
          class="ctrl-btn btn-switch material-icons"
          :class="{ pressed: sidePanel === SidePanel.Chat }"
          @click="switchSidePanel(SidePanel.Chat)"
        >
          chat
        </div>
      </div>
    </div>
    <div v-if="sidePanel !== SidePanel.None" class="side-panel">
      <RoomMessages :user-id="user.id" :messages="messages" :loading="loading" @message="sendMessage" />
    </div>
  </div>

  <div v-if="joinConfirmed">
    <audio v-for="stream in remote.audios" :key="stream.id" :srcObject="stream" autoplay></audio>
  </div>

  <ModalDialog v-if="modal === Modal.ConfirmJoin" :buttons="{ join: 'Join', leave: 'Leave' }" @click="confirmJoin">
    <p>You are about to join the talk online</p>
    <p v-if="inviteLink">
      Share this link to invite people<br />
      <CopyField :value="inviteLink">Test</CopyField>
    </p>
  </ModalDialog>
  <ModalDialog v-if="modal === Modal.ConfirmLeave" :buttons="{ leave: 'Leave', stay: 'Stay' }" @click="onModalClose">
    <p>You are about to leave the talk while presenting.</p>
    <p>If you leave, your presentation will end.</p>
  </ModalDialog>
  <InternalError v-if="modal === Modal.Error" @click="modal = Modal.None" />
</template>

<script setup lang="ts">
import { reactive, computed, ref, watch, nextTick, onMounted, onUnmounted } from "vue"
import { onBeforeRouteLeave } from "vue-router"
import { Client, LocalStream, RemoteStream, Constraints } from "ion-sdk-js"
import { EventType, PayloadPeerState } from "@/api/models"
import { userStore, RTC, Event, client, eventClient, State, Hint, Track, Policy } from "@/api"
import { RecordProcessor, BufferedProcessor } from "@/components/room"
import { Record } from "@/components/room/record"
import { MessageProcessor, Message } from "@/components/room/messages"
import InternalError from "@/components/modals/InternalError.vue"
import RoomAudience from "@/components/room/RoomAudience.vue"
import RoomMessages from "@/components/room/RoomMessages.vue"
import ModalDialog from "@/components/modals/ModalDialog.vue"
import CopyField from "@/components/fields/CopyField.vue"
import { EventOrder } from "@/api/schema"

enum Modal {
  None = "",
  Error = "error",
  ConfirmJoin = "confirm_join",
  ConfirmLeave = "confirm_leave",
}

enum SidePanel {
  None = "",
  Chat = "chat",
}

interface Remote {
  camera: MediaStream | null
  screen: MediaStream | null
  audios: MediaStream[]
}

interface Local {
  camera: LocalStream | null
  screen: LocalStream | null
  mic: LocalStream | null
}

interface Resizer {
  resize(): void
}

const emit = defineEmits<{
  (e: "join", confirmed: boolean): void
}>()

const sidePanelKey = "roomSidePanel"

const user = userStore.getState()

const props = defineProps<{
  roomId: string
  inviteLink?: string
  joinConfirmed?: boolean
}>()

const modal = ref(Modal.None)
const local = reactive<Local>({
  camera: null,
  screen: null,
  mic: null,
}) as Local
const remote = reactive<Remote>({
  camera: null,
  screen: null,
  audios: [],
})
const publishing = ref(false)
const loading = ref(false)
const sidePanel = ref(localStorage.getItem(sidePanelKey) || SidePanel.None)
const audience = ref<RecordProcessor & Resizer>()

const streamsByTrackId = {} as { [key: string]: RemoteStream }
const tracksById = {} as { [key: string]: Track }
const state = { tracks: {} } as State

let messageProcessor = ref<MessageProcessor | null>(null)
let rtcClient = null as RTC | null
let sfuClient = null as Client | null
let modalClosed: (button: string) => void = (button: string) => {}

const messages = computed((): Message[] => {
  if (!messageProcessor.value) {
    return []
  }
  return messageProcessor.value.messages()
})

watch(
  () => props.roomId,
  async (value: string) => {
    const roomId = value
    loading.value = true

    const rtc = await client.rtc(roomId)
    const sfu = new Client(rtc)

    messageProcessor.value = new MessageProcessor(rtc)

    const processors: RecordProcessor[] = [messageProcessor.value, { processRecords }]
    if (audience.value) {
      processors.push(audience.value)
    }
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
      computeRemote()
      stream.onremovetrack = () => {
        delete streamsByTrackId[id]
        computeRemote()
      }
    }

    const iter = eventClient.fetch({ roomId: value }, EventOrder.DESC, Policy.NetworkOnly)
    const events = await iter.next({ count: 500, seconds: 2 * 60 * 60 })
    buffered.put(events, false)
    buffered.autoflush = true
    buffered.flush()

    loading.value = false
  },
  { immediate: true },
)

onMounted(() => {
  if (!props.joinConfirmed) {
    modal.value = Modal.ConfirmJoin
  }
})

onUnmounted(() => {
  unshareScreen()
  unshareCamera()
  unshareMic()
  sfuClient?.close()
  rtcClient?.close()
})

onBeforeRouteLeave(async (to, from, next) => {
  if (!local.screen && !local.camera && !local.mic) {
    next()
    return
  }
  const btn = await new Promise<string>((resolve) => {
    modalClosed = (button: string) => {
      resolve(button)
    } 
    modal.value = Modal.ConfirmLeave
  })
  next(btn === "leave")
})

function onModalClose(button: string): void {
  modalClosed(button)
  modal.value = Modal.None
}

function confirmJoin(value: string) {
  emit("join", value === "join")
  modal.value = Modal.None
}

function trackId(s: MediaStream): string {
  const tracks = s.getTracks()
  if (tracks.length < 1) {
    return ""
  }
  return tracks[0].id
}

function sendMessage(message: string): void {
  if (!messageProcessor.value) {
    return
  }
  messageProcessor.value.send(user.id, message)
}

function switchSidePanel(panel: SidePanel) {
  if (sidePanel.value === panel) {
    panel = SidePanel.None
  } else {
    sidePanel.value = panel
  }
  localStorage.setItem(sidePanelKey, panel)

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
    console.warn("Failed to share media:", e)
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
    stream.removeTrack(t)
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
  computeRemote()
}

function computeRemote() {
  remote.camera = null
  remote.screen = null
  remote.audios = []

  for (const id in streamsByTrackId) {
    const track = tracksById[id]
    if (!track) {
      continue
    }
    switch (track.hint) {
      case Hint.Camera:
        remote.camera = streamsByTrackId[id]
        break
      case Hint.Screen:
        remote.screen = streamsByTrackId[id]
        break
      case Hint.UserAudio:
        remote.audios.push(streamsByTrackId[id])
        break
    }
  }
  if (local.camera) {
    remote.camera = local.camera
  }
  if (local.screen) {
    remote.screen = local.screen
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
  max-width: min(90%, 800px)
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
  width: 30%
  max-width: 450px
  max-height: 100%
  overflow: hidden

.messages
  @include theme.shadow-inset-m

  border-radius: 4px
  flex: 1
  max-height: 100%
</style>

<template>
  <div class="content">
    <div class="room">
      <div class="video-content">
        <div class="videos">
          <div class="screen video-container">
            <RoomLiveVideo
              v-if="localScreen || remoteScreen"
              class="video screen-video"
              :track="localScreen || remoteScreen"
            ></RoomLiveVideo>
            <div v-else class="video-off">
              <div class="video-off-icon material-icons">desktop_access_disabled</div>
            </div>
            <div v-if="recordingStatus === 'RECORDING'" class="rec-indicator"></div>
          </div>
          <div class="camera video-container">
            <RoomLiveVideo
              v-if="localCamera || remoteCamera"
              :track="localCamera || remoteCamera"
              class="video camera-video"
              :class="{ local: localCamera }"
              :disable-controls="true"
            ></RoomLiveVideo>
            <div v-else class="video-off">
              <div class="video-off-icon material-icons">videocam_off</div>
            </div>
          </div>
        </div>
      </div>
      <RoomAudience
        ref="audience"
        :user-id="accessStore.state.id"
        :is-loading="room.state.rtc !== 'CONNECTED'"
        :is-playing="true"
        :peers="room.state.peers"
        :statuses="room.state.statuses"
      />
    </div>
    <div class="controls">
      <div class="controls-top">
        <RoomReactions @reaction="sendReaction"></RoomReactions>
        <div
          v-if="accessStore.state.id === talk.ownerId"
          class="ctrl-btn btn-switch material-icons"
          :class="{ active: localScreen }"
          :disabled="room.state.sfu !== 'CONNECTED' ? true : null"
          @click="room.switchScreen"
        >
          {{ localScreen ? "desktop_windows" : "desktop_access_disabled" }}
        </div>
        <div
          v-if="accessStore.state.id === talk.ownerId"
          class="ctrl-btn btn-switch material-icons"
          :class="{ active: localCamera }"
          :disabled="room.state.sfu !== 'CONNECTED' ? true : null"
          @click="room.switchCamera"
        >
          {{ localCamera ? "videocam" : "videocam_off" }}
        </div>
        <div
          class="ctrl-btn btn-switch material-icons"
          :class="{ active: localMic }"
          :disabled="room.state.sfu !== 'CONNECTED' ? true : null"
          @click="room.switchMic"
        >
          {{ localMic ? "mic" : "mic_off" }}
        </div>
        <div
          v-if="recordingStatus !== 'STOPPED' && accessStore.state.id === talk.ownerId"
          class="ctrl-btn btn-switch material-icons record-icon"
          :disabled="recordingStatus === 'PENDING' ? true : null"
          @click="handleRecording"
        >
          {{ recordingStatus !== "RECORDING" ? "radio_button_checked" : "stop_circle" }}
        </div>
      </div>
      <div class="controls-bottom">
        <div v-if="sidePanel !== 'NONE'" class="ctrl-btn btn-switch material-icons" @click="switchSidePanel('NONE')">
          close
        </div>
        <div
          class="ctrl-btn btn-switch material-icons"
          :class="{ pressed: sidePanel === 'CHAT' }"
          @click="switchSidePanel('CHAT')"
        >
          chat
          <div
            v-if="room.state.messages.length && lastReadMessageId !== room.state.messages.at(-1)?.id"
            class="new-content-marker"
          ></div>
        </div>
      </div>
    </div>
    <div v-if="sidePanel !== 'NONE'" class="side-panel">
      <RoomMessages
        :user-id="accessStore.state.id"
        :messages="room.state.messages"
        :is-loading="room.state.rtc !== 'CONNECTED'"
        @sent="sendMessage"
      />
    </div>
  </div>

  <RoomLiveAudio v-for="t in remoteMics" :key="t.id" :track="t"></RoomLiveAudio>

  <ModalDialog
    :is-visible="modal.state === 'CONFIRM_SFU_CONNECT'"
    :buttons="[
      {
        text: 'Leave',
        click: () => {
          router.push(route.talk(props.confaHandle, talk.handle, 'overview'))
        },
      },
      {
        text: 'Join',
        click: () => {
          room.connectSFU(roomId)
          modal.set()
        },
      },
    ]"
  >
    <p>You are about to join a live talk</p>
    <p v-if="inviteLink">
      Share this link to invite people<br />
      <CopyField :value="inviteLink"></CopyField>
    </p>
  </ModalDialog>
  <ModalDialog
    :is-visible="modal.state === 'CONFIRM_SFU_RECONNECT'"
    :buttons="[
      {
        text: 'Leave',
        click: () => {
          router.push(route.talk(props.confaHandle, talk.handle, 'overview'))
        },
      },
      {
        text: 'Reconnect',
        click: () => {
          room.connectSFU(roomId)
          modal.set()
        },
      },
    ]"
  >
    <p>You joined this talk from another device or tab</p>
    <p>Only one active session is allowed</p>
  </ModalDialog>
  <ModalDialog
    :is-visible="modal.state === 'CONFIRM_LEAVE'"
    :ctrl="modal"
    :buttons="[
      { id: 'stay', text: 'Stay' },
      { id: 'leave', text: 'Leave' },
    ]"
  >
    <p>You are about to leave the talk while presenting.</p>
    <p>If you leave, your presentation will end.</p>
  </ModalDialog>
  <ModalDialog
    :is-visible="modal.state === 'RECORDING_FINISHED'"
    :buttons="[
      { text: 'Stay', click: () => modal.set() },
      { text: 'Go to recording', click: recordingFinished },
    ]"
  >
    <p>Recording finished.</p>
    <p>For demo purposes it is limited to 5 minutes.</p>
  </ModalDialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onUnmounted, provide } from "vue"
import { useRouter } from "vue-router"
import { onBeforeRouteLeave } from "vue-router"
import { api } from "@/api"
import { TalkClient } from "@/api/talk"
import { Talk, TalkState } from "@/api/models/talk"
import { accessStore } from "@/api/models/access"
import { LiveRoom } from "@/components/room"
import { Reaction, TrackSource } from "@/api/rtc/schema"
import { ModalController } from "@/components/modals/controller"
import { Track } from "@/components/room/live"
import RoomAudience from "@/components/room/RoomAudience.vue"
import RoomMessages from "@/components/room/RoomMessages.vue"
import RoomLiveVideo from "@/components/room/RoomLiveVideo.vue"
import RoomLiveAudio from "@/components/room/RoomLiveAudio.vue"
import RoomReactions from "@/components/room/RoomReactions.vue"
import ModalDialog from "@/components/modals/ModalDialog.vue"
import CopyField from "@/components/fields/CopyField.vue"
import { Throttler } from "@/platform/sync"
import { notificationStore } from "@/api/models/notifications"
import { route } from "@/router"

type RecordingStatus = "NONE" | "PENDING" | "RECORDING" | "STOPPED"

const modal = new ModalController<
  "CONFIRM_SFU_CONNECT" | "CONFIRM_LEAVE" | "RECORDING_FINISHED" | "CONFIRM_SFU_RECONNECT"
>("CONFIRM_SFU_CONNECT")

type SidePanel = "NONE" | "CHAT"

type Resizer = {
  resize(): void
}

const emit = defineEmits<{
  (e: "update", talk: Talk): void
}>()

const props = defineProps<{
  talk: Talk
  confaHandle: string
  inviteLink?: string
}>()

const router = useRouter()

const sidePanelKey = "roomSidePanel"
const sidePanel = ref<SidePanel>((localStorage.getItem(sidePanelKey) as SidePanel) || "NONE")
const audience = ref<Resizer>()
const room = new LiveRoom()
const recordingStatus = ref<RecordingStatus>("NONE")
const roomId = computed<string>(() => {
  return props.talk.roomId
})
const lastReadMessageId = ref<string>("")

const localScreen = computed<Track | undefined>(() => {
  return track(room.state.localTracks, TrackSource.Screen)
})
const remoteScreen = computed<Track | undefined>(() => {
  return track(room.state.remoteTracks, TrackSource.Screen)
})
const localCamera = computed<Track | undefined>(() => {
  return track(room.state.localTracks, TrackSource.Camera)
})
const remoteCamera = computed<Track | undefined>(() => {
  return track(room.state.remoteTracks, TrackSource.Camera)
})
const localMic = computed<Track | undefined>(() => {
  return track(room.state.localTracks, TrackSource.Microphone)
})
const remoteMics = computed<Track[]>(() => {
  const tracks: Track[] = []
  for (const track of room.state.remoteTracks.values()) {
    if (track.source === TrackSource.Microphone) {
      tracks.push(track)
    }
  }
  return tracks
})

const connectThrottler = new Throttler({ delayMs: 1000 })
let connectRetries = 0
connectThrottler.func = async () => {
  try {
    await room.connectRTC(roomId.value)
    if (!modal.state) {
      await room.connectSFU(roomId.value)
    }
    if (connectRetries > 0) {
      notificationStore.info("connection restored")
      connectRetries = 0
    }
  } catch (e) {
    if (connectRetries === 0) {
      notificationStore.error("connection lost")
    }
    connectRetries++
    connectThrottler.do()
  }
}

provide("attacher", room)

watch(
  [roomId, () => accessStore.state.id],
  () => {
    connectThrottler.do()
  },
  { immediate: true },
)

watch(room.state.recording, async (r) => {
  if (r.isRecording) {
    recordingStatus.value = "RECORDING"
  } else {
    recordingStatus.value = "STOPPED"
    modal.set("RECORDING_FINISHED")
  }
})

watch(
  () => room.state.rtc,
  async (state) => {
    if (state !== "ERROR") {
      return
    }
    connectThrottler.do()
  },
)

watch(
  () => room.state.sfu,
  async (state) => {
    if (state !== "DISCONNECTED") {
      return
    }
    modal.set("CONFIRM_SFU_RECONNECT")
  },
)

watch(
  () => props.talk.state,
  (state?: TalkState) => {
    if (!state) {
      return
    }

    switch (state) {
      case TalkState.RECORDING:
        recordingStatus.value = "RECORDING"
        break
      case TalkState.ENDED:
        recordingStatus.value = "STOPPED"
        break
    }
  },
  { immediate: true },
)

watch([room.state.messages, sidePanel], () => {
  if (sidePanel.value === "CHAT") {
    lastReadMessageId.value = room.state.messages.at(-1)?.id || ""
  }
})

onUnmounted(async () => {
  connectThrottler.close()
  await room.close()
})

onBeforeRouteLeave(async (to, from, next) => {
  if (room.state.localTracks.size === 0) {
    next()
    return
  }
  const btn = await modal.set("CONFIRM_LEAVE")
  next(btn === "leave")
})

function track(tracks: Map<string, Track>, source: TrackSource): Track | undefined {
  for (const t of tracks.values()) {
    if (t.source === source) {
      return t
    }
  }
  return undefined
}

function recordingFinished() {
  modal.set()
  const talk = Object.assign({}, props.talk)
  talk.state = TalkState.ENDED
  emit("update", talk)
}

function sendMessage(message: string) {
  room.send(accessStore.state.id, message)
}

function sendReaction(reaction: Reaction) {
  room.reaction(reaction)
}

function switchSidePanel(panel: SidePanel) {
  if (sidePanel.value === panel) {
    panel = "NONE"
  } else {
    sidePanel.value = panel
  }
  localStorage.setItem(sidePanelKey, panel)

  nextTick(() => {
    audience.value?.resize()
  })
}

async function handleRecording() {
  switch (recordingStatus.value) {
    case "NONE":
      recordingStatus.value = "PENDING"
      try {
        await new TalkClient(api).startRecording({ id: props.talk.id })
      } catch (e) {
        notificationStore.error("failed to start recording")
      }
      break
    case "RECORDING":
      recordingStatus.value = "PENDING"
      try {
        await new TalkClient(api).stopRecording({ id: props.talk.id })
      } catch (e) {
        notificationStore.error("failed to stop recording")
      }
      break
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
  max-width: min(90%, 700px)
  width: 100%

.video-container
  overflow: hidden
  position: relative

.rec-indicator
  position: absolute
  right: 0
  top: 0
  width: 10px
  height: 10px
  background: var(--color-red)
  border-radius: 50%
  margin: 10px

.video-content
  display: flex
  flex-direction: row
  justify-content: center
  margin-bottom: 20px

.video
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

.controls-top
  display: flex
  flex-direction: column
  align-items: center

.controls-bottom
  margin-top: auto

.ctrl-btn
  border-radius: 50%
  margin: 5px
  padding: 15px
  font-size: 25px
  position: relative
  box-sizing: border-box
  &.active
    outline: 1px solid var(--color-highlight-background)

.new-content-marker
  position: absolute
  right: 20%
  top: 20%
  width: 12px
  height: 12px
  border-radius: 50%
  background: var(--color-highlight-background)

.record-icon
  color: var(--color-red)

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

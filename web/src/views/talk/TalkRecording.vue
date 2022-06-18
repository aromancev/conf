<template>
  <div class="content">
    <div class="room">
      <div class="video-content">
        <div class="videos">
          <div class="screen video-container">
            <RoomOnlineVideo
              v-if="local.screen || remote.screen"
              class="video screen-video"
              :src="local.screen || remote.screen"
            >
            </RoomOnlineVideo>
            <div v-else class="video-off">
              <div class="video-off-icon material-icons">desktop_access_disabled</div>
            </div>
          </div>
          <div class="camera video-container">
            <video
              v-if="local.camera || remote.camera"
              class="video camera-video"
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

      <RoomAudience ref="audience" :loading="!roomJoined" :peers="room.peers" />
    </div>
    <div class="controls">
      <div class="controls-top">
        <div
          class="ctrl-btn btn-switch material-icons"
          :class="{ active: local.screen }"
          :disabled="roomJoined && !roomPublishing ? null : true"
          @click="room.switchScreen"
        >
          {{ local.screen ? "desktop_windows" : "desktop_access_disabled" }}
        </div>
        <div
          class="ctrl-btn btn-switch material-icons"
          :class="{ active: local.camera }"
          :disabled="roomJoined && !roomPublishing ? null : true"
          @click="room.switchCamera"
        >
          {{ local.camera ? "videocam" : "videocam_off" }}
        </div>
        <div
          class="ctrl-btn btn-switch material-icons"
          :class="{ active: local.mic }"
          :disabled="roomJoined && !roomPublishing ? null : true"
          @click="room.switchMic"
        >
          {{ local.mic ? "mic" : "mic_off" }}
        </div>
        <div
          class="ctrl-btn btn-switch material-icons record-icon"
          :disabled="recordingStatus === 'pending' || recordingStatus === 'stopped' ? true : null"
          @click="handleRecording"
        >
          {{ recordingStatus !== "recording" ? "radio_button_checked" : "stop_circle" }}
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
      <RoomMessages :user-id="user.id" :messages="room.messages" :loading="!roomJoined" @message="sendMessage" />
    </div>
  </div>

  <div v-if="joinConfirmed">
    <audio v-for="stream in remote.audios" :key="stream.id" :srcObject="stream" autoplay></audio>
  </div>

  <ModalDialog v-if="modal === 'confirm_join'" :buttons="{ join: 'Join', leave: 'Leave' }" @click="confirmJoin">
    <p>You are about to join the talk online</p>
    <p v-if="inviteLink">
      Share this link to invite people<br />
      <CopyField :value="inviteLink"></CopyField>
    </p>
  </ModalDialog>
  <ModalDialog v-if="modal === 'confirm_leave'" :buttons="{ leave: 'Leave', stay: 'Stay' }" @click="onModalClose">
    <p>You are about to leave the talk while presenting.</p>
    <p>If you leave, your presentation will end.</p>
  </ModalDialog>
  <InternalError v-if="modal === 'error'" @click="modal = 'none'" />
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from "vue"
import { onBeforeRouteLeave } from "vue-router"
import { talkClient } from "@/api"
import { Talk, TalkState, userStore } from "@/api/models"
import { LiveRoom } from "@/components/room"
import InternalError from "@/components/modals/InternalError.vue"
import RoomAudience from "@/components/room/RoomAudience.vue"
import RoomMessages from "@/components/room/RoomMessages.vue"
import RoomOnlineVideo from "@/components/room/RoomOnlineVideo.vue"
import ModalDialog from "@/components/modals/ModalDialog.vue"
import CopyField from "@/components/fields/CopyField.vue"

type Modal = "none" | "error" | "confirm_join" | "confirm_leave"
type RecordingStatus = "none" | "pending" | "recording" | "stopped"

enum SidePanel {
  None = "",
  Chat = "chat",
}

interface Resizer {
  resize(): void
}

const emit = defineEmits<{
  (e: "join", confirmed: boolean): void
}>()

const sidePanelKey = "roomSidePanel"

const user = userStore.state()

const props = defineProps<{
  talk: Talk
  inviteLink?: string
  joinConfirmed?: boolean
}>()

const modal = ref<Modal>("none")
const sidePanel = ref(localStorage.getItem(sidePanelKey) || SidePanel.None)
const audience = ref<Resizer>()
const room = new LiveRoom()
const roomJoined = room.isJoined()
const roomPublishing = room.isPublishing()
const roomRecording = room.isRecording()
const local = room.localStreams()
const remote = room.remoteStreams()
const recordingStatus = ref<RecordingStatus>("none")
const roomId = computed<string>(() => {
  return props.talk.roomId
})

let modalClosed: (button: string) => void = () => {} // eslint-disable-line @typescript-eslint/no-empty-function

watch(
  roomId,
  async (roomId: string) => {
    room.close()
    await room.join(roomId)
  },
  { immediate: true },
)

watch(roomRecording, (isRecording: boolean) => {
  recordingStatus.value = isRecording ? "recording" : "stopped"
})
watch(
  () => props.talk.state,
  (state?: TalkState) => {
    if (!state) {
      return
    }

    switch (state) {
      case TalkState.RECORDING:
        recordingStatus.value = "recording"
        break
      case TalkState.ENDED:
        recordingStatus.value = "stopped"
        break
    }
  },
  { immediate: true },
)

onMounted(() => {
  if (!props.joinConfirmed) {
    modal.value = "confirm_join"
  }
})

onUnmounted(() => {
  room.close()
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
    modal.value = "confirm_leave"
  })
  next(btn === "leave")
})

function onModalClose(button: string): void {
  modalClosed(button)
  modal.value = "none"
}

function confirmJoin(value: string) {
  emit("join", value === "join")
  modal.value = "none"
}

function sendMessage(message: string) {
  room.send(user.id, message)
}

function switchSidePanel(panel: SidePanel) {
  if (sidePanel.value === panel) {
    panel = SidePanel.None
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
    case "none":
      recordingStatus.value = "pending"
      try {
        await talkClient.startRecording({ id: props.talk.id })
      } catch (e) {
        modal.value = "error"
      }
      break
    case "recording":
      recordingStatus.value = "pending"
      try {
        await talkClient.stopRecording({ id: props.talk.id })
      } catch (e) {
        modal.value = "error"
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
  max-width: min(90%, 800px)
  width: 100%

.video-container
  overflow: hidden
  position: relative

.video-content
  display: flex
  flex-direction: row
  justify-content: center

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

.controls-bottom
  margin-top: auto

.ctrl-btn
  border-radius: 50%
  margin: 11px
  padding: 0.6em
  &.active
    margin: 10px
    border: 1px solid var(--color-highlight-background)

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

<template>
  <div class="content">
    <div class="room">
      <div class="video-content">
        <div class="videos">
          <div class="screen video-container">
            <RoomLiveVideo v-if="screen" class="video screen-video" :src="screen"> </RoomLiveVideo>
            <div v-else class="video-off">
              <div class="video-off-icon material-icons">desktop_access_disabled</div>
            </div>
          </div>
          <div class="camera video-container">
            <video
              v-if="camera"
              class="video camera-video"
              :class="{ local: room.state.local.camera }"
              :srcObject="camera"
              autoplay
              muted
            ></video>
            <div v-else class="video-off">
              <div class="video-off-icon material-icons">videocam_off</div>
            </div>
          </div>
        </div>
      </div>

      <RoomAudience ref="audience" :loading="room.state.isLoading" :peers="room.state.peers" />
    </div>
    <div class="controls">
      <div class="controls-top">
        <div
          class="ctrl-btn btn-switch material-icons"
          :class="{ active: room.state.local.screen }"
          :disabled="room.state.isLoading || room.state.isPublishing || !room.state.joinedMedia ? true : null"
          @click="room.switchScreen"
        >
          {{ room.state.local.screen ? "desktop_windows" : "desktop_access_disabled" }}
        </div>
        <div
          class="ctrl-btn btn-switch material-icons"
          :class="{ active: room.state.local.camera }"
          :disabled="room.state.isLoading || room.state.isPublishing || !room.state.joinedMedia ? true : null"
          @click="room.switchCamera"
        >
          {{ room.state.local.camera ? "videocam" : "videocam_off" }}
        </div>
        <div
          class="ctrl-btn btn-switch material-icons"
          :class="{ active: room.state.local.mic }"
          :disabled="room.state.isLoading || room.state.isPublishing || !room.state.joinedMedia ? true : null"
          @click="room.switchMic"
        >
          {{ room.state.local.mic ? "mic" : "mic_off" }}
        </div>
        <div
          v-if="recordingStatus !== 'stopped'"
          class="ctrl-btn btn-switch material-icons record-icon"
          :disabled="recordingStatus === 'pending' ? true : null"
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
      <RoomMessages
        :user-id="user.id"
        :messages="room.state.messages"
        :is-loading="room.state.isLoading"
        @message="sendMessage"
      />
    </div>
  </div>

  <div v-if="joinConfirmed">
    <audio v-for="stream in audios" :key="stream.id" :srcObject="stream" autoplay></audio>
  </div>

  <ModalDialog
    v-if="modal.state.current === 'confirm_join'"
    :ctrl="modal"
    :buttons="{ join: 'Join', leave: 'Leave' }"
    @click="confirmJoin"
  >
    <p>You are about to join a live talk</p>
    <p v-if="inviteLink">
      Share this link to invite people<br />
      <CopyField :value="inviteLink"></CopyField>
    </p>
  </ModalDialog>
  <ModalDialog v-if="modal.state.current === 'confirm_leave'" :ctrl="modal" :buttons="{ leave: 'Leave', stay: 'Stay' }">
    <p>You are about to leave the talk while presenting.</p>
    <p>If you leave, your presentation will end.</p>
  </ModalDialog>
  <ModalDialog
    v-if="modal.state.current === 'reconnect'"
    :ctrl="modal"
    :buttons="{ reconnect: 'Reconnect', leave: 'Leave' }"
  >
    <p>Connection lost.</p>
    <p>Please check your internet connection and try to reconnect.</p>
  </ModalDialog>
  <ModalDialog
    v-if="modal.state.current === 'recording_finished'"
    :ctrl="modal"
    :buttons="{ leave: 'Go to recording', stay: 'Stay' }"
  >
    <p>Recording finished.</p>
    <p>For demo purposes it is limited to 5 minutes.</p>
  </ModalDialog>
  <InternalError v-if="modal.state.current === 'error'" :ctrl="modal" />
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onUnmounted } from "vue"
import { onBeforeRouteLeave, useRouter } from "vue-router"
import { route } from "@/router"
import { talkClient } from "@/api"
import { Talk, TalkState, userStore } from "@/api/models"
import { LiveRoom } from "@/components/room"
import { Hint } from "@/api/room/schema"
import { ModalController } from "@/components/modals/controller"
import InternalError from "@/components/modals/InternalError.vue"
import RoomAudience from "@/components/room/RoomAudience.vue"
import RoomMessages from "@/components/room/RoomMessages.vue"
import RoomLiveVideo from "@/components/room/RoomLiveVideo.vue"
import ModalDialog from "@/components/modals/ModalDialog.vue"
import CopyField from "@/components/fields/CopyField.vue"

type RecordingStatus = "none" | "pending" | "recording" | "stopped"

const modal = new ModalController<"error" | "confirm_join" | "confirm_leave" | "reconnect" | "recording_finished">()

enum SidePanel {
  None = "",
  Chat = "chat",
}

interface Resizer {
  resize(): void
}

const emit = defineEmits<{
  (e: "join", confirmed: boolean): void
  (e: "talk_ended"): void
}>()

const sidePanelKey = "roomSidePanel"

const router = useRouter()
const user = userStore.state()

const props = defineProps<{
  talk: Talk
  confaHandle: string
  inviteLink?: string
  joinConfirmed?: boolean
}>()

const sidePanel = ref(localStorage.getItem(sidePanelKey) || SidePanel.None)
const audience = ref<Resizer>()
const room = new LiveRoom()
const recordingStatus = ref<RecordingStatus>("none")
const roomId = computed<string>(() => {
  return props.talk.roomId
})

const screen = computed<MediaStream | undefined>(() => {
  if (room.state.local.screen) {
    return room.state.local.screen
  }

  for (const stream of room.state.remote.values()) {
    if (stream.hint === Hint.Screen) {
      return stream.sourse
    }
  }
  return undefined
})
const camera = computed<MediaStream | undefined>(() => {
  if (room.state.local.camera) {
    return room.state.local.camera
  }

  for (const stream of room.state.remote.values()) {
    if (stream.hint === Hint.Camera) {
      return stream.sourse
    }
  }
  return undefined
})
const audios = computed<MediaStream[]>(() => {
  const auds: MediaStream[] = []
  for (const stream of room.state.remote.values()) {
    if (stream.hint === Hint.UserAudio) {
      auds.push(stream.sourse)
    }
  }
  return auds
})

watch(
  roomId,
  async (roomId: string) => {
    room.close()
    await room.joinRTC(roomId)
    if (!props.joinConfirmed) {
      await modal.set("confirm_join")
    }
    await room.joinMedia()
  },
  { immediate: true },
)

watch(room.state.recording, async (r) => {
  if (r.isRecording) {
    recordingStatus.value = "recording"
  } else {
    recordingStatus.value = "stopped"
    const answer = await modal.set("recording_finished")
    if (answer === "leave") {
      emit("talk_ended")
    }
  }
})
watch(
  () => room.state.error,
  async (err) => {
    if (!err) {
      return
    }

    const answer = await modal.set("reconnect")
    if (answer === "reconnect") {
      await room.joinRTC(roomId.value)
      await room.joinMedia()
    } else {
      router.push(route.talk(props.confaHandle, props.talk.handle, "overview"))
    }
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
        recordingStatus.value = "recording"
        break
      case TalkState.ENDED:
        recordingStatus.value = "stopped"
        break
    }
  },
  { immediate: true },
)

onUnmounted(() => {
  room.close()
})

onBeforeRouteLeave(async (to, from, next) => {
  if (!room.state.local.screen && !room.state.local.camera && !room.state.local.mic) {
    next()
    return
  }
  const btn = await modal.set("confirm_leave")
  next(btn === "leave")
})

function confirmJoin(value: string) {
  emit("join", value === "join")
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
        modal.set("error")
      }
      break
    case "recording":
      recordingStatus.value = "pending"
      try {
        await talkClient.stopRecording({ id: props.talk.id })
      } catch (e) {
        modal.set("error")
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

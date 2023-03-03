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
            <div v-if="recordingStatus === 'recording'" class="rec-indicator"></div>
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
      <RoomAudience
        ref="audience"
        :user-id="accessStore.state.id"
        :is-loading="!isRoomReady"
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
          :class="{ active: room.state.local.screen }"
          :disabled="!isMediaReady ? true : null"
          @click="room.switchScreen"
        >
          {{ room.state.local.screen ? "desktop_windows" : "desktop_access_disabled" }}
        </div>
        <div
          v-if="accessStore.state.id === talk.ownerId"
          class="ctrl-btn btn-switch material-icons"
          :class="{ active: room.state.local.camera }"
          :disabled="!isMediaReady ? true : null"
          @click="room.switchCamera"
        >
          {{ room.state.local.camera ? "videocam" : "videocam_off" }}
        </div>
        <div
          class="ctrl-btn btn-switch material-icons"
          :class="{ active: room.state.local.mic }"
          :disabled="!isMediaReady ? true : null"
          @click="room.switchMic"
        >
          {{ room.state.local.mic ? "mic" : "mic_off" }}
        </div>
        <div
          v-if="recordingStatus !== 'stopped' && accessStore.state.id === talk.ownerId"
          class="ctrl-btn btn-switch material-icons record-icon"
          :disabled="recordingStatus === 'pending' ? true : null"
          @click="handleRecording"
        >
          {{ recordingStatus !== "recording" ? "radio_button_checked" : "stop_circle" }}
        </div>
      </div>
      <div class="controls-bottom">
        <div v-if="sidePanel !== 'none'" class="ctrl-btn btn-switch material-icons" @click="switchSidePanel('none')">
          close
        </div>
        <div
          class="ctrl-btn btn-switch material-icons"
          :class="{ pressed: sidePanel === 'chat' }"
          @click="switchSidePanel('chat')"
        >
          chat
          <div
            v-if="room.state.messages.length && lastReadMessageId !== room.state.messages.at(-1)?.id"
            class="new-content-marker"
          ></div>
        </div>
      </div>
    </div>
    <div v-if="sidePanel !== 'none'" class="side-panel">
      <RoomMessages
        :user-id="accessStore.state.id"
        :messages="room.state.messages"
        :is-loading="!isRoomReady"
        @sent="sendMessage"
      />
    </div>
  </div>

  <div v-if="joinConfirmed">
    <audio v-for="stream in audios" :key="stream.id" :srcObject="stream" autoplay></audio>
  </div>

  <ModalDialog
    :is-visible="modal.state === 'confirm_join'"
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
  <ModalDialog :is-visible="modal.state === 'confirm_leave'" :ctrl="modal" :buttons="{ leave: 'Leave', stay: 'Stay' }">
    <p>You are about to leave the talk while presenting.</p>
    <p>If you leave, your presentation will end.</p>
  </ModalDialog>
  <ModalDialog
    :is-visible="modal.state === 'recording_finished'"
    :ctrl="modal"
    :buttons="{ leave: 'Go to recording', stay: 'Stay' }"
  >
    <p>Recording finished.</p>
    <p>For demo purposes it is limited to 5 minutes.</p>
  </ModalDialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onUnmounted } from "vue"
import { onBeforeRouteLeave } from "vue-router"
import { talkClient } from "@/api"
import { Talk, TalkState } from "@/api/models/talk"
import { accessStore } from "@/api/models/access"
import { LiveRoom } from "@/components/room"
import { Hint, Reaction } from "@/api/room/schema"
import { ModalController } from "@/components/modals/controller"
import RoomAudience from "@/components/room/RoomAudience.vue"
import RoomMessages from "@/components/room/RoomMessages.vue"
import RoomLiveVideo from "@/components/room/RoomLiveVideo.vue"
import RoomReactions from "@/components/room/RoomReactions.vue"
import ModalDialog from "@/components/modals/ModalDialog.vue"
import CopyField from "@/components/fields/CopyField.vue"
import { Backoff } from "@/platform/sync"
import { notificationStore } from "@/api/models/notifications"

type RecordingStatus = "none" | "pending" | "recording" | "stopped"

const modal = new ModalController<"confirm_join" | "confirm_leave" | "recording_finished">()

type SidePanel = "none" | "chat"

interface Resizer {
  resize(): void
}

const emit = defineEmits<{
  (e: "join", confirmed: boolean): void
  (e: "update", talk: Talk): void
}>()

const props = defineProps<{
  talk: Talk
  confaHandle: string
  inviteLink?: string
  joinConfirmed?: boolean
}>()

const sidePanelKey = "roomSidePanel"
const sidePanel = ref<SidePanel>((localStorage.getItem(sidePanelKey) as SidePanel) || "none")
const audience = ref<Resizer>()
const room = new LiveRoom()
const recordingStatus = ref<RecordingStatus>("none")
const roomId = computed<string>(() => {
  return props.talk.roomId
})
const lastReadMessageId = ref<string>("")
const connectBackoff = new Backoff(1.5, 1000, 60 * 1000, 0.2)
let connectTimeoutId: ReturnType<typeof setTimeout> = 0

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
const isRoomReady = computed<boolean>(() => {
  return !room.state.isLoading && !room.state.error
})
const isMediaReady = computed<boolean>(() => {
  return !room.state.isLoading && !room.state.error && !room.state.isPublishing && room.state.joinedMedia
})

watch(roomId, async () => connect(), { immediate: true })

watch(room.state.recording, async (r) => {
  if (r.isRecording) {
    recordingStatus.value = "recording"
  } else {
    recordingStatus.value = "stopped"
    const answer = await modal.set("recording_finished")
    if (answer === "leave") {
      const talk = Object.assign({}, props.talk)
      talk.state = TalkState.ENDED
      emit("update", talk)
    }
  }
})
watch(
  () => room.state.error,
  async (err) => {
    if (!err) {
      return
    }
    connect()
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
watch([room.state.messages, sidePanel], () => {
  if (sidePanel.value === "chat") {
    lastReadMessageId.value = room.state.messages.at(-1)?.id || ""
  }
})

onUnmounted(() => {
  clearTimeout(connectTimeoutId)
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

async function connect(): Promise<void> {
  clearTimeout(connectTimeoutId)
  try {
    room.close()
    await room.joinRTC(roomId.value)
    if (!props.joinConfirmed) {
      await modal.set("confirm_join")
    }
    await room.joinMedia()
    if (connectBackoff.retries > 0) {
      notificationStore.info("connection restored")
      connectBackoff.reset()
    }
  } catch (e) {
    if (connectBackoff.retries === 0) {
      notificationStore.error("connection lost")
    }
    connectTimeoutId = setTimeout(() => connect(), connectBackoff.next())
  }
}

function confirmJoin(value: string) {
  emit("join", value === "join")
}

function sendMessage(message: string) {
  room.send(accessStore.state.id, message)
}

function sendReaction(reaction: Reaction) {
  room.reaction(reaction)
}

function switchSidePanel(panel: SidePanel) {
  if (sidePanel.value === panel) {
    panel = "none"
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
        notificationStore.error("failed to start recording")
      }
      break
    case "recording":
      recordingStatus.value = "pending"
      try {
        await talkClient.stopRecording({ id: props.talk.id })
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
  &.active
    margin: 10px
    border: 1px solid var(--color-highlight-background)

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

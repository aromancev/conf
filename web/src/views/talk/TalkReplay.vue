<template>
  <div v-if="!state.isReady" class="loading-content">
    <PageLoader />
    <div v-if="!state.isLoading" class="processing-note">
      <p>Processing talk recording.</p>
      <p>It might take a while (especially screen sharing) becuase it runs on a very cheap server.</p>
      <p v-if="user.id === talk.ownerId">You will receive an email when it's done.</p>
    </div>
  </div>
  <div v-if="state.isReady" class="content">
    <div class="room">
      <div class="video-content">
        <div class="videos">
          <div class="screen video-container">
            <RoomReplayVideo
              :media="screen"
              :duration="room.state.duration"
              :progress="room.state.progress"
              :buffer="room.state.buffer"
              :is-playing="room.state.isPlaying"
              :is-buffering="room.state.isBuffering"
              class="video screen-video"
              @toggle-play="() => room.togglePlay()"
              @rewind="(pos) => room.rewind(pos)"
              @buffer="(bufferMs, durationMs) => room.updateMediaBuffer(screen?.id || '', bufferMs, durationMs)"
            >
            </RoomReplayVideo>
          </div>
          <div class="camera video-container">
            <RoomReplayVideo
              v-if="camera"
              :media="camera"
              :duration="room.state.duration"
              :progress="room.state.progress"
              :buffer="room.state.buffer"
              :is-playing="room.state.isPlaying"
              :is-buffering="room.state.isBuffering"
              :disable-controlls="true"
              :fit="'cover'"
              :hide-loader="true"
              class="video camera-video"
            >
            </RoomReplayVideo>
            <div v-else class="video-off">
              <div class="video-off-icon material-icons">videocam_off</div>
            </div>
          </div>
        </div>
      </div>

      <RoomAudience ref="audience" :is-loading="room.state.isLoading" :peers="room.state.peers" />
    </div>
    <div class="controls">
      <div class="controls-bottom">
        <div
          v-if="state.sidePanel !== 'none'"
          class="ctrl-btn btn-switch material-icons"
          @click="switchSidePanel('none')"
        >
          close
        </div>
        <div
          class="ctrl-btn btn-switch material-icons"
          :class="{ pressed: state.sidePanel === 'chat' }"
          @click="switchSidePanel('chat')"
        >
          chat
        </div>
      </div>
    </div>
    <div v-if="state.sidePanel !== 'none'" class="side-panel">
      <RoomMessages
        :user-id="user.id"
        :messages="room.state.messages"
        :is-loading="room.state.isLoading"
        :is-readonly="true"
      />
    </div>
    <RoomReplayAudio
      v-for="source in audios"
      :key="source.manifestUrl"
      :media="source"
      :duration="room.state.duration"
      :progress="room.state.progress"
      :is-playing="room.state.isPlaying"
      :is-buffering="room.state.isBuffering"
      @buffer="(bufferMs, durationMs) => room.updateMediaBuffer(source?.id || '', bufferMs, durationMs)"
    ></RoomReplayAudio>
  </div>
  <InternalError v-if="modal === 'error'" @click="modal = 'none'" />
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onUnmounted, reactive } from "vue"
import { recordingClient } from "@/api"
import { RecordingStatus, Talk, userStore } from "@/api/models"
import { ReplayRoom } from "@/components/room"
import InternalError from "@/components/modals/InternalError.vue"
import RoomAudience from "@/components/room/RoomAudience.vue"
import RoomMessages from "@/components/room/RoomMessages.vue"
import RoomReplayVideo from "@/components/room/RoomReplayVideo.vue"
import RoomReplayAudio from "@/components/room/RoomReplayAudio.vue"
import PageLoader from "@/components/PageLoader.vue"
import { Media } from "@/components/room/aggregators/media"
import { Hint } from "@/api/room/schema"

const READY_CHECK_INTERVAL = 10 * 1000

type Modal = "none" | "error"

type SidePanel = "none" | "chat"

interface Resizer {
  resize(): void
}

const sidePanelKey = "roomSidePanel"

const user = userStore.state()

const props = defineProps<{
  talk: Talk
}>()

interface State {
  sidePanel: SidePanel
  isReady: boolean
  isLoading: boolean
}

const state = reactive<State>({
  sidePanel: (localStorage.getItem(sidePanelKey) as SidePanel) || "none",
  isReady: false,
  isLoading: true,
})

let loadTimerId: ReturnType<typeof setTimeout> = -1
const modal = ref<Modal>("none")
const audience = ref<Resizer>()
const room = new ReplayRoom()
const roomId = computed<string>(() => {
  return props.talk.roomId
})
const screen = computed<Media | undefined>(() => {
  for (const media of room.state.medias.values()) {
    if (media.hint === Hint.Screen) {
      return media
    }
  }
  return undefined
})
const camera = computed<Media | undefined>(() => {
  for (const media of room.state.medias.values()) {
    if (media.hint === Hint.Camera) {
      return media
    }
  }
  return undefined
})
const audios = computed<Media[]>(() => {
  const auds: Media[] = []
  for (const media of room.state.medias.values()) {
    if (media.hint === Hint.UserAudio) {
      auds.push(media)
    }
  }
  return auds
})

watch(
  roomId,
  () => {
    loadRoom()
  },
  { immediate: true },
)

function switchSidePanel(panel: SidePanel) {
  if (state.sidePanel === panel) {
    state.sidePanel = "none"
  } else {
    state.sidePanel = panel
  }
  localStorage.setItem(sidePanelKey, panel)

  nextTick(() => {
    audience.value?.resize()
  })
}

onUnmounted(() => {
  room.close()
})

async function loadRoom(): Promise<void> {
  clearTimeout(loadTimerId)
  state.isReady = false
  try {
    const recording = await recordingClient.fetchOne(
      { roomId: roomId.value, key: props.talk.id },
      { policy: "no-cache" },
    )
    state.isLoading = false
    if (recording.status != RecordingStatus.READY) {
      // If recording isn't finished yet, try again later.
      loadTimerId = setTimeout(() => loadRoom(), READY_CHECK_INTERVAL)
      return
    }
    state.isReady = true
    await room.load(roomId.value, recording)
  } catch (e) {
    modal.value = "error"
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

.loading-content
  width: 100%
  height: 100%

  display: flex
  flex-direction: column
  justify-content: center
  align-items: center
  text-align: center
  padding: 30px

.processing-note
  margin: 30px

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

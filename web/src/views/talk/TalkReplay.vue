<template>
  <div v-if="!state.isReady" class="loading-content">
    <PageLoader />
    <div v-if="!state.isLoading" class="processing-note">
      <p>Processing talk recording.</p>
      <p>It might take a while (especially screen sharing) becuase it runs on a very cheap server.</p>
      <p v-if="accessStore.state.id === talk.ownerId">You will receive an email when it's done.</p>
    </div>
  </div>
  <div v-if="state.isReady" class="content">
    <div class="room">
      <div class="video-content">
        <div class="videos">
          <div class="screen video-container">
            <RoomReplayVideo
              :record="screen"
              :duration="room.state.duration"
              :progress="room.state.progress"
              :buffer="room.state.buffer"
              :is-playing="room.state.isPlaying"
              :is-buffering="room.state.isBuffering"
              class="video screen-video"
              @toggle-play="() => room.togglePlay()"
              @rewind="(pos) => room.rewind(pos)"
              @buffer="(bufferMs, durationMs) => room.updateMediaBuffer(screen?.recordId || '', bufferMs, durationMs)"
            >
            </RoomReplayVideo>
          </div>
          <div class="camera video-container">
            <RoomReplayVideo
              v-if="camera"
              :record="camera"
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

      <RoomAudience
        ref="audience"
        :user-id="accessStore.state.id"
        :is-loading="room.state.isLoading"
        :is-playing="room.state.isPlaying"
        :peers="room.state.peers"
        :statuses="room.state.statuses"
        :self-reactions="true"
      />
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
        :user-id="accessStore.state.id"
        :messages="room.state.messages"
        :is-loading="room.state.isLoading"
        :is-readonly="true"
      />
    </div>
    <RoomReplayAudio
      v-for="source in audios"
      :key="source.manifestUrl"
      :record="source"
      :duration="room.state.duration"
      :progress="room.state.progress"
      :is-playing="room.state.isPlaying"
      :is-buffering="room.state.isBuffering"
      @buffer="(bufferMs, durationMs) => room.updateMediaBuffer(source?.recordId || '', bufferMs, durationMs)"
    ></RoomReplayAudio>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onUnmounted, reactive } from "vue"
import { api } from "@/api"
import { RecordingClient } from "@/api/recording"
import { Talk } from "@/api/models/talk"
import { RecordingStatus } from "@/api/models/recording"
import { accessStore } from "@/api/models/access"
import { ReplayRoom } from "@/components/room"
import RoomAudience from "@/components/room/RoomAudience.vue"
import RoomMessages from "@/components/room/RoomMessages.vue"
import RoomReplayVideo from "@/components/room/RoomReplayVideo.vue"
import RoomReplayAudio from "@/components/room/RoomReplayAudio.vue"
import PageLoader from "@/components/PageLoader.vue"
import { TrackRecord } from "@/components/room/aggregators/record"
import { TrackSource } from "@/api/rtc/schema"
import { notificationStore } from "@/api/models/notifications"
import { Backoff } from "@/platform/sync"

type SidePanel = "none" | "chat"

interface Resizer {
  resize(): void
}

const sidePanelKey = "roomSidePanel"

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

let loadTimerId = -1
const audience = ref<Resizer>()
const room = new ReplayRoom()
const roomId = computed<string>(() => {
  return props.talk.roomId
})
const screen = computed<TrackRecord | undefined>(() => {
  for (const media of room.state.medias.values()) {
    if (media.source === TrackSource.Screen) {
      return media
    }
  }
  return undefined
})
const camera = computed<TrackRecord | undefined>(() => {
  for (const media of room.state.medias.values()) {
    if (media.source === TrackSource.Camera) {
      return media
    }
  }
  return undefined
})
const audios = computed<TrackRecord[]>(() => {
  const auds: TrackRecord[] = []
  for (const media of room.state.medias.values()) {
    if (media.source === TrackSource.Microphone) {
      auds.push(media)
    }
  }
  return auds
})
const loadBackoff = new Backoff(1.2, 3000, 3 * 60 * 1000)

watch(
  [roomId, () => accessStore.state.id],
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
    const recording = await new RecordingClient(api).fetchOne(
      { roomId: roomId.value, key: props.talk.id },
      { policy: "no-cache" },
    )
    state.isLoading = false
    if (recording.status != RecordingStatus.READY) {
      // If recording isn't finished yet, try again later.
      loadTimerId = window.setTimeout(() => loadRoom(), loadBackoff.next())
      return
    }
    state.isReady = true
    loadBackoff.reset()
    await room.load(roomId.value, recording)
  } catch (e) {
    notificationStore.error("failed to load recording")
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

import { LocalStream, Constraints } from "ion-sdk-js"
import { reactive, readonly } from "vue"
import { api } from "@/api"
import { RTCPeer } from "@/api/room"
import { BufferedAggregator } from "./aggregators/buffered"
import { ProfileRepository } from "./profiles"
import { MessageAggregator, Message } from "./aggregators/messages"
import { Peer, Status, PeerAggregator } from "./aggregators/peers"
import { Stream, StreamAggregator } from "./aggregators/streams"
import { RecordingAggregator } from "./aggregators/recording"
import { RoomEvent, Hint, Track, Reaction } from "@/api/room/schema"
import { FIFOMap } from "@/platform/cache"
import { EventClient } from "@/api/event"

const UPDATE_TIME_INTERVAL_MS = 300
const PROFILE_CACHE_SIZE = 500
const MAX_PEERS = 3000
const MAX_STREAMS = 10
const MAX_TRACKS = 10
const LOAD_EVENTS = 3000

interface State {
  peers: Map<string, Peer>
  statuses: Map<string, Status>
  messages: Message[]
  isPublishing: boolean
  isLoading: boolean
  joinedMedia: boolean
  recording: {
    isRecording: boolean
  }
  remote: Map<string, Stream>
  local: {
    camera?: LocalStream
    screen?: LocalStream
    mic?: LocalStream
  }
  error?: Error
}

type Error = "UNKNOWN"

export class LiveRoom {
  private rtc: RTCPeer
  private readonly reactive: State
  private readonly readonly: State
  private peerState: PeerState
  private profileRepo: ProfileRepository
  private messageAggregator?: MessageAggregator
  private setTimeIntervalId: ReturnType<typeof setInterval>

  constructor() {
    this.reactive = reactive<State>({
      peers: new FIFOMap(MAX_PEERS),
      statuses: new FIFOMap(MAX_PEERS),
      remote: new FIFOMap(MAX_STREAMS),
      messages: [],
      isLoading: false,
      isPublishing: false,
      joinedMedia: false,
      recording: {
        isRecording: false,
      },
      local: {},
    }) as State
    this.readonly = readonly(this.reactive) as State
    this.profileRepo = new ProfileRepository(PROFILE_CACHE_SIZE, 3000)
    this.peerState = { tracks: new FIFOMap(MAX_TRACKS) }
    this.rtc = new RTCPeer()
    this.rtc.onerror = () => {
      this.close()
      this.reactive.error = "UNKNOWN"
    }
    this.setTimeIntervalId = 0
  }

  get state(): State {
    return this.readonly
  }

  async close(): Promise<void> {
    clearInterval(this.setTimeIntervalId)
    this.unshareScreen()
    this.unshareCamera()
    this.unshareMic()
    await this.rtc.close()
    this.reactive.peers.clear()
    this.reactive.statuses.clear()
    this.reactive.remote.clear()
    this.reactive.messages.slice(0, this.reactive.messages.length)
    this.reactive.isLoading = false
    this.reactive.joinedMedia = false
    this.reactive.recording.isRecording = false
    this.peerState = { tracks: new Map() }
  }

  async joinRTC(roomId: string): Promise<void> {
    await this.close()
    this.reactive.isLoading = true

    try {
      const streams = new StreamAggregator(this.reactive.remote)
      const peers = new PeerAggregator(this.reactive.peers, this.reactive.statuses, this.profileRepo)
      const recording = new RecordingAggregator(this.reactive.recording)
      this.messageAggregator = new MessageAggregator(this.reactive.messages, this.profileRepo)
      const aggregators = new BufferedAggregator([peers, streams, recording, this.messageAggregator], 50)

      let serverNow = 0
      let serverNowAt = 0
      this.rtc.onevent = (event: RoomEvent): void => {
        if (event.createdAt > serverNow) {
          serverNow = event.createdAt
          serverNowAt = Date.now()
        }
        aggregators.put(event)
      }
      this.rtc.ontrack = (t, s) => streams.addTrack(t, s)
      await this.rtc.joinRTC(roomId)

      const iter = new EventClient(api).fetch({ roomId: roomId }, { policy: "network-only", cursor: { Asc: true } })
      const events = await iter.next(LOAD_EVENTS)

      aggregators.prepend(...events)
      aggregators.flush()

      // Update aggregators about current time on the server which is taken from received events.
      this.setTimeIntervalId = setInterval(() => {
        if (!serverNow) {
          return
        }
        const elapsed = Date.now() - serverNowAt
        aggregators.setTime(serverNow + elapsed)
      }, UPDATE_TIME_INTERVAL_MS)

      this.reactive.error = undefined
    } finally {
      this.reactive.isLoading = false
    }
  }

  async joinMedia(): Promise<void> {
    this.reactive.joinedMedia = false
    await this.rtc.joinMedia()
    this.reactive.joinedMedia = true
  }

  async send(userId: string, message: string): Promise<void> {
    if (!this.rtc || !this.messageAggregator) {
      throw new Error("Must join room before sending a message.")
    }

    const setId = this.messageAggregator.addMessage(userId, message)
    const event = await this.rtc.message({ text: message })
    setId(event.id)
  }

  async reaction(reaction: Reaction): Promise<void> {
    if (!this.rtc) {
      throw new Error("Must join room before sending a reaction.")
    }
    await this.rtc.reaction(reaction)
  }

  switchCamera() {
    if (this.reactive.local.camera) {
      this.unshareCamera()
    } else {
      this.shareCamera()
    }
  }

  switchScreen() {
    if (this.reactive.local.screen) {
      this.unshareScreen()
    } else {
      this.shareScreen()
    }
  }

  switchMic() {
    if (this.reactive.local.mic) {
      this.unshareMic()
    } else {
      this.shareMic()
    }
  }

  async shareCamera() {
    if (!this.reactive.joinedMedia) {
      return
    }
    this.reactive.local.camera = await this.share(async () => {
      return await LocalStream.getUserMedia({
        codec: "vp8",
        resolution: "vga",
        simulcast: false,
        video: true,
        audio: false,
      })
    }, Hint.Camera)
  }

  unshareCamera() {
    this.unshare(this.reactive.local.camera)
    this.reactive.local.camera = undefined
  }

  async shareScreen() {
    if (!this.reactive.joinedMedia) {
      return
    }
    this.reactive.local.screen = await this.share(async () => {
      const stream = await LocalStream.getDisplayMedia({
        codec: "vp8",
        resolution: "hd",
        simulcast: false,
        video: {
          width: { ideal: 1920, max: 1920 },
          height: { ideal: 1080, max: 1080 },
          frameRate: {
            ideal: 30,
            max: 30,
          },
        },
        audio: false,
      })
      for (const t of stream.getTracks()) {
        t.onended = () => {
          this.unshareScreen()
        }
      }
      return stream
    }, Hint.Screen)
  }

  unshareScreen() {
    this.unshare(this.reactive.local.screen)
    this.reactive.local.screen = undefined
  }

  async shareMic() {
    if (!this.reactive.joinedMedia) {
      return
    }
    this.reactive.local.mic = await this.share(() => {
      return LocalStream.getUserMedia({
        video: false,
        audio: true,
      } as Constraints)
    }, Hint.UserAudio)
  }

  unshareMic() {
    this.unshare(this.reactive.local.mic)
    this.reactive.local.mic = undefined
  }

  private async share(fetch: () => Promise<LocalStream>, hint: Hint): Promise<LocalStream | undefined> {
    if (!this.rtc || this.reactive.isLoading || this.reactive.isPublishing) {
      return undefined
    }

    this.reactive.isPublishing = true
    try {
      const stream = await fetch()
      const tId = trackId(stream)
      this.peerState.tracks.set(tId, { id: tId, hint: hint })
      await this.rtc.state({ tracks: Array.from(this.peerState.tracks.values()) })
      this.rtc.publish(stream)
      return stream
    } catch (e) {
      console.warn("Failed to share media:", e)
      return undefined
    } finally {
      this.reactive.isPublishing = false
    }
  }

  private unshare(stream: LocalStream | null | undefined) {
    if (!stream) {
      return
    }
    this.peerState.tracks.delete(trackId(stream))
    stream.unpublish()
    for (const t of stream.getTracks()) {
      t.stop()
      stream.removeTrack(t)
    }
  }
}

interface PeerState {
  tracks: Map<string, Track>
}

function trackId(s: MediaStream): string {
  const tracks = s.getTracks()
  if (tracks.length < 1) {
    return ""
  }
  return tracks[0].id
}

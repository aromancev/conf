import { LocalStream, Constraints } from "ion-sdk-js"
import { reactive, readonly } from "vue"
import { RTCPeer, eventClient } from "@/api"
import { BufferedAggregator } from "./aggregators/buffered"
import { ProfileRepository } from "./profiles"
import { MessageAggregator, Message } from "./aggregators/messages"
import { Peer, Status, PeerAggregator } from "./aggregators/peers"
import { Stream, StreamAggregator } from "./aggregators/streams"
import { RecordingAggregator } from "./aggregators/recording"
import { RoomEvent, Hint, Track, Reaction } from "@/api/room/schema"

const UPDATE_TIME_INTERVAL_MS = 300
const PROFILE_CACHE_SIZE = 500

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

type Error = "CLOSED"

export class LiveRoom {
  private rtc: RTCPeer
  private _state: State
  private readState: State
  private peerState: PeerState
  private profileRepo: ProfileRepository
  private messageAggregator?: MessageAggregator
  private setTimeIntervalId: ReturnType<typeof setInterval>

  constructor() {
    this._state = reactive<State>({
      peers: new Map(),
      statuses: new Map(),
      messages: [],
      isLoading: false,
      isPublishing: false,
      joinedMedia: false,
      recording: {
        isRecording: false,
      },
      remote: new Map(),
      local: {},
    }) as State
    this.readState = readonly(this._state) as State
    this.profileRepo = new ProfileRepository(PROFILE_CACHE_SIZE, 3000)
    this.peerState = { tracks: new Map() }
    this.rtc = new RTCPeer()
    this.rtc.onclose = () => {
      this.close()
      this._state.error = "CLOSED"
    }
    this.setTimeIntervalId = 0
  }

  get state(): State {
    return this.readState
  }

  close(): void {
    this.unshareScreen()
    this.unshareCamera()
    this.unshareMic()
    this.rtc.close()
    this.peerState = { tracks: new Map() }
    this._state.joinedMedia = false
    clearInterval(this.setTimeIntervalId)
  }

  async joinRTC(roomId: string): Promise<void> {
    this.close()
    this._state.isLoading = true

    try {
      const streams = new StreamAggregator()
      const peers = new PeerAggregator(this.profileRepo)
      const recording = new RecordingAggregator()
      this.messageAggregator = new MessageAggregator(this.profileRepo)
      const aggregators = new BufferedAggregator([peers, streams, recording, this.messageAggregator], 50)

      this._state.remote = streams.state().streams
      this._state.peers = peers.state().peers
      this._state.statuses = peers.state().statuses
      this._state.recording = recording.state()
      this._state.messages = this.messageAggregator.state().messages

      let serverNow = 0
      let serverNowAt = 0
      this.rtc.onevent = (event: RoomEvent): void => {
        serverNow = event.createdAt
        serverNowAt = Date.now()
        aggregators.put(event)
      }
      this.rtc.ontrack = (t, s) => streams.addTrack(t, s)
      await this.rtc.joinRTC(roomId)

      const iter = eventClient.fetch({ roomId: roomId }, { policy: "network-only", cursor: { Asc: true } })
      const events = await iter.next(3000)

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

      this._state.error = undefined
    } finally {
      this._state.isLoading = false
    }
  }

  async joinMedia(): Promise<void> {
    this._state.joinedMedia = false
    await this.rtc.joinMedia()
    this._state.joinedMedia = true
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
    if (this._state.local.camera) {
      this.unshareCamera()
    } else {
      this.shareCamera()
    }
  }

  switchScreen() {
    if (this._state.local.screen) {
      this.unshareScreen()
    } else {
      this.shareScreen()
    }
  }

  switchMic() {
    if (this._state.local.mic) {
      this.unshareMic()
    } else {
      this.shareMic()
    }
  }

  async shareCamera() {
    this._state.local.camera = await this.share(async () => {
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
    this.unshare(this._state.local.camera)
    this._state.local.camera = undefined
  }

  async shareScreen() {
    this._state.local.screen = await this.share(async () => {
      const stream = await LocalStream.getDisplayMedia({
        codec: "vp8",
        resolution: "hd",
        simulcast: false,
        video: {
          width: { ideal: 1920 },
          height: { ideal: 1080 },
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
    this.unshare(this._state.local.screen)
    this._state.local.screen = undefined
  }

  async shareMic() {
    this._state.local.mic = await this.share(() => {
      return LocalStream.getUserMedia({
        video: false,
        audio: true,
      } as Constraints)
    }, Hint.UserAudio)
  }

  unshareMic() {
    this.unshare(this._state.local.mic)
    this._state.local.mic = undefined
  }

  private async share(fetch: () => Promise<LocalStream>, hint: Hint): Promise<LocalStream | undefined> {
    if (!this.rtc || this._state.isLoading || this._state.isPublishing) {
      return undefined
    }

    this._state.isPublishing = true
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
      this._state.isPublishing = false
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

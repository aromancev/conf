import { LocalStream, Constraints } from "ion-sdk-js"
import { reactive, readonly } from "vue"
import { RTCPeer, eventClient } from "@/api"
import { BufferedAggregator } from "./aggregators/buffered"
import { ProfileRepository } from "./profiles"
import { MessageAggregator, Message } from "./aggregators/messages"
import { Peer, PeerAggregator } from "./aggregators/peers"
import { Stream, StreamAggregator } from "./aggregators/streams"
import { RecordingAggregator } from "./aggregators/recording"
import { RoomEvent, Hint, Track } from "@/api/room/schema"
import { EventOrder } from "@/api/schema"

interface State {
  peers: Map<string, Peer>
  messages: Message[]
  isPublishing: boolean
  isLoading: boolean
  recording: {
    isRecording: boolean
  }
  remote: Map<string, Stream>
  local: {
    camera?: LocalStream
    screen?: LocalStream
    mic?: LocalStream
  }
}

export class LiveRoom {
  private rtc: RTCPeer
  private _state: State
  private readState: State
  private peerState: PeerState
  private profileRepo: ProfileRepository
  private messageAggregator?: MessageAggregator

  constructor() {
    this._state = reactive<State>({
      peers: new Map(),
      messages: [],
      isLoading: false,
      isPublishing: false,
      recording: {
        isRecording: false,
      },
      remote: new Map(),
      local: {},
    }) as State
    this.readState = readonly(this._state) as State
    this.profileRepo = new ProfileRepository(100, 3000)
    this.rtc = new RTCPeer()
    this.peerState = { tracks: new Map() }
  }

  get state(): State {
    return this.readState
  }

  close(): void {
    this.unshareScreen()
    this.unshareCamera()
    this.unshareMic()
    this.rtc.close()
  }

  async join(roomId: string) {
    this._state.isLoading = true

    try {
      const streams = new StreamAggregator()
      const peers = new PeerAggregator(this.profileRepo)
      const recording = new RecordingAggregator()
      this.messageAggregator = new MessageAggregator(this.profileRepo)
      const aggregators = new BufferedAggregator([peers, streams, recording, this.messageAggregator], 50)

      this._state.remote = streams.state().streams
      this._state.peers = peers.state().peers
      this._state.recording = recording.state()
      this._state.messages = this.messageAggregator.state().messages

      this.rtc.onevent = (event: RoomEvent): void => {
        aggregators.put(event)
      }
      this.rtc.ontrack = (t, s) => streams.addTrack(t, s)
      await this.rtc.join(roomId, true)

      const iter = eventClient.fetch({ roomId: roomId }, { order: EventOrder.DESC, policy: "network-only" })
      const events = await iter.next({ count: 3000, seconds: 2 * 60 * 60 })
      // Sorting events to always be in chronological order.
      events.reverse()
      aggregators.prepend(...events)
      aggregators.flush()
    } finally {
      this._state.isLoading = false
    }
  }

  async send(userId: string, message: string): Promise<void> {
    if (!this.rtc || !this.messageAggregator) {
      throw new Error("Must join room before sending a message.")
    }

    const setId = this.messageAggregator.addMessage(userId, message)
    const event = await this.rtc.message({ text: message })
    setId(event.id)
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

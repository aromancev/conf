import { reactive, readonly } from "vue"
import {
  Room,
  RemoteTrack,
  RoomEvent as SFURoomEvent,
  LocalTrackPublication,
  RemoteTrackPublication,
  Track as LiveKitTrack,
  ConnectionState,
  DisconnectReason,
} from "livekit-client"
import { api } from "@/api"
import { RTCPeer } from "@/api/rtc"
import { BufferedAggregator } from "./aggregators/buffered"
import { ProfileRepository } from "./profiles"
import { MessageAggregator, Message } from "./aggregators/messages"
import { Peer, Status, PeerAggregator } from "./aggregators/peers"
import { RecordingAggregator } from "./aggregators/recording"
import { RoomEvent, Reaction, TrackSource } from "@/api/rtc/schema"
import { FIFOMap } from "@/platform/cache"
import { EventClient } from "@/api/event"
import { RoomClient } from "@/api/room"
import { notificationStore } from "@/api/models/notifications"
import { config } from "@/config"

const UPDATE_TIME_INTERVAL_MS = 300
const PROFILE_CACHE_SIZE = 500
const MAX_PEERS = 3000
const MAX_STREAMS = 10
const LOAD_EVENTS = 3000

export type Track = {
  id: string
  source: TrackSource
}

export type State = {
  peers: Map<string, Peer>
  statuses: Map<string, Status>
  messages: Message[]
  recording: {
    isRecording: boolean
  }
  remoteTracks: Map<string, Track>
  localTracks: Map<string, Track>
  sfu: SFUState
  rtc: RTCState
}

type RTCState = "DISCONNECTED" | "CONNECTING" | "CONNECTED" | "ERROR"
type SFUState = "DISCONNECTED" | "CONNECTING" | "CONNECTED" | "ERROR"

export class LiveRoom {
  private readonly reactive: State
  private readonly readonly: State
  private readonly tracks: Map<string, LiveKitTrack>
  private rtc: RTCPeer
  private sfu: Room
  private profileRepo: ProfileRepository
  private messageAggregator?: MessageAggregator
  private setTimeIntervalId: number

  constructor() {
    this.reactive = reactive<State>({
      peers: new FIFOMap(MAX_PEERS),
      statuses: new FIFOMap(MAX_PEERS),
      remoteTracks: new FIFOMap(MAX_STREAMS),
      localTracks: new FIFOMap(MAX_STREAMS),
      messages: [],
      recording: {
        isRecording: false,
      },
      sfu: "DISCONNECTED",
      rtc: "DISCONNECTED",
    }) as State
    this.readonly = readonly(this.reactive) as State
    this.tracks = new FIFOMap(MAX_STREAMS * 2)
    this.profileRepo = new ProfileRepository(PROFILE_CACHE_SIZE, 3000)
    this.rtc = new RTCPeer()
    this.rtc.onerror = () => {
      this.close()
      this.reactive.rtc = "ERROR"
    }
    this.sfu = new Room()
    this.sfu.on(SFURoomEvent.TrackSubscribed, this.addRemoteTrack.bind(this))
    this.sfu.on(SFURoomEvent.TrackUnsubscribed, this.removeRemoteTrack.bind(this))
    this.sfu.on(SFURoomEvent.LocalTrackPublished, this.addLocalTrack.bind(this))
    this.sfu.on(SFURoomEvent.LocalTrackUnpublished, this.removeLocalTrack.bind(this))
    this.sfu.on(SFURoomEvent.ConnectionStateChanged, (state: ConnectionState) => {
      switch (state) {
        case ConnectionState.Connecting:
        case ConnectionState.Reconnecting:
          this.reactive.sfu = "CONNECTING"
          break
        case ConnectionState.Connected:
          this.reactive.sfu = "CONNECTED"
      }
    })
    this.sfu.on(SFURoomEvent.Disconnected, (reason?: DisconnectReason) => {
      switch (reason) {
        case DisconnectReason.CLIENT_INITIATED:
          break
        case DisconnectReason.DUPLICATE_IDENTITY:
        case DisconnectReason.PARTICIPANT_REMOVED:
          this.reactive.sfu = "DISCONNECTED"
          break
        default:
          this.reactive.sfu = "ERROR"
          break
      }
    })
    this.setTimeIntervalId = 0
  }

  get state(): State {
    return this.readonly
  }

  async close(): Promise<void> {
    await Promise.all([this.disconnectRTC(), this.disconnectSFU()])
  }

  async connectRTC(roomId: string): Promise<void> {
    if (this.reactive.rtc === "CONNECTED") {
      return
    }

    this.reactive.rtc = "CONNECTING"

    try {
      await this.disconnectRTC()

      const peers = new PeerAggregator(this.reactive.peers, this.reactive.statuses, this.profileRepo)
      const recording = new RecordingAggregator(this.reactive.recording)
      this.messageAggregator = new MessageAggregator(this.reactive.messages, this.profileRepo)
      const aggregators = new BufferedAggregator([peers, recording, this.messageAggregator], 50)

      let serverNow = 0
      let serverNowAt = 0
      this.rtc.onevent = (event: RoomEvent): void => {
        if (event.createdAt > serverNow) {
          serverNow = event.createdAt
          serverNowAt = Date.now()
        }
        aggregators.put(event)
      }
      await this.rtc.joinRTC(roomId)

      const iter = new EventClient(api).fetch({ roomId: roomId }, { policy: "network-only", cursor: { Asc: true } })
      const events = await iter.next(LOAD_EVENTS)

      aggregators.prepend(...events)
      aggregators.flush()

      // Update aggregators about current time on the server which is taken from received events.
      this.setTimeIntervalId = window.setInterval(() => {
        if (!serverNow) {
          return
        }
        const elapsed = Date.now() - serverNowAt
        aggregators.setTime(serverNow + elapsed)
      }, UPDATE_TIME_INTERVAL_MS)

      this.reactive.rtc = "CONNECTED"
    } catch (e) {
      console.error(e)
      notificationStore.error("real time communication failed")
      this.reactive.rtc = "ERROR"
    }
  }

  async connectSFU(roomId: string): Promise<void> {
    if (this.reactive.sfu === "CONNECTED") {
      return
    }

    this.reactive.sfu = "CONNECTING"

    try {
      await this.disconnectSFU()

      const token = await new RoomClient(api).requestSFUAccess(roomId)
      await this.sfu.connect(config.sfu.url, token)
      this.reactive.sfu = "CONNECTED"
    } catch (e) {
      console.error(e)
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

  async reaction(reaction: Reaction): Promise<void> {
    if (!this.rtc) {
      throw new Error("Must join room before sending a reaction.")
    }
    await this.rtc.reaction(reaction)
  }

  async switchCamera() {
    if (this.sfu.localParticipant.isCameraEnabled) {
      await this.unshareCamera()
    } else {
      await this.shareCamera()
    }
  }

  async switchScreen() {
    if (this.sfu.localParticipant.isScreenShareEnabled) {
      await this.unshareScreen()
    } else {
      await this.shareScreen()
    }
  }

  async switchMic() {
    if (this.sfu.localParticipant.isMicrophoneEnabled) {
      await this.unshareMic()
    } else {
      await this.shareMic()
    }
  }

  async shareCamera() {
    if (this.reactive.sfu !== "CONNECTED") {
      return
    }
    try {
      await this.sfu.localParticipant.setCameraEnabled(
        true,
        {
          resolution: {
            width: 1280,
            height: 720,
            frameRate: 30,
          },
        },
      )
    } catch {
      this.reactive.sfu = "DISCONNECTED"
    }
  }

  async unshareCamera() {
    try {
      const pub = await this.sfu.localParticipant.setCameraEnabled(false)
      if (!pub?.track) {
        return
      }
      this.sfu.localParticipant.unpublishTrack(pub.track)
    } catch {
      this.reactive.sfu = "DISCONNECTED"
    }
  }

  async shareScreen() {
    try {
      if (this.reactive.sfu !== "CONNECTED") {
        return
      }
      await this.sfu.localParticipant.setScreenShareEnabled(
        true,
        {
          resolution: {
            width: 1920,
            height: 1080,
            frameRate: 30,
          },
          audio: false,
        },
      )
    } catch {
      this.reactive.sfu = "DISCONNECTED"
    }
  }

  async unshareScreen() {
    try {
      const pub = await this.sfu.localParticipant.setScreenShareEnabled(false)
      if (!pub?.track) {
        return
      }
      this.sfu.localParticipant.unpublishTrack(pub.track)
    } catch {
      this.reactive.sfu = "DISCONNECTED"
    }
  }

  async shareMic() {
    try {
      if (this.reactive.sfu !== "CONNECTED") {
        return
      }
      await this.sfu.localParticipant.setMicrophoneEnabled(true)
    } catch {
      this.reactive.sfu = "DISCONNECTED"
    }
  }

  async unshareMic() {
    try {
      const pub = await this.sfu.localParticipant.setMicrophoneEnabled(false)
      if (!pub?.track) {
        return
      }
      this.sfu.localParticipant.unpublishTrack(pub.track)
    } catch {
      this.reactive.sfu = "DISCONNECTED"
    }
  }

  attach(trackId: string, el: HTMLMediaElement) {
    const track = this.tracks.get(trackId)
    if (!track) {
      return
    }
    track.attach(el)
  }

  private async disconnectRTC(): Promise<void> {
    clearInterval(this.setTimeIntervalId)
    await this.rtc.close()
    this.reactive.peers.clear()
    this.reactive.statuses.clear()
    this.reactive.messages.slice(0, this.reactive.messages.length)
  }

  private async disconnectSFU(): Promise<void> {
    this.unshareScreen()
    this.unshareCamera()
    this.unshareMic()
    await this.sfu.disconnect()
    this.reactive.remoteTracks.clear()
    this.reactive.localTracks.clear()
    this.tracks.clear()
  }

  private addLocalTrack(publication: LocalTrackPublication) {
    const track = publication.track
    if (!track) {
      return
    }

    this.tracks.set(publication.trackSid, track)
    this.reactive.localTracks.set(publication.trackSid, {
      id: publication.trackSid,
      source: source(track.source),
    })
  }

  private removeLocalTrack(publication: LocalTrackPublication) {
    publication.track?.detach()
    this.tracks.delete(publication.trackSid)
    this.reactive.localTracks.delete(publication.trackSid)
  }

  private addRemoteTrack(track: RemoteTrack, publication: RemoteTrackPublication) {
    if (track.kind !== LiveKitTrack.Kind.Video && track.kind !== LiveKitTrack.Kind.Audio) {
      return
    }

    this.tracks.set(publication.trackSid, track)
    this.reactive.remoteTracks.set(publication.trackSid, {
      id: publication.trackSid,
      source: source(track.source),
    })
  }

  private removeRemoteTrack(track: RemoteTrack, publication: RemoteTrackPublication) {
    track.detach()
    this.tracks.delete(publication.trackSid)
    this.reactive.remoteTracks.delete(publication.trackSid)
  }
}

function source(s: LiveKitTrack.Source): TrackSource {
  switch (s) {
    case LiveKitTrack.Source.Camera:
      return TrackSource.Camera
    case LiveKitTrack.Source.Microphone:
      return TrackSource.Microphone
    case LiveKitTrack.Source.ScreenShare:
      return TrackSource.Screen
    case LiveKitTrack.Source.ScreenShareAudio:
      return TrackSource.ScreenAudio
    default:
      return TrackSource.Unknown
  }
}

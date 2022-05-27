import { LocalStream, RemoteStream, Constraints } from "ion-sdk-js"
import { computed, reactive, readonly, ref, Ref, ComputedRef } from "vue"
import { RTCPeer, eventClient } from "@/api"
import { BufferedAggregator } from "./buffered"
import { ProfileRepository } from "./profiles"
import { MessageAggregator, Message } from "./messages"
import { Peer, PeerAggregator } from "./peers"
import { RoomEvent, Hint, Track } from "@/api/room/schema"
import { EventOrder } from "@/api/schema"

interface Remote {
  camera?: MediaStream
  screen?: MediaStream
  audios: MediaStream[]
}

interface Local {
  camera?: LocalStream
  screen?: LocalStream
  mic?: LocalStream
}

interface State {
  tracks: Map<string, Track>
}

export class LiveRoom {
  messages: Message[]
  peers: Map<string, Peer>

  private local: Local
  private remote: Remote
  private joined: Ref<boolean>
  private publishing: Ref<boolean>
  private rtc: RTCPeer
  private state: State
  private streamsByTrackId: Map<string, RemoteStream>
  private tracksById: Map<string, Track>
  private profileRepo: ProfileRepository

  constructor() {
    this.local = reactive<Local>({}) as Local
    this.remote = reactive<Remote>({
      audios: [],
    })

    this.joined = ref<boolean>(false)
    this.publishing = ref<boolean>(false)
    this.messages = reactive<Message[]>([])
    this.peers = reactive<Map<string, Peer>>(new Map<string, Peer>())

    this.profileRepo = new ProfileRepository(100, 3000)
    this.rtc = new RTCPeer()
    this.state = { tracks: new Map() }
    this.streamsByTrackId = new Map()
    this.tracksById = new Map()
  }

  close(): void {
    this.unshareScreen()
    this.unshareCamera()
    this.unshareMic()
    this.rtc.close()
  }

  remoteStreams(): Remote {
    return readonly(this.remote) as Remote
  }

  localStreams(): Local {
    return readonly(this.local) as Local
  }

  isJoined(): ComputedRef<boolean> {
    return computed(() => this.joined.value)
  }

  isPublishing(): ComputedRef<boolean> {
    return computed(() => this.publishing.value)
  }

  async join(roomId: string) {
    this.joined.value = false

    // A workaround to avoid creating a new class to hide `put` method from the public API.
    const thisAggregator = {
      put: (event: RoomEvent) => {
        this.put(event)
      },
    }

    const aggregators = new BufferedAggregator(
      [
        new MessageAggregator(this.profileRepo, this.messages),
        new PeerAggregator(this.profileRepo, this.peers),
        thisAggregator,
      ],
      50,
    )

    this.rtc.onevent = (event: RoomEvent): void => {
      aggregators.put(event)
    }
    this.rtc.ontrack = (track: MediaStreamTrack, stream: RemoteStream): void => {
      if (track.kind !== "video" && track.kind !== "audio") {
        return
      }

      const id = trackId(stream)
      this.streamsByTrackId.set(id, stream)
      this.computeRemote()
      stream.onremovetrack = () => {
        this.streamsByTrackId.delete(id)
        this.computeRemote()
      }
    }
    await this.rtc.join(roomId, true)

    const iter = eventClient.fetch({ roomId: roomId }, { order: EventOrder.DESC, policy: "network-only" })
    const events = await iter.next({ count: 500, seconds: 2 * 60 * 60 })
    // Sorting events to always be in chronological order.
    events.sort((l: RoomEvent, r: RoomEvent): number => {
      return l.createdAt - r.createdAt
    })
    aggregators.prepend(...events)
    aggregators.flush()

    this.joined.value = true
  }

  async send(userId: string, message: string): Promise<void> {
    if (!this.rtc) {
      throw new Error("Must join room before sending a message.")
    }

    const msg: Message = {
      id: "",
      fromId: userId,
      text: message,
      accepted: false,
      profile: this.profileRepo.profile(userId),
    }
    this.messages.push(msg)
    const event = await this.rtc.message({ text: message })
    msg.id = event.id
  }

  switchCamera() {
    if (this.local.camera) {
      this.unshareCamera()
    } else {
      this.shareCamera()
    }
  }

  switchScreen() {
    if (this.local.screen) {
      this.unshareScreen()
    } else {
      this.shareScreen()
    }
  }

  switchMic() {
    if (this.local.mic) {
      this.unshareMic()
    } else {
      this.shareMic()
    }
  }

  async shareCamera() {
    this.local.camera = await this.share(async () => {
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
    this.unshare(this.local.camera)
    this.local.camera = undefined
  }

  async shareScreen() {
    this.local.screen = await this.share(async () => {
      const stream = await LocalStream.getDisplayMedia({
        codec: "vp8",
        resolution: "hd",
        simulcast: false,
        video: {
          width: { ideal: 2560 },
          height: { ideal: 1440 },
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
    this.unshare(this.local.screen)
    this.local.screen = undefined
  }

  async shareMic() {
    this.local.mic = await this.share(() => {
      return LocalStream.getUserMedia({
        video: false,
        audio: true,
      } as Constraints)
    }, Hint.UserAudio)
  }

  unshareMic() {
    this.unshare(this.local.mic)
    this.local.mic = undefined
  }

  private async share(fetch: () => Promise<LocalStream>, hint: Hint): Promise<LocalStream | undefined> {
    if (!this.rtc || !this.joined.value || this.publishing.value) {
      return undefined
    }

    try {
      this.publishing.value = true
      const stream = await fetch()
      const tId = trackId(stream)
      this.state.tracks.set(tId, { id: tId, hint: hint })
      await this.rtc.state({ tracks: Array.from(this.state.tracks.values()) })
      this.rtc.publish(stream)
      return stream
    } catch (e) {
      console.warn("Failed to share media:", e)
      return undefined
    } finally {
      this.publishing.value = false
    }
  }

  private unshare(stream: LocalStream | null | undefined) {
    if (!stream) {
      return
    }
    this.state.tracks.delete(trackId(stream))
    stream.unpublish()
    for (const t of stream.getTracks()) {
      t.stop()
      stream.removeTrack(t)
    }
  }

  private put(event: RoomEvent): void {
    const payload = event.payload.peerState
    if (!payload?.tracks || payload.tracks.length === 0) {
      return
    }
    for (const t of payload.tracks) {
      this.tracksById.set(t.id, t)
    }
    this.computeRemote()
  }

  computeRemote() {
    this.remote.camera = undefined
    this.remote.screen = undefined
    this.remote.audios = []

    this.streamsByTrackId.forEach((stream: RemoteStream, trackId: string) => {
      const track = this.tracksById.get(trackId)
      if (!track) {
        return
      }
      switch (track.hint) {
        case Hint.Camera:
          this.remote.camera = stream
          break
        case Hint.Screen:
          this.remote.screen = stream
          break
        case Hint.UserAudio:
          this.remote.audios.push(stream)
          break
      }
    })
  }
}

function trackId(s: MediaStream): string {
  const tracks = s.getTracks()
  if (tracks.length < 1) {
    return ""
  }
  return tracks[0].id
}

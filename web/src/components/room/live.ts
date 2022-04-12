import { LocalStream, RemoteStream, Constraints } from "ion-sdk-js"
import { computed, reactive, readonly, ref, Ref, ComputedRef, watch } from "vue"
import { RTCPeer, eventClient, Policy } from "@/api"
import { BufferedAggregator, MessageAggregator, PeerAggregator, Message, Peer } from "@/api/room"
import { RoomEvent, Hint, Track } from "@/api/room/schema"
import { EventOrder } from "@/api/schema"
import { ProfileHydrator } from "./profiles"

interface Remote {
  camera: MediaStream | null
  screen: MediaStream | null
  audios: MediaStream[]
}

interface Local {
  camera: LocalStream | null
  screen: LocalStream | null
  mic: LocalStream | null
}

interface State {
  tracks: { [key: string]: Track }
}

export class LiveRoom {
  messages: Message[]
  peers: Peer[]

  private local: Local
  private remote: Remote
  private joined: Ref<boolean>
  private publishing: Ref<boolean>
  private rtc: RTCPeer
  private state: State
  private streamsByTrackId: { [key: string]: RemoteStream }
  private tracksById: { [key: string]: Track }

  constructor() {
    this.local = reactive<Local>({
      camera: null,
      screen: null,
      mic: null,
    }) as Local

    this.remote = reactive<Remote>({
      camera: null,
      screen: null,
      audios: [],
    })

    this.joined = ref<boolean>(false)
    this.publishing = ref<boolean>(false)
    this.messages = reactive<Message[]>([])
    this.peers = reactive<Peer[]>([])

    this.rtc = new RTCPeer()
    this.state = { tracks: {} }
    this.streamsByTrackId = {}
    this.tracksById = {}

    const profileHydrator = new ProfileHydrator(this.peers, 1000)
    watch(
      this.peers,
      () => {
        profileHydrator.hydrate()
      },
      { immediate: false, deep: false },
    )
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
      [new MessageAggregator(this.messages), new PeerAggregator(this.peers), thisAggregator],
      500,
    )

    this.rtc.onevent = (event: RoomEvent): void => {
      aggregators.put(event)
    }
    this.rtc.ontrack = (track: MediaStreamTrack, stream: RemoteStream): void => {
      if (track.kind !== "video" && track.kind !== "audio") {
        return
      }

      const id = trackId(stream)
      this.streamsByTrackId[id] = stream
      this.computeRemote()
      stream.onremovetrack = () => {
        delete this.streamsByTrackId[id]
        this.computeRemote()
      }
    }
    await this.rtc.join(roomId, true)

    const iter = eventClient.fetch({ roomId: roomId }, EventOrder.DESC, Policy.NetworkOnly)
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
        simulcast: true,
        video: true,
        audio: false,
      })
    }, Hint.Camera)
  }

  unshareCamera() {
    this.unshare(this.local.camera)
    this.local.camera = null
  }

  async shareScreen() {
    this.local.screen = await this.share(async () => {
      const stream = await LocalStream.getDisplayMedia({
        codec: "vp8",
        resolution: "hd",
        simulcast: true,
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
    this.local.screen = null
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
    this.local.mic = null
  }

  private async share(fetch: () => Promise<LocalStream>, hint: Hint): Promise<LocalStream | null> {
    if (!this.rtc || !this.joined.value || this.publishing.value) {
      return null
    }

    try {
      this.publishing.value = true
      const stream = await fetch()
      const tId = trackId(stream)
      this.state.tracks[tId] = { id: tId, hint: hint }
      await this.rtc.state({ tracks: Object.values(this.state.tracks) })
      this.rtc.publish(stream)
      return stream
    } catch (e) {
      console.warn("Failed to share media:", e)
      return null
    } finally {
      this.publishing.value = false
    }
  }

  private unshare(stream: LocalStream | null | undefined) {
    if (!stream) {
      return
    }
    delete this.state.tracks[trackId(stream)]
    stream.unpublish()
    for (const t of stream.getTracks()) {
      t.stop()
      stream.removeTrack(t)
    }
  }

  private put(event: RoomEvent): void {
    const payload = event.payload.peerState
    if (!payload?.tracks) {
      return
    }
    for (const t of payload.tracks) {
      this.tracksById[t.id] = t
    }
    this.computeRemote()
  }

  computeRemote() {
    this.remote.camera = null
    this.remote.screen = null
    this.remote.audios = []

    for (const id in this.streamsByTrackId) {
      const track = this.tracksById[id]
      if (!track) {
        continue
      }
      switch (track.hint) {
        case Hint.Camera:
          this.remote.camera = this.streamsByTrackId[id]
          break
        case Hint.Screen:
          this.remote.screen = this.streamsByTrackId[id]
          break
        case Hint.UserAudio:
          this.remote.audios.push(this.streamsByTrackId[id])
          break
      }
    }
  }
}

function trackId(s: MediaStream): string {
  const tracks = s.getTracks()
  if (tracks.length < 1) {
    return ""
  }
  return tracks[0].id
}

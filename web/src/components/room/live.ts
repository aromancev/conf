import { Client as SFU, LocalStream, RemoteStream, Constraints } from "ion-sdk-js"
import { computed, reactive, readonly, ref, Ref, ComputedRef } from "vue"
import { config } from "@/config"
import { RTC, Event, client, eventClient, State, Hint, Track, Policy } from "@/api"
import { EventType, PayloadPeerState } from "@/api/models"
import { BufferedAggregator, MessageAggregator, PeerAggregator, Message, Peer } from "@/api/room"
import { EventOrder } from "@/api/schema"

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

export class LiveRoom {
  messages: Message[]
  peers: Peer[]

  private local: Local
  private remote: Remote
  private joined: Ref<boolean>
  private publishing: Ref<boolean>
  private rtc: RTC | null
  private sfu: SFU | null
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

    this.rtc = null
    this.sfu = null
    this.state = { tracks: {} }
    this.streamsByTrackId = {}
    this.tracksById = {}
  }

  close(): void {
    this.unshareScreen()
    this.unshareCamera()
    this.unshareMic()
    this.sfu?.close()
    this.rtc?.close()
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

  async join(userId: string, roomId: string) {
    this.joined.value = false

    this.rtc = await client.rtc(roomId)
    const iceServers: RTCIceServer[] = []
    if (config.sfu.stunURLs.length) {
      iceServers.push({
        urls: config.sfu.stunURLs,
      })
    }
    if (config.sfu.turnURLs.length) {
      iceServers.push({
        urls: config.sfu.turnURLs,
        credentialType: "password",
        username: this.rtc.token,
        credential: "confa.io",
      })
    }
    this.sfu = new SFU(this.rtc, {
      codec: "vp8",
      iceServers: iceServers,
    })

    const aggregators = new BufferedAggregator(
      [new MessageAggregator(this.messages), new PeerAggregator(this.peers), this],
      500,
    )

    this.rtc.onevent = (event: Event) => {
      aggregators.put(event, true)
    }
    this.rtc.onopen = async () => {
      await this.sfu?.join(roomId, userId)
      this.joined.value = true
    }
    this.sfu.ontrack = (track: MediaStreamTrack, stream: RemoteStream) => {
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

    const iter = eventClient.fetch({ roomId: roomId }, EventOrder.DESC, Policy.NetworkOnly)
    const events = await iter.next({ count: 500, seconds: 2 * 60 * 60 })
    for (const ev of events) {
      aggregators.put(ev, true)
    }
    aggregators.autoflush = true
    aggregators.flush()
  }

  async send(userId: string, message: string): Promise<void> {
    if (!this.rtc) {
      throw new Error("Must join room before sending a message.")
    }

    const msg: Message = {
      id: "",
      from: userId,
      text: message,
      accepted: false,
    }
    const ev = {
      payload: {
        type: EventType.Message,
        payload: {
          text: message,
        },
      },
    }
    this.messages.push(msg)
    msg.id = await this.rtc.event(ev)
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
    if (!this.rtc || !this.sfu || this.publishing.value) {
      return null
    }

    try {
      this.publishing.value = true
      const stream = await fetch()
      this.state.tracks[trackId(stream)] = { hint: hint }
      await this.rtc.state(this.state)
      this.sfu.publish(stream)
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

  // TODO: remove from public api.
  put(event: Event): void {
    if (event.payload.type !== EventType.PeerState) {
      return
    }
    const payload = event.payload.payload as PayloadPeerState
    if (!payload.tracks) {
      return
    }
    Object.assign(this.tracksById, payload.tracks)
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
    if (this.local.camera) {
      this.remote.camera = this.local.camera
    }
    if (this.local.screen) {
      this.remote.screen = this.local.screen
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

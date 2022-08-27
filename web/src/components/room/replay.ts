import { reactive, readonly } from "vue"
import { eventClient, recordingClient } from "@/api"
import { ProfileRepository } from "./profiles"
import { MessageAggregator, Message } from "./aggregators/messages"
import { Peer, PeerAggregator } from "./aggregators/peers"
import { Media, MediaAggregator } from "./aggregators/media"
import { RoomEvent } from "@/api/room/schema"
import { duration } from "@/platform/time"

export interface State {
  isLoaded: boolean
  isPlaying: boolean
  duration: number
  delta: number
  unpausedAt: number
  messages: Message[]
  peers: Map<string, Peer>
  medias: Map<string, Media>
}

export class ReplayRoom {
  private _state: State
  private profileRepo: ProfileRepository
  private events: RoomEvent[]
  private aggregators: Aggregator[]
  private recordingStartedAt: number
  private replayTimeout: ReturnType<typeof setTimeout>
  private eventsPrepeared: number
  private eventsConsumed: number
  private stopped: boolean

  constructor() {
    this._state = reactive<State>({
      isLoaded: false,
      isPlaying: false,
      duration: 0,
      delta: 0,
      unpausedAt: 0,
      messages: [],
      peers: new Map(),
      medias: new Map(),
    })
    this.profileRepo = new ProfileRepository(100, 3000)
    this.recordingStartedAt = 0
    this.eventsPrepeared = 0
    this.eventsConsumed = 0
    this.events = []
    this.aggregators = []
    this.replayTimeout = -1
    this.stopped = false
  }

  state(): State {
    return readonly(this._state) as State
  }

  async load(talkId: string, roomId: string) {
    this._state.isLoaded = false

    try {
      const recording = await recordingClient.fetchOne({ roomId: roomId, key: talkId })
      if (!recording.stoppedAt) {
        throw new Error("Recording is not stopped.")
      }

      const media = new MediaAggregator(roomId, recording.startedAt)
      this._state.medias = media.state().medias
      this.aggregators = [
        new MessageAggregator(this.profileRepo, this._state.messages),
        new PeerAggregator(this.profileRepo, this._state.peers),
        media,
      ]

      const eventsFrom = recording.startedAt - duration({ minutes: 15 })
      const iter = eventClient.fetch(
        {
          roomId: roomId,
        },
        {
          from: {
            createdAt: eventsFrom.toString(),
            id: "",
          },
        },
      )
      this.events = await iter.next({ count: 3000 })
      this.recordingStartedAt = recording.startedAt
      this._state.duration = recording.stoppedAt - recording.startedAt
      this._state.delta = 0
      this.prepare()
      this.consumeUntil(recording.startedAt)
    } finally {
      this._state.isLoaded = true
    }
  }

  play(): void {
    if (!this._state.isLoaded) {
      return
    }
    // If unpaused is unset, it means we are playing from the beginning.
    if (this.stopped) {
      this.rewind(0)
      this.stopped = false
    }
    clearTimeout(this.replayTimeout)
    this._state.isPlaying = true
    this._state.unpausedAt = Date.now()
    this.iterate()
  }

  pause(): void {
    if (!this._state.isLoaded) {
      return
    }
    this._state.isPlaying = false
    clearTimeout(this.replayTimeout)
    this._state.delta = this._state.delta + Date.now() - this._state.unpausedAt
  }

  stop(): void {
    if (!this._state.isLoaded) {
      return
    }
    this._state.isPlaying = false
    clearTimeout(this.replayTimeout)
    this._state.delta = 0
    this._state.unpausedAt = 0
    this.stopped = true
  }

  togglePlay(): void {
    if (this._state.isPlaying) {
      this.pause()
    } else {
      this.play()
    }
  }

  rewind(pos: number): void {
    clearTimeout(this.replayTimeout)
    this._state.delta = 0
    this._state.unpausedAt = 0
    for (const agg of this.aggregators) {
      if (agg.reset) {
        agg.reset()
      }
    }
    this._state.messages.splice(0, this._state.messages.length)
    this._state.peers.clear()
    this.eventsConsumed = 0
    this.consumeUntil(this.recordingStartedAt + pos)
    this._state.delta = pos
    if (this._state.isPlaying) {
      this._state.unpausedAt = Date.now()
      this.iterate()
    }
  }

  private iterate(): void {
    const progress = Date.now() - this._state.unpausedAt + this._state.delta
    if (progress >= this._state.duration) {
      this.stop()
      return
    }
    const nextEventAt = this.consumeUntil(this.recordingStartedAt + progress)
    if (nextEventAt === 0) {
      return
    }
    const iterateIn = nextEventAt - this.recordingStartedAt - progress + 100
    this.replayTimeout = setTimeout(() => this.iterate(), iterateIn)
  }

  private consumeUntil(stopAt: number): number {
    for (let i = this.eventsConsumed; i < this.events.length; i++) {
      const ev = this.events[i]
      if (ev.createdAt > stopAt) {
        return ev.createdAt
      }

      for (const agg of this.aggregators) {
        agg.put(ev)
      }
      this.eventsConsumed += 1
    }

    return 0
  }

  private prepare(): void {
    const toPrepare = this.events.slice(this.eventsPrepeared, this.events.length)
    for (const agg of this.aggregators) {
      if (agg.prepare) {
        agg.prepare(toPrepare)
      }
    }
    this.eventsPrepeared = this.events.length
  }
}

interface Aggregator {
  put(event: RoomEvent): void
  prepare?(events: RoomEvent[]): void
  reset?(): void
}

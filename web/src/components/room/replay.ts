import { reactive, readonly } from "vue"
import { eventClient, recordingClient } from "@/api"
import { EventIterator } from "@/api/event"
import { ProfileRepository } from "./profiles"
import { MessageAggregator, Message } from "./aggregators/messages"
import { Peer, PeerAggregator } from "./aggregators/peers"
import { Media, MediaAggregator } from "./aggregators/media"
import { RoomEvent } from "@/api/room/schema"
import { duration } from "@/platform/time"
import { Throttler } from "@/platform/sync"

export interface State {
  isLoading: boolean
  isPlaying: boolean
  isBuffering: boolean
  duration: number
  buffer: number
  progress: Progress
  messages: Message[]
  peers: Map<string, Peer>
  medias: Map<string, Media>
}

export interface Progress {
  value: number
  increasingSince: number
}

export class ReplayRoom {
  private _state: State
  private readState: State
  private roomId: string
  private aggregators: Aggregator[]
  private recordingStartedAt: number
  private profileRepo: ProfileRepository
  private eventIter?: EventIterator
  private eventBatch: RoomEvent[]
  private putFromIndex: number
  private buffers: Map<string, number>
  private stopped: boolean
  private processEvents: Throttler<void>
  private fetchEvents: Throttler<void>
  private processIntervalId: ReturnType<typeof setInterval>
  private deferredProcessTimeoutId: ReturnType<typeof setTimeout>

  constructor() {
    this._state = reactive<State>({
      isLoading: true,
      isPlaying: false,
      isBuffering: true,
      duration: 0,
      buffer: 0,
      messages: [],
      progress: {
        value: 0,
        increasingSince: 0,
      },
      peers: new Map(),
      medias: new Map(),
    })
    this.readState = readonly(this._state) as State
    this.profileRepo = new ProfileRepository(100, 3 * 1000)
    this.recordingStartedAt = 0
    this.putFromIndex = 0
    this.eventBatch = []
    this.aggregators = []
    this.stopped = false
    this.buffers = new Map()
    this.roomId = ""
    this.processEvents = new Throttler({ delayMs: MIN_EVENT_DELAY_MS })
    this.fetchEvents = new Throttler({ delayMs: MIN_FETCH_DELAY_MS })
    this.processEvents.func = () => this.processEventsFunc()
    this.fetchEvents.func = () => this.fetchEventsFunc()
    this.processIntervalId = setInterval(() => this.processEvents.do(), MAX_EVENT_DELAY_MS)
    this.processIntervalId = 0
    this.deferredProcessTimeoutId = -1
  }

  get state(): State {
    return this.readState
  }

  async load(talkId: string, roomId: string) {
    this._state.isLoading = true

    this.roomId = roomId
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

      this.resetState()
      this.resetEventFetching()
      this.recordingStartedAt = recording.startedAt
      this._state.duration = recording.stoppedAt - recording.startedAt
      this._state.progress.value = 0
      this._state.progress.increasingSince = 0
      while (this.eventBuffer() <= Math.min(FETCH_ADVANCE_MS, this._state.duration)) {
        await this.fetchEvents.do()
      }
    } finally {
      this._state.isLoading = false
    }
  }

  play(): void {
    if (this._state.isLoading) {
      return
    }
    if (this.stopped) {
      this.rewind(0)
      this.stopped = false
    }
    this._state.isPlaying = true
    if (!this._state.isBuffering) {
      this._state.progress.increasingSince = Date.now()
    }
    this.processEvents.do()
  }

  pause(): void {
    if (this._state.isLoading) {
      return
    }
    this._state.isPlaying = false
    if (!this._state.progress.increasingSince) {
      return
    }
    this._state.progress.value = this._state.progress.value + Date.now() - this._state.progress.increasingSince
    this._state.progress.increasingSince = 0
  }

  stop(): void {
    if (this._state.isLoading) {
      return
    }
    this._state.isPlaying = false
    this._state.progress.value = this._state.duration
    this._state.progress.increasingSince = 0
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
    if (this._state.isLoading) {
      return
    }

    const progress = this.progressFor(Date.now())
    if (pos < progress || this.stopped) {
      this.resetState()
      if (this.eventIter && this.eventIter.pagesIterated() > 1) {
        this.resetEventFetching()
      }
    }
    this.stopped = false

    this._state.progress.value = pos
    if (this._state.isPlaying) {
      this._state.progress.increasingSince = Date.now()
    } else {
      this._state.progress.increasingSince = 0
    }
    this.processEvents.do()
  }

  updateMediaBuffer(id: string, ms: number): void {
    const media = this._state.medias.get(id)
    if (!media) {
      return
    }
    this.updateBuffer(id, media.startsAt + ms)
  }

  close(): void {
    this.stop()
    clearInterval(this.processIntervalId)
  }

  private updateBuffer(id: string, ms: number): void {
    this.buffers.set(id, ms)
    this._state.buffer = Math.min(...Array.from(this.buffers.values()))
    if (this._state.buffer > this._state.duration) {
      this._state.buffer = this._state.duration
    }

    this.processEvents.do()
  }

  private processEventsFunc(): void {
    clearTimeout(this.deferredProcessTimeoutId)

    const now = Date.now()
    const progress = this.progressFor(now)

    // Kick fetching loop just in case. It will doearly return anyway.
    this.fetchEvents.do()

    // Put due events.
    const nextEventAt = this.putEventsUntil(this.recordingStartedAt + progress)
    // Only stop AFTER the events were consumed.
    if (progress >= this._state.duration) {
      this.stop()
      return
    }

    // Update buffering state.
    const wasBuffering = this._state.isBuffering
    this._state.isBuffering = progress >= this._state.buffer
    if (this._state.isBuffering) {
      this._state.progress.increasingSince = 0
    }
    // Stopped buffering and is playing.
    if (wasBuffering && !this._state.isBuffering && this._state.isPlaying) {
      this._state.progress.increasingSince = now
    }
    // Started buffering.
    if (!wasBuffering && this._state.isBuffering) {
      this._state.progress.value = progress
    }

    if (!nextEventAt || this._state.isBuffering || !this._state.isPlaying) {
      return
    }

    const untilNextEvent = nextEventAt - this.recordingStartedAt - progress
    const untilBufferRunsOut = this._state.buffer - progress
    this.deferredProcessTimeoutId = setTimeout(
      () => this.processEvents.do(),
      Math.min(untilNextEvent, untilBufferRunsOut),
    )
  }

  private async fetchEventsFunc(): Promise<void> {
    const eventBuffer = this.eventBuffer()
    if (eventBuffer > this._state.duration) {
      return
    }
    const progress = this.progressFor(Date.now())
    if (progress < eventBuffer - FETCH_ADVANCE_MS) {
      return
    }

    if (!this.eventIter) {
      const eventsFrom = this.recordingStartedAt - duration({ minutes: 15 })
      this.eventIter = eventClient.fetch(
        {
          roomId: this.roomId,
        },
        {
          from: {
            createdAt: eventsFrom.toString(),
            id: "",
          },
        },
      )
    }

    const fetched = await this.eventIter.next({ count: EVENT_BATCH })
    if (!fetched.length) {
      // Didn't fetch anything or the fetch was aborted (iterator reset).
      return
    }
    // Append the next batch of events and remove all events that are already put.
    this.eventBatch = this.eventBatch.slice(this.putFromIndex).concat(fetched)
    this.putFromIndex = 0
    // Prepare all fetched events.
    for (const agg of this.aggregators) {
      if (agg.prepare) {
        agg.prepare(fetched)
      }
    }
    const lastAt = fetched[fetched.length - 1].createdAt
    this.updateBuffer(EVENTS_BUFFER_ID, lastAt - this.recordingStartedAt)
    return
  }

  private putEventsUntil(stopAt: number): number {
    for (let i = this.putFromIndex; i < this.eventBatch.length; i++) {
      const ev = this.eventBatch[i]
      if (ev.createdAt > stopAt) {
        return ev.createdAt
      }

      for (const agg of this.aggregators) {
        agg.put(ev)
      }
      this.putFromIndex += 1
    }

    return 0
  }

  private resetState(): void {
    for (const agg of this.aggregators) {
      if (agg.reset) {
        agg.reset()
      }
    }
    this._state.messages.splice(0, this._state.messages.length)
    this._state.peers.clear()
    this.putFromIndex = 0
  }

  private resetEventFetching(): void {
    this.buffers.clear()
    this._state.buffer = 0
    this.eventIter = undefined
    this.eventBatch = []
  }

  private progressFor(time: number): number {
    if (!this._state.progress.increasingSince) {
      return this._state.progress.value
    }
    const timeProgress = time - this._state.progress.increasingSince + this._state.progress.value
    return Math.min(timeProgress, this._state.duration)
  }

  private eventBuffer(): number {
    return this.buffers.get(EVENTS_BUFFER_ID) || 0
  }
}

const EVENT_BATCH = 3000
const FETCH_ADVANCE_MS = 60 * 1000
const MIN_EVENT_DELAY_MS = 100
const MAX_EVENT_DELAY_MS = 3 * 1000
const MIN_FETCH_DELAY_MS = 1 * 1000
const EVENTS_BUFFER_ID = "events"

interface Aggregator {
  put(event: RoomEvent): void
  prepare?(events: RoomEvent[]): void
  reset?(): void
}

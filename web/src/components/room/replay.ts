import { reactive, readonly } from "vue"
import { eventClient } from "@/api"
import { Recording } from "@/api/models/recording"
import { EventIterator } from "@/api/event"
import { ProfileRepository } from "./profiles"
import { MessageAggregator, Message } from "./aggregators/messages"
import { Peer, Status, PeerAggregator } from "./aggregators/peers"
import { Media, MediaAggregator } from "./aggregators/media"
import { RoomEvent } from "@/api/room/schema"
import { duration } from "@/platform/time"
import { Throttler } from "@/platform/sync"
import { FIFOMap } from "@/platform/cache"

const PROFILE_CACHE_SIZE = 500
const MAX_PEERS = 3000
const MAX_MEDIAS = 10

interface State {
  isLoading: boolean
  isPlaying: boolean
  isBuffering: boolean
  duration: number
  buffer: number
  progress: Progress
  messages: Message[]
  peers: Map<string, Peer>
  statuses: Map<string, Status>
  medias: Map<string, Media>
}

export interface Progress {
  value: number
  increasingSince: number
}

export class ReplayRoom {
  private readonly reactive: State
  private readonly readonly: State
  private roomId: string
  private aggregators: Aggregator[]
  private recordingStartedAt: number
  private profileRepo: ProfileRepository
  private eventIter?: EventIterator
  private eventBatch: RoomEvent[]
  private putFromIndex: number
  private buffers: Map<string, MediaBuffer>
  private stopped: boolean
  private processEvents: Throttler<void>
  private fetchEvents: Throttler<void>
  private processIntervalId: ReturnType<typeof setInterval>
  private deferredProcessTimeoutId: ReturnType<typeof setTimeout>

  constructor() {
    this.reactive = reactive<State>({
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
      peers: new FIFOMap(MAX_PEERS),
      statuses: new FIFOMap(MAX_PEERS),
      medias: new FIFOMap(MAX_MEDIAS),
    })
    this.readonly = readonly(this.reactive) as State
    this.profileRepo = new ProfileRepository(PROFILE_CACHE_SIZE, 3 * 1000)
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
    this.deferredProcessTimeoutId = -1
  }

  get state(): State {
    return this.readonly
  }

  async load(roomId: string, recording: Recording) {
    this.reactive.isLoading = true

    this.roomId = roomId
    try {
      if (!recording.stoppedAt) {
        throw new Error("Recording is not finished.")
      }

      const media = new MediaAggregator(this.reactive.medias, roomId, recording.startedAt)
      const peers = new PeerAggregator(this.reactive.peers, this.reactive.statuses, this.profileRepo)
      const messages = new MessageAggregator(this.reactive.messages, this.profileRepo)
      this.aggregators = [messages, peers, media]

      this.resetState()
      this.resetEventFetching()
      this.recordingStartedAt = recording.startedAt
      this.reactive.duration = recording.stoppedAt - recording.startedAt
      this.reactive.progress.value = 0
      this.reactive.progress.increasingSince = 0
      while (this.eventBuffer() <= Math.min(FETCH_ADVANCE_MS, this.reactive.duration)) {
        await this.fetchEvents.do()
      }
      await this.processEvents.do()
    } finally {
      this.reactive.isLoading = false
    }
  }

  play(): void {
    if (this.reactive.isLoading) {
      return
    }
    if (this.stopped) {
      this.rewind(0)
      this.stopped = false
    }
    this.reactive.isPlaying = true
    if (!this.reactive.isBuffering) {
      this.reactive.progress.increasingSince = Date.now()
    }
    this.processEvents.do()
  }

  pause(): void {
    if (this.reactive.isLoading) {
      return
    }
    this.reactive.isPlaying = false
    if (!this.reactive.progress.increasingSince) {
      return
    }
    this.reactive.progress.value = this.reactive.progress.value + Date.now() - this.reactive.progress.increasingSince
    this.reactive.progress.increasingSince = 0
  }

  stop(): void {
    if (this.reactive.isLoading) {
      return
    }
    this.reactive.isPlaying = false
    this.reactive.progress.value = this.reactive.duration
    this.reactive.progress.increasingSince = 0
    this.stopped = true
  }

  togglePlay(): void {
    if (this.reactive.isPlaying) {
      this.pause()
    } else {
      this.play()
    }
  }

  rewind(pos: number): void {
    if (this.reactive.isLoading) {
      return
    }

    const progress = this.progressForTime(Date.now())
    if (pos < progress || this.stopped) {
      this.resetState()
      if (this.eventIter && this.eventIter.pagesIterated() > 1) {
        this.resetEventFetching()
      }
    }
    this.stopped = false

    this.reactive.progress.value = pos
    if (this.reactive.isPlaying) {
      this.reactive.progress.increasingSince = Date.now()
    } else {
      this.reactive.progress.increasingSince = 0
    }
    this.processEvents.do()
  }

  updateMediaBuffer(id: string, bufferMs: number, durationMs: number): void {
    const media = this.reactive.medias.get(id)
    if (!media) {
      return
    }
    this.updateBuffer(id, {
      bufferMs: media.startsAt + bufferMs,
      durationMs: durationMs,
    })
  }

  close(): void {
    this.stop()
    clearInterval(this.processIntervalId)
  }

  private updateBuffer(id: string, buffer: MediaBuffer): void {
    this.buffers.set(id, buffer)
    this.processEvents.do()
  }

  private processEventsFunc(): void {
    clearTimeout(this.deferredProcessTimeoutId)

    const now = Date.now()
    const progress = this.progressForTime(now)

    // Kick fetching loop just in case. It will doearly return anyway.
    this.fetchEvents.do()

    // Put due events.
    const nextEventAt = this.putEventsUntil(this.recordingStartedAt + progress)

    // Update time for aggregators.
    this.setAggregatorsTime(this.recordingStartedAt + progress)

    // Only stop AFTER the events were consumed.
    if (progress >= this.reactive.duration) {
      this.stop()
      return
    }

    // Update buffering state.
    this.reactive.buffer = this.bufferForProgress(progress)
    const wasBuffering = this.reactive.isBuffering
    this.reactive.isBuffering = progress >= this.reactive.buffer
    if (this.reactive.isBuffering) {
      this.reactive.progress.increasingSince = 0
    }
    // Stopped buffering and is playing.
    if (wasBuffering && !this.reactive.isBuffering && this.reactive.isPlaying) {
      this.reactive.progress.increasingSince = now
    }
    // Started buffering.
    if (!wasBuffering && this.reactive.isBuffering) {
      this.reactive.progress.value = progress
    }

    if (!nextEventAt || this.reactive.isBuffering || !this.reactive.isPlaying) {
      return
    }

    const untilNextEvent = nextEventAt - this.recordingStartedAt - progress
    const untilBufferRunsOut = this.reactive.buffer - progress
    this.deferredProcessTimeoutId = setTimeout(
      () => this.processEvents.do(),
      Math.min(untilNextEvent, untilBufferRunsOut),
    )
  }

  private async fetchEventsFunc(): Promise<void> {
    const eventBuffer = this.eventBuffer()
    if (eventBuffer > this.reactive.duration) {
      return
    }
    const progress = this.progressForTime(Date.now())
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
          cursor: {
            createdAt: eventsFrom.toString(),
            Asc: true,
          },
        },
      )
    }

    const fetched = await this.eventIter.next(EVENT_BATCH)
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
    this.updateBuffer(EVENTS_BUFFER_ID, {
      bufferMs: lastAt - this.recordingStartedAt,
      durationMs: this.reactive.duration,
    })
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

  private setAggregatorsTime(time: number): void {
    for (const agg of this.aggregators) {
      if (!agg.setTime) {
        continue
      }
      agg.setTime(time)
    }
  }

  private resetState(): void {
    for (const agg of this.aggregators) {
      if (agg.reset) {
        agg.reset()
      }
    }
    this.putFromIndex = 0
  }

  private resetEventFetching(): void {
    this.buffers.clear()
    this.reactive.buffer = 0
    this.eventIter = undefined
    this.eventBatch = []
  }

  private progressForTime(time: number): number {
    if (!this.reactive.progress.increasingSince) {
      return this.reactive.progress.value
    }
    const timeProgress = time - this.reactive.progress.increasingSince + this.reactive.progress.value
    return Math.min(timeProgress, this.reactive.duration)
  }

  private bufferForProgress(progress: number): number {
    let min = Infinity
    this.buffers.forEach((buf: MediaBuffer, id: string) => {
      const media = this.reactive.medias.get(id)
      if (!media) {
        return
      }
      if (media.startsAt + buf.durationMs <= progress) {
        return
      }
      min = Math.min(min, media.startsAt + buf.bufferMs)
    })
    return min
  }

  private eventBuffer(): number {
    return this.buffers.get(EVENTS_BUFFER_ID)?.bufferMs || 0
  }
}

const EVENT_BATCH = 3000
const FETCH_ADVANCE_MS = 60 * 1000
const MIN_EVENT_DELAY_MS = 200
const MAX_EVENT_DELAY_MS = 1 * 500
const MIN_FETCH_DELAY_MS = 1 * 1000
const EVENTS_BUFFER_ID = "events"

interface Aggregator {
  put(event: RoomEvent): void
  setTime?(time: number): void
  prepare?(events: RoomEvent[]): void
  reset?(): void
}

interface MediaBuffer {
  bufferMs: number
  durationMs: number
}

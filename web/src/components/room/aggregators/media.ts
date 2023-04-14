import { RoomEvent, Track, Hint } from "@/api/room/schema"
import { config } from "@/config"
import { FIFOMap } from "@/platform/cache"

export interface Media {
  id: string
  manifestUrl: string
  hint: Hint
  startedAt: number
  isLive: boolean
}

export class MediaAggregator {
  private readonly medias: Map<string, Media>
  private roomId: string
  private startedAt: number
  private recordings: FIFOMap<string, Recording>
  private tracks: FIFOMap<string, Track>

  constructor(medias: Map<string, Media>, roomId: string, startedAt: number) {
    this.medias = medias
    this.roomId = roomId
    this.startedAt = startedAt
    this.recordings = new FIFOMap(CAPACITY)
    this.tracks = new FIFOMap(CAPACITY)
  }

  prepare(events: RoomEvent[]): void {
    for (const event of events) {
      const recording = event.payload.trackRecording
      if (recording) {
        this.recordings.set(recording.trackId, {
          id: recording.id,
          trackId: recording.trackId,
          startedAt: event.createdAt - this.startedAt,
          isLive: false,
        })
        continue
      }

      const peerState = event.payload.peerState
      if (peerState) {
        if (!peerState?.tracks || peerState.tracks.length === 0) {
          continue
        }
        for (const t of peerState.tracks) {
          this.tracks.set(t.id, t)
        }
        continue
      }
    }
    this.calculateMedias()
  }

  put(event: RoomEvent): void {
    const recording = event.payload.trackRecording
    if (!recording) {
      return
    }
    const rec = this.recordings.get(recording.trackId)
    if (!rec) {
      return
    }
    rec.isLive = true
    this.calculateMedias()
  }

  reset(): void {
    for (const recording of this.recordings.values()) {
      recording.isLive = false
    }
    this.calculateMedias()
  }

  private calculateMedias(): void {
    this.medias.clear()

    for (const rec of this.recordings.values()) {
      const track = this.tracks.get(rec.trackId)
      if (!track) {
        break
      }
      this.medias.set(rec.id, {
        id: rec.id,
        manifestUrl: `${config.storage.baseURL}/confa-tracks-public/${this.roomId}/${rec.id}/manifest`,
        hint: track.hint,
        startedAt: rec.startedAt,
        isLive: rec.isLive,
      })
    }
  }
}

const CAPACITY = 10

interface Recording {
  id: string
  trackId: string
  startedAt: number
  isLive: boolean
}

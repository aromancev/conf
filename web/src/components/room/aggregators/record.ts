import { RoomEvent, TrackKind, TrackSource } from "@/api/rtc/schema"
import { config } from "@/config"

export interface TrackRecord {
  recordId: string
  kind: TrackKind
  source: TrackSource
  manifestUrl: string
  startedAt: number
  isLive: boolean
}

export class MediaAggregator {
  private readonly tracks: Map<string, TrackRecord>
  private roomId: string
  private startedAt: number

  constructor(tracks: Map<string, TrackRecord>, roomId: string, startedAt: number) {
    this.tracks = tracks
    this.roomId = roomId
    this.startedAt = startedAt
  }

  prepare(events: RoomEvent[]): void {
    for (const event of events) {
      const pl = event.payload.trackRecord
      if (!pl) {
        continue
      }
      this.tracks.set(pl.recordId, {
        recordId: pl.recordId,
        kind: pl.kind,
        source: pl.source,
        startedAt: event.createdAt - this.startedAt,
        manifestUrl: `${config.storage.baseURL}/confa-tracks-public/${this.roomId}/${pl.recordId}/manifest`,
        isLive: false,
      })
    }
  }

  put(event: RoomEvent): void {
    const record = event.payload.trackRecord
    if (!record) {
      return
    }
    const track = this.tracks.get(record.recordId)
    if (!track) {
      return
    }
    track.isLive = true
  }

  reset(): void {
    for (const track of this.tracks.values()) {
      track.isLive = false
    }
  }
}

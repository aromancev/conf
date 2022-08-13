import { reactive, readonly } from "vue"
import { RoomEvent, Track, Hint } from "@/api/room/schema"
import { config } from "@/config"

export interface Media {
  manifestUrl: string
  hint: Hint
  startsAt?: number
}

export class MediaAggregator {
  private roomId: string
  private startedAt: number
  private recordings: Map<string, Recording>
  private tracks: Map<string, Track>
  private _state: State

  constructor(roomId: string, startedAt: number) {
    this.roomId = roomId
    this.startedAt = startedAt
    this.recordings = new Map()
    this.tracks = new Map()
    this._state = reactive({
      medias: new Map(),
    })
  }

  state(): State {
    return readonly(this._state) as State
  }

  prepare(events: RoomEvent[]): void {
    for (const event of events) {
      const recording = event.payload.trackRecording
      if (recording) {
        this.recordings.set(recording.trackId, {
          id: recording.id,
          trackId: recording.trackId,
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

    rec.startsAt = event.createdAt - this.startedAt
    this.calculateMedias()
    return
  }

  reset(): void {
    for (const recording of this.recordings.values()) {
      recording.startsAt = undefined
    }
    this.calculateMedias()
  }

  private calculateMedias(): void {
    this._state.medias.clear()

    for (const rec of this.recordings.values()) {
      const track = this.tracks.get(rec.trackId)
      if (!track) {
        break
      }
      this._state.medias.set(rec.id, {
        manifestUrl: `${config.storage.baseURL}/confa-tracks-public/${this.roomId}/${rec.id}/manifest`,
        hint: track.hint,
        startsAt: rec.startsAt,
      })
    }
  }
}

interface Recording {
  id: string
  trackId: string
  startsAt?: number
}

interface State {
  medias: Map<string, Media>
}

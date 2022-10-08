import { reactive, readonly } from "vue"
import { RemoteStream } from "ion-sdk-js"
import { RoomEvent, Track, Hint } from "@/api/room/schema"
import { FIFO } from "@/platform/cache"

export interface Stream {
  id: string
  hint: Hint
  sourse: MediaStream
}

export interface State {
  streams: Map<string, Stream>
}

export class StreamAggregator {
  private tracks: FIFO<string, Track>
  private streams: FIFO<string, MediaStream>
  private _state: State

  constructor() {
    this.tracks = new FIFO(CAPACITY)
    this.streams = new FIFO(CAPACITY)
    this._state = reactive({
      streams: new FIFO(CAPACITY),
    })
  }

  state(): State {
    return readonly(this._state) as State
  }

  put(event: RoomEvent): void {
    const peerState = event.payload.peerState
    if (!peerState) {
      return
    }
    if (!peerState?.tracks || peerState.tracks.length === 0) {
      return
    }
    for (const t of peerState.tracks) {
      this.tracks.set(t.id, t)
    }
    this.computeState()
  }

  reset(): void {
    this.tracks.clear()
    this.streams.clear()
    this._state.streams.clear()
  }

  addTrack(track: MediaStreamTrack, stream: RemoteStream): void {
    if (track.kind !== "video" && track.kind !== "audio") {
      return
    }

    const id = trackId(stream)
    this.streams.set(id, stream)
    this.computeState()
    stream.onremovetrack = () => {
      this.streams.delete(id)
      this.computeState()
    }
  }

  private computeState(): void {
    this._state.streams.clear()

    for (const stream of this.streams.values()) {
      const id = trackId(stream)
      const track = this.tracks.get(id)
      if (!track) {
        break
      }
      this._state.streams.set(id, {
        id: id,
        hint: track.hint,
        sourse: stream,
      })
    }
  }
}

const CAPACITY = 10

function trackId(s: MediaStream): string {
  const tracks = s.getTracks()
  if (tracks.length < 1) {
    return ""
  }
  return tracks[0].id
}

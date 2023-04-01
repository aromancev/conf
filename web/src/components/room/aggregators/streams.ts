import { RemoteStream } from "ion-sdk-js"
import { RoomEvent, Track, Hint } from "@/api/room/schema"
import { FIFOMap } from "@/platform/cache"

export interface Stream {
  id: string
  hint: Hint
  sourse: MediaStream
}

export class StreamAggregator {
  private readonly streams: Map<string, Stream>
  private readonly tracks: FIFOMap<string, Track>
  private readonly mediaStreams: FIFOMap<string, MediaStream>

  constructor(streams: Map<string, Stream>) {
    this.streams = streams
    this.tracks = new FIFOMap(CAPACITY)
    this.mediaStreams = new FIFOMap(CAPACITY)
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
    this.streams.clear()
    this.tracks.clear()
    this.mediaStreams.clear()
  }

  addTrack(track: MediaStreamTrack, stream: RemoteStream): void {
    if (track.kind !== "video" && track.kind !== "audio") {
      return
    }

    const id = trackId(stream)
    this.mediaStreams.set(id, stream)
    this.computeState()
    stream.onremovetrack = () => {
      this.mediaStreams.delete(id)
      this.computeState()
    }
  }

  private computeState(): void {
    this.streams.clear()

    for (const stream of this.mediaStreams.values()) {
      const id = trackId(stream)
      const track = this.tracks.get(id)
      if (!track) {
        break
      }
      this.streams.set(id, {
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

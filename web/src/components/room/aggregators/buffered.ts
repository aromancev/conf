import { RoomEvent } from "@/api/room/schema"
import { FIFOMap } from "@/platform/cache"

interface Aggregator {
  put(event: RoomEvent): void
}

export class BufferedAggregator {
  private autoflush: boolean
  private aggregators: Aggregator[]
  private cache: FIFOMap<string, RoomEvent>
  private cap: number
  private buffered: RoomEvent[]

  constructor(aggregators: Aggregator[], cap: number) {
    this.aggregators = aggregators
    this.cache = new FIFOMap(cap)
    this.cap = cap
    this.buffered = []
    this.autoflush = false
  }

  flush(): void {
    for (const event of this.buffered) {
      for (const aggregator of this.aggregators) {
        aggregator.put(event)
      }
    }
    this.buffered = []
    this.autoflush = true
  }

  put(event: RoomEvent): void {
    if (this.cache.has(event.id)) {
      return
    }

    this.cache.set(event.id, event)

    this.buffered.push(event)
    if (this.buffered.length > this.cap - this.cache.size) {
      this.buffered.shift()
    }

    if (this.autoflush) {
      this.flush()
    }
  }

  prepend(...events: RoomEvent[]): void {
    this.buffered = this.buffered.concat(events)
  }
}

import { RoomEvent } from "@/api/rtc/schema"
import { FIFOMap } from "@/platform/cache"

interface Aggregator {
  put(event: RoomEvent): void
  setTime?(time: number): void
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
      for (const agg of this.aggregators) {
        agg.put(event)
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
    if (this.buffered.length > this.cap) {
      this.buffered.shift()
    }

    if (this.autoflush) {
      this.flush()
    }
  }

  setTime(time: number): void {
    for (const agg of this.aggregators) {
      if (!agg.setTime) {
        continue
      }
      agg.setTime(time)
    }
  }

  prepend(...events: RoomEvent[]): void {
    this.buffered = this.buffered.concat(events)
  }
}

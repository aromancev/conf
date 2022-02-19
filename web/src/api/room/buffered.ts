import { Event } from "@/api/models" // TODO: Resolve events dependency so this module can be shared with BE.

export interface Aggregator {
  put(event: Event, forwards: boolean): void
}

interface BufferedEvent {
  event: Event
  forwards: boolean
}

export class BufferedAggregator {
  autoflush: boolean

  private aggregators: Aggregator[]
  private cap: number
  private byId: { [key: string]: Event }
  private ordered: Event[]
  private buffered: BufferedEvent[]

  constructor(aggregators: Aggregator[], cap: number) {
    this.aggregators = aggregators
    this.cap = cap
    this.byId = {}
    this.ordered = []
    this.buffered = []
    this.autoflush = false
  }

  flush(): void {
    for (const buff of this.buffered) {
      for (const aggregator of this.aggregators) {
        aggregator.put(buff.event, buff.forwards)
      }
    }
    this.buffered = []
  }

  put(event: Event, forwards: boolean): void {
    if (this.byId[event.id || ""]) {
      return
    }

    this.byId[event.id || ""] = event
    this.ordered.push(event)
    if (this.ordered.length > this.cap) {
      delete this.byId[this.ordered[0].id || ""]
      this.ordered.shift()
    }

    this.buffered.push({ event: event, forwards: forwards })
    if (this.buffered.length > this.cap) {
      this.buffered.shift()
    }

    if (this.autoflush) {
      this.flush()
    }
  }
}

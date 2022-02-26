import { Event } from "@/api/models"

export interface Aggregator {
  put(event: Event, forwards: boolean): void
}

export class BufferedAggregator {
  private autoflush: boolean
  private aggregators: Aggregator[]
  private cap: number
  private byId: { [key: string]: Event }
  private ordered: Event[]
  private buffered: Event[]

  constructor(aggregators: Aggregator[], cap: number) {
    this.aggregators = aggregators
    this.cap = cap
    this.byId = {}
    this.ordered = []
    this.buffered = []
    this.autoflush = false
  }

  flush(): void {
    for (const event of this.buffered) {
      for (const aggregator of this.aggregators) {
        aggregator.put(event, true)
      }
    }
    this.buffered = []
    this.autoflush = true
  }

  put(event: Event): void {
    if (this.byId[event.id || ""]) {
      return
    }

    this.byId[event.id || ""] = event
    this.ordered.push(event)
    if (this.ordered.length > this.cap) {
      delete this.byId[this.ordered[0].id || ""]
      this.ordered.shift()
    }

    this.buffered.push(event)
    if (this.buffered.length > this.cap) {
      this.buffered.shift()
    }

    if (this.autoflush) {
      this.flush()
    }
  }

  prepend(...events: Event[]): void {
    this.buffered = this.buffered.concat(events)
  }
}

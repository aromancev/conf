import { Event } from "@/api/models"

export interface Record {
  event: Event
  forward: boolean
  live: boolean
}

export interface RecordProcessor {
  processRecords(records: Record[]): void
}

export class BufferedProcessor {
  autoflush: boolean

  private processors: RecordProcessor[]
  private cap: number
  private byId: { [key: string]: Event }
  private ordered: Event[]
  private buffered: Record[]

  constructor(procs: RecordProcessor[], cap: number) {
    this.processors = procs
    this.cap = cap
    this.byId = {}
    this.ordered = []
    this.buffered = []
    this.autoflush = false
  }

  flush(): void {
    for (const p of this.processors) {
      p.processRecords(this.buffered)
    }
    this.buffered = []
  }

  put(events: Event[], live: boolean): void {
    for (const e of events) {
      if (this.byId[e.id || ""]) {
        continue
      }

      this.byId[e.id || ""] = e
      this.ordered.push(e)
      if (this.ordered.length > this.cap) {
        delete this.byId[this.ordered[0].id || ""]
        this.ordered.shift()
      }

      this.buffered.push({
        event: e,
        live: live,
        forward: true,
      })
      if (this.buffered.length > this.cap) {
        this.buffered.shift()
      }
    }

    if (this.autoflush) {
      this.flush()
    }
  }
}

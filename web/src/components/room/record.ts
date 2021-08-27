import { Event } from "@/api/models"

export interface Record {
  event: Event
  forward: boolean
  live: boolean
}

export interface RecordProcessor {
  processRecords(records: Record[]): void
}

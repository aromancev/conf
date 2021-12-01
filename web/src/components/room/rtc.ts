import { Event } from "@/api/models"

export interface Emitter {
  event(event: Event): Promise<string>
}

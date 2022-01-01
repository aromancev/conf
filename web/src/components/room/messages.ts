import { Record } from "./record"
import { EventType, PayloadMessage } from "@/api/models"
import { genName, genAvatar } from "@/platform/gen"
import { reactive, readonly } from "vue"
import { Event } from "@/api"

const maxMessages = 100

export interface Emitter {
  event(event: Event): Promise<string>
}

export interface Message {
  id: string
  from: string
  fromName: string
  avatar: string
  text: string
  isSent: boolean
  isFirstFrom: boolean
  isLatestFrom: boolean
}

export class MessageProcessor {
  private ordered: Message[]
  private byId: { [key: string]: Message }
  private emitter: Emitter

  constructor(emitter: Emitter) {
    this.ordered = reactive<Message[]>([])
    this.byId = {}
    this.emitter = emitter
  }

  messages(): Message[] {
    return readonly(this.ordered) as Message[]
  }

  async send(userId: string, message: string): Promise<void> {
    const msg: Message = {
      id: "",
      from: userId,
      fromName: "",
      avatar: "",
      text: message,
      isSent: false,
      isFirstFrom: false,
      isLatestFrom: true,
    }
    const ev = {
      payload: {
        type: EventType.Message,
        payload: {
          text: message,
        },
      },
    }
    this.ordered.push(msg)
    msg.id = await this.emitter.event(ev)
    this.byId[msg.id] = msg
  }

  processRecords(records: Record[]) {
    for (const r of records) {
      if (r.event.payload.type !== EventType.Message) {
        continue
      }

      const payload = r.event.payload.payload as PayloadMessage

      let isFirstFrom = true
      if (this.ordered.length) {
        const latest = this.ordered[this.ordered.length - 1]
        if (latest.from === r.event.ownerId) {
          latest.isLatestFrom = false
          isFirstFrom = false
        }
      }
      const msg: Message = {
        id: r.event.id || "",
        from: r.event.ownerId || "",
        fromName: genName(r.event.ownerId || ""),
        avatar: genAvatar(r.event.ownerId || "", 32 + 1),
        text: payload.text,
        isSent: true,
        isFirstFrom: isFirstFrom,
        isLatestFrom: true,
      }

      const existing = this.byId[msg.id]
      if (existing) {
        existing.isSent = true
      } else {
        this.byId[msg.id] = msg
        this.ordered.push(msg)
      }

      if (this.ordered.length > maxMessages) {
        delete this.byId[this.ordered[0].id]
        this.ordered.shift()
      }
    }
  }
}

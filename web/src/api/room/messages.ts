import { RoomEvent } from "./schema"

const maxMessages = 100

export interface Message {
  id: string
  fromId: string
  text: string
  accepted: boolean
}

export class MessageAggregator {
  private messages: Message[]

  constructor(messages: Message[]) {
    this.messages = messages
  }

  put(event: RoomEvent): void {
    if (!event.payload.message) {
      return
    }

    const msg: Message = {
      id: event.id || "",
      fromId: event.ownerId || "",
      text: event.payload.message.text,
      accepted: true,
    }

    const existing = this.find(msg.id)
    if (existing) {
      existing.fromId = msg.fromId
      existing.text = msg.text
      existing.accepted = msg.accepted
    } else {
      this.messages.push(msg)
    }

    // Remove outstanding message.
    if (this.messages.length > maxMessages) {
      this.messages.shift()
    }
  }

  // Since the number of messages is not too large, we don't use a separate map for simplicity.
  private find(id: string): Message | null {
    for (const m of this.messages) {
      if (m.id === id) {
        return m
      }
    }
    return null
  }
}

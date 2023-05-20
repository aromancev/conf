import { RoomEvent } from "@/api/rtc/schema"

interface Profile {
  handle: string
  name: string
  avatar: string
}

interface Repo {
  profile(id: string): Profile
}

export interface Message {
  id: string
  fromId: string
  text: string
  accepted: boolean
  profile: Profile
}

export class MessageAggregator {
  private readonly messages: Message[]
  private readonly repo: Repo

  constructor(messages: Message[], repo: Repo) {
    this.messages = messages
    this.repo = repo
  }

  put(event: RoomEvent): void {
    const message = event.payload.message
    if (!message) {
      return
    }

    const msg: Message = {
      id: event.id || "",
      fromId: message.fromId,
      text: message.text,
      accepted: true,
      profile: this.repo.profile(message.fromId),
    }

    const existing = this.find(msg.id)
    if (existing) {
      existing.fromId = msg.fromId
      existing.text = msg.text
      existing.accepted = msg.accepted
    } else {
      this.messages.push(msg)
    }

    // Drop outstanding message.
    if (this.messages.length > CAPACITY) {
      this.messages.shift()
    }
  }

  reset(): void {
    this.messages.splice(0, this.messages.length)
  }

  // addMessage returns a function that can be called to provide message id for created message.
  addMessage(userId: string, text: string): (id: string) => void {
    const msg: Message = {
      id: "",
      fromId: userId,
      text: text,
      accepted: false,
      profile: this.repo.profile(userId),
    }
    this.messages.push(msg)
    // Drop outstanding message.
    if (this.messages.length > CAPACITY) {
      this.messages.shift()
    }
    return (id: string) => {
      msg.id = id
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

const CAPACITY = 500

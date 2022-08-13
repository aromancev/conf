import { RoomEvent } from "@/api/room/schema"

const maxMessages = 100

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
  private messages: Message[]
  private repo: Repo

  constructor(repo: Repo, messages: Message[]) {
    this.repo = repo
    this.messages = messages
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

import { reactive, readonly } from "vue"
import { RoomEvent } from "@/api/room/schema"

interface Profile {
  handle: string
  name: string
  avatar: string
}

interface Repo {
  profile(id: string): Profile
}

export interface State {
  messages: Message[]
}

export interface Message {
  id: string
  fromId: string
  text: string
  accepted: boolean
  profile: Profile
}

export class MessageAggregator {
  private _state: State
  private repo: Repo

  constructor(repo: Repo) {
    this._state = reactive({
      messages: [],
    })
    this.repo = repo
  }

  state(): State {
    return readonly(this._state) as State
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
      this._state.messages.push(msg)
    }

    // Drop outstanding message.
    if (this._state.messages.length > CAPACITY) {
      this._state.messages.shift()
    }
  }

  reset(): void {
    this._state.messages.splice(0, this._state.messages.length)
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
    this._state.messages.push(msg)
    // Drop outstanding message.
    if (this._state.messages.length > CAPACITY) {
      this._state.messages.shift()
    }
    return (id: string) => {
      msg.id = id
    }
  }

  // Since the number of messages is not too large, we don't use a separate map for simplicity.
  private find(id: string): Message | null {
    for (const m of this._state.messages) {
      if (m.id === id) {
        return m
      }
    }
    return null
  }
}

const CAPACITY = 500

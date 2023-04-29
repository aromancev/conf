import { api } from "@/api"
import { config } from "@/config"
import { PeerMessage, Message, Reaction, MessagePayload, RoomEvent } from "./schema"

export class RTCPeer {
  private readonly socket: RoomWebSocket

  constructor() {
    this.socket = new RoomWebSocket()
  }

  set onevent(val: (event: RoomEvent) => void) {
    this.socket.onevent = val
  }

  set onclose(val: () => void) {
    this.socket.onclose = val
  }

  set onerror(val: () => void) {
    this.socket.onerror = val
  }

  async message(msg: PeerMessage): Promise<RoomEvent> {
    return this.socket.message(msg)
  }

  async reaction(reaction: Reaction): Promise<RoomEvent> {
    return this.socket.reaction(reaction)
  }

  async joinRTC(roomId: string): Promise<void> {
    const token = await api.token()
    await this.socket.connect(roomId, token)
  }

  async close(): Promise<void> {
    await this.socket.close()
  }
}

enum CloseCode {
  NORMAL_CLOSURE = 1000,
}

const requestTimeout = 10 * 1000

class RoomWebSocket {
  onevent?: (event: RoomEvent) => void
  onclose?: () => void
  onerror?: () => void

  private socket?: WebSocket
  private requestId = 0
  private pendingRequests: Map<string, (msg: Message) => void> = new Map()
  private connected?: Promise<void>

  async connect(roomId: string, token: string): Promise<void> {
    if (this.socket) {
      await this.close()
    }
    this.socket = new WebSocket(`${config.rtc.room.baseURL}/${roomId}?t=${token}`)
    this.socket.onmessage = (resp) => {
      const msg = JSON.parse(resp.data) as Message

      if (msg.responseId) {
        this.closePending(msg)
        return
      }

      const event = msg.payload.event
      if (event) {
        if (this.onevent) {
          this.onevent(event)
        }
        return
      }
    }

    this.socket.onclose = (e: CloseEvent) => {
      if (e.code === CloseCode.NORMAL_CLOSURE) {
        if (this.onclose) {
          this.onclose()
        }
      } else {
        if (this.onerror) {
          this.onerror()
        }
      }
    }

    const sock = this.socket
    const opened = new Promise<void>((resolve) => {
      sock.onopen = () => {
        resolve()
      }
    })
    const failed = new Promise<void>((resolve) => {
      sock.onerror = () => {
        resolve()
      }
    })
    this.connected = Promise.race([opened, failed])
    await this.connected
    if (this.socket.readyState !== this.socket.OPEN) {
      throw new Error("Failed to connect to websocket.")
    }
  }

  async message(msg: PeerMessage): Promise<RoomEvent> {
    const resp = await this.send({
      peerMessage: msg,
    })

    if (!resp.payload.event) {
      throw new Error("Unexpected response from RTC.")
    }

    return resp.payload.event
  }

  async reaction(reaction: Reaction): Promise<RoomEvent> {
    const resp = await this.send({
      reaction: reaction,
    })
    if (!resp.payload.event) {
      throw new Error("Unexpected response from RTC.")
    }
    return resp.payload.event
  }

  async close(): Promise<void> {
    await this.connected
    this.socket?.close(CloseCode.NORMAL_CLOSURE)
  }

  private send(payload: MessagePayload): Promise<Message> {
    return new Promise<Message>((resolve, reject) => {
      const id = this.openPending((msg: Message) => {
        resolve(msg)
      }, reject)

      this.notify({
        requestId: id,
        payload: payload,
      })
    })
  }

  private notify(req: Message): void {
    if (!this.socket) {
      throw new Error("RTC not connected.")
    }
    this.socket.send(JSON.stringify(req))
  }

  private openPending(resolve: (msg: Message) => void, reject: (reason: string) => void): string {
    this.requestId++
    const pendingId = this.requestId

    this.pendingRequests.set(pendingId.toString(), resolve)

    setTimeout(() => {
      if (pendingId in this.pendingRequests) {
        this.pendingRequests.delete(pendingId.toString())
        reject("Message to RTC timed out.")
      }
    }, requestTimeout)
    return pendingId.toString()
  }

  private closePending(msg: Message): void {
    const pendingId = msg.responseId
    if (!pendingId) {
      return
    }
    const pending = this.pendingRequests.get(pendingId)
    if (pending) {
      pending(msg)
      this.pendingRequests.delete(pendingId)
    }
  }
}

import { Trickle } from "ion-sdk-js"
import { Event, Track } from "./models/event"

const requestTimeout = 10 * 1000

enum Type {
  Join = "join",
  Offer = "offer",
  Answer = "answer",
  Trickle = "trickle",
  Event = "event",
  EventAck = "event_ack",
  State = "state",
}

interface Message {
  requestId?: string
  type: Type
  payload: Join | Answer | Offer | Trickle | Event | EventAck | State
}

interface Join {
  sessionId: string
  userId: string
  description: RTCSessionDescriptionInit
}

interface Answer {
  description: RTCSessionDescriptionInit
}

interface Offer {
  description: RTCSessionDescriptionInit
}

interface EventAck {
  eventId: string
}

export interface State {
  tracks: { [key: string]: Track }
}

export class RTC {
  onnegotiate?: (jsep: RTCSessionDescriptionInit) => void
  ontrickle?: (trickle: Trickle) => void
  onevent?: (event: Event) => void

  token: string

  private _onopen?: () => void
  private socket: WebSocket
  private onSignalAnswer: ((desc: RTCSessionDescriptionInit) => void) | null
  private requestId = 0
  private pendingRequests = {} as { [key: string]: (msg: Message) => void }

  constructor(roomId: string, token: string) {
    let protocol = "wss"
    if (process.env.NODE_ENV == "development") {
      protocol = "ws"
    }
    this.onSignalAnswer = null
    this.token = token

    this.socket = new WebSocket(`${protocol}://${window.location.hostname}/api/rtc/room/${roomId}?t=${token}`)
    this.socket.onopen = () => {
      if (this._onopen) {
        this._onopen()
      }
    }
    this.socket.onmessage = (msg) => {
      const resp = JSON.parse(msg.data) as Message
      switch (resp.type) {
        case Type.Answer:
          if (this.onSignalAnswer) {
            this.onSignalAnswer((resp.payload as Answer).description)
          }
          break
        case Type.Offer:
          if (this.onnegotiate) {
            this.onnegotiate((resp.payload as Offer).description)
          }
          break
        case Type.Trickle:
          if (this.ontrickle) {
            this.ontrickle(resp.payload as Trickle)
          }
          break
        case Type.Event:
          if (this.onevent) {
            this.onevent(resp.payload as Event)
          }
          break
        case Type.EventAck:
        case Type.State:
          this.closePending(resp)
          break
      }
    }
  }

  set onopen(onopen: () => void) {
    if (this.socket.readyState === WebSocket.OPEN) {
      onopen()
    }
    this._onopen = onopen
  }

  join(sid: string, uid: string, offer: RTCSessionDescriptionInit): Promise<RTCSessionDescriptionInit> {
    this.send({
      type: Type.Join,
      payload: {
        sessionId: sid,
        userId: uid,
        description: offer,
      },
    })
    return new Promise((resolve) => {
      this.onSignalAnswer = (desc) => {
        this.onSignalAnswer = null
        resolve(desc)
      }
    })
  }

  offer(offer: RTCSessionDescriptionInit): Promise<RTCSessionDescriptionInit> {
    this.send({
      type: Type.Offer,
      payload: { description: offer },
    })
    return new Promise((resolve) => {
      this.onSignalAnswer = (desc) => {
        this.onSignalAnswer = null
        resolve(desc)
      }
    })
  }

  answer(answer: RTCSessionDescriptionInit): void {
    this.send({
      type: Type.Answer,
      payload: { description: answer },
    })
  }

  trickle(trickle: Trickle): void {
    this.send({
      type: Type.Trickle,
      payload: trickle,
    })
  }

  event(event: Event): Promise<string> {
    return new Promise<string>((resolve, reject) => {
      const id = this.openPending((msg: Message) => {
        resolve((msg.payload as EventAck).eventId)
      }, reject)

      this.send({
        requestId: id,
        type: Type.Event,
        payload: event,
      })
    })
  }

  state(state: State): Promise<State> {
    return new Promise<State>((resolve, reject) => {
      const id = this.openPending((msg: Message) => {
        resolve(msg.payload as State)
      }, reject)

      this.send({
        requestId: id,
        type: Type.State,
        payload: state,
      })
    })
  }

  close(): void {
    this.socket.close()
  }

  private send(req: Message): void {
    this.socket.send(JSON.stringify(req))
  }

  private openPending(resolve: (msg: Message) => void, reject: (reason: string) => void): string {
    this.requestId++
    const pendingId = this.requestId

    this.pendingRequests[pendingId] = resolve

    setTimeout(() => {
      if (pendingId in this.pendingRequests) {
        delete this.pendingRequests[pendingId]
        reject("Message to RTC timed out.")
      }
    }, requestTimeout)
    return pendingId.toString()
  }

  private closePending(msg: Message): void {
    if (!msg.requestId) {
      return
    }
    if (msg.requestId in this.pendingRequests) {
      this.pendingRequests[msg.requestId](msg)
      delete this.pendingRequests[msg.requestId]
    }
  }
}

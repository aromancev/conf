import { Trickle } from "ion-sdk-js"
import { Event } from "./models/event"

enum Type {
  Join = "join",
  Offer = "offer",
  Answer = "answer",
  Trickle = "trickle",
  Event = "event",
  EventAck = "event_ack",
}

interface Message {
  requestId?: string
  type: Type
  payload: Join | Answer | Offer | Trickle | Event | EventAck
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

export class RTC {
  onnegotiate?: (jsep: RTCSessionDescriptionInit) => void
  ontrickle?: (trickle: Trickle) => void
  onevent?: (event: Event) => void

  private _onopen?: () => void
  private socket: WebSocket
  private onSignalAnswer: ((desc: RTCSessionDescriptionInit) => void) | null
  private requestId = 0
  private pendingEvents = {} as { [key: string]: (ack: EventAck) => void }

  constructor(roomId: string, token: string) {
    let protocol = "wss"
    if (process.env.NODE_ENV == "development") {
      protocol = "ws"
    }
    this.onSignalAnswer = null

    this.socket = new WebSocket(
      `${protocol}://${window.location.hostname}/api/rtc/ws/${roomId}?t=${token}`,
    )
    this.socket.onopen = () => {
      if (this._onopen) {
        this._onopen()
      }
    }
    this.socket.onmessage = msg => {
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
          if (!resp.requestId) {
            return
          }
          if (resp.requestId in this.pendingEvents) {
            this.pendingEvents[resp.requestId](resp.payload as EventAck)
          }
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

  join(
    sid: string,
    uid: string,
    offer: RTCSessionDescriptionInit,
  ): Promise<RTCSessionDescriptionInit> {
    this.send({
      type: Type.Join,
      payload: {
        sessionId: sid,
        userId: uid,
        description: offer,
      },
    })
    return new Promise(resolve => {
      this.onSignalAnswer = desc => {
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
    return new Promise(resolve => {
      this.onSignalAnswer = desc => {
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
    this.requestId++
    const requestId = this.requestId

    return new Promise<string>((resolve) => {
      this.pendingEvents[requestId] = (ack: EventAck) => {
        resolve(ack.eventId)
      }

      this.send({
        requestId: requestId.toString(),
        type: Type.Event,
        payload: event,
      })
    })
  }

  close(): void {
    this.socket.close()
  }

  private send(req: Message): void {
    this.socket.send(
      JSON.stringify(req),
    )
  }
}

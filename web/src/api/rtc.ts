import { Trickle } from "ion-sdk-js"

enum Type {
  Join = "join",
  Offer = "offer",
  Answer = "answer",
  Trickle = "trickle",
}

interface Request {
  type: Type
  payload: Join | RTCSessionDescriptionInit | Trickle
}

interface Join {
  sid: string
  uid: string
  offer: RTCSessionDescriptionInit
}

interface Response {
  type: Type
  payload: RTCSessionDescriptionInit | Trickle
}

export class Signal {
  onnegotiate?: (jsep: RTCSessionDescriptionInit) => void
  ontrickle?: (trickle: Trickle) => void

  private _onopen?: () => void
  private socket: WebSocket
  private onSignalAnswer: ((desc: RTCSessionDescriptionInit) => void) | null

  constructor() {
    let protocol = "wss"
    if (process.env.NODE_ENV == "development") {
      protocol = "ws"
    }
    this.onSignalAnswer = null

    this.socket = new WebSocket(
      `${protocol}://${window.location.hostname}/api/rtc/v1/ws`,
    )
    this.socket.onopen = () => {
      if (this._onopen) {
        this._onopen()
      }
    }
    this.socket.onmessage = msg => {
      const resp = JSON.parse(msg.data) as Response
      switch (resp.type) {
        case Type.Answer:
          if (this.onSignalAnswer) {
            this.onSignalAnswer(resp.payload as RTCSessionDescriptionInit)
          }
          break
        case Type.Offer:
          if (this.onnegotiate) {
            this.onnegotiate(resp.payload as RTCSessionDescriptionInit)
          }
          break
        case Type.Trickle:
          if (this.ontrickle) {
            this.ontrickle(resp.payload as Trickle)
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
        sid: sid,
        uid: uid,
        offer: offer,
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
      payload: offer,
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
      payload: answer,
    })
  }

  trickle(trickle: Trickle): void {
    this.send({
      type: Type.Trickle,
      payload: trickle,
    })
  }

  close(): void {
    this.socket.close()
  }

  private send(req: Request): void {
    this.socket.send(
      JSON.stringify({
        type: req.type,
        payload: req.payload,
      }),
    )
  }
}

import { Trickle } from "ion-sdk-js"

const maxPending = 10

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
  offer: RTCSessionDescriptionInit
}

interface Response {
  id?: number
  type: Type
  payload: RTCSessionDescriptionInit | Trickle
}

export class Signal {
  onnegotiate?: (jsep: RTCSessionDescriptionInit) => void
  ontrickle?: (trickle: Trickle) => void
  onopen?: () => void

  private socket: WebSocket
  private requestId: number
  private pendingRequests: Record<number, (resp: Response) => void>

  constructor(url: string) {
    this.requestId = 0
    this.pendingRequests = {}

    this.socket = new WebSocket(url)
    this.socket.onopen = () => {
      if (this.onopen) {
        this.onopen()
      }
    }
    this.socket.onmessage = msg => {
      const resp = JSON.parse(msg.data) as Response
      if (resp.id !== undefined) {
        if (resp.id in this.pendingRequests) {
          this.pendingRequests[resp.id](resp)
        } else {
          throw new Error("unexpected reply from rtc")
        }
      } else {
        if (this.onnegotiate && resp.type === Type.Offer) {
          this.onnegotiate(resp.payload as RTCSessionDescriptionInit)
          return
        }
        if (this.ontrickle && resp.type === Type.Trickle) {
          this.ontrickle(resp.payload as Trickle)
          return
        }
      }
    }
  }

  join(
    sid: string,
    offer: RTCSessionDescriptionInit,
  ): Promise<RTCSessionDescriptionInit> {
    const send = this.send({
      type: Type.Join,
      payload: {
        sid: sid,
        offer: offer,
      },
    })
    return new Promise(resolve => {
      send.then(resp => {
        resolve(resp.payload as RTCSessionDescriptionInit)
      })
    })
  }

  offer(offer: RTCSessionDescriptionInit): Promise<RTCSessionDescriptionInit> {
    const send = this.send({
      type: Type.Offer,
      payload: offer,
    })
    return new Promise(resolve => {
      send.then(resp => {
        resolve(resp.payload as RTCSessionDescriptionInit)
      })
    })
  }

  answer(answer: RTCSessionDescriptionInit): void {
    this.notify({
      type: Type.Answer,
      payload: answer,
    })
  }

  trickle(trickle: Trickle): void {
    this.notify({
      type: Type.Trickle,
      payload: trickle,
    })
  }

  close(): void {
    this.socket.close()
  }

  private notify(req: Request): void {
    this.requestId++
    const id = this.requestId
    this.socket.send(
      JSON.stringify({
        id: id,
        type: req.type,
        payload: req.payload,
      }),
    )
  }

  private send(req: Request): Promise<Response> {
    if (Object.keys(this.pendingRequests).length > maxPending) {
      throw new Error("too many pending requests")
    }
    this.requestId++
    const id = this.requestId
    this.socket.send(
      JSON.stringify({
        id: id,
        type: req.type,
        payload: req.payload,
      }),
    )
    return new Promise(resolve => {
      this.pendingRequests[id] = (resp: Response): void => {
        delete this.pendingRequests[id]
        resolve(resp)
      }
    })
  }
}

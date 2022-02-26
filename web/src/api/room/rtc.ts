import {
  Client as IonClient,
  Trickle,
  RemoteStream as IonRemoteStream,
  LocalStream as IonLocalStream,
} from "ion-sdk-js"
import { Event, Track, userStore } from "../models/"
import { client } from "@/api"
import { config } from "@/config"

export type RemoteStream = IonRemoteStream
export type LocalStream = IonLocalStream

export interface State {
  tracks: { [key: string]: Track }
}

export class RTCPeer {
  ontrack?: (track: MediaStreamTrack, stream: RemoteStream) => void

  private signal: SignalRTC
  private sfu?: IonClient

  constructor() {
    this.signal = new SignalRTC()
  }

  set onevent(val: (event: Event) => void) {
    this.signal.onevent = val
  }

  async sendEvent(event: Event): Promise<string> {
    return this.signal.sendEvent(event)
  }

  async sendState(state: State): Promise<State> {
    return this.signal.sendState(state)
  }

  async join(roomId: string, media: boolean): Promise<void> {
    const token = await client.token()
    await this.signal.connect(roomId, token)

    if (!media) {
      return
    }
    const iceServers: RTCIceServer[] = []
    if (config.sfu.stunURLs) {
      iceServers.push({
        urls: config.sfu.stunURLs,
      })
    }
    if (config.sfu.turnURLs) {
      iceServers.push({
        urls: config.sfu.turnURLs,
        credentialType: "password",
        username: token,
        credential: "confa.io",
      })
    }
    this.sfu = new IonClient(this.signal, {
      codec: "vp8",
      iceServers: iceServers,
    })
    this.sfu.ontrack = this.ontrack
    await this.sfu.join(roomId, userStore.getState().id)
  }

  publish(stream: LocalStream, encodingParams?: RTCRtpEncodingParameters[]): void {
    if (!this.sfu) {
      throw new Error("Peer not joined to media.")
    }
    this.sfu.publish(stream, encodingParams)
  }

  close(): void {
    this.sfu?.close()
    this.signal.close()
  }
}

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

class SignalRTC {
  onnegotiate?: (jsep: RTCSessionDescriptionInit) => void
  ontrickle?: (trickle: Trickle) => void
  onevent?: (event: Event) => void

  private socket?: WebSocket
  private onSignalAnswer?: (desc: RTCSessionDescriptionInit) => void
  private requestId = 0
  private pendingRequests = {} as { [key: string]: (msg: Message) => void }

  async connect(roomId: string, token: string): Promise<void> {
    const socket = new WebSocket(`${config.rtc.room.baseURL}/${roomId}?t=${token}`)
    socket.onmessage = (msg) => {
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

    await new Promise<void>((resolve) => {
      socket.onopen = () => {
        resolve()
      }
    })

    this.socket = socket
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
        this.onSignalAnswer = undefined
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
        this.onSignalAnswer = undefined
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

  sendEvent(event: Event): Promise<string> {
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

  sendState(state: State): Promise<State> {
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
    this.socket?.close()
  }

  private send(req: Message): void {
    if (!this.socket) {
      throw new Error("RTC not connected.")
    }
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

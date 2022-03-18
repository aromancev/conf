import {
  Client as IonClient,
  Trickle,
  RemoteStream as IonRemoteStream,
  LocalStream as IonLocalStream,
} from "ion-sdk-js"
import { userStore } from "../models"
import { client } from "@/api"
import { config } from "@/config"
import { PeerMessage, PeerState, Message, MessagePayload, RoomEvent, SDPType } from "./schema"

export type RemoteStream = IonRemoteStream
export type LocalStream = IonLocalStream

export class RTCPeer {
  ontrack?: (track: MediaStreamTrack, stream: RemoteStream) => void

  private socket: RoomWebSocket
  private sfu?: IonClient

  constructor() {
    this.socket = new RoomWebSocket()
  }

  set onevent(val: (event: RoomEvent) => void) {
    this.socket.onevent = val
  }

  async message(msg: PeerMessage): Promise<RoomEvent> {
    return this.socket.message(msg)
  }

  async state(state: PeerState): Promise<PeerState> {
    return this.socket.state(state)
  }

  async join(roomId: string, media: boolean): Promise<void> {
    const token = await client.token()
    await this.socket.connect(roomId, token, media)

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
    this.sfu = new IonClient(this.socket, {
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
    this.socket.close()
  }
}

const requestTimeout = 10 * 1000

class RoomWebSocket {
  onnegotiate?: (jsep: RTCSessionDescriptionInit) => void
  ontrickle?: (trickle: Trickle) => void
  onevent?: (event: RoomEvent) => void

  private socket?: WebSocket
  private onSignalAnswer?: (desc: RTCSessionDescriptionInit) => void
  private requestId = 0
  private pendingRequests = {} as { [key: string]: (msg: Message) => void }

  async connect(roomId: string, token: string, media: boolean): Promise<void> {
    const socket = new WebSocket(`${config.rtc.room.baseURL}/${roomId}?t=${token}&media=${media ? "true" : "false"}`)
    socket.onmessage = (resp) => {
      const msg = JSON.parse(resp.data) as Message

      if (msg.responseId) {
        this.closePending(msg)
        return
      }

      const signal = msg.payload.signal
      if (signal?.answer) {
        if (this.onSignalAnswer) {
          this.onSignalAnswer(signal.answer.description)
        }
        return
      }
      if (signal?.offer) {
        if (this.onnegotiate) {
          this.onnegotiate(signal.offer.description)
        }
        return
      }
      if (signal?.trickle) {
        if (this.ontrickle) {
          this.ontrickle(signal.trickle)
        }
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

    await new Promise<void>((resolve) => {
      socket.onopen = () => {
        resolve()
      }
    })

    this.socket = socket
  }

  async join(sid: string, uid: string, offer: RTCSessionDescriptionInit): Promise<RTCSessionDescriptionInit> {
    this.notify({
      payload: {
        signal: {
          join: {
            sessionId: sid,
            userId: uid,
            description: {
              sdp: offer.sdp || "",
              type: offer.type as SDPType,
            },
          },
        },
      },
    })
    return new Promise((resolve) => {
      this.onSignalAnswer = (desc) => {
        this.onSignalAnswer = undefined
        resolve(desc)
      }
    })
  }

  async offer(offer: RTCSessionDescriptionInit): Promise<RTCSessionDescriptionInit> {
    this.notify({
      payload: {
        signal: {
          offer: {
            description: {
              sdp: offer.sdp || "",
              type: offer.type as SDPType,
            },
          },
        },
      },
    })
    return new Promise((resolve) => {
      this.onSignalAnswer = (desc) => {
        this.onSignalAnswer = undefined
        resolve(desc)
      }
    })
  }

  answer(answer: RTCSessionDescriptionInit): void {
    this.notify({
      payload: {
        signal: {
          answer: {
            description: {
              sdp: answer.sdp || "",
              type: answer.type as SDPType,
            },
          },
        },
      },
    })
  }

  trickle(trickle: Trickle): void {
    this.notify({
      payload: {
        signal: {
          trickle: {
            candidate: {
              candidate: trickle.candidate.candidate || "",
              sdpMid: trickle.candidate.sdpMid || undefined,
              sdpMLineIndex: trickle.candidate.sdpMLineIndex || undefined,
              usernameFragment: trickle.candidate.usernameFragment || undefined,
            },
            target: trickle.target,
          },
        },
      },
    })
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

  async state(state: PeerState): Promise<PeerState> {
    const resp = await this.send({
      state: state,
    })

    if (!resp.payload.state) {
      throw new Error("Unexpected response from RTC.")
    }

    return resp.payload.state
  }

  close(): void {
    this.socket?.close()
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
    if (!msg.responseId) {
      return
    }
    if (msg.responseId in this.pendingRequests) {
      this.pendingRequests[msg.responseId](msg)
      delete this.pendingRequests[msg.responseId]
    }
  }
}

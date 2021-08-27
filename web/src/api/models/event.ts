export enum EventType {
  PeerStatus = "peer_status",
  Message = "message",
}

export type EventPayload = PayloadPeerStatus | PayloadMessage

export interface Event {
  id?: string
  ownerId?: string
  roomId?: string
  createdAt?: string
  payload: {
    type: EventType
    payload: EventPayload
  }
}

export interface EventProcessor {
  forward(event: Event): void
  backward(event: Event): void
}

export enum PeerStatus {
  Joined = "joined",
  Left = "left",
}

export interface PayloadPeerStatus {
  status: PeerStatus
}

export interface PayloadMessage {
  text: string
}

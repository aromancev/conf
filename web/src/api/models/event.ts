export enum EventType {
  PeerStatus = "peer_status",
}

export interface Event {
  id: string
  ownerId: string
  roomId: string
  payload: {
    type: EventType
    payload: PayloadPeerStatus
  }
}

export enum PeerStatus {
  Joined = "joined",
  Left = "left",
}

export interface PayloadPeerStatus {
  status: PeerStatus
}

export interface EventProcessor {
  forward(event: Event): void
  backward(event: Event): void
}
